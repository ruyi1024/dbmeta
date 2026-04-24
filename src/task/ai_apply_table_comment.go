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
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/setting"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"github.com/ruyi1024/dbmeta/src/utils"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go aiApplyTableCommentCrontabTask()
}

func aiApplyTableCommentCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "ai_apply_table_comment").Take(&record)
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "ai_apply_table_comment").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='ai_apply_table_comment'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doAiApplyTableCommentTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='ai_apply_table_comment'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func doAiApplyTableCommentTask() {
	logger := log.Logger
	logger.Info("开始执行AI应用表注释任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("ai_apply_table_comment")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	// 查询待应用的AI注释表
	var tables []model.MetaTable
	result := database.DB.Where("ai_comment IS NOT NULL AND ai_comment != '' AND ai_fixed = 2 AND is_deleted = 0").Find(&tables)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("查询待应用AI注释的表失败: %v", result.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	if len(tables) == 0 {
		successMsg := "没有需要应用AI注释的表"
		logger.Info(successMsg)
		taskLogger.Success(successMsg)
		return
	}

	logger.Info("找到需要应用AI注释的表", zap.Int("count", len(tables)))
	taskLogger.UpdateResult(fmt.Sprintf("找到 %d 个需要应用AI注释的表", len(tables)))

	// 按数据源分组
	datasourceGroups := groupTablesByDatasource(tables)

	successCount := 0
	failedCount := 0
	errorDetails := []string{}

	for datasourceKey, tableGroup := range datasourceGroups {
		logger.Info("处理数据源", zap.String("datasource", datasourceKey), zap.Int("table_count", len(tableGroup)))

		// 获取数据源连接信息
		datasource, err := getDatasourceInfo(tableGroup[0])
		if err != nil {
			errorMsg := fmt.Sprintf("获取数据源 %s 信息失败: %v", datasourceKey, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount += len(tableGroup)
			continue
		}

		// 连接到目标数据库
		dbCon, err := connectToTargetDatabase(datasource)
		if err != nil {
			errorMsg := fmt.Sprintf("连接数据源 %s 失败: %v", datasourceKey, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount += len(tableGroup)
			continue
		}
		defer dbCon.Close()

		// 处理该数据源下的所有表
		for i, table := range tableGroup {
			logger.Info("应用表注释", zap.Int("index", i+1), zap.Int("total", len(tableGroup)),
				zap.String("datasource_type", table.DatasourceType),
				zap.String("database_name", table.DatabaseName),
				zap.String("table_name", table.TableNameX))

			err := applyTableComment(dbCon, table)
			if err != nil {
				errorMsg := fmt.Sprintf("应用表 %s.%s 注释失败: %v", table.DatabaseName, table.TableNameX, err)
				logger.Error(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				failedCount++
			} else {
				// 更新状态为已应用
				updateResult := database.DB.Model(&model.MetaTable{}).Where("id = ?", table.Id).Update("ai_fixed", 3)
				if updateResult.Error != nil {
					errorMsg := fmt.Sprintf("更新表 %s.%s 状态失败: %v", table.DatabaseName, table.TableNameX, updateResult.Error)
					logger.Error(errorMsg)
					errorDetails = append(errorDetails, errorMsg)
					failedCount++
				} else {
					successCount++
					logger.Info("成功应用表注释", zap.String("database_name", table.DatabaseName),
						zap.String("table_name", table.TableNameX), zap.String("comment", table.AiComment))
				}
			}

			// 更新进度
			progressMsg := fmt.Sprintf("已处理 %d/%d 个表 (成功: %d, 失败: %d)", successCount+failedCount, len(tables), successCount, failedCount)
			taskLogger.UpdateResult(progressMsg)
		}
	}

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 总计: %d, 成功: %d, 失败: %d", len(tables), successCount, failedCount)
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0]) // 只记录第一个错误详情
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

// 按数据源分组
func groupTablesByDatasource(tables []model.MetaTable) map[string][]model.MetaTable {
	groups := make(map[string][]model.MetaTable)

	for _, table := range tables {
		key := fmt.Sprintf("%s_%s_%s", table.DatasourceType, table.Host, table.Port)
		groups[key] = append(groups[key], table)
	}

	return groups
}

// 获取数据源信息
func getDatasourceInfo(table model.MetaTable) (*model.Datasource, error) {
	var datasource model.Datasource
	result := database.DB.Where("type = ? AND host = ? AND port = ? AND enable = 1",
		table.DatasourceType, table.Host, table.Port).First(&datasource)

	if result.Error != nil {
		return nil, fmt.Errorf("数据源不存在或已禁用: %v", result.Error)
	}

	return &datasource, nil
}

// 连接到目标数据库
func connectToTargetDatabase(datasource *model.Datasource) (*sql.DB, error) {
	// 解密密码
	var origPass string
	if datasource.Pass != "" {
		var err error
		origPass, err = utils.AesPassDecode(datasource.Pass, setting.Setting.DbPassKey)
		if err != nil {
			return nil, fmt.Errorf("密码解密失败: %v", err)
		}
	}

	var dbCon *sql.DB
	var err error

	// 根据数据库类型连接
	if datasource.Type == "MySQL" || datasource.Type == "TiDB" || datasource.Type == "Doris" || datasource.Type == "MariaDB" || datasource.Type == "GreatSQL" || datasource.Type == "OceanBase" || datasource.Type == "PostgreSQL" {

		dbCon, err = database.Connect(
			database.WithDriver("mysql"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(origPass),
			database.WithDatabase("information_schema"))

	} else if datasource.Type == "ClickHouse" {
		dbCon, err = database.Connect(
			database.WithDriver("clickhouse"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(origPass),
			database.WithDatabase("system"))
	} else {
		return nil, fmt.Errorf("不支持的数据库类型: %s", datasource.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	return dbCon, nil
}

// 应用表注释
func applyTableComment(dbCon *sql.DB, table model.MetaTable) error {
	var alterSQL string

	// 转义注释内容，防止SQL注入
	escapedComment := strings.ReplaceAll(table.AiComment, "'", "''")

	// 根据数据库类型生成不同的ALTER语句
	if table.DatasourceType == "MySQL" || table.DatasourceType == "TiDB" || table.DatasourceType == "Doris" || table.DatasourceType == "MariaDB" || table.DatasourceType == "GreatSQL" || table.DatasourceType == "OceanBase" || table.DatasourceType == "PostgreSQL" {

		alterSQL = fmt.Sprintf("ALTER TABLE `%s`.`%s` COMMENT = '%s'",
			table.DatabaseName, table.TableNameX, escapedComment)

	} else if table.DatasourceType == "ClickHouse" {
		alterSQL = fmt.Sprintf("ALTER TABLE `%s`.`%s` MODIFY COMMENT '%s'",
			table.DatabaseName, table.TableNameX, escapedComment)
	} else {
		return fmt.Errorf("不支持的数据库类型: %s", table.DatasourceType)
	}

	// 执行ALTER语句
	_, err := dbCon.Exec(alterSQL)
	if err != nil {
		return fmt.Errorf("执行ALTER语句失败: %s, 错误: %v", alterSQL, err)
	}

	return nil
}

// ExecuteAiApplyTableCommentTask 手动触发，与定时任务逻辑一致（计划任务平台「手工运行」）
func ExecuteAiApplyTableCommentTask() {
	doAiApplyTableCommentTask()
}
