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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/setting"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/libary/mongodb"
	"github.com/ruyi1024/dbmeta/src/model"
	"github.com/ruyi1024/dbmeta/src/utils"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

		err := doPumpkinCollectorTask(datasourceType, host, port, user, origPass, dbid)
		if err != nil {
			errorMsg := fmt.Sprintf("数据源 %s:%s 表容量采集失败: %v", host, port, err)
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

	// 清理过期数据
	// logger.Info("开始清理过期表容量数据")
	// expireTime := time.Now().Add(-time.Hour * 24).Format("2006-01-02 15:04:05")

	// cleanupResult := database.DB.Model(model.PumpkinTableSize{}).Where("gmt_updated <= ?", expireTime).Delete(&model.PumpkinTableSize{})

	// cleanupMsg := fmt.Sprintf("清理过期表容量数据: %d 条记录", cleanupResult.RowsAffected)
	// logger.Info(cleanupMsg)

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 数据源总计: %d, 成功: %d, 失败: %d",
		len(dataList), successCount, failedCount)
	finalResult += fmt.Sprintf("。实例状态: %s", summarizeInstanceStatuses(instanceStatuses, 20, 1300))
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0])
		if len(errorDetails) > 1 {
			finalResult += fmt.Sprintf(" 等 %d 个错误", len(errorDetails))
		}
	}

	// 记录最终结果：脚本完整跑完即视为任务成功；单个数据源失败只记日志与结果摘要，不将整任务标为失败
	taskLogger.Success(finalResult)
	logger.Info(finalResult)
}

func getPumpkinDbCon(datasourceType, host, port, user, origPass, dbid string) *sql.DB {
	var dbCon *sql.DB
	var err error

	t := strings.TrimSpace(datasourceType)
	switch {
	case strings.EqualFold(t, "MySQL") || strings.EqualFold(t, "TiDB") || strings.EqualFold(t, "Doris") || strings.EqualFold(t, "MariaDB") || strings.EqualFold(t, "GreatSQL") || strings.EqualFold(t, "OceanBase"):
		dbCon, err = database.Connect(database.WithDriver("mysql"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("information_schema"))
	case strings.EqualFold(t, "PostgreSQL") || strings.EqualFold(t, "Postgres"):
		pgDatabase := dbid
		if pgDatabase == "" {
			pgDatabase = "postgres"
		}
		dbCon, err = database.Connect(database.WithDriver("postgres"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase(pgDatabase))
	case strings.EqualFold(t, "SQLServer") || strings.EqualFold(t, "Mssql"):
		sqlServerDatabase := dbid
		if sqlServerDatabase == "" {
			sqlServerDatabase = "master"
		}
		dbCon, err = database.Connect(database.WithDriver("mssql"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase(sqlServerDatabase))
	case strings.EqualFold(t, "ClickHouse"):
		dbCon, err = database.Connect(database.WithDriver("clickhouse"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase("system"))
	case t == "达梦数据库":
		dbCon, err = database.Connect(database.WithDriver("dm"), database.WithHost(host), database.WithPort(port), database.WithUsername(user), database.WithPassword(origPass), database.WithDatabase(dbid))
	}

	if err != nil {
		log.Logger.Error(fmt.Sprintf("Can't connect server on %s:%s, %s", host, port, err))
		return nil
	}
	return dbCon
}

// pumpkinTableGatherMeta 引擎侧表时间（可选），用于填充 pumpkin_table_lifecycle
type pumpkinTableGatherMeta struct {
	TableCreateTime *time.Time
	TableUpdateTime *time.Time
}

func parsePumpkinOptionalTime(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case time.Time:
		if t.IsZero() {
			return nil
		}
		return &t
	case []byte:
		return parsePumpkinOptionalTime(string(t))
	case string:
		s := strings.TrimSpace(t)
		if s == "" || strings.HasPrefix(s, "0000-00-00") {
			return nil
		}
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02T15:04:05",
			time.RFC3339,
		}
		for _, layout := range layouts {
			if parsed, err := time.ParseInLocation(layout, s, time.Local); err == nil {
				return &parsed
			}
		}
	default:
		return parsePumpkinOptionalTime(formatPumpkinInterface(v))
	}
	return nil
}

func metaFromPumpkinRow(item map[string]interface{}) *pumpkinTableGatherMeta {
	if item == nil {
		return nil
	}
	var c, u *time.Time
	if v, ok := item["table_create_time"]; ok {
		c = parsePumpkinOptionalTime(v)
	}
	if v, ok := item["table_update_time"]; ok {
		u = parsePumpkinOptionalTime(v)
	}
	if c == nil && u == nil {
		return nil
	}
	return &pumpkinTableGatherMeta{TableCreateTime: c, TableUpdateTime: u}
}

func normalizePumpkinRowKeysToLower(data []map[string]interface{}) []map[string]interface{} {
	if len(data) == 0 {
		return data
	}
	result := make([]map[string]interface{}, 0, len(data))
	for _, row := range data {
		if row == nil {
			continue
		}
		normalized := make(map[string]interface{}, len(row))
		for k, v := range row {
			normalized[strings.ToLower(k)] = v
		}
		result = append(result, normalized)
	}
	return result
}

// savePumpkinTableSizeSnapshot 写入历史表，并在 pumpkin_table_size 中 upsert 该表最新一条容量（同事务）。
// 表生命周期由独立任务 gather_table_lifecycle 维护，见 gather_table_lifecycle.go。
func savePumpkinTableSizeSnapshot(db *gorm.DB, record *model.PumpkinTableSize, snapshotAt time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		h := model.PumpkinTableSizeHistory{
			DatasourceType: record.DatasourceType,
			Host:           record.Host,
			Port:           record.Port,
			DatabaseName:   record.DatabaseName,
			TableNameField: record.TableNameField,
			DataSize:       record.DataSize,
			IndexSize:      record.IndexSize,
			FreeSize:       record.FreeSize,
			TableRows:      record.TableRows,
			AvgRowLength:   record.AvgRowLength,
			SnapshotAt:     snapshotAt,
		}
		if err := tx.Create(&h).Error; err != nil {
			return err
		}

		var existing model.PumpkinTableSize
		err := tx.Where("datasource_type = ? AND host = ? AND port = ? AND database_name = ? AND table_name = ?",
			record.DatasourceType, record.Host, record.Port, record.DatabaseName, record.TableNameField).Take(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(record).Error
		}
		if err != nil {
			return err
		}
		now := time.Now()
		return tx.Model(&existing).Updates(map[string]interface{}{
			"data_size":      record.DataSize,
			"index_size":     record.IndexSize,
			"free_size":      record.FreeSize,
			"table_rows":     record.TableRows,
			"avg_row_length": record.AvgRowLength,
			"gmt_updated":    now,
		}).Error
	})
}

func doPumpkinCollectorTask(datasourceType, host, port, user, origPass, dbid string) error {
	if strings.EqualFold(strings.TrimSpace(datasourceType), "MongoDB") {
		return doMongoPumpkinCollectorTask(host, port, user, origPass, dbid)
	}

	tableSizeList, err := queryPumpkinRemoteTableRows(datasourceType, host, port, user, origPass, dbid)
	if err != nil {
		return err
	}

	snapshotAt := time.Now()
	if err := database.EnsurePumpkinTableSizeHistorySchema(database.DB); err != nil {
		return fmt.Errorf("容量历史表就绪失败: %w", err)
	}

	for _, item := range tableSizeList {
		record, _, ok := pumpkinRemoteRowToPumpkinRecord(datasourceType, host, port, item)
		if !ok {
			continue
		}
		if err := savePumpkinTableSizeSnapshot(database.DB, &record, snapshotAt); err != nil {
			log.Logger.Error(fmt.Sprintf("Can't save table size record on %s:%s, %s", host, port, err.Error()))
			return fmt.Errorf("保存表容量记录失败: %s", err.Error())
		}
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}

// ExecutePumpkinTask 导出函数，用于手动执行任务
func ExecutePumpkinTask() {
	doPumpkinTask()
}

func doMongoPumpkinCollectorTask(host, port, user, origPass, dbid string) error {
	ctx := context.Background()
	client, err := mongodb.Connect(host, port, user, origPass, dbid)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer client.Disconnect(ctx)

	databaseNames, err := mongodb.ListDatabase(client)
	if err != nil {
		return fmt.Errorf("查询MongoDB数据库列表失败: %v", err)
	}

	snapshotAt := time.Now()
	if err := database.EnsurePumpkinTableSizeHistorySchema(database.DB); err != nil {
		return fmt.Errorf("容量历史表就绪失败: %w", err)
	}
	insertedCount := 0
	for _, databaseName := range databaseNames {
		if isMongoSystemDatabase(databaseName) {
			continue
		}

		collectionNames, err := mongodb.ListCollection(client, databaseName)
		if err != nil {
			return fmt.Errorf("查询MongoDB库[%s]集合列表失败: %v", databaseName, err)
		}

		for _, collectionName := range collectionNames {
			var stats bson.M
			if err := client.Database(databaseName).RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats); err != nil {
				log.Logger.Warn("查询MongoDB集合统计失败，已跳过",
					zap.String("host", host),
					zap.String("port", port),
					zap.String("database", databaseName),
					zap.String("collection", collectionName),
					zap.Error(err))
				continue
			}

			storageSize := mongoNumberToInt64(stats["storageSize"])
			dataSize := storageSize
			if dataSize == 0 {
				dataSize = mongoNumberToInt64(stats["size"])
			}
			indexSize := mongoNumberToInt64(stats["totalIndexSize"])
			tableRows := mongoNumberToInt64(stats["count"])
			avgRowLength := mongoNumberToInt64(stats["avgObjSize"])
			freeSize := int64(0)
			rawSize := mongoNumberToInt64(stats["size"])
			if storageSize > rawSize && rawSize >= 0 {
				freeSize = storageSize - rawSize
			}

			record := model.PumpkinTableSize{
				DatasourceType: "MongoDB",
				Host:           host,
				Port:           port,
				DatabaseName:   databaseName,
				TableNameField: collectionName,
				DataSize:       dataSize,
				IndexSize:      indexSize,
				FreeSize:       freeSize,
				TableRows:      tableRows,
				AvgRowLength:   avgRowLength,
			}
			if err := savePumpkinTableSizeSnapshot(database.DB, &record, snapshotAt); err != nil {
				return fmt.Errorf("保存MongoDB表容量记录失败: %s", err.Error())
			}
			insertedCount++
			time.Sleep(1 * time.Millisecond)
		}
	}

	if insertedCount == 0 {
		return fmt.Errorf("未采集到MongoDB表容量数据，请检查权限或集合数据")
	}
	return nil
}

func mongoNumberToInt64(v interface{}) int64 {
	switch val := v.(type) {
	case nil:
		return 0
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case float32:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		return utils.StrToInt64(val)
	default:
		return utils.StrToInt64(formatPumpkinInterface(val))
	}
}
