/*
Copyright 2026 The Dbmeta Team Group, website: https://www.dbmeta.com
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
You may not use this file except in compliance with the License.
*/

package task

import (
	"context"
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

// pumpkinLifecycleRowTimeBoundsPerTableTimeout 单表 MIN/MAX 与 information_schema 查询的总超时，
// 避免超大表无索引全表扫描导致整数据源长时间阻塞、任务无法结束。
const pumpkinLifecycleRowTimeBoundsPerTableTimeout = 40 * time.Second

// pumpkinLifecycleTableProgressEvery 每处理多少张表更新一次 task_log.result，便于前端看到进度。
const pumpkinLifecycleTableProgressEvery = 80

func init() {
	go tableLifecycleCrontabTask()
}

func tableLifecycleCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	const defaultCrontab = "15 * * * *"
	if err := db.Where("task_key=?", "gather_table_lifecycle").Take(&record).Error; err != nil {
		log.Logger.Warn("gather_table_lifecycle 任务配置不存在，使用默认 crontab", zap.Error(err))
		record.Crontab = defaultCrontab
	} else if strings.TrimSpace(record.Crontab) == "" {
		record.Crontab = defaultCrontab
	}
	c := cron.New()
	entryID, err := c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "gather_table_lifecycle").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_table_lifecycle'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doTableLifecycleTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='gather_table_lifecycle'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	if err != nil {
		log.Logger.Error("gather_table_lifecycle cron 注册失败", zap.Error(err), zap.String("crontab", record.Crontab))
		return
	}
	_ = entryID
	c.Start()
}

func doTableLifecycleTask() {
	logger := log.Logger
	logger.Info("开始执行表生命周期采集任务")

	taskLogger := NewTaskLogger("gather_table_lifecycle")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	if err := database.EnsurePumpkinTableLifecycleSchema(database.DB); err != nil {
		logger.Error("pumpkin_table_lifecycle 表就绪失败", zap.Error(err))
		taskLogger.Failed(err.Error())
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

	logger.Info("表生命周期：找到数据源", zap.Int("count", len(dataList)))
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

		logger.Info("表生命周期处理数据源", zap.Int("index", i+1), zap.Int("total", len(dataList)),
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

		err := doTableLifecycleCollectorTask(datasourceType, host, port, user, origPass, dbid, taskLogger)
		if err != nil {
			errorMsg := fmt.Sprintf("数据源 %s:%s 表生命周期采集失败: %v", host, port, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			instanceStatuses = append(instanceStatuses, fmt.Sprintf("[失败] %s: %s", formatDatasourceInstance(datasource), truncateText(err.Error(), 100)))
			failedCount++
		} else {
			instanceStatuses = append(instanceStatuses, fmt.Sprintf("[成功] %s", formatDatasourceInstance(datasource)))
			successCount++
		}

		progressMsg := fmt.Sprintf("已处理 %d/%d 个数据源 (成功: %d, 失败: %d)", i+1, len(dataList), successCount, failedCount)
		taskLogger.UpdateResult(progressMsg)
	}

	finalResult := fmt.Sprintf("表生命周期任务完成 - 数据源总计: %d, 成功: %d, 失败: %d",
		len(dataList), successCount, failedCount)
	finalResult += fmt.Sprintf("。实例状态: %s", summarizeInstanceStatuses(instanceStatuses, 20, 1300))
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0])
		if len(errorDetails) > 1 {
			finalResult += fmt.Sprintf(" 等 %d 个错误", len(errorDetails))
		}
	}

	taskLogger.Success(finalResult)
	logger.Info(finalResult)
}

func doTableLifecycleCollectorTask(datasourceType, host, port, user, origPass, dbid string, taskLogger *TaskLogger) error {
	if strings.EqualFold(strings.TrimSpace(datasourceType), "MongoDB") {
		return doMongoTableLifecycleCollectorTask(host, port, user, origPass, dbid, taskLogger)
	}

	dbCon := getPumpkinDbCon(datasourceType, host, port, user, origPass, dbid)
	if dbCon == nil {
		return fmt.Errorf("无法连接到数据库 %s:%s", host, port)
	}
	defer dbCon.Close()

	tableSizeList, err := queryPumpkinRemoteTableRowsFromConn(dbCon, datasourceType)
	if err != nil {
		return err
	}
	log.Logger.Info("表生命周期远程查询完成",
		zap.String("type", datasourceType),
		zap.String("host", host),
		zap.String("port", port),
		zap.Int("rowCount", len(tableSizeList)))

	if taskLogger != nil && len(tableSizeList) > 0 {
		_ = taskLogger.UpdateResult(fmt.Sprintf("表生命周期 %s:%s %s：共 %d 张表，正在写入…", datasourceType, host, port, len(tableSizeList)))
	}

	for i, item := range tableSizeList {
		record, meta, ok := pumpkinRemoteRowToPumpkinRecord(datasourceType, host, port, item)
		if !ok {
			continue
		}

		ctxBounds, cancelBounds := context.WithTimeout(context.Background(), pumpkinLifecycleRowTimeBoundsPerTableTimeout)
		rowMinAt, rowMaxAt, berr := queryPumpkinTableRowTimeBoundsFromConn(ctxBounds, dbCon, datasourceType, record.DatabaseName, record.TableNameField)
		cancelBounds()
		if berr != nil {
			if errors.Is(berr, context.DeadlineExceeded) {
				log.Logger.Warn("表行时间边界查询超时，已跳过该表 MIN/MAX（仍写入生命周期其它字段）",
					zap.String("host", host), zap.String("port", port),
					zap.String("database", record.DatabaseName), zap.String("table", record.TableNameField),
					zap.Duration("timeout", pumpkinLifecycleRowTimeBoundsPerTableTimeout))
			} else {
				log.Logger.Debug("表行时间边界查询失败",
					zap.String("host", host), zap.String("db", record.DatabaseName), zap.String("table", record.TableNameField), zap.Error(berr))
			}
		}

		var oldRows, oldDataSize, oldIndexSize int64
		var existing model.PumpkinTableSize
		takeErr := database.DB.Where("datasource_type = ? AND host = ? AND port = ? AND database_name = ? AND table_name = ?",
			record.DatasourceType, record.Host, record.Port, record.DatabaseName, record.TableNameField).Take(&existing).Error
		if takeErr == nil {
			oldRows = existing.TableRows
			oldDataSize = existing.DataSize
			oldIndexSize = existing.IndexSize
		} else if !errors.Is(takeErr, gorm.ErrRecordNotFound) {
			return takeErr
		}

		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			return upsertPumpkinTableLifecycle(tx, &record, meta, oldRows, oldDataSize, oldIndexSize, rowMinAt, rowMaxAt)
		}); err != nil {
			log.Logger.Error(fmt.Sprintf("表生命周期写入失败 %s:%s %s", host, port, err.Error()))
			return fmt.Errorf("表生命周期写入失败: %w", err)
		}
		time.Sleep(1 * time.Millisecond)

		if taskLogger != nil && len(tableSizeList) > 0 {
			n := i + 1
			if n%pumpkinLifecycleTableProgressEvery == 0 || n == len(tableSizeList) {
				_ = taskLogger.UpdateResult(fmt.Sprintf("表生命周期 %s:%s %s：已处理 %d/%d 张表", datasourceType, host, port, n, len(tableSizeList)))
			}
		}
	}

	return nil
}

func doMongoTableLifecycleCollectorTask(host, port, user, origPass, dbid string, taskLogger *TaskLogger) error {
	ctx := context.Background()
	client, err := mongodb.Connect(host, port, user, origPass, dbid)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}
	defer client.Disconnect(ctx)

	if taskLogger != nil {
		_ = taskLogger.UpdateResult(fmt.Sprintf("MongoDB %s:%s：正在枚举库与集合…", host, port))
	}

	databaseNames, err := mongodb.ListDatabase(client)
	if err != nil {
		return fmt.Errorf("查询MongoDB数据库列表失败: %v", err)
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
				log.Logger.Warn("MongoDB 表生命周期：集合统计失败，已跳过",
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

			var oldRows, oldDataSize, oldIndexSize int64
			var existing model.PumpkinTableSize
			takeErr := database.DB.Where("datasource_type = ? AND host = ? AND port = ? AND database_name = ? AND table_name = ?",
				record.DatasourceType, record.Host, record.Port, record.DatabaseName, record.TableNameField).Take(&existing).Error
			if takeErr == nil {
				oldRows = existing.TableRows
				oldDataSize = existing.DataSize
				oldIndexSize = existing.IndexSize
			} else if !errors.Is(takeErr, gorm.ErrRecordNotFound) {
				return takeErr
			}

			if err := database.DB.Transaction(func(tx *gorm.DB) error {
				return upsertPumpkinTableLifecycle(tx, &record, nil, oldRows, oldDataSize, oldIndexSize, nil, nil)
			}); err != nil {
				return fmt.Errorf("MongoDB 表生命周期写入失败: %w", err)
			}
			insertedCount++
			time.Sleep(1 * time.Millisecond)
		}
	}

	if insertedCount == 0 {
		return fmt.Errorf("未采集到MongoDB表生命周期数据，请检查权限或集合数据")
	}
	return nil
}

// upsertPumpkinTableLifecycle 按数据源+库+表维度维护生命周期。
// rowDataMinAt / rowDataMaxAt 来自业务表时间列的 MIN/MAX（若有）；last_write_at 更新时优先采用 rowDataMaxAt，
// 不与「采集时刻」因容量增长混用取大，以免偏离真实数据最后写入时间。下线中/已下线状态不自动改回使用中。
func upsertPumpkinTableLifecycle(tx *gorm.DB, record *model.PumpkinTableSize, meta *pumpkinTableGatherMeta, oldRows, oldDataSize, oldIndexSize int64, rowDataMinAt, rowDataMaxAt *time.Time) error {
	now := time.Now()
	totalSize := record.DataSize + record.IndexSize
	oldTotal := oldDataSize + oldIndexSize
	sizeGrew := record.TableRows > oldRows || totalSize > oldTotal

	var lc model.PumpkinTableLifecycle
	err := tx.Where("datasource_type = ? AND host = ? AND port = ? AND database_name = ? AND table_name = ?",
		record.DatasourceType, record.Host, record.Port, record.DatabaseName, record.TableNameField).Take(&lc).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		nlc := model.PumpkinTableLifecycle{
			DatasourceType:  record.DatasourceType,
			Host:            record.Host,
			Port:            record.Port,
			DatabaseName:    record.DatabaseName,
			TableNameField:  record.TableNameField,
			LifecycleStatus: model.TableLifecycleUnused,
		}
		if meta != nil && meta.TableCreateTime != nil && !meta.TableCreateTime.IsZero() {
			t := *meta.TableCreateTime
			nlc.TableCreatedAt = &t
		}
		hasData := record.TableRows > 0 || totalSize > 0
		if rowDataMinAt != nil && !rowDataMinAt.IsZero() {
			t := *rowDataMinAt
			nlc.DataWriteStartAt = &t
		} else if hasData {
			nlc.DataWriteStartAt = &now
		}
		if rowDataMaxAt != nil && !rowDataMaxAt.IsZero() {
			t := *rowDataMaxAt
			nlc.LastWriteAt = &t
		} else if hasData {
			nlc.LastWriteAt = &now
		}
		if meta != nil && meta.TableUpdateTime != nil && !meta.TableUpdateTime.IsZero() {
			nlc.LastWriteAt = pumpkinMaxTimePtr(nlc.LastWriteAt, meta.TableUpdateTime)
		}
		if hasData {
			nlc.LifecycleStatus = model.TableLifecycleInUse
		}
		return tx.Create(&nlc).Error
	}
	if err != nil {
		return err
	}

	manualLock := lc.LifecycleStatus == model.TableLifecycleDecommissioning || lc.LifecycleStatus == model.TableLifecycleDecommissioned

	updates := map[string]interface{}{
		"gmt_updated": now,
	}
	if lc.TableCreatedAt == nil && meta != nil && meta.TableCreateTime != nil && !meta.TableCreateTime.IsZero() {
		updates["table_created_at"] = *meta.TableCreateTime
	}
	if rowDataMinAt != nil && !rowDataMinAt.IsZero() {
		if lc.DataWriteStartAt == nil || rowDataMinAt.Before(*lc.DataWriteStartAt) {
			updates["data_write_start_at"] = *rowDataMinAt
		}
	} else if lc.DataWriteStartAt == nil && (record.TableRows > 0 || totalSize > 0) {
		updates["data_write_start_at"] = now
	}

	// last_write_at 以业务表时间列 MAX（及引擎表更新时间）为准；不要用「采集时刻 now」与 MAX 取大，
	// 否则只要行数/容量增长就会把最后写入顶成当前任务运行时间，偏离真实数据最后写入时间。
	var candLast *time.Time
	if rowDataMaxAt != nil && !rowDataMaxAt.IsZero() {
		t := *rowDataMaxAt
		candLast = &t
	}
	if meta != nil && meta.TableUpdateTime != nil && !meta.TableUpdateTime.IsZero() {
		candLast = pumpkinMaxTimePtr(candLast, meta.TableUpdateTime)
	}
	if lc.LastWriteAt != nil && !lc.LastWriteAt.IsZero() {
		t := *lc.LastWriteAt
		candLast = pumpkinMaxTimePtr(candLast, &t)
	}
	if candLast == nil || candLast.IsZero() {
		if sizeGrew {
			candLast = &now
		}
	}
	if candLast != nil && !candLast.IsZero() {
		if lc.LastWriteAt == nil || candLast.After(*lc.LastWriteAt) {
			updates["last_write_at"] = *candLast
		}
	}

	if !manualLock {
		if record.TableRows == 0 && totalSize == 0 {
			updates["lifecycle_status"] = model.TableLifecycleUnused
		} else {
			updates["lifecycle_status"] = model.TableLifecycleInUse
		}
	}
	return tx.Model(&lc).Updates(updates).Error
}

// ExecuteTableLifecycleTask 手动执行表生命周期采集（与定时任务 gather_table_lifecycle 相同逻辑）
func ExecuteTableLifecycleTask() {
	doTableLifecycleTask()
}
