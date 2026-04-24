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
	go aiApplyColumnCommentCrontabTask()
}

func aiApplyColumnCommentCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	logger := log.Logger
	var record model.TaskOption
	if err := db.Select("crontab").Where("task_key=?", "ai_apply_column_comment").Take(&record).Error; err != nil {
		logger.Warn("failed to load ai_apply_column_comment schedule, fallback to default cron",
			zap.Error(err), zap.String("default_cron", "*/30 * * * *"))
		record.Crontab = "*/30 * * * *"
	}
	if strings.TrimSpace(record.Crontab) == "" {
		record.Crontab = "*/30 * * * *"
		logger.Warn("ai_apply_column_comment crontab is empty, fallback to default cron",
			zap.String("default_cron", record.Crontab))
	}
	c := cron.New()
	if _, err := c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "ai_apply_column_comment").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='ai_apply_column_comment'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doAiApplyColumnCommentTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='ai_apply_column_comment'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	}); err != nil {
		logger.Error("failed to register ai_apply_column_comment cron task",
			zap.String("crontab", record.Crontab), zap.Error(err))
		return
	}
	logger.Info("registered ai_apply_column_comment cron task", zap.String("crontab", record.Crontab))
	c.Start()
}

func doAiApplyColumnCommentTask() {
	logger := log.Logger
	logger.Info("start ai_apply_column_comment task")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("ai_apply_column_comment")
	if err := taskLogger.Start(); err != nil {
		logger.Error("failed to create task log", zap.Error(err))
		return
	}

	// 查询待应用的AI注释字段
	var columns []model.MetaColumn
	result := database.DB.Where("ai_comment IS NOT NULL AND ai_comment != '' AND ai_fixed = 2 ").Find(&columns)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("failed to query columns pending ai comment apply: %v", result.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	if len(columns) == 0 {
		successMsg := "no columns pending ai comment apply"
		logger.Info(successMsg)
		taskLogger.Success(successMsg)
		return
	}

	logger.Info("found columns pending ai comment apply", zap.Int("count", len(columns)))
	taskLogger.UpdateResult(fmt.Sprintf("found %d columns pending ai comment apply", len(columns)))

	// 按数据源分组
	datasourceGroups := groupColumnsByDatasource(columns)

	successCount := 0
	failedCount := 0
	errorDetails := []string{}

	for datasourceKey, columnGroup := range datasourceGroups {
		logger.Info("processing datasource", zap.String("datasource", datasourceKey), zap.Int("column_count", len(columnGroup)))

		// 获取数据源连接信息
		datasource, err := getDatasourceInfoForColumn(columnGroup[0])
		if err != nil {
			errorMsg := fmt.Sprintf("failed to get datasource info %s: %v", datasourceKey, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount += len(columnGroup)
			continue
		}

		// 连接到目标数据库
		dbCon, err := connectToTargetDatabaseForColumn(datasource)
		if err != nil {
			errorMsg := fmt.Sprintf("failed to connect datasource %s: %v", datasourceKey, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount += len(columnGroup)
			continue
		}
		defer dbCon.Close()

		// 处理该数据源下的所有字段
		for i, column := range columnGroup {
			logger.Info("applying column comment", zap.Int("index", i+1), zap.Int("total", len(columnGroup)),
				zap.String("datasource_type", column.DatasourceType),
				zap.String("database_name", column.DatabaseName),
				zap.String("table_name", column.TableNameX),
				zap.String("column_name", column.ColumnName))

			err := applyColumnComment(dbCon, column)
			if err != nil {
				errorMsg := fmt.Sprintf("failed to apply comment for column %s.%s.%s: %v", column.DatabaseName, column.TableNameX, column.ColumnName, err)
				logger.Error(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				failedCount++
			} else {
				// 更新状态为已应用
				updateResult := database.DB.Model(&model.MetaColumn{}).Where("id = ?", column.Id).Update("ai_fixed", 3)
				if updateResult.Error != nil {
					errorMsg := fmt.Sprintf("failed to update ai_fixed for column %s.%s.%s: %v", column.DatabaseName, column.TableNameX, column.ColumnName, updateResult.Error)
					logger.Error(errorMsg)
					errorDetails = append(errorDetails, errorMsg)
					failedCount++
				} else {
					successCount++
					logger.Info("column comment applied", zap.String("database_name", column.DatabaseName),
						zap.String("table_name", column.TableNameX), zap.String("column_name", column.ColumnName),
						zap.String("comment", column.AiComment))
				}
			}

			// 更新进度
			progressMsg := fmt.Sprintf("processed %d/%d columns (success: %d, failed: %d)", successCount+failedCount, len(columns), successCount, failedCount)
			taskLogger.UpdateResult(progressMsg)
		}
	}

	// 记录最终结果
	finalResult := fmt.Sprintf("task finished - total: %d, success: %d, failed: %d", len(columns), successCount, failedCount)
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf(". first error: %s", errorDetails[0])
		if len(errorDetails) > 1 {
			finalResult += fmt.Sprintf(" and %d more errors", len(errorDetails))
		}
	}

	if failedCount == 0 {
		taskLogger.Success(finalResult)
	} else {
		taskLogger.Failed(finalResult)
	}

	logger.Info(finalResult)
}

// 按数据源分组字段
func groupColumnsByDatasource(columns []model.MetaColumn) map[string][]model.MetaColumn {
	groups := make(map[string][]model.MetaColumn)

	for _, column := range columns {
		key := fmt.Sprintf("%s_%s_%s", column.DatasourceType, column.Host, column.Port)
		groups[key] = append(groups[key], column)
	}

	return groups
}

// 获取数据源信息
func getDatasourceInfoForColumn(column model.MetaColumn) (*model.Datasource, error) {
	var datasource model.Datasource
	result := database.DB.Where("type = ? AND host = ? AND port = ? AND enable = 1",
		column.DatasourceType, column.Host, column.Port).First(&datasource)

	if result.Error != nil {
		return nil, fmt.Errorf("datasource not found or disabled: %v", result.Error)
	}

	return &datasource, nil
}

// 连接到目标数据库
func connectToTargetDatabaseForColumn(datasource *model.Datasource) (*sql.DB, error) {
	// 解密密码
	var origPass string
	if datasource.Pass != "" {
		var err error
		origPass, err = utils.AesPassDecode(datasource.Pass, setting.Setting.DbPassKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt datasource password: %v", err)
		}
	}

	var dbCon *sql.DB
	var err error

	// 根据数据库类型连接
	if datasource.Type == "MySQL" || datasource.Type == "TiDB" || datasource.Type == "Doris" ||
		datasource.Type == "MariaDB" || datasource.Type == "GreatSQL" || datasource.Type == "OceanBase" || datasource.Type == "PostgreSQL" {

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
		return nil, fmt.Errorf("unsupported datasource type: %s", datasource.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect target database: %v", err)
	}

	return dbCon, nil
}

// 从CREATE TABLE语句中提取指定字段的完整定义
func extractColumnDefinition(createTableSQL, columnName string) (string, error) {
	// 按行分割CREATE TABLE语句
	lines := strings.Split(createTableSQL, "\n")

	// 1. 找到包含目标字段的行
	var targetLine string
	for _, line := range lines {
		if strings.Contains(line, "`"+columnName+"`") {
			targetLine = line
			break
		}
	}

	if targetLine == "" {
		return "", fmt.Errorf("column definition not found for %s", columnName)
	}

	// 2. 找到字段名在行中的位置
	columnStart := strings.Index(targetLine, "`"+columnName+"`")
	if columnStart == -1 {
		return "", fmt.Errorf("column definition not found for %s", columnName)
	}

	// 3. 从字段名开始，跳过字段名和空白字符
	start := columnStart + len("`"+columnName+"`")
	for start < len(targetLine) && (targetLine[start] == ' ' || targetLine[start] == '\t') {
		start++
	}

	// 4. 检查该行是否包含COMMENT
	upperLine := strings.ToUpper(targetLine)
	commentIndex := strings.Index(upperLine, "COMMENT")

	var end int
	if commentIndex != -1 {
		// 如果包含COMMENT，提取字段名到COMMENT之间的部分
		end = commentIndex
		// 向前查找字段定义的开始位置，去除尾部空白
		for end > start && (targetLine[end-1] == ' ' || targetLine[end-1] == '\t') {
			end--
		}
	} else {
		// 如果不包含COMMENT，提取字段名到逗号或行尾的内容
		end = start
		for end < len(targetLine) {
			if targetLine[end] == ',' {
				break
			}
			end++
		}
		// 去除尾部空白
		for end > start && (targetLine[end-1] == ' ' || targetLine[end-1] == '\t') {
			end--
		}
	}

	// 5. 提取字段定义
	if start >= end {
		return "", fmt.Errorf("failed to parse column definition for %s", columnName)
	}

	columnDef := strings.TrimSpace(targetLine[start:end])
	if columnDef == "" {
		return "", fmt.Errorf("failed to parse column definition for %s", columnName)
	}

	return columnDef, nil
}

// 应用字段注释
func applyColumnComment(dbCon *sql.DB, column model.MetaColumn) error {
	// 转义注释内容，防止SQL注入
	escapedComment := strings.ReplaceAll(column.AiComment, "'", "''")

	// 根据数据库类型生成不同的ALTER语句
	if column.DatasourceType == "MySQL" || column.DatasourceType == "TiDB" || column.DatasourceType == "Doris" ||
		column.DatasourceType == "MariaDB" || column.DatasourceType == "GreatSQL" || column.DatasourceType == "OceanBase" || column.DatasourceType == "PostgreSQL" {

		// 1. SHOW CREATE TABLE 获取结果
		showCreateSQL := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", column.DatabaseName, column.TableNameX)

		var tableName, createTableSQL string
		err := dbCon.QueryRow(showCreateSQL).Scan(&tableName, &createTableSQL)
		if err != nil {
			return fmt.Errorf("failed to fetch table ddl: %v", err)
		}

		// 2. 找到该字段的整行数据并提取字段定义
		columnDefinition, err := extractColumnDefinition(createTableSQL, column.ColumnName)
		if err != nil {
			return fmt.Errorf("failed to parse column definition: %v", err)
		}

		// 3. 生成完整的ALTER语句
		alterSQL := fmt.Sprintf("ALTER TABLE `%s`.`%s` MODIFY COLUMN `%s` %s COMMENT '%s'",
			column.DatabaseName, column.TableNameX, column.ColumnName, columnDefinition, escapedComment)

		fmt.Println(alterSQL)
		log.Logger.Info(alterSQL)

		// 4. 执行ALTER语句
		_, err = dbCon.Exec(alterSQL)
		if err != nil {
			return fmt.Errorf("failed to execute ALTER SQL: %s, error: %v", alterSQL, err)
		}

	} else if column.DatasourceType == "ClickHouse" {
		// ClickHouse 修改字段注释
		alterSQL := fmt.Sprintf("ALTER TABLE `%s`.`%s` COMMENT COLUMN `%s` '%s'",
			column.DatabaseName, column.TableNameX, column.ColumnName, escapedComment)

		_, err := dbCon.Exec(alterSQL)
		if err != nil {
			return fmt.Errorf("failed to execute ALTER SQL: %s, error: %v", alterSQL, err)
		}
	} else {
		return fmt.Errorf("unsupported datasource type: %s", column.DatasourceType)
	}

	return nil
}

// ExecuteAiApplyColumnCommentTask 手动触发，与定时任务逻辑一致（计划任务平台「手工运行」）
func ExecuteAiApplyColumnCommentTask() {
	doAiApplyColumnCommentTask()
}
