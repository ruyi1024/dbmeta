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
	"database/sql"
	"dbmcloud/log"
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ExecuteDataQualityTask 异步执行数据质量评估任务
func ExecuteDataQualityTask(taskId int64) {
	go executeDataQualityTaskAsync(taskId)
}

// executeDataQualityTaskAsync 异步执行数据质量评估任务
func executeDataQualityTaskAsync(taskId int64) {
	logger := log.Logger
	logger.Info("开始执行数据质量评估任务", zap.Int64("task_id", taskId))

	// 查询任务
	var task model.DataQualityTask
	if err := database.DB.First(&task, taskId).Error; err != nil {
		logger.Error("查询任务失败", zap.Int64("task_id", taskId), zap.Error(err))
		return
	}

	// 更新任务状态为运行中
	now := time.Now()
	task.Status = "running"
	task.StartTime = &now
	if err := database.DB.Save(&task).Error; err != nil {
		logger.Error("更新任务状态失败", zap.Error(err))
		return
	}

	defer func() {
		// 更新任务结束时间和执行时长
		endTime := time.Now()
		task.EndTime = &endTime
		if task.StartTime != nil {
			task.Duration = int(endTime.Sub(*task.StartTime).Seconds())
		}
		database.DB.Save(&task)
	}()

	// 获取数据库信息
	var dbInfo model.MetaDatabase
	if task.DatasourceId != nil {
		if err := database.DB.Where("id = ? AND is_deleted = 0", *task.DatasourceId).First(&dbInfo).Error; err != nil {
			errorMsg := fmt.Sprintf("查询数据库信息失败: %v", err)
			logger.Error(errorMsg)
			task.Status = "failed"
			task.ErrorMessage = errorMsg
			return
		}
	} else if task.DatabaseName != "" {
		if err := database.DB.Where("database_name = ? AND is_deleted = 0", task.DatabaseName).First(&dbInfo).Error; err != nil {
			errorMsg := fmt.Sprintf("查询数据库信息失败: %v", err)
			logger.Error(errorMsg)
			task.Status = "failed"
			task.ErrorMessage = errorMsg
			return
		}
	} else {
		errorMsg := "任务缺少数据库信息"
		logger.Error(errorMsg)
		task.Status = "failed"
		task.ErrorMessage = errorMsg
		return
	}

	// 获取数据源信息
	var datasource model.Datasource
	if err := database.DB.Where("host = ? AND port = ? AND enable = 1", dbInfo.Host, dbInfo.Port).First(&datasource).Error; err != nil {
		errorMsg := fmt.Sprintf("查询数据源失败: %v", err)
		logger.Error(errorMsg)
		task.Status = "failed"
		task.ErrorMessage = errorMsg
		return
	}

	// 执行数据质量评估
	assessment, err := performDataQualityAssessment(task, dbInfo, datasource)
	if err != nil {
		errorMsg := fmt.Sprintf("执行数据质量评估失败: %v", err)
		logger.Error(errorMsg)
		task.Status = "failed"
		task.ErrorMessage = errorMsg
		return
	}

	// 保存评估结果
	if err := saveAssessmentResult(assessment, task); err != nil {
		errorMsg := fmt.Sprintf("保存评估结果失败: %v", err)
		logger.Error(errorMsg)
		task.Status = "failed"
		task.ErrorMessage = errorMsg
		return
	}

	// 更新任务状态为成功
	task.Status = "success"
	task.ResultSummary = fmt.Sprintf("评估完成，共发现 %d 个质量问题", assessment.TotalIssues)
	logger.Info("数据质量评估任务执行成功", zap.Int64("task_id", taskId), zap.Int("total_issues", assessment.TotalIssues))
}

// performDataQualityAssessment 执行数据质量评估
func performDataQualityAssessment(task model.DataQualityTask, dbInfo model.MetaDatabase, datasource model.Datasource) (*model.DataQualityAssessment, error) {
	logger := log.Logger

	// 解密密码
	var origPass string
	if datasource.Pass != "" {
		var err error
		origPass, err = utils.AesPassDecode(datasource.Pass, setting.Setting.DbPassKey)
		if err != nil {
			return nil, fmt.Errorf("密码解密失败: %v", err)
		}
	}

	// 连接数据库（使用analysis.go中的connectToDatabase函数）
	dbConn, err := connectToDatabase(datasource, origPass)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	// 如果提供了数据库名，执行 USE database 语句（对于MySQL等）
	if dbInfo.DatabaseName != "" && (strings.ToUpper(datasource.Type) == "MYSQL" || strings.ToUpper(datasource.Type) == "MARIADB" || strings.ToUpper(datasource.Type) == "GREATSQL" || strings.ToUpper(datasource.Type) == "TIDB" || strings.ToUpper(datasource.Type) == "DORIS" || strings.ToUpper(datasource.Type) == "OCEANBASE") {
		useSQL := fmt.Sprintf("USE `%s`", dbInfo.DatabaseName)
		_, err = dbConn.Exec(useSQL)
		if err != nil {
			return nil, fmt.Errorf("选择数据库失败: %v", err)
		}
	}

	// 获取表列表
	tables, err := getTableList(dbConn, dbInfo.DatabaseName, datasource.Type, task.TableFilter)
	if err != nil {
		return nil, fmt.Errorf("获取表列表失败: %v", err)
	}

	logger.Info("开始评估数据质量", zap.Int("table_count", len(tables)))

	// 创建评估记录
	assessment := &model.DataQualityAssessment{
		AssessmentTime: time.Now(),
		DatasourceId:   &dbInfo.Id,
		DatabaseName:   dbInfo.DatabaseName,
		TotalTables:    len(tables),
		Status:         1,
	}

	var totalColumns int
	var totalIssues int
	var issues []model.DataQualityIssue

	// 获取启用的质量规则
	var rules []model.DataQualityRule
	if err := database.DB.Where("enabled = 1").Find(&rules).Error; err != nil {
		logger.Warn("获取质量规则失败，使用默认规则", zap.Error(err))
	}

	// 遍历每个表进行评估
	for _, tableName := range tables {
		// 获取表的列信息
		columns, err := getTableColumns(dbConn, dbInfo.DatabaseName, tableName, datasource.Type)
		if err != nil {
			logger.Warn("获取表列信息失败", zap.String("table", tableName), zap.Error(err))
			continue
		}

		totalColumns += len(columns)

		// 对每个列进行质量检查
		for _, column := range columns {
			// 执行质量规则检查
			columnIssues := checkColumnQuality(dbConn, dbInfo.DatabaseName, tableName, column, datasource.Type, rules)
			issues = append(issues, columnIssues...)
			totalIssues += len(columnIssues)
		}
	}

	assessment.TotalColumns = totalColumns
	assessment.TotalIssues = totalIssues

	// 计算质量指标
	assessment.FieldCompleteness = calculateCompleteness(issues, totalColumns)
	assessment.FieldAccuracy = calculateAccuracy(issues, totalColumns)
	assessment.TableCompleteness = calculateTableCompleteness(tables, issues)
	assessment.DataConsistency = calculateConsistency(issues, totalColumns)
	assessment.DataUniqueness = calculateUniqueness(issues, totalColumns)
	assessment.DataTimeliness = calculateTimeliness(issues, totalColumns)

	// 计算总体评分
	assessment.OverallScore = (assessment.FieldCompleteness + assessment.FieldAccuracy +
		assessment.TableCompleteness + assessment.DataConsistency +
		assessment.DataUniqueness + assessment.DataTimeliness) / 6.0

	// 设置质量等级
	assessment.OverallLevel = getQualityLevel(assessment.OverallScore)

	// 将问题列表保存到assessment的临时字段（用于后续保存）
	assessment.Issues = issues

	return assessment, nil
}

// getTableList 获取表列表
func getTableList(dbConn *sql.DB, databaseName, dbType string, tableFilter string) ([]string, error) {
	var tables []string
	var query string

	// 解析表过滤条件
	var includeTables []string
	if tableFilter != "" {
		var filter map[string]interface{}
		if err := json.Unmarshal([]byte(tableFilter), &filter); err == nil {
			if includes, ok := filter["include"].([]interface{}); ok {
				for _, t := range includes {
					if tableName, ok := t.(string); ok {
						includeTables = append(includeTables, tableName)
					}
				}
			}
		}
	}

	switch strings.ToUpper(dbType) {
	case "MYSQL", "MARIADB", "GREATSQL", "TIDB", "DORIS", "OCEANBASE":
		if len(includeTables) > 0 {
			placeholders := strings.Repeat("?,", len(includeTables))
			placeholders = placeholders[:len(placeholders)-1]
			query = fmt.Sprintf("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME IN (%s) AND TABLE_TYPE = 'BASE TABLE'", placeholders)
			args := []interface{}{databaseName}
			for _, t := range includeTables {
				args = append(args, t)
			}
			rows, err := dbConn.Query(query, args...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					continue
				}
				tables = append(tables, tableName)
			}
		} else {
			query = "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'"
			rows, err := dbConn.Query(query, databaseName)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					continue
				}
				tables = append(tables, tableName)
			}
		}
	case "POSTGRESQL":
		if len(includeTables) > 0 {
			// PostgreSQL使用$1, $2等占位符
			placeholders := ""
			for i := 0; i < len(includeTables); i++ {
				if i > 0 {
					placeholders += ", "
				}
				placeholders += fmt.Sprintf("$%d", i+2)
			}
			query = fmt.Sprintf("SELECT tablename FROM pg_tables WHERE schemaname = $1 AND tablename IN (%s)", placeholders)
			args := []interface{}{databaseName}
			for _, t := range includeTables {
				args = append(args, t)
			}
			rows, err := dbConn.Query(query, args...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					continue
				}
				tables = append(tables, tableName)
			}
		} else {
			query = "SELECT tablename FROM pg_tables WHERE schemaname = $1"
			rows, err := dbConn.Query(query, databaseName)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					continue
				}
				tables = append(tables, tableName)
			}
		}
	default:
		return nil, fmt.Errorf("不支持的数据源类型: %s", dbType)
	}

	return tables, nil
}

// getTableColumns 获取表的列信息
func getTableColumns(dbConn *sql.DB, databaseName, tableName, dbType string) ([]string, error) {
	var columns []string
	var query string

	switch strings.ToUpper(dbType) {
	case "MYSQL", "MARIADB", "GREATSQL", "TIDB", "DORIS", "OCEANBASE":
		query = "SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		rows, err := dbConn.Query(query, databaseName, tableName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var columnName string
			if err := rows.Scan(&columnName); err != nil {
				continue
			}
			columns = append(columns, columnName)
		}
	case "POSTGRESQL":
		query = "SELECT column_name FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2 ORDER BY ordinal_position"
		rows, err := dbConn.Query(query, databaseName, tableName)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var columnName string
			if err := rows.Scan(&columnName); err != nil {
				continue
			}
			columns = append(columns, columnName)
		}
	default:
		return nil, fmt.Errorf("不支持的数据源类型: %s", dbType)
	}

	return columns, nil
}

// checkColumnQuality 检查列的数据质量
func checkColumnQuality(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, rules []model.DataQualityRule) []model.DataQualityIssue {
	var issues []model.DataQualityIssue

	// 如果没有规则，使用默认检查
	if len(rules) == 0 {
		// 默认检查：空值率
		nullRate, nullCount, err := checkNullRate(dbConn, databaseName, tableName, columnName, dbType)
		if err == nil && nullRate > 0.2 {
			issues = append(issues, model.DataQualityIssue{
				DatabaseName: databaseName,
				TableNameX:   tableName,
				ColumnName:   columnName,
				IssueType:    "完整性",
				IssueLevel:   "high",
				IssueDesc:    fmt.Sprintf("字段空值率过高: %.2f%%, 空值数量: %d", nullRate*100, nullCount),
				IssueCount:   int(nullCount),
				IssueRate:    nullRate * 100,
				CheckTime:    time.Now(),
				Status:       1,
			})
		}
		return issues
	}

	// 根据规则进行检查
	for _, rule := range rules {
		var ruleConfig map[string]interface{}
		if err := json.Unmarshal([]byte(rule.RuleConfig), &ruleConfig); err != nil {
			continue
		}

		switch rule.RuleType {
		case "完整性":
			issues = append(issues, checkCompletenessRule(dbConn, databaseName, tableName, columnName, dbType, rule, ruleConfig)...)
		case "准确性":
			issues = append(issues, checkAccuracyRule(dbConn, databaseName, tableName, columnName, dbType, rule, ruleConfig)...)
		case "唯一性":
			issues = append(issues, checkUniquenessRule(dbConn, databaseName, tableName, columnName, dbType, rule, ruleConfig)...)
		case "一致性":
			issues = append(issues, checkConsistencyRule(dbConn, databaseName, tableName, columnName, dbType, rule, ruleConfig)...)
		case "及时性":
			issues = append(issues, checkTimelinessRule(dbConn, databaseName, tableName, columnName, dbType, rule, ruleConfig)...)
		}
	}

	return issues
}

// checkNullRate 检查空值率
func checkNullRate(dbConn *sql.DB, databaseName, tableName, columnName, dbType string) (float64, int64, error) {
	// 根据数据库类型构建不同的SQL
	var query string

	// 转义标识符以防止SQL注入
	if strings.ToUpper(dbType) == "POSTGRESQL" {
		// PostgreSQL使用双引号
		escapedColumn := strings.ReplaceAll(columnName, `"`, `""`)
		escapedTable := strings.ReplaceAll(tableName, `"`, `""`)
		escapedDb := strings.ReplaceAll(databaseName, `"`, `""`)
		query = fmt.Sprintf(`SELECT COUNT(*) as total, SUM(CASE WHEN "%s" IS NULL THEN 1 ELSE 0 END) as null_count FROM "%s"."%s"`, escapedColumn, escapedDb, escapedTable)
	} else {
		// MySQL等使用反引号
		escapedColumn := strings.ReplaceAll(columnName, "`", "``")
		escapedTable := strings.ReplaceAll(tableName, "`", "``")
		escapedDb := strings.ReplaceAll(databaseName, "`", "``")
		query = fmt.Sprintf("SELECT COUNT(*) as total, SUM(CASE WHEN `%s` IS NULL THEN 1 ELSE 0 END) as null_count FROM `%s`.`%s`", escapedColumn, escapedDb, escapedTable)
	}

	var total, nullCount int64
	err := dbConn.QueryRow(query).Scan(&total, &nullCount)
	if err != nil {
		return 0, 0, err
	}

	if total == 0 {
		return 0, 0, nil
	}

	rate := float64(nullCount) / float64(total)
	return rate, nullCount, nil
}

// checkCompletenessRule 检查完整性规则
func checkCompletenessRule(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, rule model.DataQualityRule, config map[string]interface{}) []model.DataQualityIssue {
	var issues []model.DataQualityIssue

	// 检查空值率
	if maxNullRate, ok := config["max_null_rate"].(float64); ok {
		nullRate, nullCount, err := checkNullRate(dbConn, databaseName, tableName, columnName, dbType)
		if err == nil && nullRate > maxNullRate {
			issues = append(issues, model.DataQualityIssue{
				DatabaseName: databaseName,
				TableNameX:   tableName,
				ColumnName:   columnName,
				IssueType:    "完整性",
				IssueLevel:   rule.Severity,
				IssueDesc:    fmt.Sprintf("字段空值率超过阈值: %.2f%% > %.2f%%, 空值数量: %d", nullRate*100, maxNullRate*100, nullCount),
				IssueCount:   int(nullCount),
				IssueRate:    nullRate * 100,
				CheckTime:    time.Now(),
				Status:       1,
			})
		}
	}

	return issues
}

// checkAccuracyRule 检查准确性规则
func checkAccuracyRule(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, rule model.DataQualityRule, config map[string]interface{}) []model.DataQualityIssue {
	var issues []model.DataQualityIssue
	logger := log.Logger

	// 1. 格式验证（正则表达式）
	if pattern, ok := config["pattern"].(string); ok {
		if columns, ok := config["columns"].([]interface{}); ok {
			// 检查当前列是否在配置的列名列表中
			columnMatched := false
			for _, col := range columns {
				if colStr, ok := col.(string); ok && strings.EqualFold(colStr, columnName) {
					columnMatched = true
					break
				}
			}
			if !columnMatched {
				return issues // 当前列不在检查列表中
			}
		}

		// 执行格式验证
		invalidCount, err := checkPatternMatch(dbConn, databaseName, tableName, columnName, dbType, pattern)
		if err != nil {
			logger.Warn("格式验证检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
		} else if invalidCount > 0 {
			totalCount, _ := getTotalCount(dbConn, databaseName, tableName, dbType)
			invalidRate := float64(invalidCount) / float64(totalCount)
			if totalCount > 0 && invalidRate > rule.Threshold/100.0 {
				issues = append(issues, model.DataQualityIssue{
					DatabaseName: databaseName,
					TableNameX:   tableName,
					ColumnName:   columnName,
					IssueType:    "准确性",
					IssueLevel:   rule.Severity,
					IssueDesc:    fmt.Sprintf("字段格式不符合规范: 不符合格式的记录数 %d, 占比 %.2f%%", invalidCount, invalidRate*100),
					IssueCount:   int(invalidCount),
					IssueRate:    invalidRate * 100,
					CheckTime:    time.Now(),
					Status:       1,
				})
			}
		}
	}

	// 2. 数值范围检查
	if rangeConfig, ok := config[columnName].(map[string]interface{}); ok {
		minVal, hasMin := rangeConfig["min"]
		maxVal, hasMax := rangeConfig["max"]
		if hasMin || hasMax {
			invalidCount, err := checkNumericRange(dbConn, databaseName, tableName, columnName, dbType, minVal, maxVal)
			if err != nil {
				logger.Warn("数值范围检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
			} else if invalidCount > 0 {
				totalCount, _ := getTotalCount(dbConn, databaseName, tableName, dbType)
				invalidRate := float64(invalidCount) / float64(totalCount)
				if totalCount > 0 && invalidRate > rule.Threshold/100.0 {
					issues = append(issues, model.DataQualityIssue{
						DatabaseName: databaseName,
						TableNameX:   tableName,
						ColumnName:   columnName,
						IssueType:    "准确性",
						IssueLevel:   rule.Severity,
						IssueDesc:    fmt.Sprintf("字段数值超出合理范围: 超出范围的记录数 %d, 占比 %.2f%%", invalidCount, invalidRate*100),
						IssueCount:   int(invalidCount),
						IssueRate:    invalidRate * 100,
						CheckTime:    time.Now(),
						Status:       1,
					})
				}
			}
		}
	}

	// 3. 日期有效性检查 - 只有当字段是日期/时间类型时才检查
	if checkFutureDate, ok := config["check_future_date"].(bool); ok && checkFutureDate {
		// 先检查当前字段是否是日期/时间类型
		isDateTime, err := isDateTimeColumn(dbConn, databaseName, tableName, columnName, dbType)
		if err != nil {
			logger.Warn("检查字段是否为日期类型失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
		} else if isDateTime {
			// 只有确认是日期/时间字段才进行日期有效性检查
			invalidCount, err := checkFutureDateValues(dbConn, databaseName, tableName, columnName, dbType)
			if err != nil {
				logger.Warn("日期有效性检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
			} else if invalidCount > 0 {
				totalCount, _ := getTotalCount(dbConn, databaseName, tableName, dbType)
				invalidRate := float64(invalidCount) / float64(totalCount)
				if totalCount > 0 && invalidRate > rule.Threshold/100.0 {
					issues = append(issues, model.DataQualityIssue{
						DatabaseName: databaseName,
						TableNameX:   tableName,
						ColumnName:   columnName,
						IssueType:    "准确性",
						IssueLevel:   rule.Severity,
						IssueDesc:    fmt.Sprintf("日期字段包含未来日期: 无效记录数 %d, 占比 %.2f%%", invalidCount, invalidRate*100),
						IssueCount:   int(invalidCount),
						IssueRate:    invalidRate * 100,
						CheckTime:    time.Now(),
						Status:       1,
					})
				}
			}
		}
		// 如果不是日期/时间字段，直接跳过日期有效性检查
	}

	return issues
}

// checkUniquenessRule 检查唯一性规则
func checkUniquenessRule(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, rule model.DataQualityRule, config map[string]interface{}) []model.DataQualityIssue {
	var issues []model.DataQualityIssue
	logger := log.Logger

	// 1. 主键唯一性检查 - 只有当字段确实是主键时才检查
	if checkPrimaryKey, ok := config["check_primary_key"].(bool); ok && checkPrimaryKey {
		// 先检查当前字段是否是主键
		isPrimaryKey, err := isPrimaryKeyColumn(dbConn, databaseName, tableName, columnName, dbType)
		if err != nil {
			logger.Warn("检查字段是否为主键失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
		} else if isPrimaryKey {
			// 只有确认是主键字段才进行唯一性检查
			duplicateCount, err := checkPrimaryKeyUniqueness(dbConn, databaseName, tableName, columnName, dbType)
			if err != nil {
				logger.Warn("主键唯一性检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
			} else if duplicateCount > 0 {
				issues = append(issues, model.DataQualityIssue{
					DatabaseName: databaseName,
					TableNameX:   tableName,
					ColumnName:   columnName,
					IssueType:    "唯一性",
					IssueLevel:   rule.Severity,
					IssueDesc:    fmt.Sprintf("主键字段存在重复值: 重复记录数 %d", duplicateCount),
					IssueCount:   int(duplicateCount),
					IssueRate:    0, // 唯一性违规是严重问题，即使只有一条也违规
					CheckTime:    time.Now(),
					Status:       1,
				})
			}
		}
		// 如果不是主键字段，直接跳过主键唯一性检查
	}

	// 2. 业务唯一性检查
	if uniqueFields, ok := config["unique_fields"].([]interface{}); ok {
		columnMatched := false
		for _, field := range uniqueFields {
			if fieldStr, ok := field.(string); ok && strings.EqualFold(fieldStr, columnName) {
				columnMatched = true
				break
			}
		}
		if columnMatched {
			duplicateCount, err := checkColumnUniqueness(dbConn, databaseName, tableName, columnName, dbType)
			if err != nil {
				logger.Warn("业务唯一性检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
			} else if duplicateCount > 0 {
				totalCount, _ := getTotalCount(dbConn, databaseName, tableName, dbType)
				duplicateRate := float64(duplicateCount) / float64(totalCount)
				if totalCount > 0 && duplicateRate > rule.Threshold/100.0 {
					issues = append(issues, model.DataQualityIssue{
						DatabaseName: databaseName,
						TableNameX:   tableName,
						ColumnName:   columnName,
						IssueType:    "唯一性",
						IssueLevel:   rule.Severity,
						IssueDesc:    fmt.Sprintf("业务唯一字段存在重复值: 重复记录数 %d, 占比 %.2f%%", duplicateCount, duplicateRate*100),
						IssueCount:   int(duplicateCount),
						IssueRate:    duplicateRate * 100,
						CheckTime:    time.Now(),
						Status:       1,
					})
				}
			}
		}
	}

	return issues
}

// checkConsistencyRule 检查一致性规则
func checkConsistencyRule(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, rule model.DataQualityRule, config map[string]interface{}) []model.DataQualityIssue {
	var issues []model.DataQualityIssue
	logger := log.Logger

	// 1. 枚举值一致性检查
	if enumFields, ok := config["enum_fields"].(map[string]interface{}); ok {
		if enumValues, ok := enumFields[columnName].([]interface{}); ok {
			invalidCount, err := checkEnumValues(dbConn, databaseName, tableName, columnName, dbType, enumValues)
			if err != nil {
				logger.Warn("枚举值一致性检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
			} else if invalidCount > 0 {
				totalCount, _ := getTotalCount(dbConn, databaseName, tableName, dbType)
				invalidRate := float64(invalidCount) / float64(totalCount)
				if totalCount > 0 && invalidRate > rule.Threshold/100.0 {
					issues = append(issues, model.DataQualityIssue{
						DatabaseName: databaseName,
						TableNameX:   tableName,
						ColumnName:   columnName,
						IssueType:    "一致性",
						IssueLevel:   rule.Severity,
						IssueDesc:    fmt.Sprintf("字段值不在预定义的枚举值范围内: 无效记录数 %d, 占比 %.2f%%", invalidCount, invalidRate*100),
						IssueCount:   int(invalidCount),
						IssueRate:    invalidRate * 100,
						CheckTime:    time.Now(),
						Status:       1,
					})
				}
			}
		}
	}

	// 注意：外键一致性和数据关联一致性检查需要跨表查询，实现较复杂，这里先实现枚举值检查
	// 外键一致性检查需要知道关联表信息，可以在后续版本中实现

	return issues
}

// checkTimelinessRule 检查及时性规则
func checkTimelinessRule(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, rule model.DataQualityRule, config map[string]interface{}) []model.DataQualityIssue {
	var issues []model.DataQualityIssue
	logger := log.Logger

	// 1. 数据更新时效性检查
	if maxIntervalHours, ok := config["max_update_interval_hours"].(float64); ok {
		if checkFields, ok := config["check_fields"].([]interface{}); ok {
			columnMatched := false
			for _, field := range checkFields {
				if fieldStr, ok := field.(string); ok && strings.EqualFold(fieldStr, columnName) {
					columnMatched = true
					break
				}
			}
			if !columnMatched {
				return issues // 当前列不在检查列表中
			}
		}

		// 检查最后更新时间
		staleCount, err := checkUpdateTimeliness(dbConn, databaseName, tableName, columnName, dbType, int(maxIntervalHours))
		if err != nil {
			logger.Warn("数据更新时效性检查失败", zap.String("table", tableName), zap.String("column", columnName), zap.Error(err))
		} else if staleCount > 0 {
			totalCount, _ := getTotalCount(dbConn, databaseName, tableName, dbType)
			staleRate := float64(staleCount) / float64(totalCount)
			if totalCount > 0 && staleRate > rule.Threshold/100.0 {
				issues = append(issues, model.DataQualityIssue{
					DatabaseName: databaseName,
					TableNameX:   tableName,
					ColumnName:   columnName,
					IssueType:    "及时性",
					IssueLevel:   rule.Severity,
					IssueDesc:    fmt.Sprintf("数据更新时效性不足: 超过%d小时未更新的记录数 %d, 占比 %.2f%%", int(maxIntervalHours), staleCount, staleRate*100),
					IssueCount:   int(staleCount),
					IssueRate:    staleRate * 100,
					CheckTime:    time.Now(),
					Status:       1,
				})
			}
		}
	}

	return issues
}

// 计算质量指标的函数
func calculateCompleteness(issues []model.DataQualityIssue, totalColumns int) float64 {
	if totalColumns == 0 {
		return 100.0
	}
	completenessIssues := 0
	for _, issue := range issues {
		if issue.IssueType == "完整性" {
			completenessIssues++
		}
	}
	return (1.0 - float64(completenessIssues)/float64(totalColumns)) * 100.0
}

func calculateAccuracy(issues []model.DataQualityIssue, totalColumns int) float64 {
	if totalColumns == 0 {
		return 100.0
	}
	accuracyIssues := 0
	for _, issue := range issues {
		if issue.IssueType == "准确性" {
			accuracyIssues++
		}
	}
	return (1.0 - float64(accuracyIssues)/float64(totalColumns)) * 100.0
}

func calculateTableCompleteness(tables []string, issues []model.DataQualityIssue) float64 {
	if len(tables) == 0 {
		return 100.0
	}
	// 简化计算：基于有问题的表数量
	problemTables := make(map[string]bool)
	for _, issue := range issues {
		problemTables[issue.TableNameX] = true
	}
	return (1.0 - float64(len(problemTables))/float64(len(tables))) * 100.0
}

func calculateConsistency(issues []model.DataQualityIssue, totalColumns int) float64 {
	if totalColumns == 0 {
		return 100.0
	}
	consistencyIssues := 0
	for _, issue := range issues {
		if issue.IssueType == "一致性" {
			consistencyIssues++
		}
	}
	return (1.0 - float64(consistencyIssues)/float64(totalColumns)) * 100.0
}

func calculateUniqueness(issues []model.DataQualityIssue, totalColumns int) float64 {
	if totalColumns == 0 {
		return 100.0
	}
	uniquenessIssues := 0
	for _, issue := range issues {
		if issue.IssueType == "唯一性" {
			uniquenessIssues++
		}
	}
	return (1.0 - float64(uniquenessIssues)/float64(totalColumns)) * 100.0
}

func calculateTimeliness(issues []model.DataQualityIssue, totalColumns int) float64 {
	if totalColumns == 0 {
		return 100.0
	}
	timelinessIssues := 0
	for _, issue := range issues {
		if issue.IssueType == "及时性" {
			timelinessIssues++
		}
	}
	return (1.0 - float64(timelinessIssues)/float64(totalColumns)) * 100.0
}

func getQualityLevel(score float64) string {
	if score >= 90 {
		return "优秀"
	} else if score >= 80 {
		return "良好"
	} else if score >= 70 {
		return "一般"
	} else if score >= 60 {
		return "较差"
	}
	return "差"
}

// saveAssessmentResult 保存评估结果
func saveAssessmentResult(assessment *model.DataQualityAssessment, task model.DataQualityTask) error {
	// 保存评估记录（先创建，获取ID）
	assessmentRecord := *assessment
	assessmentRecord.Issues = nil // 清除临时字段
	if err := database.DB.Create(&assessmentRecord).Error; err != nil {
		return fmt.Errorf("保存评估记录失败: %v", err)
	}

	// 保存问题列表（从临时字段中获取）
	issues := assessment.Issues
	for i := range issues {
		issues[i].AssessmentId = assessmentRecord.Id
		if err := database.DB.Create(&issues[i]).Error; err != nil {
			log.Logger.Warn("保存质量问题失败", zap.Error(err))
		}
	}

	// 保存到历史表
	history := model.DataQualityHistory{
		AssessmentId:      assessmentRecord.Id,
		DatasourceId:      assessmentRecord.DatasourceId,
		DatabaseName:      assessmentRecord.DatabaseName,
		AssessmentDate:    assessmentRecord.AssessmentTime,
		OverallScore:      assessmentRecord.OverallScore,
		FieldCompleteness: assessmentRecord.FieldCompleteness,
		FieldAccuracy:     assessmentRecord.FieldAccuracy,
		TableCompleteness: assessmentRecord.TableCompleteness,
		DataConsistency:   assessmentRecord.DataConsistency,
		DataUniqueness:    assessmentRecord.DataUniqueness,
		DataTimeliness:    assessmentRecord.DataTimeliness,
		TotalIssues:       assessmentRecord.TotalIssues,
	}
	if err := database.DB.Create(&history).Error; err != nil {
		log.Logger.Warn("保存评估历史失败", zap.Error(err))
	}

	// TODO: 生成AI分析结果（可以后续集成AI功能）

	return nil
}

// ========== 辅助函数：准确性检查 ==========

// checkPatternMatch 检查字段值是否符合正则表达式模式
// 返回不符合模式的数量
func checkPatternMatch(dbConn *sql.DB, databaseName, tableName, columnName, dbType, pattern string) (int64, error) {
	// 注意：数据库层面的正则匹配需要根据数据库类型使用不同的函数
	// MySQL使用REGEXP，PostgreSQL使用~或~*
	var query string

	escapedColumn := escapeIdentifier(columnName, dbType)
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		// PostgreSQL使用!~进行正则不匹配（区分大小写）或!~*（不区分大小写）
		// 这里使用!~*不区分大小写
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s" WHERE "%s" IS NOT NULL AND "%s" !~* $1`, escapedDb, escapedTable, escapedColumn, escapedColumn)
		var count int64
		err := dbConn.QueryRow(query, pattern).Scan(&count)
		return count, err
	} else {
		// MySQL使用NOT REGEXP或NOT RLIKE
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE `%s` IS NOT NULL AND `%s` NOT REGEXP ?", escapedDb, escapedTable, escapedColumn, escapedColumn)
		var count int64
		err := dbConn.QueryRow(query, pattern).Scan(&count)
		return count, err
	}
}

// checkNumericRange 检查数值范围
func checkNumericRange(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, minVal, maxVal interface{}) (int64, error) {
	var query string
	escapedColumn := escapeIdentifier(columnName, dbType)
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	conditions := []string{}
	args := []interface{}{}

	if minVal != nil {
		if strings.ToUpper(dbType) == "POSTGRESQL" {
			conditions = append(conditions, fmt.Sprintf(`"%s" < $%d`, escapedColumn, len(args)+1))
		} else {
			conditions = append(conditions, fmt.Sprintf("`%s` < ?", escapedColumn))
		}
		args = append(args, minVal)
	}
	if maxVal != nil {
		if strings.ToUpper(dbType) == "POSTGRESQL" {
			conditions = append(conditions, fmt.Sprintf(`"%s" > $%d`, escapedColumn, len(args)+1))
		} else {
			conditions = append(conditions, fmt.Sprintf("`%s` > ?", escapedColumn))
		}
		args = append(args, maxVal)
	}

	if len(conditions) == 0 {
		return 0, nil
	}

	whereClause := strings.Join(conditions, " OR ")
	if strings.ToUpper(dbType) == "POSTGRESQL" {
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s" WHERE "%s" IS NOT NULL AND (%s)`, escapedDb, escapedTable, escapedColumn, whereClause)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE `%s` IS NOT NULL AND (%s)", escapedDb, escapedTable, escapedColumn, whereClause)
	}

	var count int64
	var err error
	if strings.ToUpper(dbType) == "POSTGRESQL" {
		err = dbConn.QueryRow(query, args...).Scan(&count)
	} else {
		err = dbConn.QueryRow(query, args...).Scan(&count)
	}
	return count, err
}

// checkFutureDateValues 检查未来日期
func checkFutureDateValues(dbConn *sql.DB, databaseName, tableName, columnName, dbType string) (int64, error) {
	var query string
	escapedColumn := escapeIdentifier(columnName, dbType)
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	now := time.Now().Format("2006-01-02 15:04:05")

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s" WHERE "%s" IS NOT NULL AND "%s" > $1`, escapedDb, escapedTable, escapedColumn, escapedColumn)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE `%s` IS NOT NULL AND `%s` > ?", escapedDb, escapedTable, escapedColumn, escapedColumn)
	}

	var count int64
	err := dbConn.QueryRow(query, now).Scan(&count)
	return count, err
}

// ========== 辅助函数：唯一性检查 ==========

// checkPrimaryKeyUniqueness 检查主键唯一性
func checkPrimaryKeyUniqueness(dbConn *sql.DB, databaseName, tableName, columnName, dbType string) (int64, error) {
	var query string
	escapedColumn := escapeIdentifier(columnName, dbType)
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		query = fmt.Sprintf(`SELECT COUNT(*) - COUNT(DISTINCT "%s") FROM "%s"."%s" WHERE "%s" IS NOT NULL`, escapedColumn, escapedDb, escapedTable, escapedColumn)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) - COUNT(DISTINCT `%s`) FROM `%s`.`%s` WHERE `%s` IS NOT NULL", escapedColumn, escapedDb, escapedTable, escapedColumn)
	}

	var duplicateCount int64
	err := dbConn.QueryRow(query).Scan(&duplicateCount)
	if duplicateCount < 0 {
		duplicateCount = 0
	}
	return duplicateCount, err
}

// checkColumnUniqueness 检查列唯一性
func checkColumnUniqueness(dbConn *sql.DB, databaseName, tableName, columnName, dbType string) (int64, error) {
	// 与主键唯一性检查相同
	return checkPrimaryKeyUniqueness(dbConn, databaseName, tableName, columnName, dbType)
}

// ========== 辅助函数：一致性检查 ==========

// checkEnumValues 检查枚举值
func checkEnumValues(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, enumValues []interface{}) (int64, error) {
	if len(enumValues) == 0 {
		return 0, nil
	}

	var query string
	escapedColumn := escapeIdentifier(columnName, dbType)
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	// 构建IN子句
	placeholders := []string{}
	args := []interface{}{}
	for i, val := range enumValues {
		if strings.ToUpper(dbType) == "POSTGRESQL" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		} else {
			placeholders = append(placeholders, "?")
		}
		args = append(args, val)
	}

	inClause := strings.Join(placeholders, ",")

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s" WHERE "%s" IS NOT NULL AND "%s" NOT IN (%s)`, escapedDb, escapedTable, escapedColumn, escapedColumn, inClause)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE `%s` IS NOT NULL AND `%s` NOT IN (%s)", escapedDb, escapedTable, escapedColumn, escapedColumn, inClause)
	}

	var count int64
	err := dbConn.QueryRow(query, args...).Scan(&count)
	return count, err
}

// ========== 辅助函数：及时性检查 ==========

// checkUpdateTimeliness 检查更新时效性
func checkUpdateTimeliness(dbConn *sql.DB, databaseName, tableName, columnName, dbType string, maxIntervalHours int) (int64, error) {
	var query string
	escapedColumn := escapeIdentifier(columnName, dbType)
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	cutoffTime := time.Now().Add(-time.Duration(maxIntervalHours) * time.Hour).Format("2006-01-02 15:04:05")

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s" WHERE "%s" IS NOT NULL AND "%s" < $1`, escapedDb, escapedTable, escapedColumn, escapedColumn)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s` WHERE `%s` IS NOT NULL AND `%s` < ?", escapedDb, escapedTable, escapedColumn, escapedColumn)
	}

	var count int64
	err := dbConn.QueryRow(query, cutoffTime).Scan(&count)
	return count, err
}

// ========== 通用辅助函数 ==========

// getTotalCount 获取表的总记录数
func getTotalCount(dbConn *sql.DB, databaseName, tableName, dbType string) (int64, error) {
	var query string
	escapedTable := escapeIdentifier(tableName, dbType)
	escapedDb := escapeIdentifier(databaseName, dbType)

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		query = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s"`, escapedDb, escapedTable)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", escapedDb, escapedTable)
	}

	var count int64
	err := dbConn.QueryRow(query).Scan(&count)
	return count, err
}

// escapeIdentifier 转义数据库标识符
func escapeIdentifier(identifier, dbType string) string {
	if strings.ToUpper(dbType) == "POSTGRESQL" {
		// PostgreSQL使用双引号
		return strings.ReplaceAll(identifier, `"`, `""`)
	} else {
		// MySQL等使用反引号
		return strings.ReplaceAll(identifier, "`", "``")
	}
}

// isPrimaryKeyColumn 检查字段是否是主键
func isPrimaryKeyColumn(dbConn *sql.DB, databaseName, tableName, columnName, dbType string) (bool, error) {
	var query string
	var args []interface{}

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		// PostgreSQL查询主键 - 使用pg_constraint和pg_attribute
		query = `SELECT COUNT(*) FROM information_schema.table_constraints tc
			JOIN information_schema.constraint_column_usage AS ccu ON tc.constraint_name = ccu.constraint_name
			WHERE tc.table_schema = $1 
			AND tc.table_name = $2 
			AND tc.constraint_type = 'PRIMARY KEY'
			AND ccu.column_name = $3`
		args = []interface{}{databaseName, tableName, columnName}
	} else {
		// MySQL等使用KEY_COLUMN_USAGE表
		query = `SELECT COUNT(*) FROM information_schema.KEY_COLUMN_USAGE
			WHERE TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
			AND COLUMN_NAME = ?
			AND CONSTRAINT_NAME = 'PRIMARY'`
		args = []interface{}{databaseName, tableName, columnName}
	}

	var count int64
	err := dbConn.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// isDateTimeColumn 检查字段是否是日期/时间类型
func isDateTimeColumn(dbConn *sql.DB, databaseName, tableName, columnName, dbType string) (bool, error) {
	var query string
	var args []interface{}

	if strings.ToUpper(dbType) == "POSTGRESQL" {
		// PostgreSQL查询字段类型
		query = `SELECT LOWER(data_type) FROM information_schema.columns
			WHERE table_schema = $1 
			AND table_name = $2 
			AND column_name = $3`
		args = []interface{}{databaseName, tableName, columnName}
	} else {
		// MySQL等查询字段类型
		query = `SELECT LOWER(DATA_TYPE) FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = ?
			AND TABLE_NAME = ?
			AND COLUMN_NAME = ?`
		args = []interface{}{databaseName, tableName, columnName}
	}

	var dataType string
	err := dbConn.QueryRow(query, args...).Scan(&dataType)
	if err != nil {
		return false, err
	}

	// 定义日期/时间类型列表（转换为小写进行比较）
	dataTypeLower := strings.ToLower(dataType)
	dateTypeKeywords := []string{
		"date",
		"datetime",
		"timestamp",
		"time",
		"year",
	}

	// 检查是否是日期/时间类型
	for _, keyword := range dateTypeKeywords {
		if strings.Contains(dataTypeLower, keyword) {
			return true, nil
		}
	}

	return false, nil
}
