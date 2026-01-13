/*
Copyright 2014-2022 The Lepus Team Group, website: https://www.lepus.cc
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.gnu.org/licenses/gpl-3.0.html
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
Special note:
Please do not use this source code for any commercial purpose,
or use it for commercial purposes after secondary development, otherwise you may bear legal risks.
*/

package task

import (
	"dbmcloud/log"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go pumpkinGrowthCrontabTask()
}

func pumpkinGrowthCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "gather_pumpkin_growth").Take(&record)

	// 如果任务配置不存在，使用默认的cron表达式（每天凌晨2点执行）
	if record.Crontab == "" {
		record.Crontab = "0 2 * * *" // 每天凌晨2点执行
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

// doPumpkinGrowthTask 执行容量增长计算任务
func doPumpkinGrowthTask() {
	logger := log.Logger
	logger.Info("开始执行容量增长计算任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("gather_pumpkin_growth")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	// 计算24小时前的时间点
	now := time.Now()
	time24HoursAgo := now.Add(-24 * time.Hour)
	time24HoursAgoStr := time24HoursAgo.Format("2006-01-02 15:04:05.999")
	nowStr := now.Format("2006-01-02 15:04:05.999")

	logger.Info("计算时间范围", zap.String("开始时间", time24HoursAgoStr), zap.String("结束时间", nowStr))
	taskLogger.UpdateResult(fmt.Sprintf("计算时间范围: %s 至 %s", time24HoursAgoStr, nowStr))

	// 第一步：从 pumpkin_table_size 计算 pumpkin_table_growth
	err := calculateTableGrowth(time24HoursAgoStr, nowStr)
	if err != nil {
		errorMsg := fmt.Sprintf("计算表容量增长失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	logger.Info("表容量增长计算完成")
	taskLogger.UpdateResult("表容量增长计算完成")

	// 第二步：从 pumpkin_table_growth 计算 pumpkin_database_growth
	err = calculateDatabaseGrowth(nowStr)
	if err != nil {
		errorMsg := fmt.Sprintf("计算数据库容量增长失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	successMsg := "容量增长计算任务完成"
	logger.Info(successMsg)
	taskLogger.Success(successMsg)
}

// calculateTableGrowth 计算表容量增长
func calculateTableGrowth(time24HoursAgoStr, nowStr string) error {
	var db = database.DB
	logger := log.Logger

	// 查询当前时间点的表容量数据（最近1小时内的数据，按表分组取最新的一条）
	// 使用子查询获取每个表的最新记录
	currentDataSQL := `
		SELECT 
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			(t1.data_size + t1.index_size + t1.free_size) as table_size,
			t1.table_rows as table_rows,
			t1.gmt_created as max_created
		FROM pumpkin_table_size t1
		INNER JOIN (
			SELECT 
				datasource_type, host, port, database_name, table_name,
				MAX(gmt_created) as max_created
			FROM pumpkin_table_size
			WHERE gmt_created >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.table_name = t2.table_name 
			AND t1.gmt_created = t2.max_created
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

	if err := db.Raw(currentDataSQL).Scan(&currentData).Error; err != nil {
		return fmt.Errorf("查询当前表容量数据失败: %v", err)
	}

	logger.Info("查询到当前表容量数据", zap.Int("count", len(currentData)))

	// 查询24小时前的表容量数据（获取24小时前最近1小时内的最新数据）
	previousTime := time.Now().Add(-24 * time.Hour)
	previousTimeStart := previousTime.Add(-1 * time.Hour).Format("2006-01-02 15:04:05.999")
	previousTimeEnd := previousTime.Format("2006-01-02 15:04:05.999")

	previousDataSQL := `
		SELECT 
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			(t1.data_size + t1.index_size + t1.free_size) as table_size,
			t1.table_rows as table_rows
		FROM pumpkin_table_size t1
		INNER JOIN (
			SELECT 
				datasource_type, host, port, database_name, table_name,
				MAX(gmt_created) as max_created
			FROM pumpkin_table_size
			WHERE gmt_created >= ? AND gmt_created <= ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.table_name = t2.table_name 
			AND t1.gmt_created = t2.max_created
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

	if err := db.Raw(previousDataSQL, previousTimeStart, previousTimeEnd).Scan(&previousData).Error; err != nil {
		return fmt.Errorf("查询24小时前表容量数据失败: %v", err)
	}

	logger.Info("查询到24小时前表容量数据", zap.Int("count", len(previousData)))

	// 创建映射以便快速查找24小时前的数据
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

	// 计算增量并保存到 pumpkin_table_growth
	successCount := 0
	failedCount := 0

	for _, current := range currentData {
		key := fmt.Sprintf("%s|%s|%s|%s|%s", current.DatasourceType, current.Host, current.Port, current.DatabaseName, current.TableName)

		var tableSizeIncr int64 = 0
		var tableRowsIncr int64 = 0

		if previous, exists := previousMap[key]; exists {
			tableSizeIncr = current.TableSize - previous.TableSize
			tableRowsIncr = current.TableRows - previous.TableRows
		} else {
			// 如果没有24小时前的数据，增量就是当前值
			tableSizeIncr = current.TableSize
			tableRowsIncr = current.TableRows
		}

		// 插入新的增长记录（每次都插入新记录，不更新）
		growthData := map[string]interface{}{
			"datasource_type": current.DatasourceType,
			"host":            current.Host,
			"port":            current.Port,
			"database_name":   current.DatabaseName,
			"table_name":      current.TableName,
			"table_size":      current.TableSize,
			"table_rows":      current.TableRows,
			"table_size_incr": tableSizeIncr,
			"table_rows_incr": tableRowsIncr,
			"gmt_created":     time.Now(),
			"gmt_updated":     time.Now(),
		}
		if err := db.Table("pumpkin_table_growth").Create(growthData).Error; err != nil {
			logger.Error("插入表容量增长记录失败", zap.Error(err), zap.String("table", current.TableName))
			failedCount++
			continue
		}
		successCount++
	}

	logger.Info("表容量增长计算完成", zap.Int("成功", successCount), zap.Int("失败", failedCount))
	return nil
}

// calculateDatabaseGrowth 计算数据库容量增长
func calculateDatabaseGrowth(nowStr string) error {
	var db = database.DB
	logger := log.Logger

	// 从 pumpkin_table_growth 聚合计算数据库容量增长
	// 查询今天的数据（每次执行任务都会插入新记录）
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	aggregateSQL := `
		SELECT 
			datasource_type,
			host,
			port,
			database_name,
			SUM(table_size) as database_size,
			SUM(table_rows) as database_rows,
			COUNT(DISTINCT table_name) as table_count,
			SUM(table_size_incr) as database_size_incr,
			COALESCE(SUM(table_rows_incr), 0) as database_rows_incr
		FROM pumpkin_table_growth
		WHERE gmt_created >= ?
		GROUP BY datasource_type, host, port, database_name
	`

	var aggregateData []struct {
		DatasourceType   string `gorm:"column:datasource_type"`
		Host             string `gorm:"column:host"`
		Port             string `gorm:"column:port"`
		DatabaseName     string `gorm:"column:database_name"`
		DatabaseSize     int64  `gorm:"column:database_size"`
		DatabaseRows     int64  `gorm:"column:database_rows"`
		TableCount       int64  `gorm:"column:table_count"`
		DatabaseSizeIncr int64  `gorm:"column:database_size_incr"`
		DatabaseRowsIncr int64  `gorm:"column:database_rows_incr"`
	}

	if err := db.Raw(aggregateSQL, todayStart).Scan(&aggregateData).Error; err != nil {
		return fmt.Errorf("聚合数据库容量增长数据失败: %v", err)
	}

	logger.Info("聚合到数据库容量增长数据", zap.Int("count", len(aggregateData)))

	successCount := 0
	failedCount := 0

	for _, item := range aggregateData {
		// 插入新的数据库容量增长记录（每次都插入新记录，不更新）
		growth := model.PumpkinDatabaseGrowth{
			DatasourceType:   item.DatasourceType,
			Host:             item.Host,
			Port:             item.Port,
			DatabaseName:     item.DatabaseName,
			DatabaseSize:     item.DatabaseSize,
			DatabaseRows:     item.DatabaseRows,
			TableCount:       item.TableCount,
			DatabaseSizeIncr: item.DatabaseSizeIncr,
			DatabaseRowsIncr: item.DatabaseRowsIncr,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := db.Create(&growth).Error; err != nil {
			logger.Error("插入数据库容量增长记录失败", zap.Error(err), zap.String("database", item.DatabaseName))
			failedCount++
			continue
		}
		successCount++
	}

	logger.Info("数据库容量增长计算完成", zap.Int("成功", successCount), zap.Int("失败", failedCount))
	return nil
}

// ExecutePumpkinGrowthTask 导出函数，用于手动执行任务
func ExecutePumpkinGrowthTask() {
	doPumpkinGrowthTask()
}
