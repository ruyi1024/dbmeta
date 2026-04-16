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
	"database/sql"
	"dbmeta-core/log"
	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/model"
	"dbmeta-core/src/utils"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go pumpkinCrontabTask()
}

func pumpkinCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "gather_pumpkin").Take(&record)
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "gather_pumpkin").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_pumpkin'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doPumpkinTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_pumpkin'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func formatPumpkinInterface(inter interface{}) string {
	if inter == nil {
		return ""
	}

	switch v := inter.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func doPumpkinTask() {
	logger := log.Logger
	logger.Info("开始执行表容量采集任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("gather_pumpkin")
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
				failedCount++
				continue
			}
		}

		err := doPumpkinCollectorTask(datasourceType, host, port, user, origPass, dbid)
		if err != nil {
			errorMsg := fmt.Sprintf("数据源 %s:%s 表容量采集失败: %v", host, port, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount++
		} else {
			successCount++
		}

		// 更新进度
		progressMsg := fmt.Sprintf("已处理 %d/%d 个数据源 (成功: %d, 失败: %d)", i+1, len(dataList), successCount, failedCount)
		taskLogger.UpdateResult(progressMsg)
	}

	// 清理过期数据
	// logger.Info("开始清理过期表容量数据")
	// expireTime := time.Now().Add(-time.Hour * 24).Format("2006-01-02 15:04:05")

	// cleanupResult := database.DB.Model(model.PumpkinTableSize{}).Where("gmt_updated <= ?", expireTime).Delete(&model.PumpkinTableSize{})

	// cleanupMsg := fmt.Sprintf("清理过期表容量数据: %d 条记录", cleanupResult.RowsAffected)
	// logger.Info(cleanupMsg)

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 数据源总计: %d, 成功: %d, 失败: %d",
		len(dataList), successCount, failedCount)
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

func getPumpkinDbCon(datasourceType, host, port, user, origPass, dbid string) *sql.DB {
	var dbCon *sql.DB
	var err error

	switch datasourceType {
	case "MySQL", "TiDB", "Doris", "MariaDB", "GreatSQL", "OceanBase":
		dbCon, err = database.Connect(database.WithDriver("mysql"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("information_schema"))
	case "ClickHouse":
		dbCon, err = database.Connect(database.WithDriver("clickhouse"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("system"))
	}

	if err != nil {
		log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
		return nil
	}
	return dbCon
}

func doPumpkinCollectorTask(datasourceType, host, port, user, origPass, dbid string) error {
	//var db = database.DB
	var queryTableSizeSql string

	switch datasourceType {
	case "MySQL", "TiDB", "Doris", "MariaDB", "GreatSQL", "OceanBase":
		// MySQL系列数据库的表容量查询SQL
		queryTableSizeSql = `
			SELECT 
				table_schema as database_name,
				table_name as table_name,
				data_length as data_size,
				index_length as index_size,
				data_free as free_size,
				table_rows as table_rows,
				avg_row_length as avg_row_length
			FROM information_schema.tables 
			WHERE table_schema NOT IN ('information_schema', 'performance_schema', 'sys', 'mysql', 'metrics_schema', '__internal_schema', 'sys_audit', 'lbacsys', 'oceanbase', 'ocs', 'oraauditor')
			AND table_rows > 0
			and table_type='BASE TABLE'
		`
	// case "ClickHouse":
	// 	// ClickHouse的表容量查询SQL
	// 	queryTableSizeSql = `
	// 		SELECT
	// 			database as database_name,
	// 			name as table_name,
	// 			ROUND((data_compressed_bytes / 1024 / 1024), 2) as data_size,
	// 			ROUND((data_uncompressed_bytes / 1024 / 1024), 2) as index_size,
	// 			0 as free_size,
	// 			rows as table_rows,
	// 			ROUND((data_uncompressed_bytes / rows), 2) as avg_row_length
	// 		FROM system.tables
	// 		WHERE database NOT IN ('information_schema', 'INFORMATION_SCHEMA', 'system')
	// 		AND rows > 0
	// 		ORDER BY data_compressed_bytes DESC
	//	`
	default:
		return fmt.Errorf("不支持的数据库类型: %s", datasourceType)
	}

	// 连接数据库
	dbCon := getPumpkinDbCon(datasourceType, host, port, user, origPass, dbid)
	if dbCon == nil {
		return fmt.Errorf("无法连接到数据库 %s:%s", host, port)
	}
	defer dbCon.Close()

	// 查询表容量数据
	tableSizeList, err := database.QueryRemote(dbCon, queryTableSizeSql)
	if err != nil {
		return fmt.Errorf("查询表容量数据失败: %v", err)
	}

	// 处理查询结果
	for _, item := range tableSizeList {
		// 检查必要字段是否存在
		if item["database_name"] == nil || item["table_name"] == nil {
			log.Logger.Warn("跳过无效记录：缺少必要字段", zap.Any("item", item))
			continue
		}

		databaseName := formatPumpkinInterface(item["database_name"])
		tableName := formatPumpkinInterface(item["table_name"])

		if databaseName == "" || tableName == "" {
			log.Logger.Warn("跳过无效记录：字段为空", zap.String("database_name", databaseName), zap.String("table_name", tableName))
			continue
		}

		// 安全转换数据类型，处理nil值
		var dataSize int64 = 0
		if item["data_size"] != nil {
			dataSize = utils.StrToInt64(formatPumpkinInterface(item["data_size"]))
		}

		var indexSize int64 = 0
		if item["index_size"] != nil {
			indexSize = utils.StrToInt64(formatPumpkinInterface(item["index_size"]))
		}

		var freeSize int64 = 0
		if item["free_size"] != nil {
			freeSize = utils.StrToInt64(formatPumpkinInterface(item["free_size"]))
		}

		var tableRows int64 = 0
		if item["table_rows"] != nil {
			tableRows = utils.StrToInt64(formatPumpkinInterface(item["table_rows"]))
		}

		var avgRowLength int64 = 0
		if item["avg_row_length"] != nil {
			avgRowLength = utils.StrToInt64(formatPumpkinInterface(item["avg_row_length"]))
		}

		var record model.PumpkinTableSize
		record.DatasourceType = datasourceType
		record.Host = host
		record.Port = port
		record.DatabaseName = databaseName
		record.TableNameField = tableName
		record.DataSize = dataSize
		record.IndexSize = indexSize
		record.FreeSize = freeSize
		record.TableRows = tableRows
		record.AvgRowLength = avgRowLength

		result := database.DB.Create(&record)
		if result.Error != nil {
			log.Logger.Error(fmt.Sprintf("Can't create table size record on %s:%s, %s", host, port, result.Error.Error()))
			return fmt.Errorf("创建表容量记录失败: %s", result.Error.Error())
		}

		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// ExecutePumpkinTask 导出函数，用于手动执行任务
func ExecutePumpkinTask() {
	doPumpkinTask()
}
