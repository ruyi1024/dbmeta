/*
Copyright 2026 The Dbmeta Team Group, website: https://www.dbmeta.com
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.gnu.org/licenses/gpl-3.0.html
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package task

import (
	"context"
	"database/sql"
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/setting"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/libary/mongodb"
	"github.com/ruyi1024/dbmeta/src/model"
	"github.com/ruyi1024/dbmeta/src/utils"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var dbCon *sql.DB
var err error

func init() {
	go dbMetaCrontabTask()
}

func dbMetaCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "gather_dbmeta").Take(&record)
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "gather_dbmeta").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_dbmeta'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doDbMetaTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_dbmeta'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func formatInterface(inter interface{}) string {
	if inter != nil {
		return inter.(string)
	} else {
		return ""
	}
}

func doDbMetaTask() {
	logger := log.Logger
	logger.Info("开始执行数据库元数据采集任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("gather_dbmeta")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	var db = database.DB
	var dataList []model.Datasource
	result := db.Where("enable=1").Where("dbmeta_enable=1").Order("type asc").Find(&dataList)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("查询数据源失败: %v", result.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	if len(dataList) == 0 {
		successMsg := "没有找到启用的数据源"
		logger.Info(successMsg)
		taskLogger.Success(successMsg)
		return
	}

	logger.Info("找到数据源", zap.Int("count", len(dataList)))
	taskLogger.UpdateResult(fmt.Sprintf("找到 %d 个启用的数据源", len(dataList)))

	successCount := 0
	failedCount := 0
	errorDetails := []string{}
	instanceStatuses := []string{}

	for i, datasource := range dataList {
		datasourceType := datasource.Type
		host := datasource.Host
		port := datasource.Port
		user := datasource.User
		pass := datasource.Pass
		dbid := datasource.Dbid

		logger.Info("处理数据源", zap.Int("index", i+1), zap.Int("total", len(dataList)),
			zap.String("type", datasourceType), zap.String("host", host), zap.String("port", port))

		var origPass string
		if pass != "" {
			var err error
			origPass, err = utils.AesPassDecode(pass, setting.Setting.DbPassKey)
			if err != nil {
				errorMsg := fmt.Sprintf("数据源 %s:%s 密码解密失败: %v", host, port, err)
				logger.Error(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				instanceStatuses = append(instanceStatuses, fmt.Sprintf("[失败] %s: 密码解密失败", formatDatasourceInstance(datasource)))
				failedCount++
				continue
			}
		}

		err := doDbMetaCollectorTask(datasourceType, host, port, user, origPass, dbid)
		if err != nil {
			errorMsg := fmt.Sprintf("数据源 %s:%s 元数据采集失败: %v", host, port, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			instanceStatuses = append(instanceStatuses, fmt.Sprintf("[失败] %s: %s", formatDatasourceInstance(datasource), truncateText(err.Error(), 100)))
			failedCount++
		} else {
			instanceStatuses = append(instanceStatuses, fmt.Sprintf("[成功] %s", formatDatasourceInstance(datasource)))
			successCount++
		}

		// 更新进度
		progressMsg := fmt.Sprintf("已处理 %d/%d 个数据源 (成功: %d, 失败: %d)", i+1, len(dataList), successCount, failedCount)
		taskLogger.UpdateResult(progressMsg)
	}

	// 清理过期元数据
	logger.Info("开始清理过期元数据")
	expireTime := time.Now().Add(-time.Minute * 600).Format("2006-01-02 15:04:05")

	dbResult := database.DB.Model(model.MetaDatabase{}).Where("gmt_updated <= ?", expireTime).Updates(map[string]interface{}{"is_deleted": 1})
	tableResult := database.DB.Model(model.MetaTable{}).Where("gmt_updated <= ?", expireTime).Updates(map[string]interface{}{"is_deleted": 1})
	columnResult := database.DB.Model(model.MetaColumn{}).Where("gmt_updated <= ?", expireTime).Updates(map[string]interface{}{"is_deleted": 1})

	cleanupMsg := fmt.Sprintf("清理过期元数据 - 数据库: %d, 表: %d, 字段: %d",
		dbResult.RowsAffected, tableResult.RowsAffected, columnResult.RowsAffected)
	logger.Info(cleanupMsg)

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 数据源总计: %d, 成功: %d, 失败: %d, %s",
		len(dataList), successCount, failedCount, cleanupMsg)
	finalResult += fmt.Sprintf("。实例状态: %s", summarizeInstanceStatuses(instanceStatuses, 20, 1300))
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0])
		if len(errorDetails) > 1 {
			finalResult += fmt.Sprintf(" 等 %d 个错误", len(errorDetails))
		}
	}

	if failedCount == 0 {
		taskLogger.Success(finalResult)
	} else {
		taskLogger.Failed(finalResult)
	}

	logger.Info(finalResult)
}

func getDbCon(datasourceType, host, port, user, origPass, dbid string) *sql.DB {
	if datasourceType == "MySQL" || datasourceType == "TiDB" || datasourceType == "Doris" || datasourceType == "MariaDB" || datasourceType == "GreatSQL" || datasourceType == "OceanBase" {
		dbCon, err = database.Connect(database.WithDriver("mysql"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("information_schema"))
		if err != nil {
			log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
			return nil
		}
	} else if datasourceType == "PostgreSQL" {
		pgDatabase := dbid
		if pgDatabase == "" {
			pgDatabase = "postgres"
		}
		dbCon, err = database.Connect(database.WithDriver("postgres"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase(pgDatabase))
		if err != nil {
			log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
			return nil
		}
	} else if datasourceType == "ClickHouse" {
		dbCon, err = database.Connect(database.WithDriver("clickhouse"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("system"))
		if err != nil {
			log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
			return nil
		}
	}
	return dbCon
}

func doDbMetaCollectorTask(datasourceType, host, port, user, origPass, dbid string) error {
	if datasourceType == "MongoDB" {
		return doMongoMetaCollectorTask(host, port, user, origPass, dbid)
	}

	var db = database.DB
	var (
		queryDatabaseSql string
		queryTableSql    string
		queryColumnSql   string
	)
	if datasourceType == "MySQL" || datasourceType == "TiDB" || datasourceType == "Doris" || datasourceType == "MariaDB" || datasourceType == "GreatSQL" || datasourceType == "OceanBase" {
		queryDatabaseSql = "select lower(schema_name) as database_name,lower(schema_name) as schema_name,default_character_set_name as characters from information_schema.schemata where lower(schema_name) not in ('information_schema','performance_schema','sys','mysql','metrics_schema','__internal_schema','sys_audit','lbacsys','oceanbase','ocs','oraauditor') order by database_name asc"
		queryTableSql = "select table_type as table_type,lower(table_schema) as database_name,lower(table_name) as table_name,table_comment as table_comment,table_collation as characters from information_schema.tables where lower(table_schema) not in ('information_schema','performance_schema','sys','mysql','metrics_schema','__internal_schema','sys_audit','lbacsys','oceanbase','ocs','oraauditor') order by database_name asc,table_name asc"
		queryColumnSql = "select lower(table_schema)  as database_name,lower(table_name) as table_name,lower(column_name) as column_name,  lower(column_comment) as column_comment, lower(data_type) as data_type,lower(is_nullable) as is_nullable,lower(column_default) as default_value,lower(ordinal_position) as ordinal_position,lower(collation_name) as characters from information_schema.COLUMNS where lower(table_schema) not in ('information_schema','performance_schema','sys','mysql','metrics_schema','__internal_schema','sys_audit','lbacsys','oceanbase','ocs','oraauditor')  order by table_name asc,ordinal_position asc"
	} else if datasourceType == "PostgreSQL" {
		queryDatabaseSql = "select lower(current_database()) as database_name,lower(nspname) as schema_name,'' as characters from pg_namespace where nspname not in ('pg_catalog','information_schema') and nspname not like 'pg_toast%' and nspname not like 'pg_temp_%' order by schema_name asc"
		queryTableSql = "select t.table_type as table_type,lower(current_database()) as database_name,lower(concat(t.table_schema,'.',t.table_name)) as table_name,coalesce(obj_description(c.oid,'pg_class'),'') as table_comment,'' as characters from information_schema.tables t join pg_namespace n on n.nspname=t.table_schema join pg_class c on c.relname=t.table_name and c.relnamespace=n.oid where t.table_schema not in ('pg_catalog','information_schema') and t.table_type in ('BASE TABLE','VIEW') order by table_name asc"
		queryColumnSql = "select lower(current_database()) as database_name,lower(concat(c.table_schema,'.',c.table_name)) as table_name,lower(c.column_name) as column_name,coalesce(col_description(pc.oid,c.ordinal_position),'') as column_comment,lower(c.data_type) as data_type,lower(c.is_nullable) as is_nullable,coalesce(c.column_default,'') as default_value,lower(c.ordinal_position::text) as ordinal_position,coalesce(c.collation_name,'') as characters from information_schema.columns c join pg_namespace n on n.nspname=c.table_schema join pg_class pc on pc.relname=c.table_name and pc.relnamespace=n.oid where c.table_schema not in ('pg_catalog','information_schema') order by table_name asc,c.ordinal_position asc"
	} else if datasourceType == "ClickHouse" {
		queryDatabaseSql = "select lower(name) as database_name,lower(name) as schema_name,'' as characters from databases where lower(name) not in ('information_schema','INFORMATION_SCHEMA','system') order by name asc"
		//queryTableSql = "select engine as table_type,lower(`database`) as database_name,lower(name) as table_name,comment as table_comment,'' as characters from tables where database_name not in ('information_schema','INFORMATION_SCHEMA','system') order by database_name asc,table_name asc limit 100"
		//queryColumnSql = "select lower(`database`) as database_name,lower(`table`) as table_name, lower(name) as column_name,comment as column_comment,type as data_type,'' as is_nullable, '' as default_value, toString(position) as ordinal_position,'' as characters from columns where database_name not in ('information_schema','INFORMATION_SCHEMA','system') order by database_name asc,table_name asc,ordinal_position asc"
		queryTableSql = "select engine as table_type,lower(`database`) as database_name,name as table_name,comment as table_comment,'' as characters from tables where database_name not in ('information_schema','INFORMATION_SCHEMA','system')  order by database_name asc,table_name asc limit 100"
		queryColumnSql = "select lower(`database`) as database_name,lower(`table`) as table_name, lower(name) as column_name,comment as column_comment,type as data_type,'' as is_nullable, '' as default_value, toString(position) as ordinal_position,'' as characters from columns where database_name not in ('information_schema','INFORMATION_SCHEMA','system') order by database_name asc,table_name asc,ordinal_position asc"

		// } else if datasourceType == "PostgreSQL" {
		// 	queryDatabaseSql = "select pg_database.datname as database_name,pg_database.datname as schema_name,pg_encoding_to_char(encoding) as characters from pg_database where datname not in ('postgres','template0','template1') order by database_name asc"
		// 	queryTableSql = ""
		// 	dbCon, err = database.Connect(database.WithDriver("postgres"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("postgres"))
		// 	if err != nil {
		// 		log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
		// 		return
		// 	}
		// 	//defer dbCon.Close()
		// } else if datasourceType == "Oracle" {
		// 	queryDatabaseSql = "select username as database_name,username as schema_name,'' as characters from dba_users where username not in ('SYSTEM','SYS') order by username asc;"
		// 	queryTableSql = ""
		// 	dbCon, err = database.Connect(database.WithDriver("godror"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase(dbid))
		// 	if err != nil {
		// 		log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
		// 		return
		// 	}
		// 	//defer dbCon.Close()
		// } else if datasourceType == "SQLServer" {
		// 	queryDatabaseSql = "SELECT name as database_name,name as schema_name,collation_name as characters FROM sys.databases where name not in ('master','tempdb','msdb','model') order by name asc"
		// 	queryTableSql = "SELECT o.type_desc AS table_type, DB_NAME() AS database_name,  o.name AS table_name, CAST(ep.value AS NVARCHAR(MAX)) AS table_comment,'' as characters FROM  sys.objects o LEFT JOIN  sys.extended_properties ep ON o.object_id = ep.major_id AND ep.name = 'MS_Description'   WHERE  o.type IN ('U')  AND o.is_ms_shipped = 0"
		// 	queryColumnSql = "SELECT DB_NAME() AS database_name,  t.name AS table_Name, c.name AS column_name, '' as column_comment,ty.name AS data_type, '' as is_nullable, OBJECT_DEFINITION(c.default_object_id) AS default_value,'' as ordinal_position,'' as characters FROM sys.tables t INNER JOIN sys.columns c ON t.object_id = c.object_id LEFT JOIN sys.types ty ON c.system_type_id = ty.system_type_id AND c.user_type_id = ty.user_type_id WHERE t.is_ms_shipped = 0"
		// 	dbCon, err = database.Connect(database.WithDriver("mssql"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("master"))
		// 	fmt.Println(queryTableSql)
		// 	fmt.Println(dbCon)
		// 	if err != nil {
		// 		log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
		// 		return
		// 	}
		// 	//defer dbCon.Close()
	} else {
		return fmt.Errorf("不支持的数据库类型: %s", datasourceType)
	}

	//采集数据库列表
	dbCon = getDbCon(datasourceType, host, port, user, origPass, dbid)
	if dbCon == nil {
		return fmt.Errorf("无法连接到数据库 %s:%s", host, port)
	}
	defer dbCon.Close()

	databaseList, err := database.QueryRemote(dbCon, queryDatabaseSql)
	if err != nil {
		return fmt.Errorf("查询数据库列表失败: %v", err)
	}
	for _, item := range databaseList {
		var dataList []model.MetaDatabase
		if item["database_name"] == nil || item["schema_name"] == nil {
			return fmt.Errorf("数据库元数据格式错误: %v", item)
		}
		db.Where("host=?", host).Where("port=?", port).Where("database_name=?", item["database_name"].(string)).Where("schema_name=?", item["schema_name"].(string)).Find(&dataList)
		if (len(dataList)) == 0 {
			var record model.MetaDatabase
			record.DatasourceType = datasourceType
			record.Host = host
			record.Port = port
			record.DatabaseName = item["database_name"].(string)
			record.SchemaName = item["schema_name"].(string)
			if item["characters"] != nil {
				record.Characters = item["characters"].(string)
			} else {
				record.Characters = ""
			}
			result := database.DB.Create(&record)
			if result.Error != nil {
				fmt.Println(result.Error.Error())
				log.Logger.Error(fmt.Sprintf("Can't collector database on %s:%s, %s", host, port, result.Error.Error()))
				return fmt.Errorf("创建数据库元数据失败: %s", result.Error.Error())
			}
		} else {
			var record model.MetaDatabase
			record.Characters = formatInterface(item["characters"])
			record.IsDeleted = 0
			result := db.Model(&record).Select("characters", "is_deleted").Omit("id").Where("host=?", host).Where("port=?", port).Where("database_name=?", item["database_name"].(string)).Where("schema_name=?", item["schema_name"].(string)).Updates(&record)
			if result.Error != nil {
				log.Logger.Error(fmt.Sprintf("Can't collector database on %s:%s, %s", host, port, result.Error.Error()))
				return fmt.Errorf("更新数据库元数据失败: %s", result.Error.Error())
			}
		}
		time.Sleep(1 * time.Millisecond)
	}

	//采集数据表列表
	dbCon = getDbCon(datasourceType, host, port, user, origPass, dbid)
	if dbCon == nil {
		return fmt.Errorf("无法连接到数据库 %s:%s", host, port)
	}
	defer dbCon.Close()
	tableList, err := database.QueryRemote(dbCon, queryTableSql)
	if err != nil {
		fmt.Println(err)
		log.Logger.Error(fmt.Sprintf("Can't query table meta on %s:%s, %s", host, port, err))
		return fmt.Errorf("查询数据表列表失败: %v", err)
	}

	for _, item := range tableList {
		var dataList []model.MetaTable
		db.Where("host=?", host).Where("port=?", port).Where("database_name=?", item["database_name"].(string)).Where("table_name=?", item["table_name"].(string)).Find(&dataList)
		if (len(dataList)) == 0 {
			//fmt.Println(dataList)
			var record model.MetaTable
			record.DatasourceType = datasourceType
			record.Host = host
			record.Port = port
			record.DatabaseName = item["database_name"].(string)
			record.TableType = formatInterface(item["table_type"])
			record.TableNameX = item["table_name"].(string)
			record.TableComment = formatInterface(item["table_comment"])
			record.Characters = formatInterface(item["characters"])
			result := database.DB.Create(&record)
			if result.Error != nil {
				log.Logger.Error(fmt.Sprintf("Can't collector table on %s:%s, %s", host, port, result.Error.Error()))
				return fmt.Errorf("创建数据表元数据失败: %s", result.Error.Error())
			}
		} else {
			var record model.MetaTable
			record.TableType = formatInterface(item["table_type"])
			record.TableComment = formatInterface(item["table_comment"])
			record.Characters = formatInterface(item["characters"])
			result := db.Model(&record).Select("table_comment", "table_type", "characters").Omit("id").Where("host=?", host).Where("port=?", port).Where("database_name=?", item["database_name"].(string)).Where("table_name=?", item["table_name"].(string)).Updates(&record)
			if result.Error != nil {
				log.Logger.Error(fmt.Sprintf("Can't collector table on %s:%s, %s", host, port, result.Error.Error()))
				return fmt.Errorf("更新数据表元数据失败: %s", result.Error.Error())
			}
		}
		time.Sleep(1 * time.Millisecond)
	}

	//采集字段列表
	dbCon = getDbCon(datasourceType, host, port, user, origPass, dbid)
	if dbCon == nil {
		return fmt.Errorf("无法连接到数据库 %s:%s", host, port)
	}
	defer dbCon.Close()
	columnList, err := database.QueryRemote(dbCon, queryColumnSql)
	if err != nil {
		fmt.Println(err)
		log.Logger.Error(fmt.Sprintf("Can't query column meta on %s:%s, %s", host, port, err))
		return fmt.Errorf("查询字段列表失败: %v", err)
	}
	for _, item := range columnList {
		var dataList []model.MetaColumn
		db.Where("host=?", host).Where("port=?", port).Where("database_name=?", item["database_name"].(string)).Where("table_name=?", item["table_name"].(string)).Where("column_name=?", item["column_name"].(string)).Find(&dataList)
		if (len(dataList)) == 0 {
			var record model.MetaColumn
			record.DatasourceType = datasourceType
			record.Host = host
			record.Port = port
			record.DatabaseName = item["database_name"].(string)
			record.TableNameX = item["table_name"].(string)
			record.ColumnName = item["column_name"].(string)
			record.ColumnComment = formatInterface(item["column_comment"])
			record.DataType = formatInterface(item["data_type"])
			record.IsNullable = formatInterface(item["is_nullable"])
			record.DefaultValue = formatInterface(item["default_value"])
			record.Ordinal_Position = utils.StrToInt(item["ordinal_position"].(string))
			record.Characters = formatInterface(item["characters"])
			result := database.DB.Create(&record)
			if result.Error != nil {
				log.Logger.Error(fmt.Sprintf("Can't collector column on %s:%s, %s", host, port, result.Error.Error()))
				return fmt.Errorf("创建字段元数据失败: %s", result.Error.Error())
			}
		} else {
			var record model.MetaColumn
			record.ColumnComment = formatInterface(item["column_comment"])
			record.DataType = formatInterface(item["data_type"])
			record.IsNullable = formatInterface(item["is_nullable"])
			record.DefaultValue = formatInterface(item["default_value"])
			record.Ordinal_Position = utils.StrToInt(item["ordinal_position"].(string))
			record.Characters = formatInterface(item["characters"])
			result := db.Model(&record).Select("column_comment", "data_type", "is_nullable", "default_value", "ordinal_position", "characters").Omit("id").Where("host=?", host).Where("port=?", port).Where("database_name=?", item["database_name"].(string)).Where("table_name=?", item["table_name"].(string)).Where("column_name=?", item["column_name"].(string)).Updates(&record)
			if result.Error != nil {
				log.Logger.Error(fmt.Sprintf("Can't collector column on %s:%s, %s", host, port, result.Error.Error()))
				return fmt.Errorf("更新字段元数据失败: %s", result.Error.Error())
			}
		}
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// ExecuteDbMetaTask 导出函数，用于手动执行任务
func ExecuteDbMetaTask() {
	doDbMetaTask()
}

func doMongoMetaCollectorTask(host, port, user, origPass, dbid string) error {
	var db = database.DB
	client, err := mongodb.Connect(host, port, user, origPass, dbid)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer client.Disconnect(context.Background())

	databaseNames, err := mongodb.ListDatabase(client)
	if err != nil {
		return fmt.Errorf("查询MongoDB数据库列表失败: %v", err)
	}

	for _, databaseName := range databaseNames {
		if isMongoSystemDatabase(databaseName) {
			continue
		}
		var dataList []model.MetaDatabase
		db.Where("host=?", host).Where("port=?", port).Where("database_name=?", databaseName).Where("schema_name=?", databaseName).Find(&dataList)
		if len(dataList) == 0 {
			record := model.MetaDatabase{
				DatasourceType: "MongoDB",
				Host:           host,
				Port:           port,
				DatabaseName:   databaseName,
				SchemaName:     databaseName,
				Characters:     "",
			}
			result := db.Create(&record)
			if result.Error != nil {
				return fmt.Errorf("创建MongoDB数据库元数据失败: %s", result.Error.Error())
			}
		} else {
			record := model.MetaDatabase{
				Characters: "",
				IsDeleted:  0,
			}
			result := db.Model(&record).Select("characters", "is_deleted").Omit("id").Where("host=?", host).Where("port=?", port).Where("database_name=?", databaseName).Where("schema_name=?", databaseName).Updates(&record)
			if result.Error != nil {
				return fmt.Errorf("更新MongoDB数据库元数据失败: %s", result.Error.Error())
			}
		}
		time.Sleep(1 * time.Millisecond)
	}

	for _, databaseName := range databaseNames {
		if isMongoSystemDatabase(databaseName) {
			continue
		}
		collectionNames, err := mongodb.ListCollection(client, databaseName)
		if err != nil {
			return fmt.Errorf("查询MongoDB数据表列表失败(%s): %v", databaseName, err)
		}
		for _, collectionName := range collectionNames {
			var dataList []model.MetaTable
			db.Where("host=?", host).Where("port=?", port).Where("database_name=?", databaseName).Where("table_name=?", collectionName).Find(&dataList)
			if len(dataList) == 0 {
				record := model.MetaTable{
					DatasourceType: "MongoDB",
					Host:           host,
					Port:           port,
					DatabaseName:   databaseName,
					TableType:      "collection",
					TableNameX:     collectionName,
					TableComment:   "",
					Characters:     "",
					IsDeleted:      0,
				}
				result := db.Create(&record)
				if result.Error != nil {
					return fmt.Errorf("创建MongoDB数据表元数据失败: %s", result.Error.Error())
				}
			} else {
				record := model.MetaTable{
					TableType:    "collection",
					TableComment: "",
					Characters:   "",
					IsDeleted:    0,
				}
				result := db.Model(&record).Select("table_comment", "table_type", "characters", "is_deleted").Omit("id").Where("host=?", host).Where("port=?", port).Where("database_name=?", databaseName).Where("table_name=?", collectionName).Updates(&record)
				if result.Error != nil {
					return fmt.Errorf("更新MongoDB数据表元数据失败: %s", result.Error.Error())
				}
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
	return nil
}

func isMongoSystemDatabase(databaseName string) bool {
	name := strings.ToLower(strings.TrimSpace(databaseName))
	return name == "admin" || name == "config" || name == "local"
}

func formatDatasourceInstance(datasource model.Datasource) string {
	if datasource.Dbid != "" {
		return fmt.Sprintf("%s(%s %s:%s/%s)", datasource.Name, datasource.Type, datasource.Host, datasource.Port, datasource.Dbid)
	}
	return fmt.Sprintf("%s(%s %s:%s)", datasource.Name, datasource.Type, datasource.Host, datasource.Port)
}

func truncateText(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}

func summarizeInstanceStatuses(statuses []string, maxItems, maxRunes int) string {
	if len(statuses) == 0 {
		return "无"
	}
	items := statuses
	omitted := 0
	if maxItems > 0 && len(statuses) > maxItems {
		items = statuses[:maxItems]
		omitted = len(statuses) - maxItems
	}
	result := strings.Join(items, "；")
	if omitted > 0 {
		result += fmt.Sprintf("；... 其余 %d 个实例", omitted)
	}
	return truncateText(result, maxRunes)
}
