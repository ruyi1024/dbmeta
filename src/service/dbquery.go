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
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DbQueryRequest 智能查数请求
type DbQueryRequest struct {
	Question       string
	Page           int
	PageSize       int
	ModelId        int    // SQLCoder模型ID（可选，默认使用启用的SQLCoder模型）
	DatasourceId   int    // 可选，指定数据源ID
	DatabaseName   string // 可选，指定数据库名
	DatasourceType string // 可选，指定数据源类型
	Host           string // 可选，指定主机
	Port           string // 可选，指定端口
	TableName      string // 可选，指定表名
}

// DbQueryResult 智能查数结果
type DbQueryResult struct {
	SQLQuery    string
	QueryResult []map[string]interface{}
	Total       int
	Page        int
	PageSize    int
}

// ProcessDbQuery 处理智能查数请求
func ProcessDbQuery(req *DbQueryRequest) (*DbQueryResult, error) {
	log.Logger.Info("开始处理智能查数请求",
		zap.String("question", req.Question),
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize))

	// 检查用户输入是否是SQL语句（以SELECT、SHOW、DESC开头，不区分大小写）
	questionTrimmed := strings.TrimSpace(req.Question)
	isSQL := strings.HasPrefix(strings.ToUpper(questionTrimmed), "SELECT") || strings.HasPrefix(strings.ToUpper(questionTrimmed), "SHOW") || strings.HasPrefix(strings.ToUpper(questionTrimmed), "DESC") || strings.HasPrefix(strings.ToUpper(questionTrimmed), "DESCRIBE")

	if isSQL {
		log.Logger.Info("检测到SQL语句，直接执行SQL查询")
		return processDirectSQL(req)
	}

	// 如果请求中提供了完整的数据库信息（database_name, datasource_type, host, port），则跳过数据库识别
	var databaseName, tableName string
	if req.DatabaseName != "" && req.DatasourceType != "" && req.Host != "" && req.Port != "" {
		log.Logger.Info("使用请求中提供的数据库信息，跳过数据库识别",
			zap.String("database", req.DatabaseName),
			zap.String("datasource_type", req.DatasourceType),
			zap.String("host", req.Host),
			zap.String("port", req.Port))
		databaseName = req.DatabaseName
		// 只识别表名，不识别数据库名
		_, tableName = ExtractDatabaseAndTableFromQuery(req.Question)
		log.Logger.Info("识别表名",
			zap.String("table", tableName))
	} else {
		// 1. 识别数据库和表
		databaseName, tableName = ExtractDatabaseAndTableFromQuery(req.Question)
		log.Logger.Info("识别数据库和表",
			zap.String("database", databaseName),
			zap.String("table", tableName))

		fmt.Println("databaseName", databaseName)
		fmt.Println("tableName", tableName)

		// 如果请求中指定了数据库名，优先使用
		if req.DatabaseName != "" {
			databaseName = req.DatabaseName
			log.Logger.Info("使用请求中指定的数据库名", zap.String("database", databaseName))
		}

		// 如果无法识别，尝试使用AI识别
		if tableName == "" {
			log.Logger.Info("正则识别失败，尝试使用AI识别")
			var err error
			aiDatabaseName, aiTableName, err := ExtractDatabaseAndTableWithAI(req.Question)
			if err != nil {
				log.Logger.Warn("AI识别也失败", zap.Error(err))
				// 继续使用空值，后续会提示用户
			} else {
				log.Logger.Info("AI识别成功",
					zap.String("database", aiDatabaseName),
					zap.String("table", aiTableName))
				if databaseName == "" {
					databaseName = aiDatabaseName
				}
				if tableName == "" {
					tableName = aiTableName
				}
			}
		}
	}
	fmt.Println("database---:", databaseName)
	fmt.Println("table---:", tableName)

	// 10. 获取元数据
	metadata, err := GetMetadataInfo(databaseName, tableName)
	fmt.Println("---------------metadata--------------")
	fmt.Println(metadata)
	if err != nil {
		return nil, fmt.Errorf("获取元数据失败: %v", err)
	}

	// 4. 构建查询上下文
	context, err := BuildQueryContext("", databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("构建查询上下文失败: %v", err)
	}
	// 设置元数据
	if context != nil {
		context.Metadata = metadata
	}
	fmt.Println("------------------xontenxt------------")
	fmt.Println("context", context)
	fmt.Println(context.Metadata)
	log.Logger.Info(formatMetadataResult(context.Metadata.Tables))
	log.Logger.Info(formatMetadataResult(context.Metadata.Columns))

	// 5. 获取SQLCoder模型
	var aiModel *model.AIModel
	if req.ModelId > 0 {
		// 使用指定的模型
		model, err := GetModelById(req.ModelId)
		if err != nil {
			return nil, fmt.Errorf("获取AI模型失败: %v", err)
		}
		aiModel = model
	} else {
		// 优先使用“智能生成 SQL”场景配置的默认模型
		defaultSQLModel, err := GetDefaultAIModelByScenario(model.AIModelScenarioSQLGeneration)
		if err != nil {
			log.Logger.Warn("读取智能生成SQL默认模型失败，回退到自动选择", zap.Error(err))
		} else if defaultSQLModel != nil && defaultSQLModel.Enabled == 1 {
			aiModel = defaultSQLModel
		}

		if aiModel == nil {
			// 查找启用的SQLCoder模型（模型名称包含sqlcoder）
			models, err := GetEnabledModels()
			if err != nil {
				return nil, fmt.Errorf("获取启用的模型失败: %v", err)
			}

			// 优先查找名称包含sqlcoder的模型
			for _, m := range models {
				if strings.Contains(strings.ToLower(m.Name), "sqlcoder") ||
					strings.Contains(strings.ToLower(m.ModelName), "sqlcoder") {
					aiModel = &m
					break
				}
			}

			// 如果没有找到SQLCoder模型，使用第一个启用的模型
			if aiModel == nil && len(models) > 0 {
				aiModel = &models[0]
				log.Logger.Warn("未找到SQLCoder模型，使用默认模型", zap.String("model", aiModel.Name))
			}
		}
	}

	if aiModel == nil {
		return nil, fmt.Errorf("未找到可用的AI模型")
	}

	fmt.Println("aiModel", aiModel)
	log.Logger.Info("使用AI模型", zap.String("model", aiModel.Name))

	// 6. 构建SQLCoder Prompt
	prompt := BuildSQLCoderPrompt(req.Question, context)
	log.Logger.Info("构建的Prompt", zap.String("prompt", prompt))

	// 7. 调用SQLCoder生成SQL
	client, err := NewAIClient(aiModel)
	if err != nil {
		return nil, fmt.Errorf("创建AI客户端失败: %v", err)
	}

	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := client.Chat(messages, &ChatOptions{
		Temperature: aiModel.Temperature,
		MaxTokens:   aiModel.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("调用AI模型失败: %v", err)
	}

	// 提取SQL（可能包含markdown代码块）
	sqlQuery := extractSQLFromResponse(response.Content)
	log.Logger.Info("生成的SQL", zap.String("sql", sqlQuery))
	fmt.Println("sqlQuery", sqlQuery)

	// 9. 验证SQL
	if err := ValidateSQL(sqlQuery); err != nil {
		return nil, fmt.Errorf("SQL验证失败: %v", err)
	}

	// 9. 如果SQL没有LIMIT，自动添加LIMIT 500
	sqlQuery = addLimitIfNotExists(sqlQuery, 500)
	log.Logger.Debug("SQL添加LIMIT后", zap.String("sql", sqlQuery))

	// 10. 从元数据中提取数据源信息并查找数据源ID
	// 如果请求中指定了数据源ID，优先使用
	var datasourceId int
	if req.DatasourceId > 0 {
		datasourceId = req.DatasourceId
		log.Logger.Info("使用请求中指定的数据源ID", zap.Int("datasource_id", datasourceId))
	} else if req.Host != "" && req.Port != "" {
		// 如果请求中提供了host和port，直接使用它们查找数据源
		var datasource model.Datasource
		result := database.DB.Where("host = ? AND port = ? AND enable = 1", req.Host, req.Port).First(&datasource)
		if result.Error != nil {
			return nil, fmt.Errorf("根据host和port查找数据源失败: %v", result.Error)
		}
		datasourceId = datasource.Id
		log.Logger.Info("根据请求中的host和port找到数据源",
			zap.Int("datasource_id", datasourceId),
			zap.String("host", req.Host),
			zap.String("port", req.Port))
	} else {
		// 否则从元数据中查找
		var err error
		datasourceId, err = findDatasourceFromMetadata(context)
		if err != nil {
			return nil, fmt.Errorf("查找数据源失败: %v", err)
		}
		log.Logger.Info("从元数据中找到数据源", zap.Int("datasource_id", datasourceId))
	}

	// 11. 执行SQL查询（分页）
	queryResult, total, err := ExecuteQueryWithPagination(sqlQuery, datasourceId, req.Page, req.PageSize, req.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("执行SQL查询失败: %v", err)
	}

	log.Logger.Info("查询执行成功",
		zap.Int("total", total),
		zap.Int("result_count", len(queryResult)))

	return &DbQueryResult{
		SQLQuery:    sqlQuery,
		QueryResult: queryResult,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// ExtractDatabaseAndTableWithAI 使用AI识别数据库和表
func ExtractDatabaseAndTableWithAI(question string) (databaseName, tableName string, err error) {
	// 构建AI识别Prompt
	prompt := fmt.Sprintf(`分析以下问题，提取数据库名和表名。返回JSON格式：
{"database": "数据库名或空字符串", "table": "表名或空字符串", "confidence": 0.0-1.0}
问题：%s
只返回JSON，不要包含其他文字，表名和库名不需要出现库和表等字样。`, question)

	// 使用默认模型进行识别
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := CallWithFailover(messages, &ChatOptions{
		Temperature: 0.3, // 使用较低温度以获得更确定性的结果
		MaxTokens:   1000,
	})
	if err != nil {
		return "", "", fmt.Errorf("AI识别失败: %v", err)
	}

	// 解析JSON
	var result struct {
		Database   string  `json:"database"`
		Table      string  `json:"table"`
		Confidence float64 `json:"confidence"`
	}

	content := strings.TrimSpace(response.Content)
	// 移除可能的markdown代码块
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	// 尝试解析JSON
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Logger.Warn("解析AI返回的JSON失败", zap.String("content", response.Content), zap.Error(err))
		return "", "", fmt.Errorf("解析AI返回结果失败: %v", err)
	}

	// 检查置信度
	if result.Confidence < 0.5 {
		log.Logger.Info("AI识别置信度较低", zap.Float64("confidence", result.Confidence))
	}

	return result.Database, result.Table, nil
}

// extractSQLFromResponse 从AI响应中提取SQL
func extractSQLFromResponse(response string) string {
	response = strings.TrimSpace(response)

	// 移除markdown代码块
	if strings.HasPrefix(response, "```sql") {
		response = strings.TrimPrefix(response, "```sql")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	// 移除可能的解释文字（查找第一个SELECT、SHOW等SQL关键字）
	sqlKeywords := []string{"SELECT", "SHOW", "WITH", "DESC", "DESCRIBE"}
	for _, keyword := range sqlKeywords {
		idx := strings.Index(strings.ToUpper(response), keyword)
		if idx >= 0 {
			response = response[idx:]
			break
		}
	}

	// 移除末尾的非SQL字符（如解释文字）
	// 查找可能的SQL结束位置（分号或换行后的非SQL内容）
	lines := strings.Split(response, "\n")
	var sqlLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 如果遇到明显的非SQL内容（如"解释："、"说明："等），停止
		upperLine := strings.ToUpper(line)
		if strings.Contains(upperLine, "解释") || strings.Contains(upperLine, "说明") ||
			strings.Contains(upperLine, "EXPLANATION") || strings.Contains(upperLine, "NOTE") {
			break
		}
		sqlLines = append(sqlLines, line)
	}

	response = strings.Join(sqlLines, "\n")
	response = strings.TrimSpace(response)

	// 清理SQL中的中文字符（在表名和字段名位置）
	response = cleanSQLFromChinese(response)

	return response
}

// cleanSQLFromChinese 清理SQL中的中文字符，确保表名和字段名只包含英文、数字和下划线
func cleanSQLFromChinese(sql string) string {
	// 使用正则表达式匹配并清理SQL中的中文表名和字段名
	// 匹配 FROM 或 JOIN 后的表名（可能包含中文）
	// 匹配 SELECT 后的字段名（可能包含中文）

	// 匹配 FROM table_name 或 FROM database.table_name 中的中文
	re := regexp.MustCompile(`(?i)(FROM|JOIN)\s+([\w.]+)?[\p{Han}]+([\w.]*)`)
	sql = re.ReplaceAllStringFunc(sql, func(match string) string {
		// 移除匹配中的中文字符
		reChinese := regexp.MustCompile(`[\p{Han}]+`)
		return reChinese.ReplaceAllString(match, "")
	})

	// 匹配 SELECT 字段列表中的中文
	re = regexp.MustCompile(`(?i)SELECT\s+([\w\s,.*]+[\p{Han}]+[\w\s,.*]*)`)
	sql = re.ReplaceAllStringFunc(sql, func(match string) string {
		// 移除匹配中的中文字符
		reChinese := regexp.MustCompile(`[\p{Han}]+`)
		return reChinese.ReplaceAllString(match, "")
	})

	// 清理多余的空格
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(sql, " ")
	sql = strings.TrimSpace(sql)

	return sql
}

// processDirectSQL 处理直接的SQL查询（不通过AI转换）
func processDirectSQL(req *DbQueryRequest) (*DbQueryResult, error) {
	sqlQuery := req.Question
	log.Logger.Info("处理直接SQL查询", zap.String("sql", sqlQuery))

	// 1. 查找数据源
	var datasourceId int
	var err error

	// 如果请求中已经提供了host、port和database_name，直接使用这些信息查找数据源
	if req.Host != "" && req.Port != "" && req.DatabaseName != "" {
		log.Logger.Info("使用请求中提供的数据库信息查找数据源",
			zap.String("host", req.Host),
			zap.String("port", req.Port),
			zap.String("database", req.DatabaseName))

		// 根据host和port查找数据源ID
		var datasource model.Datasource
		result := database.DB.Where("host = ? AND port = ? AND enable = 1", req.Host, req.Port).First(&datasource)
		if result.Error != nil {
			return nil, fmt.Errorf("根据host和port查找数据源失败: %v", result.Error)
		}
		datasourceId = datasource.Id
		log.Logger.Info("找到数据源", zap.Int("datasource_id", datasourceId))
	} else {
		return nil, fmt.Errorf("请先选择数据库，或使用 SELECT * FROM 数据库名.表名 格式")
	}

	// 2. 验证SQL
	if err := ValidateSQL(sqlQuery); err != nil {
		return nil, fmt.Errorf("SQL验证失败: %v", err)
	}

	// 3. 如果SQL没有LIMIT，自动添加LIMIT 500
	// 仅在以SELECT开头且不包含LIMIT时自动添加LIMIT 500
	sqlQueryTrimUpper := strings.ToUpper(strings.TrimSpace(sqlQuery))
	if strings.HasPrefix(sqlQueryTrimUpper, "SELECT") && !strings.Contains(sqlQueryTrimUpper, "LIMIT") {
		sqlQuery = addLimitIfNotExists(sqlQuery, 500)
		log.Logger.Debug("SQL添加LIMIT后", zap.String("sql", sqlQuery))
	}

	// 4. 执行SQL查询（分页）
	queryResult, total, err := ExecuteQueryWithPagination(sqlQuery, datasourceId, req.Page, req.PageSize, req.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("执行SQL查询失败: %v", err)
	}

	log.Logger.Info("SQL查询执行成功",
		zap.Int("total", total),
		zap.Int("result_count", len(queryResult)))

	return &DbQueryResult{
		SQLQuery:    sqlQuery,
		QueryResult: queryResult,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

// FindDatasourcesByTableName 根据表名查找所有包含该表的数据源ID，并返回数据源ID到数据库名的映射
func FindDatasourcesByTableName(tableName string) ([]int, map[int]string, error) {
	// 从meta_table表查找所有包含该表名的记录
	var metaTables []model.MetaTable
	result := database.DB.Where("table_name = ? AND is_deleted = 0", tableName).Find(&metaTables)
	if result.Error != nil {
		return nil, nil, fmt.Errorf("查询表元数据失败: %v", result.Error)
	}

	if len(metaTables) == 0 {
		return []int{}, make(map[int]string), nil
	}

	// 收集所有唯一的 (host, port) 组合
	hostPortMap := make(map[string]bool)
	var datasourceIds []int
	dbTableMap := make(map[int]string) // 数据源ID -> 数据库名

	for _, metaTable := range metaTables {
		if metaTable.Host == "" || metaTable.Port == "" {
			continue
		}

		hostPortKey := fmt.Sprintf("%s:%s", metaTable.Host, metaTable.Port)
		if hostPortMap[hostPortKey] {
			continue // 已经处理过这个数据源
		}
		hostPortMap[hostPortKey] = true

		// 通过host和port在datasource表中查找数据源ID
		var datasource model.Datasource
		result := database.DB.Where("host = ? AND port = ? AND enable = 1", metaTable.Host, metaTable.Port).First(&datasource)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				log.Logger.Warn("未找到对应的数据源", zap.String("host", metaTable.Host), zap.String("port", metaTable.Port))
				continue
			}
			log.Logger.Warn("查询数据源失败", zap.String("host", metaTable.Host), zap.String("port", metaTable.Port), zap.Error(result.Error))
			continue
		}

		datasourceIds = append(datasourceIds, datasource.Id)
		// 保存数据库名映射
		if metaTable.DatabaseName != "" {
			dbTableMap[datasource.Id] = metaTable.DatabaseName
		}
	}

	return datasourceIds, dbTableMap, nil
}

// addDatabaseToSQL 在SQL语句中将 FROM table 替换为 FROM database.table
func addDatabaseToSQL(sqlQuery, databaseName, tableName string) string {
	// 使用正则表达式替换 FROM table 为 FROM database.table
	// 匹配模式：FROM table 或 from table（不区分大小写）
	// 需要确保只替换第一个匹配的，并且table名要匹配

	// 先尝试匹配 FROM table（后面可能有空格、换行、WHERE等）
	pattern := fmt.Sprintf(`(?i)(FROM\s+)%s(\s|$|WHERE|JOIN|GROUP|ORDER|LIMIT|;|\n)`, regexp.QuoteMeta(tableName))
	replacement := fmt.Sprintf("${1}%s.%s${2}", databaseName, tableName)

	re := regexp.MustCompile(pattern)
	result := re.ReplaceAllString(sqlQuery, replacement)

	// 如果替换失败（可能是表名在引号中或其他格式），尝试更宽松的匹配
	if result == sqlQuery {
		// 尝试匹配 FROM `table` 或 FROM 'table'
		quotePattern := "`'\""
		pattern2 := fmt.Sprintf(`(?i)(FROM\s+[%s])%s([%s])`, quotePattern, regexp.QuoteMeta(tableName), quotePattern)
		replacement2 := fmt.Sprintf("${1}%s.${2}%s${2}", databaseName, tableName)
		re2 := regexp.MustCompile(pattern2)
		result = re2.ReplaceAllString(sqlQuery, replacement2)
	}

	return result
}

// addLimitIfNotExists 如果SQL中没有LIMIT，添加LIMIT限制
func addLimitIfNotExists(sqlQuery string, maxRows int) string {
	sqlLower := strings.ToLower(strings.TrimSpace(sqlQuery))

	// 对SHOW语句不自动添加LIMIT
	if strings.HasPrefix(sqlLower, "show") {
		return sqlQuery
	}

	// 检查是否已有LIMIT
	if strings.Contains(sqlLower, " limit ") {
		return sqlQuery
	}

	// 移除末尾的分号（如果有）
	sqlQuery = strings.TrimRight(sqlQuery, ";")
	sqlQuery = strings.TrimSpace(sqlQuery)

	// 检查是否有ORDER BY（如果有，LIMIT应该加在ORDER BY之后）
	if strings.Contains(sqlLower, " order by ") {
		// 在ORDER BY之后添加LIMIT
		re := regexp.MustCompile(`(?i)(order\s+by\s+[^;]+)`)
		if re.MatchString(sqlQuery) {
			return re.ReplaceAllString(sqlQuery, fmt.Sprintf("$1 LIMIT %d", maxRows))
		}
	}

	// 如果没有ORDER BY，在末尾添加LIMIT
	return fmt.Sprintf("%s LIMIT %d", sqlQuery, maxRows)
}

// validateSQLTableNames 验证SQL中的表名是否在元数据中存在
func validateSQLTableNames(sqlQuery string, context *QueryContext) error {
	if context == nil || context.Metadata == nil {
		return nil // 如果没有元数据，跳过验证
	}

	// 从SQL中提取表名
	sqlDbName, sqlTableName := ExtractDatabaseAndTableFromSQL(sqlQuery)

	// 如果SQL中没有表名，无法验证
	if sqlTableName == "" {
		return nil
	}

	// 收集所有可用的表名
	availableTables := make(map[string]bool)
	availableDbTables := make(map[string]string) // tableName -> databaseName

	if len(context.Metadata.Tables) > 0 {
		for _, table := range context.Metadata.Tables {
			tableName, _ := table["table_name"].(string)
			dbName, _ := table["database_name"].(string)
			if tableName != "" {
				availableTables[tableName] = true
				if dbName != "" {
					availableDbTables[tableName] = dbName
				}
			}
		}
	}

	// 验证表名是否存在（不区分大小写，并检查是否包含中文）
	// 检查表名是否包含中文字符
	reChinese := regexp.MustCompile(`[\p{Han}]+`)
	if reChinese.MatchString(sqlTableName) {
		var availableTableList []string
		for tableName := range availableTables {
			availableTableList = append(availableTableList, tableName)
		}
		return fmt.Errorf("生成的SQL中使用的表名 '%s' 包含中文字符，这是不允许的。可用表名: %v。请使用元数据中提供的英文表名", sqlTableName, availableTableList)
	}

	sqlTableNameLower := strings.ToLower(sqlTableName)
	tableNameFound := false
	var matchedTableName string
	for tableName := range availableTables {
		if strings.ToLower(tableName) == sqlTableNameLower {
			tableNameFound = true
			matchedTableName = tableName
			break
		}
	}

	if !tableNameFound {
		// 表名不存在，构建错误信息
		var availableTableList []string
		for tableName := range availableTables {
			availableTableList = append(availableTableList, tableName)
		}
		return fmt.Errorf("生成的SQL中使用的表名 '%s' 不在元数据中。可用表名: %v。请确保使用元数据中提供的准确表名，不能使用中文表名", sqlTableName, availableTableList)
	}

	// 如果表名大小写不匹配，记录警告
	if matchedTableName != sqlTableName {
		log.Logger.Warn("SQL表名大小写不匹配",
			zap.String("sql_table", sqlTableName),
			zap.String("correct_table", matchedTableName))
	}

	// 如果SQL中指定了数据库名，验证是否匹配
	if sqlDbName != "" {
		expectedDbName, exists := availableDbTables[sqlTableName]
		if exists && expectedDbName != "" && expectedDbName != sqlDbName {
			log.Logger.Warn("SQL中的数据库名与元数据不匹配",
				zap.String("sql_db", sqlDbName),
				zap.String("expected_db", expectedDbName),
				zap.String("table", sqlTableName))
			// 不返回错误，只记录警告，因为可能是跨数据库查询
		}
	}

	return nil
}

// validateSQLColumnNames 验证SQL中的字段名是否在元数据中存在
func validateSQLColumnNames(sqlQuery string, context *QueryContext) error {
	if context == nil || context.Metadata == nil {
		return nil // 如果没有元数据，跳过验证
	}

	// 从SQL中提取表名
	_, sqlTableName := ExtractDatabaseAndTableFromSQL(sqlQuery)
	if sqlTableName == "" {
		return nil // 如果没有表名，无法验证字段
	}

	// 收集该表的所有可用字段名（不区分大小写）
	availableColumns := make(map[string]string) // lowercase -> actual name

	// 优先从新格式中获取字段（从 table["columns"] 中获取）
	for _, table := range context.Metadata.Tables {
		tableName, _ := table["table_name"].(string)
		if tableName == sqlTableName {
			if cols, ok := table["columns"].([]interface{}); ok {
				for _, colInterface := range cols {
					if col, ok := colInterface.(map[string]interface{}); ok {
						if colName, ok := col["column_name"].(string); ok && colName != "" {
							availableColumns[strings.ToLower(colName)] = colName
						}
					}
				}
			}
			break
		}
	}

	// 如果没有从新格式获取到，尝试从旧的 Columns 格式获取（向后兼容）
	if len(availableColumns) == 0 {
		for _, col := range context.Metadata.Columns {
			colTableName, _ := col["table_name"].(string)
			colName, _ := col["column_name"].(string)
			if colTableName == sqlTableName && colName != "" {
				availableColumns[strings.ToLower(colName)] = colName
			}
		}
	}

	if len(availableColumns) == 0 {
		// 如果没有字段信息，跳过验证
		return nil
	}

	// 从SQL中提取字段名（SELECT、WHERE、ORDER BY等）
	// 使用正则表达式提取字段名
	sqlUpper := strings.ToUpper(sqlQuery)

	// 提取SELECT子句中的字段（跳过SELECT *）
	if !strings.Contains(sqlUpper, "SELECT *") {
		selectPattern := regexp.MustCompile(`(?i)SELECT\s+([^FROM]+)FROM`)
		selectMatch := selectPattern.FindStringSubmatch(sqlQuery)
		if len(selectMatch) > 1 {
			selectFields := selectMatch[1]
			// 分割字段（考虑逗号、AS别名等）
			fieldPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
			fields := fieldPattern.FindAllString(selectFields, -1)
			for _, field := range fields {
				fieldLower := strings.ToLower(field)
				// 跳过SQL关键字
				if isSQLKeyword(fieldLower) {
					continue
				}
				if _, exists := availableColumns[fieldLower]; !exists {
					var columnList []string
					for _, colName := range availableColumns {
						columnList = append(columnList, colName)
					}
					return fmt.Errorf("生成的SQL中使用的字段名 '%s' 不在表 '%s' 的元数据中。可用字段: %v。请确保使用元数据中提供的准确字段名", field, sqlTableName, columnList)
				}
			}
		}
	}

	// 提取WHERE子句中的字段
	wherePattern := regexp.MustCompile(`(?i)WHERE\s+([^ORDER|GROUP|LIMIT]+)`)
	whereMatch := wherePattern.FindStringSubmatch(sqlQuery)
	if len(whereMatch) > 1 {
		whereClause := whereMatch[1]
		// 提取字段名（在比较操作符之前）
		fieldPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s*[=<>!]+`)
		fields := fieldPattern.FindAllStringSubmatch(whereClause, -1)
		for _, match := range fields {
			if len(match) > 1 {
				field := match[1]
				fieldLower := strings.ToLower(field)
				if isSQLKeyword(fieldLower) {
					continue
				}
				if _, exists := availableColumns[fieldLower]; !exists {
					var columnList []string
					for _, colName := range availableColumns {
						columnList = append(columnList, colName)
					}
					return fmt.Errorf("生成的SQL中WHERE子句使用的字段名 '%s' 不在表 '%s' 的元数据中。可用字段: %v。请确保使用元数据中提供的准确字段名", field, sqlTableName, columnList)
				}
			}
		}
	}

	// 提取ORDER BY子句中的字段
	orderPattern := regexp.MustCompile(`(?i)ORDER\s+BY\s+([^LIMIT]+)`)
	orderMatch := orderPattern.FindStringSubmatch(sqlQuery)
	if len(orderMatch) > 1 {
		orderClause := orderMatch[1]
		fieldPattern := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
		fields := fieldPattern.FindAllString(orderClause, -1)
		for _, field := range fields {
			fieldLower := strings.ToLower(field)
			if isSQLKeyword(fieldLower) || fieldLower == "desc" || fieldLower == "asc" {
				continue
			}
			if _, exists := availableColumns[fieldLower]; !exists {
				var columnList []string
				for _, colName := range availableColumns {
					columnList = append(columnList, colName)
				}
				return fmt.Errorf("生成的SQL中ORDER BY子句使用的字段名 '%s' 不在表 '%s' 的元数据中。可用字段: %v。请确保使用元数据中提供的准确字段名", field, sqlTableName, columnList)
			}
		}
	}

	return nil
}

// isSQLKeyword 检查是否是SQL关键字
func isSQLKeyword(word string) bool {
	keywords := []string{
		"select", "from", "where", "order", "by", "group", "having",
		"and", "or", "not", "in", "like", "between", "is", "null",
		"as", "distinct", "count", "sum", "avg", "max", "min",
		"case", "when", "then", "else", "end", "if", "exists",
		"join", "inner", "left", "right", "outer", "on", "union",
		"limit", "offset", "desc", "asc",
	}
	wordLower := strings.ToLower(word)
	for _, keyword := range keywords {
		if wordLower == keyword {
			return true
		}
	}
	return false
}

// findDatasourceFromMetadata 从元数据中提取数据源信息并查找数据源ID
func findDatasourceFromMetadata(context *QueryContext) (int, error) {
	if context == nil || context.Metadata == nil {
		return 0, fmt.Errorf("上下文或元数据为空")
	}

	// 从元数据中提取数据源信息
	datasourceInfo := extractDatasourceInfo(context)
	if datasourceInfo == nil {
		return 0, fmt.Errorf("无法从元数据中提取数据源信息")
	}

	host, _ := datasourceInfo["host"].(string)
	port, _ := datasourceInfo["port"].(string)

	if host == "" || port == "" {
		return 0, fmt.Errorf("元数据中缺少host或port信息")
	}

	// 根据host和port查找数据源ID
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

// ensureDatabaseTableFormat 确保SQL使用database.table格式
func ensureDatabaseTableFormat(sqlQuery string, context *QueryContext) (string, error) {
	if context == nil || context.Metadata == nil {
		return sqlQuery, nil
	}

	// 从SQL中提取数据库名和表名
	sqlDbName, sqlTableName := ExtractDatabaseAndTableFromSQL(sqlQuery)

	// 如果SQL中已经有数据库名，直接返回
	if sqlDbName != "" && sqlTableName != "" {
		return sqlQuery, nil
	}

	// 如果只有表名，需要从元数据中查找对应的数据库名
	if sqlTableName != "" {
		// 从元数据中查找表对应的数据库名
		var targetDbName string
		if len(context.Metadata.Tables) > 0 {
			for _, table := range context.Metadata.Tables {
				tableName, _ := table["table_name"].(string)
				dbName, _ := table["database_name"].(string)
				if tableName == sqlTableName && dbName != "" {
					targetDbName = dbName
					break
				}
			}
		}

		// 如果从表中没找到，尝试从context中获取
		if targetDbName == "" && context.DatabaseName != "" {
			targetDbName = context.DatabaseName
		}

		// 如果还是没找到，尝试从数据库中获取
		if targetDbName == "" && len(context.Metadata.Databases) > 0 {
			// 使用第一个数据库
			if dbName, ok := context.Metadata.Databases[0]["database_name"].(string); ok && dbName != "" {
				targetDbName = dbName
			}
		}

		if targetDbName == "" {
			return sqlQuery, fmt.Errorf("无法从元数据中确定表 %s 所属的数据库名", sqlTableName)
		}

		// 使用addDatabaseToSQL函数添加数据库名
		sqlQuery = addDatabaseToSQL(sqlQuery, targetDbName, sqlTableName)
		log.Logger.Info("为SQL添加数据库名",
			zap.String("database", targetDbName),
			zap.String("table", sqlTableName),
			zap.String("sql", sqlQuery))
	}

	return sqlQuery, nil
}
