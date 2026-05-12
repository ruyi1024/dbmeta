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
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func init() {
	go pumpkinGrowthCrontabTask()
}

func pumpkinGrowthCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "gather_pumpkin_growth").Take(&record)

	// 若任务配置中 crontab 为空，使用默认：每小时第 30 分执行（与 gather_pumpkin 整点采集错开）
	if record.Crontab == "" {
		record.Crontab = "30 * * * *"
	}

	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "gather_pumpkin_growth").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_pumpkin_growth'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doPumpkinGrowthTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_pumpkin_growth'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

// doPumpkinGrowthTask 执行容量增长计算任务：基于 pumpkin_table_size_history，
// 对比「上一完整小时」与「当前小时内截至任务执行时刻」各表最新快照，写入表级与库级增长（stat_hour 为当前日历小时整点）。
func doPumpkinGrowthTask() {
	logger := log.Logger
	logger.Info("开始执行容量增长计算任务")

	taskLogger := NewTaskLogger("gather_pumpkin_growth")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	now := time.Now()
	loc := now.Location()
	statHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc)
	prevHourStart := statHour.Add(-time.Hour)

	prevStartStr := prevHourStart.Format("2006-01-02 15:04:05.999")
	prevEndStr := statHour.Format("2006-01-02 15:04:05.999")
	currStartStr := statHour.Format("2006-01-02 15:04:05.999")
	currEndStr := now.Format("2006-01-02 15:04:05.999")

	logger.Info("环比小时窗口",
		zap.String("上一小时快照区间", fmt.Sprintf("[%s, %s)", prevStartStr, prevEndStr)),
		zap.String("当前小时快照区间", fmt.Sprintf("[%s, %s]", currStartStr, currEndStr)),
		zap.Time("stat_hour", statHour),
	)
	taskLogger.UpdateResult(fmt.Sprintf("环比: 上一小时 [%s,%s) vs 当前小时 [%s,%s], stat_hour=%s",
		prevStartStr, prevEndStr, currStartStr, currEndStr, statHour.Format("2006-01-02 15:04:05")))

	err := calculatePumpkinHourlyGrowth(statHour, prevStartStr, prevEndStr, currStartStr, currEndStr)
	if err != nil {
		errorMsg := fmt.Sprintf("容量增长计算失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	successMsg := "容量增长计算任务完成"
	logger.Info(successMsg)
	taskLogger.Success(successMsg)
}

type dbGrowthAgg struct {
	datasourceType   string
	host             string
	port             string
	databaseName     string
	databaseSize     int64
	databaseRows     int64
	databaseSizeIncr int64
	databaseRowsIncr int64
	tableNames       map[string]struct{}
}

func dbAggKey(dt, host, port, db string) string {
	return fmt.Sprintf("%s|%s|%s|%s", dt, host, port, db)
}

// calculatePumpkinHourlyGrowth 从 history 取上一小时末与当前小时内最新快照，算增量并写入 pumpkin_table_growth / pumpkin_database_growth。
func calculatePumpkinHourlyGrowth(statHour time.Time, prevStartStr, prevEndStr, currStartStr, currEndStr string) error {
	var db = database.DB
	logger := log.Logger

	currentDataSQL := `
		SELECT 
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			(t1.data_size + t1.index_size + t1.free_size) as table_size,
			t1.table_rows as table_rows,
			t1.snapshot_at as max_created
		FROM pumpkin_table_size_history t1
		INNER JOIN (
			SELECT 
				datasource_type, host, port, database_name, table_name,
				MAX(snapshot_at) as max_created
			FROM pumpkin_table_size_history
			WHERE snapshot_at >= ? AND snapshot_at <= ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.table_name = t2.table_name 
			AND t1.snapshot_at = t2.max_created
	`

	var currentData []struct {
		DatasourceType string    `gorm:"column:datasource_type"`
		Host           string    `gorm:"column:host"`
		Port           string    `gorm:"column:port"`
		DatabaseName   string    `gorm:"column:database_name"`
		TableName      string    `gorm:"column:table_name"`
		TableSize      int64     `gorm:"column:table_size"`
		TableRows      int64     `gorm:"column:table_rows"`
		MaxCreated     time.Time `gorm:"column:max_created"`
	}

	if err := db.Raw(currentDataSQL, currStartStr, currEndStr).Scan(&currentData).Error; err != nil {
		return fmt.Errorf("查询当前小时表容量快照失败: %w", err)
	}
	logger.Info("当前小时表快照条数", zap.Int("count", len(currentData)))

	previousDataSQL := `
		SELECT 
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			(t1.data_size + t1.index_size + t1.free_size) as table_size,
			t1.table_rows as table_rows
		FROM pumpkin_table_size_history t1
		INNER JOIN (
			SELECT 
				datasource_type, host, port, database_name, table_name,
				MAX(snapshot_at) as max_created
			FROM pumpkin_table_size_history
			WHERE snapshot_at >= ? AND snapshot_at < ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.table_name = t2.table_name 
			AND t1.snapshot_at = t2.max_created
	`

	var previousData []struct {
		DatasourceType string `gorm:"column:datasource_type"`
		Host           string `gorm:"column:host"`
		Port           string `gorm:"column:port"`
		DatabaseName   string `gorm:"column:database_name"`
		TableName      string `gorm:"column:table_name"`
		TableSize      int64  `gorm:"column:table_size"`
		TableRows      int64  `gorm:"column:table_rows"`
	}

	if err := db.Raw(previousDataSQL, prevStartStr, prevEndStr).Scan(&previousData).Error; err != nil {
		return fmt.Errorf("查询上一小时表容量快照失败: %w", err)
	}
	logger.Info("上一小时表快照条数", zap.Int("count", len(previousData)))

	previousMap := make(map[string]struct {
		TableSize int64
		TableRows int64
	})
	for _, item := range previousData {
		key := fmt.Sprintf("%s|%s|%s|%s|%s", item.DatasourceType, item.Host, item.Port, item.DatabaseName, item.TableName)
		previousMap[key] = struct {
			TableSize int64
			TableRows int64
		}{
			TableSize: item.TableSize,
			TableRows: item.TableRows,
		}
	}

	statHourPtr := &statHour
	ts := time.Now()

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM pumpkin_table_growth WHERE stat_hour = ?", statHour).Error; err != nil {
			return fmt.Errorf("清理本小时旧表级增长记录失败: %w", err)
		}
		if err := tx.Exec("DELETE FROM pumpkin_database_growth WHERE stat_hour = ?", statHour).Error; err != nil {
			return fmt.Errorf("清理本小时旧库级增长记录失败: %w", err)
		}

		dbAggs := make(map[string]*dbGrowthAgg)
		successCount := 0
		failedCount := 0

		for _, current := range currentData {
			rowKey := fmt.Sprintf("%s|%s|%s|%s|%s", current.DatasourceType, current.Host, current.Port, current.DatabaseName, current.TableName)

			var tableSizeIncr int64
			var tableRowsIncr int64
			if previous, exists := previousMap[rowKey]; exists {
				tableSizeIncr = current.TableSize - previous.TableSize
				tableRowsIncr = current.TableRows - previous.TableRows
			}

			rowsIncr := tableRowsIncr
			row := model.PumpkinTableGrowth{
				DatasourceType: current.DatasourceType,
				Host:           current.Host,
				Port:           current.Port,
				DatabaseName:   current.DatabaseName,
				TableNameX:     current.TableName,
				TableSize:      current.TableSize,
				TableRows:      current.TableRows,
				TableSizeIncr:  tableSizeIncr,
				TableRowsIncr:  &rowsIncr,
				StatHour:       statHourPtr,
				CreatedAt:      ts,
				UpdatedAt:      ts,
			}
			if err := tx.Create(&row).Error; err != nil {
				logger.Error("插入表容量增长记录失败", zap.Error(err), zap.String("table", current.TableName))
				failedCount++
				continue
			}
			successCount++

			dk := dbAggKey(current.DatasourceType, current.Host, current.Port, current.DatabaseName)
			agg, ok := dbAggs[dk]
			if !ok {
				agg = &dbGrowthAgg{
					datasourceType: current.DatasourceType,
					host:           current.Host,
					port:           current.Port,
					databaseName:   current.DatabaseName,
					tableNames:     make(map[string]struct{}),
				}
				dbAggs[dk] = agg
			}
			agg.databaseSize += current.TableSize
			agg.databaseRows += current.TableRows
			agg.databaseSizeIncr += tableSizeIncr
			agg.databaseRowsIncr += tableRowsIncr
			agg.tableNames[current.TableName] = struct{}{}
		}

		logger.Info("表容量增长写入完成", zap.Int("成功", successCount), zap.Int("失败", failedCount))

		dbSuccess := 0
		dbFailed := 0
		for _, agg := range dbAggs {
			growth := model.PumpkinDatabaseGrowth{
				DatasourceType:   agg.datasourceType,
				Host:             agg.host,
				Port:             agg.port,
				DatabaseName:     agg.databaseName,
				DatabaseSize:     agg.databaseSize,
				DatabaseRows:     agg.databaseRows,
				TableCount:       int64(len(agg.tableNames)),
				DatabaseSizeIncr: agg.databaseSizeIncr,
				DatabaseRowsIncr: agg.databaseRowsIncr,
				StatHour:         statHourPtr,
				CreatedAt:        ts,
				UpdatedAt:        ts,
			}
			if err := tx.Create(&growth).Error; err != nil {
				logger.Error("插入数据库容量增长记录失败", zap.Error(err), zap.String("database", agg.databaseName))
				dbFailed++
				continue
			}
			dbSuccess++
		}
		logger.Info("库容量增长写入完成", zap.Int("成功", dbSuccess), zap.Int("失败", dbFailed))
		return nil
	})
}

// ExecutePumpkinGrowthTask 导出函数，用于手动执行任务
func ExecutePumpkinGrowthTask() {
	doPumpkinGrowthTask()
}
