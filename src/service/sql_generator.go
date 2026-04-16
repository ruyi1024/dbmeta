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

package service

import (
	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/model"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// QueryContext 查询上下文
type QueryContext struct {
	DatasourceId   int
	DatasourceType string
	Host           string
	Port           string
	DatabaseName   string
	TableName      string
	Metadata       *MetadataInfo
	History        []model.ChatMessage
	UserName       string
}

// MetadataInfo 元数据信息
type MetadataInfo struct {
	Databases []map[string]interface{}
	Tables    []map[string]interface{} // 新格式：每个表包含 table_name, table_comment, columns
	Columns   []map[string]interface{} // 保留用于向后兼容
}

// RuleMatchResult 规则匹配结果
type RuleMatchResult struct {
	Rule   *model.SemanticSqlRule
	Params map[string]string
}

// MatchRule 使用AI大模型进行语义匹配规则并提取参数
// 优化：优先检查rule_name的精确匹配，如果匹配则跳过LLM判断
func MatchRule(userQuery string) (*RuleMatchResult, error) {
	// 获取所有启用的规则，按优先级排序
	var rules []model.SemanticSqlRule
	result := database.DB.Where("enabled = 1").
		Order("priority DESC").
		Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("查询规则失败: %v", result.Error)
	}

	if len(rules) == 0 {
		return nil, nil // 没有启用的规则
	}

	// 优化：优先检查rule_name的精确匹配（不区分大小写，去除首尾空格）
	userQueryTrimmed := strings.TrimSpace(userQuery)
	for _, rule := range rules {
		if strings.EqualFold(strings.TrimSpace(rule.RuleName), userQueryTrimmed) {
			zap.L().Info("规则名称精确匹配成功，跳过LLM判断",
				zap.String("rule_name", rule.RuleName),
				zap.String("user_query", userQuery))
			// 精确匹配时，尝试从用户输入中提取参数（如果有）
			params := extractCommonParams(userQuery)
			return &RuleMatchResult{
				Rule:   &rule,
				Params: params,
			}, nil
		}
	}

	// 如果没有精确匹配，使用AI大模型进行语义匹配
	matchedRule, params, err := matchRuleWithAI(userQuery, rules)
	if err != nil {
		zap.L().Warn("AI语义匹配失败，回退到正则匹配", zap.Error(err))
		// 如果AI匹配失败，回退到原来的正则匹配方式
		return matchRuleWithRegex(userQuery, rules)
	}

	if matchedRule == nil {
		return nil, nil // 没有匹配的规则
	}

	return &RuleMatchResult{
		Rule:   matchedRule,
		Params: params,
	}, nil
}

// matchRuleWithAI 使用AI大模型进行语义匹配
func matchRuleWithAI(userQuery string, rules []model.SemanticSqlRule) (*model.SemanticSqlRule, map[string]string, error) {
	// 构建规则列表的描述，用于AI分析
	var rulesDescription strings.Builder
	rulesDescription.WriteString("以下是可用的SQL规则列表：\n\n")
	for i, rule := range rules {
		rulesDescription.WriteString(fmt.Sprintf("规则%d:\n", i+1))
		rulesDescription.WriteString(fmt.Sprintf("  规则名称: %s\n", rule.RuleName))
		rulesDescription.WriteString(fmt.Sprintf("  语义模式: %s\n", rule.SemanticPattern))
		rulesDescription.WriteString(fmt.Sprintf("  查询类型: %s\n", rule.QueryType))
		if rule.Description != "" {
			rulesDescription.WriteString(fmt.Sprintf("  描述: %s\n", rule.Description))
		}
		rulesDescription.WriteString("\n")
	}

	// 构建AI提示词
	prompt := fmt.Sprintf(`你是一个SQL规则匹配专家。请分析用户的问题，并从以下规则列表中找到最匹配的规则。

用户问题: %s

%s

请仔细分析用户问题的语义，与每个规则的语义模式进行匹配。返回格式必须是JSON，包含以下字段：
{
  "rule_index": <匹配的规则索引（从1开始，如果没有匹配则返回0）>,
  "confidence": <匹配置信度，0-1之间的浮点数>,
  "params": {
    <从用户问题中提取的参数，键值对形式>
  },
  "reason": "<匹配原因说明>"
}

注意：
1. 如果用户问题的语义与某个规则的语义模式高度匹配，返回该规则的索引
2. 如果没有匹配的规则，rule_index返回0
3. params字段包含从用户问题中提取的参数，如host、port、database、table等
4. 只返回JSON，不要包含其他文字说明`, userQuery, rulesDescription.String())

	// 调用AI模型
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个专业的SQL规则匹配助手，擅长分析用户问题的语义并匹配到最合适的SQL规则。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := CallWithFailover(messages, &ChatOptions{
		Temperature: 0.3, // 使用较低的温度以获得更确定性的结果
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("AI模型调用失败: %v", err)
	}

	// 解析AI返回的JSON
	var aiResult struct {
		RuleIndex  int               `json:"rule_index"`
		Confidence float64           `json:"confidence"`
		Params     map[string]string `json:"params"`
		Reason     string            `json:"reason"`
	}

	// 尝试从响应中提取JSON（可能包含markdown代码块）
	content := strings.TrimSpace(response.Content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	if err := json.Unmarshal([]byte(content), &aiResult); err != nil {
		zap.L().Warn("解析AI返回的JSON失败", zap.String("content", response.Content), zap.Error(err))
		return nil, nil, fmt.Errorf("解析AI返回结果失败: %v", err)
	}

	// 检查是否有匹配的规则
	if aiResult.RuleIndex <= 0 || aiResult.RuleIndex > len(rules) {
		zap.L().Info("AI未找到匹配的规则", zap.Int("rule_index", aiResult.RuleIndex), zap.Float64("confidence", aiResult.Confidence))
		return nil, nil, nil
	}

	// 检查置信度（如果置信度太低，可能不匹配）
	if aiResult.Confidence < 0.5 {
		zap.L().Info("AI匹配置信度太低", zap.Int("rule_index", aiResult.RuleIndex), zap.Float64("confidence", aiResult.Confidence))
		return nil, nil, nil
	}

	matchedRule := &rules[aiResult.RuleIndex-1]
	zap.L().Info("AI语义匹配成功",
		zap.String("rule", matchedRule.RuleName),
		zap.Int("rule_index", aiResult.RuleIndex),
		zap.Float64("confidence", aiResult.Confidence),
		zap.String("reason", aiResult.Reason))

	// 如果AI没有提取参数，尝试从用户问题中提取
	params := aiResult.Params
	if params == nil {
		params = make(map[string]string)
	}
	if len(params) == 0 {
		params = extractCommonParams(userQuery)
	}

	return matchedRule, params, nil
}

// matchRuleWithRegex 使用正则表达式匹配规则（回退方案）
func matchRuleWithRegex(userQuery string, rules []model.SemanticSqlRule) (*RuleMatchResult, error) {
	// 遍历规则进行匹配
	for _, rule := range rules {
		matched, params := matchPattern(rule.SemanticPattern, userQuery)
		if matched {
			return &RuleMatchResult{
				Rule:   &rule,
				Params: params,
			}, nil
		}
	}
	return nil, nil // 没有匹配的规则
}

// matchPattern 匹配模式并提取参数
func matchPattern(pattern, text string) (bool, map[string]string) {
	// 编译正则表达式
	re, err := regexp.Compile("(?i)" + pattern) // 不区分大小写
	if err != nil {
		zap.L().Error("正则表达式编译失败", zap.String("pattern", pattern), zap.Error(err))
		return false, nil
	}

	// 检查是否匹配
	if !re.MatchString(text) {
		return false, nil
	}

	// 提取命名组参数
	matches := re.FindStringSubmatch(text)
	names := re.SubexpNames()
	params := make(map[string]string)

	for i, name := range names {
		if i > 0 && name != "" && i < len(matches) {
			params[name] = matches[i]
		}
	}

	// 如果没有命名组，尝试提取一些常见参数
	if len(params) == 0 {
		params = extractCommonParams(text)
	}

	return true, params
}

// extractCommonParams 提取常见参数
func extractCommonParams(text string) map[string]string {
	params := make(map[string]string)

	// 提取主机和端口
	hostPortRe := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)[:：](\d+)`)
	if matches := hostPortRe.FindStringSubmatch(text); len(matches) >= 3 {
		params["host"] = matches[1]
		params["port"] = matches[2]
	}

	// 提取数据库名
	dbRe := regexp.MustCompile(`(?:数据库|database|db)[:：]?['"]?(\w+)['"]?`)
	if matches := dbRe.FindStringSubmatch(text); len(matches) >= 2 {
		params["database"] = matches[1]
	}

	// 提取表名
	tableRe := regexp.MustCompile(`(?:表|table)[:：]?['"]?(\w+)['"]?`)
	if matches := tableRe.FindStringSubmatch(text); len(matches) >= 2 {
		params["table"] = matches[1]
	}

	return params
}

// GenerateSQLFromRule 根据规则生成SQL
func GenerateSQLFromRule(rule *model.SemanticSqlRule, params map[string]string, context *QueryContext) string {
	sql := rule.SqlTemplate

	// 替换模板中的参数占位符 {param_name}
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		// 检查是否是数据库名或表名参数（这些参数不应该加引号，因为它们是标识符）
		keyLower := strings.ToLower(key)
		isIdentifier := keyLower == "db_name" || keyLower == "database_name" || keyLower == "database" || keyLower == "db" ||
			keyLower == "table_name" || keyLower == "table" || keyLower == "schema_name" || keyLower == "schema"

		var finalValue string
		if isIdentifier {
			// 数据库名和表名是标识符，不加引号，直接使用
			finalValue = value
		} else {
			// 判断是否需要加引号，如果需要则添加引号
			finalValue = formatSQLValue(value)
		}
		sql = strings.ReplaceAll(sql, placeholder, finalValue)
	}

	// 替换上下文参数
	if context != nil {
		if context.Host != "" {
			sql = strings.ReplaceAll(sql, "{host}", formatSQLValue(context.Host))
		}
		if context.Port != "" {
			sql = strings.ReplaceAll(sql, "{port}", formatSQLValue(context.Port))
		}
		if context.DatabaseName != "" {
			// 数据库名是标识符，不加引号
			sql = strings.ReplaceAll(sql, "{database}", context.DatabaseName)
		}
		if context.TableName != "" {
			// 表名是标识符，不加引号
			sql = strings.ReplaceAll(sql, "{table}", context.TableName)
		}
		if context.DatasourceType != "" {
			sql = strings.ReplaceAll(sql, "{datasource_type}", formatSQLValue(context.DatasourceType))
		}
	}

	return sql
}

// formatSQLValue 格式化SQL值，为字符串值添加引号
func formatSQLValue(value string) string {
	// 判断是否需要加引号
	if needsSQLQuotesForValue(value) {
		// 转义SQL注入风险字符，并添加单引号
		escapedValue := strings.ReplaceAll(value, "'", "''")
		return fmt.Sprintf("'%s'", escapedValue)
	}
	// 数字或SQL表达式，直接返回
	return value
}

// needsSQLQuotesForValue 判断SQL值是否需要加引号
// 返回true表示需要加引号（字符串值），false表示不需要（数字、NULL、SQL表达式等）
func needsSQLQuotesForValue(value string) bool {
	value = strings.TrimSpace(value)

	// 空值
	if value == "" {
		return true
	}

	// NULL值（不区分大小写）
	if strings.EqualFold(value, "NULL") {
		return false
	}

	// 纯数字（整数或小数）
	numberRegex := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	if numberRegex.MatchString(value) {
		return false
	}

	// SQL函数或表达式（如 NOW(), COUNT(*), 1+1 等）
	// 如果包含括号、运算符等，可能是SQL表达式
	if strings.Contains(value, "(") || strings.Contains(value, "+") ||
		strings.Contains(value, "-") || strings.Contains(value, "*") ||
		strings.Contains(value, "/") || strings.Contains(value, "=") {
		// 进一步检查是否是常见的SQL函数
		sqlFunctions := []string{"NOW()", "CURRENT_TIMESTAMP", "COUNT", "SUM", "AVG", "MAX", "MIN", "CASE", "WHEN", "THEN", "ELSE", "END"}
		upperValue := strings.ToUpper(value)
		for _, fn := range sqlFunctions {
			if strings.Contains(upperValue, fn) {
				return false
			}
		}
		// 如果包含运算符且看起来像表达式，不加引号
		if strings.ContainsAny(value, "+-*/=") && !strings.Contains(value, "'") {
			return false
		}
	}

	// 布尔值
	if strings.EqualFold(value, "TRUE") || strings.EqualFold(value, "FALSE") {
		return false
	}

	// 默认情况下，字符串值需要加引号
	return true
}

// ParseSqlSet 解析规则中的sql_set字段
func ParseSqlSet(rule *model.SemanticSqlRule) ([]model.SqlSetItem, error) {
	if rule == nil || strings.TrimSpace(rule.SqlSet) == "" {
		return nil, nil
	}
	var sqlSet []model.SqlSetItem
	if err := json.Unmarshal([]byte(rule.SqlSet), &sqlSet); err != nil {
		return nil, fmt.Errorf("解析sql_set失败: %v", err)
	}
	return sqlSet, nil
}

// RenderReport 根据report_template和查询数据使用AI生成报告
// metrics 为从查询结果聚合的关键指标，sqls 为执行的SQL列表
// queryResults 为完整的查询结果数据（用于AI分析）
func RenderReport(reportTemplate string, metrics map[string]interface{}, sqls []string, queryResults ...[]map[string]interface{}) string {
	// 构建查询数据摘要
	var dataSummary strings.Builder
	if len(metrics) > 0 {
		dataSummary.WriteString("## 关键指标数据：\n")
		for k, v := range metrics {
			dataSummary.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
		dataSummary.WriteString("\n")
	}

	// 如果有完整的查询结果，也包含进去
	if len(queryResults) > 0 && len(queryResults[0]) > 0 {
		dataSummary.WriteString("## 查询结果数据：\n")
		// 只取前10条数据作为示例
		maxRows := 10
		if len(queryResults[0]) < maxRows {
			maxRows = len(queryResults[0])
		}
		for i := 0; i < maxRows; i++ {
			row := queryResults[0][i]
			dataSummary.WriteString(fmt.Sprintf("### 记录 %d:\n", i+1))
			for k, v := range row {
				dataSummary.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
			}
			dataSummary.WriteString("\n")
		}
		if len(queryResults[0]) > maxRows {
			dataSummary.WriteString(fmt.Sprintf("... 还有 %d 条记录\n\n", len(queryResults[0])-maxRows))
		}
	}

	// 构建AI提示词
	var prompt strings.Builder
	prompt.WriteString("你是一个专业的数据库分析报告生成助手。请根据以下查询数据和报告模板要求，生成一份专业、清晰、美观的分析报告。\n\n")

	// 如果有报告模板，使用模板作为指导
	tpl := strings.TrimSpace(reportTemplate)
	if tpl != "" {
		prompt.WriteString("## 报告模板要求：\n")
		prompt.WriteString(tpl)
		prompt.WriteString("\n\n")
		prompt.WriteString("**注意**：请根据模板要求生成报告，但不要直接复制模板内容，而是基于实际数据进行分析和总结。\n\n")
	} else {
		prompt.WriteString("## 报告要求：\n")
		prompt.WriteString("请生成一份专业的数据库分析报告，包括：\n")
		prompt.WriteString("1. 报告摘要和总体评估\n")
		prompt.WriteString("2. 关键指标分析（基于提供的数据）\n")
		prompt.WriteString("3. 数据趋势和异常情况（如果有）\n")
		prompt.WriteString("4. 结论和建议\n\n")
	}

	// 添加查询数据
	prompt.WriteString(dataSummary.String())

	// 添加执行的SQL信息（仅作为参考，不要包含在报告中）
	if len(sqls) > 0 {
		prompt.WriteString("## 执行的SQL查询（仅供参考，不要包含在报告中）：\n")
		for i, s := range sqls {
			prompt.WriteString(fmt.Sprintf("%d. ```sql\n%s\n```\n", i+1, s))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("## 生成要求：\n")
	prompt.WriteString("1. 报告应使用Markdown格式，美观易读\n")
	prompt.WriteString("2. 使用表格、列表等格式展示关键数据\n")
	prompt.WriteString("3. 对数据进行专业分析和解读\n")
	prompt.WriteString("4. 提供有价值的结论和建议\n")
	prompt.WriteString("5. 报告应简洁明了，重点突出\n")
	prompt.WriteString("6. 只返回报告内容，不要包含其他说明文字\n")
	prompt.WriteString("7. **重要**：不要在报告中包含SQL查询语句，SQL会单独显示\n")
	prompt.WriteString("8. **重要**：不要在报告中使用\"分析报告\"、\"关键指标\"等标题，直接开始分析内容\n")
	prompt.WriteString("9. **重要**：不要在报告中重复显示查询结果数据，只进行分析和解读，原始数据会单独显示\n")

	// 调用AI生成报告
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个专业的数据库分析报告生成助手，擅长将查询数据转化为清晰、专业的分析报告。",
		},
		{
			Role:    "user",
			Content: prompt.String(),
		},
	}

	response, err := CallWithFailover(messages, &ChatOptions{
		Temperature: 0.7,
		MaxTokens:   2000,
	})
	if err != nil {
		zap.L().Warn("AI生成报告失败，使用默认格式", zap.Error(err))
		// 如果AI生成失败，回退到简单格式
		return renderDefaultReport(metrics, sqls)
	}

	report := strings.TrimSpace(response.Content)
	if report == "" {
		return renderDefaultReport(metrics, sqls)
	}

	// 清理报告（移除可能的markdown代码块标记）
	report = strings.TrimPrefix(report, "```markdown")
	report = strings.TrimPrefix(report, "```")
	report = strings.TrimSuffix(report, "```")
	report = strings.TrimSpace(report)

	return report
}

// renderDefaultReport 生成默认格式的报告（AI生成失败时的回退方案）
// 注意：不在报告中包含SQL和标题，SQL和查询结果会由前端单独显示
func renderDefaultReport(metrics map[string]interface{}, sqls []string) string {
	var sb strings.Builder

	// 不生成"分析报告"和"关键指标"标题，因为前端会单独显示查询结果
	// 只生成简单的数据摘要
	if len(metrics) > 0 {
		sb.WriteString("**数据摘要：**\n\n")
		for k, v := range metrics {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("查询已完成，请查看下方的查询结果。\n")
	}

	// 不在默认报告中包含SQL，SQL会由前端单独显示
	// 这样可以避免重复显示

	return sb.String()
}

// GetMetadataInfo 获取元数据信息
func GetMetadataInfo(databaseName, tableName string) (*MetadataInfo, error) {
	metadata := &MetadataInfo{}

	// 查询数据表列表（按数据库名过滤）
	if databaseName != "" {
		var tables []model.MetaTable
		tx := database.DB.Where("database_name = ? AND is_deleted = 0", databaseName)

		// 如果提供了表名，进一步过滤
		if tableName != "" {
			// 使用 fmt.Sprintf 构建 LIKE 模式，避免 % 被误解析为格式化占位符
			likePattern := fmt.Sprintf("%%%s%%", tableName)
			tx = tx.Where("(table_name like ? OR table_comment like ? or ai_comment like ?)", likePattern, likePattern, likePattern)
		}

		result := tx.Limit(100).Find(&tables)
		if result.Error == nil {
			// 获取所有表的表名，用于查询字段
			var tableNames []string
			tableMap := make(map[string]model.MetaTable) // tableName -> MetaTable
			for _, table := range tables {
				tableNames = append(tableNames, table.TableNameX)
				tableMap[table.TableNameX] = table
			}

			// 查询字段列表（按数据库名和表名过滤）
			var columns []model.MetaColumn
			if len(tableNames) > 0 {
				tx := database.DB.Model(&model.MetaColumn{}).
					Where("database_name = ? AND table_name IN ? AND is_deleted = 0", databaseName, tableNames).
					Order("ordinal_position ASC").
					Limit(2000)
				result := tx.Find(&columns)
				if result.Error != nil {
					zap.L().Warn("查询字段列表失败", zap.Error(result.Error))
				}
			}

			// 按表分组字段
			columnsByTable := make(map[string][]map[string]interface{})
			for _, col := range columns {
				columnComment := col.ColumnComment
				if columnComment == "" && col.AiComment != "" {
					columnComment = col.AiComment
					zap.L().Debug("使用AI生成的字段注释",
						zap.String("table", col.TableNameX),
						zap.String("column", col.ColumnName),
						zap.String("ai_comment", col.AiComment))
				}
				columnInfo := map[string]interface{}{
					"column_name":    col.ColumnName,
					"column_comment": columnComment,
					"data_type":      col.DataType,
					"is_nullable":    col.IsNullable,
					"default_value":  col.DefaultValue,
				}
				columnsByTable[col.TableNameX] = append(columnsByTable[col.TableNameX], columnInfo)
			}

			// 构建新的表格式：每个表包含 table_name, table_comment, columns
			metadata.Tables = make([]map[string]interface{}, len(tables))
			for i, table := range tables {
				tableComment := table.TableComment
				if tableComment == "" && table.AiComment != "" {
					tableComment = table.AiComment
				}

				tableInfo := map[string]interface{}{
					"table_name":    table.TableNameX,
					"table_comment": tableComment,
					"columns":       columnsByTable[table.TableNameX],
				}
				metadata.Tables[i] = tableInfo
			}

			// 保留旧的 Columns 格式用于向后兼容
			metadata.Columns = []map[string]interface{}{}
			for _, col := range columns {
				columnComment := col.ColumnComment
				if columnComment == "" && col.AiComment != "" {
					columnComment = col.AiComment
				}
				metadata.Columns = append(metadata.Columns, map[string]interface{}{
					"database_name":  col.DatabaseName,
					"table_name":     col.TableNameX,
					"column_name":    col.ColumnName,
					"data_type":      col.DataType,
					"column_comment": columnComment,
					"is_nullable":    col.IsNullable,
					"default_value":  col.DefaultValue,
				})
			}

			zap.L().Info("批量获取表结构信息",
				zap.Int("table_count", len(tables)),
				zap.Int("column_count", len(columns)))
		}
	}
	return metadata, nil
}

// FindDatasourceByDatabaseAndTable 根据数据库名和表名从元数据表查找数据源ID
func FindDatasourceByDatabaseAndTable(databaseName, tableName string) (int, error) {
	var host, port string

	if tableName != "" && databaseName != "" {
		// 如果有表名，从meta_table表查找
		var metaTable model.MetaTable
		result := database.DB.Where("database_name = ? AND table_name = ? AND is_deleted = 0", databaseName, tableName).First(&metaTable)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return 0, fmt.Errorf("未找到数据库 %s 中表 %s 的元数据信息", databaseName, tableName)
			}
			return 0, fmt.Errorf("查询表元数据失败: %v", result.Error)
		}
		host = metaTable.Host
		port = metaTable.Port
	} else if tableName != "" {
		// 如果只有表名，从meta_table表查找
		var metaTable model.MetaTable
		result := database.DB.Where("table_name = ? AND is_deleted = 0", tableName).First(&metaTable)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return 0, fmt.Errorf("未找到数据库 %s 的元数据信息", databaseName)
			}
			return 0, fmt.Errorf("查询数据库元数据失败: %v", result.Error)
		}
		host = metaTable.Host
		port = metaTable.Port
	} else {
		return 0, fmt.Errorf("数据源表名不能为空")
	}

	if host == "" || port == "" {
		return 0, fmt.Errorf("元数据中缺少host或port信息")
	}

	// 通过host和port在datasource表中查找数据源ID
	var datasource model.Datasource
	result := database.DB.Where("host = ? AND port = ? AND enable = 1", host, port).First(&datasource)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("未找到对应的数据源 (host: %s, port: %s)", host, port)
		}
		return 0, fmt.Errorf("查询数据源失败: %v", result.Error)
	}

	return datasource.Id, nil
}

// ExtractDatabaseAndTableFromQuery 从查询中提取数据库名和表名
func ExtractDatabaseAndTableFromQuery(question string) (databaseName, tableName string) {
	// 使用正则表达式提取数据库名和表名
	// 匹配模式：数据库名、表名等
	// \w+ 匹配字母、数字、下划线

	// 优先匹配"查询xx库xx表"或"查看xx库xx表"格式（同时包含库和表）
	// 注意：允许"查看"后无空格，允许"表"后有其他内容（如"的数据"）
	//优先提取库和表的格式，如果提取不到再单独提取数据表的格式

	dbTablePatterns := []string{
		`(?:查询|查看|查找|查一下|看一下|看下|在|从)\s*([\w\p{Han}]+)\s*(?:数据库|数据库)\s*([\w\p{Han}]+)\s*(?:数据表|表)`, // 查询xx库xx表 或 查看xx库xx表（允许"查看"后无空格，允许"库"和表名之间无空格）
		`([\w\p{Han}]+)\s*(?:数据库|库)\s*([\w\p{Han}]+)\s*(?:数据表|表)`,                                 // xx库xx表（允许"库"和表名之间无空格）
	}
	for _, pattern := range dbTablePatterns {
		re := regexp.MustCompile("(?i)" + pattern)
		matches := re.FindStringSubmatch(question)
		if len(matches) >= 3 && matches[1] != "" && matches[2] != "" {
			// 去除库名和表名前后的空格
			databaseName = strings.TrimSpace(matches[1])
			tableName = strings.TrimSpace(matches[2])
			if databaseName != "" && tableName != "" {
				return databaseName, tableName
			}
		}
	}

	// // 提取数据库名（优先匹配更具体的模式）
	// dbPatterns := []string{
	// 	`(?:查询|查看|查找|查一下|看一下|看下)\s*(\w+)\s*(?:库|数据库)`,           // 查询xx库、查看xx库、查找xx库、查一下xx库、看一下xx表、看下xx表
	// 	`(?:在|从)\s+(\w+)\s*(?:库|数据库)`,                           // 在xxx表、从xxx表
	// 	`(?:查询|查看|显示|查找)\s+(?:库|数据库)[:：]?['"]?\s*(\w+)\s*['"]?`, // 查询库xxx
	// }

	// for _, pattern := range dbPatterns {
	// 	re := regexp.MustCompile("(?i)" + pattern)
	// 	matches := re.FindStringSubmatch(question)
	// 	if len(matches) >= 2 && matches[1] != "" {
	// 		// 去除库名前后的空格
	// 		databaseName = strings.TrimSpace(matches[1])
	// 		if databaseName != "" {
	// 			break
	// 		}
	// 	}
	// }

	// 提取表名（优先匹配更具体的模式）
	// 使用 [\w\p{Han}]+ 来匹配字母、数字、下划线和中文字符（\p{Han} 匹配汉字）
	tablePatterns := []string{
		`(?:库|数据库)\s*(?:里|的)?\s*([\w\p{Han}]+)\s*(?:数据表|表)`,                   // 库user表、数据库user表、数据库里user表、库里user表、库的user表、库邀请码表
		`(?:查询|查看|查找|分析|查一下|看一下|看下|在|从)\s*([\w\p{Han}]+)\s*(?:数据表|表|的数据|数据集)`, // 查询xx表、查看xx表、查找xx表、查一下xx表、看一下xx表、看下xx表、在xxx表、从xxx表、查询邀请码表
	}

	for _, pattern := range tablePatterns {
		re := regexp.MustCompile("(?i)" + pattern)
		matches := re.FindStringSubmatch(question)
		if len(matches) >= 2 && matches[1] != "" {
			// 去除表名前后的空格
			extractedTableName := strings.TrimSpace(matches[1])
			// 避免将数据库名误识别为表名
			if extractedTableName != "" && extractedTableName != databaseName {
				tableName = extractedTableName
				break
			}
		}
	}

	// 如果还没有找到，尝试从"数据库.表"格式中提取
	if databaseName == "" && tableName == "" {
		dbTablePattern := regexp.MustCompile(`(?i)([\w\p{Han}]+)\.([\w\p{Han}]+)`)
		matches := dbTablePattern.FindStringSubmatch(question)
		if len(matches) >= 3 {
			databaseName = matches[1]
			tableName = matches[2]
		}
	}

	return databaseName, tableName
}

// ExtractDatabaseAndTableFromSQL 从SQL语句中提取数据库名和表名
func ExtractDatabaseAndTableFromSQL(sqlQuery string) (databaseName, tableName string) {
	// 移除注释和多余空格
	sql := strings.TrimSpace(sqlQuery)
	sql = regexp.MustCompile(`--.*`).ReplaceAllString(sql, "")
	sql = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(sql, "")
	sql = strings.TrimSpace(sql)

	// 提取数据库名和表名（格式：database.table）
	dbTablePattern := regexp.MustCompile(`(?i)from\s+(\w+)\.(\w+)`)
	matches := dbTablePattern.FindStringSubmatch(sql)
	if len(matches) >= 3 {
		databaseName = matches[1]
		tableName = matches[2]
		return
	}

	// 提取表名（格式：from table）
	tablePattern := regexp.MustCompile(`(?i)from\s+(\w+)`)
	matches = tablePattern.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		tableName = matches[1]
	}

	// 提取join中的表名
	joinPattern := regexp.MustCompile(`(?i)join\s+(?:\w+\.)?(\w+)`)
	matches = joinPattern.FindStringSubmatch(sql)
	if len(matches) >= 2 && tableName == "" {
		tableName = matches[1]
	}

	return databaseName, tableName
}

// GetLocalDatabaseName 获取本地MySQL数据库名（从配置文件）
func GetLocalDatabaseName() string {
	if setting.Setting.Database != "" {
		return setting.Setting.Database
	}
	// 如果配置文件中没有，返回空字符串
	return ""
}

// BuildQueryContext 构建查询上下文
func BuildQueryContext(sessionId string, databaseName, tableName string) (*QueryContext, error) {

	// 获取元数据信息（如果提供了表名，确保获取完整的表结构）
	metadata, err := GetMetadataInfo(databaseName, tableName)
	if err != nil {
		zap.L().Warn("获取元数据信息失败", zap.Error(err))
		metadata = &MetadataInfo{}
	} else {
		// 如果提供了表名，验证表结构是否已获取
		if tableName != "" && databaseName != "" {
			if len(metadata.Columns) == 0 {
				zap.L().Warn("表结构信息为空，尝试重新获取",
					zap.String("database", databaseName),
					zap.String("table", tableName))
				// 尝试重新获取表结构
				retryMetadata, retryErr := GetMetadataInfo(databaseName, tableName)
				if retryErr == nil && len(retryMetadata.Columns) > 0 {
					metadata = retryMetadata
					zap.L().Info("重新获取表结构成功",
						zap.String("database", databaseName),
						zap.String("table", tableName),
						zap.Int("column_count", len(retryMetadata.Columns)))
				}
			} else {
				zap.L().Info("表结构信息已获取",
					zap.String("database", databaseName),
					zap.String("table", tableName),
					zap.Int("column_count", len(metadata.Columns)))
			}
		}
	}

	// 获取历史消息（最近5条）
	history, _ := GetSessionMessages(sessionId, 5)

	context := &QueryContext{
		DatabaseName: databaseName,
		TableName:    tableName,
		Metadata:     metadata,
		History:      history,
	}

	return context, nil
}

// AISQLGenerator AI SQL生成器
type AISQLGenerator struct{}

// GenerateSQL 使用AI生成SQL
func (g *AISQLGenerator) GenerateSQL(userQuery string, context *QueryContext) (string, error) {
	prompt := g.BuildPrompt(userQuery, context, nil)
	return g.callDeepSeekAPI(prompt)
}

// GenerateSQLWithRule 使用AI生成SQL，结合规则SQL模板作为参考
func (g *AISQLGenerator) GenerateSQLWithRule(userQuery string, context *QueryContext, rule *model.SemanticSqlRule, ruleParams map[string]string) (string, error) {
	// 生成规则SQL作为参考
	ruleSQL := GenerateSQLFromRule(rule, ruleParams, context)
	prompt := g.BuildPrompt(userQuery, context, &ruleSQL)
	return g.callDeepSeekAPI(prompt)
}

// BuildPrompt 构建包含元数据信息的Prompt
// ruleSQLTemplate: 规则SQL模板（可选），作为参考信息
func (g *AISQLGenerator) BuildPrompt(userQuery string, context *QueryContext, ruleSQLTemplate *string) string {
	var prompt strings.Builder

	// 系统提示
	prompt.WriteString("你是一个专业的SQL生成助手，专门帮助用户将自然语言转换为SQL查询语句。\n\n")
	prompt.WriteString("## 重要规则：\n")
	prompt.WriteString("1. 只能生成SELECT查询语句，禁止生成INSERT、UPDATE、DELETE、DROP等修改数据的语句\n")
	prompt.WriteString("2. SQL语句必须符合数据库语法规范\n")
	prompt.WriteString("3. 如果用户查询涉及表或字段，必须使用实际存在的表名和字段名\n")
	prompt.WriteString("4. 生成的SQL应该简洁、高效\n")
	prompt.WriteString("5. 只返回SQL语句，不要包含其他解释性文字\n\n")

	// 数据源信息
	if context != nil {
		prompt.WriteString("## 数据源信息：\n")
		prompt.WriteString(fmt.Sprintf("- 数据源类型: %s\n", context.DatasourceType))
		prompt.WriteString(fmt.Sprintf("- 主机: %s\n", context.Host))
		prompt.WriteString(fmt.Sprintf("- 端口: %s\n", context.Port))
		if context.DatabaseName != "" {
			prompt.WriteString(fmt.Sprintf("- 数据库: %s\n", context.DatabaseName))
		}
		if context.TableName != "" {
			prompt.WriteString(fmt.Sprintf("- 表: %s\n", context.TableName))
		}
		prompt.WriteString("\n")

		// 元数据信息
		if context.Metadata != nil {
			if len(context.Metadata.Databases) > 0 {
				prompt.WriteString("## 可用数据库列表：\n")
				for _, db := range context.Metadata.Databases {
					prompt.WriteString(fmt.Sprintf("- %s\n", db["database_name"]))
				}
				prompt.WriteString("\n")
			}

			// 如果指定了表名，优先显示该表的详细信息
			if context.TableName != "" && len(context.Metadata.Columns) > 0 {
				// 查找表注释
				var tableComment string
				if len(context.Metadata.Tables) > 0 {
					for _, table := range context.Metadata.Tables {
						if tableName, ok := table["table_name"].(string); ok && tableName == context.TableName {
							if comment, ok := table["table_comment"].(string); ok && comment != "" {
								tableComment = comment
							}
							break
						}
					}
				}

				prompt.WriteString("## 目标表结构信息（重要！请仔细阅读字段注释）：\n")
				prompt.WriteString(fmt.Sprintf("### 表名: %s\n", context.TableName))
				if tableComment != "" {
					prompt.WriteString(fmt.Sprintf("### 表注释: %s\n", tableComment))
					prompt.WriteString("**注意：表注释说明了表的业务含义，请根据注释理解表的用途。**\n")
				}
				prompt.WriteString("\n### 字段列表（字段注释非常重要，请优先使用注释来理解字段的业务含义）：\n\n")

				// 统计有注释和没有注释的字段
				columnsWithComment := 0
				for _, col := range context.Metadata.Columns {
					if comment, ok := col["column_comment"].(string); ok && comment != "" {
						columnsWithComment++
					}
				}

				if columnsWithComment > 0 {
					prompt.WriteString(fmt.Sprintf("**共有 %d 个字段，其中 %d 个字段有注释。字段注释说明了字段的业务含义，请优先参考注释来理解字段用途。**\n\n",
						len(context.Metadata.Columns), columnsWithComment))
				}

				for _, col := range context.Metadata.Columns {
					colName := col["column_name"].(string)
					dataType := ""
					if dt, ok := col["data_type"].(string); ok {
						dataType = dt
					}
					colComment := ""
					if comment, ok := col["column_comment"].(string); ok {
						colComment = comment
					}
					isNullable := ""
					if nullable, ok := col["is_nullable"].(string); ok {
						isNullable = nullable
					}
					defaultValue := ""
					if defVal, ok := col["default_value"].(string); ok && defVal != "" {
						defaultValue = defVal
					}

					// 详细展示字段信息
					if colComment != "" {
						// 有注释的字段，突出显示注释
						prompt.WriteString(fmt.Sprintf("- **%s** (%s)", colName, dataType))
						if isNullable == "YES" {
							prompt.WriteString(" [可空]")
						}
						if defaultValue != "" {
							prompt.WriteString(fmt.Sprintf(" [默认值: %s]", defaultValue))
						}
						prompt.WriteString(fmt.Sprintf("\n  **注释（重要）**: %s\n", colComment))
					} else {
						// 没有注释的字段
						prompt.WriteString(fmt.Sprintf("- %s (%s)", colName, dataType))
						if isNullable == "YES" {
							prompt.WriteString(" [可空]")
						}
						if defaultValue != "" {
							prompt.WriteString(fmt.Sprintf(" [默认值: %s]", defaultValue))
						}
						prompt.WriteString("\n")
					}
				}
				prompt.WriteString("\n**重要提示：生成SQL时，请优先参考字段注释来理解字段的业务含义，确保生成的SQL符合业务逻辑。**\n\n")
			} else if len(context.Metadata.Tables) > 0 {
				// 如果没有指定表名，显示表列表
				prompt.WriteString("## 可用表列表：\n")
				for _, table := range context.Metadata.Tables {
					tableName := table["table_name"].(string)
					tableComment := ""
					if comment, ok := table["table_comment"].(string); ok {
						tableComment = comment
					}
					if tableComment != "" {
						prompt.WriteString(fmt.Sprintf("- %s - **注释**: %s\n", tableName, tableComment))
					} else {
						prompt.WriteString(fmt.Sprintf("- %s\n", tableName))
					}
				}
				prompt.WriteString("\n")
			}

			// 如果指定了表名但没有字段信息，提示需要表结构
			if context.TableName != "" && len(context.Metadata.Columns) == 0 {
				prompt.WriteString("## 警告：\n")
				prompt.WriteString(fmt.Sprintf("表 %s 的结构信息未获取到，可能影响SQL生成的准确性。\n\n", context.TableName))
			}
		}

		// 历史对话上下文（最近3轮）
		if len(context.History) > 0 {
			prompt.WriteString("## 历史对话上下文：\n")
			startIdx := 0
			if len(context.History) > 6 {
				startIdx = len(context.History) - 6
			}
			for i := startIdx; i < len(context.History); i++ {
				msg := context.History[i]
				if msg.Role == "user" {
					prompt.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
				} else {
					prompt.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
					if msg.SqlQuery != "" {
						prompt.WriteString(fmt.Sprintf("  (SQL: %s)\n", msg.SqlQuery))
					}
				}
			}
			prompt.WriteString("\n")
		}
	}

	// 规则SQL模板（如果提供，作为参考）
	if ruleSQLTemplate != nil && *ruleSQLTemplate != "" {
		prompt.WriteString("## 参考SQL模板（必须基于此模板生成最终SQL）：\n")
		prompt.WriteString("```sql\n")
		prompt.WriteString(*ruleSQLTemplate)
		prompt.WriteString("\n```\n\n")
		prompt.WriteString("**重要说明**：\n")
		prompt.WriteString("1. 上面的SQL模板是基础模板，你必须基于此模板生成最终SQL\n")
		prompt.WriteString("2. 必须保持规则SQL的核心结构（如FROM的表、基本查询逻辑）\n")
		prompt.WriteString("3. 可以根据用户的具体需求调整：\n")
		prompt.WriteString("   - SELECT字段（根据用户查询的具体字段需求）\n")
		prompt.WriteString("   - WHERE条件（根据用户的查询条件）\n")
		prompt.WriteString("   - GROUP BY分组（如果用户需要分组统计）\n")
		prompt.WriteString("   - ORDER BY排序（如果用户需要排序）\n")
		prompt.WriteString("   - LIMIT限制（如果用户需要限制结果数量）\n")
		prompt.WriteString("4. 生成的SQL必须可以直接执行，且必须是完整的单条SQL语句\n\n")
	}

	// 用户查询
	prompt.WriteString("## 用户查询：\n")
	prompt.WriteString(userQuery)
	prompt.WriteString("\n\n")

	if ruleSQLTemplate != nil && *ruleSQLTemplate != "" {
		prompt.WriteString("## 生成要求：\n")
		prompt.WriteString("请基于参考SQL模板，结合用户查询语义和表结构信息，生成一条完整的、可执行的SQL查询语句。\n\n")
		prompt.WriteString("**严格要求**：\n")
		prompt.WriteString("1. 必须基于参考SQL模板的结构生成，保持核心逻辑不变\n")
		prompt.WriteString("2. 只能返回一条完整的SQL语句，不能返回多条SQL（即使使用分号分隔）\n")
		prompt.WriteString("3. 生成的SQL必须可以直接执行，包含完整的SELECT、FROM等必要部分\n")
		prompt.WriteString("4. 只返回SQL语句本身，不要包含任何解释性文字、注释或其他内容\n")
		prompt.WriteString("5. 不要使用markdown代码块格式，直接返回SQL语句\n\n")
		prompt.WriteString("**示例格式**：\n")
		prompt.WriteString("```\n")
		prompt.WriteString("SELECT COUNT(*) FROM information_schema.SCHEMATA\n")
		prompt.WriteString("```\n")
	} else {
		prompt.WriteString("请根据以上信息生成一条完整的、可执行的SQL查询语句。\n")
		prompt.WriteString("**严格要求**：\n")
		prompt.WriteString("1. 只能返回一条SQL语句，不能返回多条SQL\n")
		prompt.WriteString("2. 生成的SQL必须可以直接执行\n")
		prompt.WriteString("3. 只返回SQL语句本身，不要包含任何解释性文字或其他内容\n")
		prompt.WriteString("4. 不要使用markdown代码块格式，直接返回SQL语句\n")
	}

	return prompt.String()
}

// callDeepSeekAPI 调用AI API生成SQL（使用新的模型服务，支持故障转移）
func (g *AISQLGenerator) callDeepSeekAPI(prompt string) (string, error) {
	// 使用新的模型服务
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := CallWithFailover(messages, nil)
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(response.Content)
	if content == "" {
		return "", fmt.Errorf("API返回的内容为空")
	}

	// 使用extractSingleSQL提取单条SQL
	sql := extractSingleSQL(content)
	if sql == "" {
		return "", fmt.Errorf("无法从AI返回的内容中提取有效的SQL语句")
	}

	// 验证SQL的完整性
	sqlLower := strings.ToLower(sql)
	if !strings.Contains(sqlLower, "select") {
		return "", fmt.Errorf("生成的SQL缺少SELECT关键字")
	}
	if !strings.Contains(sqlLower, "from") {
		return "", fmt.Errorf("生成的SQL缺少FROM关键字")
	}

	// 最终清理
	sql = strings.TrimSpace(sql)

	return sql, nil
}

// extractSingleSQL 从AI返回的内容中提取单条SQL语句
func extractSingleSQL(content string) string {
	// 移除markdown代码块标记
	sql := strings.TrimSpace(content)
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	// 如果包含多条SQL（用分号分隔），只提取第一条
	// 查找第一个分号（不在字符串中的分号）
	firstSemicolon := -1
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		// 检查是否在字符串中
		if !inString && (char == '\'' || char == '"' || char == '`') {
			inString = true
			stringChar = char
		} else if inString && char == stringChar {
			// 检查是否是转义的引号
			if i == 0 || sql[i-1] != '\\' {
				inString = false
			}
		} else if !inString && char == ';' {
			firstSemicolon = i
			break
		}
	}

	// 如果找到分号，只取分号之前的内容
	if firstSemicolon >= 0 {
		sql = sql[:firstSemicolon]
		sql = strings.TrimSpace(sql)
	}

	// 验证SQL的完整性（至少包含SELECT和FROM）
	sqlLower := strings.ToLower(sql)
	hasSelect := strings.Contains(sqlLower, "select")
	hasFrom := strings.Contains(sqlLower, "from")

	// 如果SQL不完整，尝试从内容中查找更完整的SQL
	if !hasSelect || !hasFrom {
		// 尝试使用正则表达式提取SQL语句
		sqlPattern := regexp.MustCompile(`(?i)select\s+.*?\s+from\s+[^;]+`)
		matches := sqlPattern.FindString(content)
		if matches != "" {
			sql = strings.TrimSpace(matches)
		}
	}

	return strings.TrimSpace(sql)
}

// cleanSQL 清理SQL语句
func cleanSQL(sql string) string {
	// 使用extractSingleSQL提取单条SQL
	sql = extractSingleSQL(sql)
	return sql
}
