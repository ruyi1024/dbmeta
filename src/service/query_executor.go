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
	"database/sql"
	"github.com/ruyi1024/dbmeta/setting"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/libary/db"
	"github.com/ruyi1024/dbmeta/src/model"
	"github.com/ruyi1024/dbmeta/src/utils"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	MaxQueryRows = 1000 // 最大查询行数
)

// ValidateSQL SQL安全性验证
func ValidateSQL(sqlQuery string) error {
	// 移除注释和多余空格
	sqlQuery = strings.TrimSpace(sqlQuery)
	sqlQuery = regexp.MustCompile(`--.*`).ReplaceAllString(sqlQuery, "")
	sqlQuery = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(sqlQuery, "")
	sqlQuery = strings.TrimSpace(sqlQuery)

	if sqlQuery == "" {
		return fmt.Errorf("SQL语句不能为空")
	}

	// 转换为小写进行检测（不区分大小写）
	sqlLower := strings.ToLower(sqlQuery)

	// 检查是否以SHOW开头（SHOW语句都是安全的查询语句，即使包含CREATE等关键词）
	isShowStatement := strings.HasPrefix(sqlLower, "show")

	// 如果不是SHOW语句，检查禁止的危险操作
	if !isShowStatement {
		dangerousKeywords := []string{
			"drop", "delete", "truncate", "alter", "create",
			"insert", "update", "replace", "grant", "revoke",
			"flush", "lock", "unlock", "kill", "exec", "execute",
		}

		for _, keyword := range dangerousKeywords {
			// 使用单词边界匹配，避免误判
			pattern := fmt.Sprintf(`\b%s\b`, keyword)
			matched, _ := regexp.MatchString(pattern, sqlLower)
			if matched {
				return fmt.Errorf("禁止执行 %s 操作，仅允许SELECT查询", strings.ToUpper(keyword))
			}
		}
	}

	// 必须是以SELECT或SHOW开头
	if !(strings.HasPrefix(sqlLower, "select") || strings.HasPrefix(sqlLower, "show")) {
		return fmt.Errorf("仅允许执行SELECT或SHOW查询语句")
	}

	// 检查是否有子查询中包含危险操作
	if strings.Contains(sqlLower, "union") && (strings.Contains(sqlLower, "drop") || strings.Contains(sqlLower, "delete")) {
		return fmt.Errorf("检测到潜在的SQL注入风险")
	}

	return nil
}

// ExecuteQuery 执行查询（支持可选的数据库名）
func ExecuteQuery(sqlQuery string, datasourceId int, databaseName ...string) ([]map[string]interface{}, error) {
	// 验证SQL
	if err := ValidateSQL(sqlQuery); err != nil {
		return nil, err
	}

	// 查询数据源信息
	var datasource model.Datasource
	result := database.DB.Where("id = ? AND enable = 1", datasourceId).First(&datasource)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("数据源不存在或已禁用")
		}
		return nil, fmt.Errorf("查询数据源失败: %v", result.Error)
	}

	// 解密密码
	var origPass string
	if datasource.Pass != "" {
		var err error
		origPass, err = utils.AesPassDecode(datasource.Pass, setting.Setting.DbPassKey)
		if err != nil {
			return nil, fmt.Errorf("密码解密失败: %v", err)
		}
	}

	// 获取数据库名（如果提供）
	var dbName string
	if len(databaseName) > 0 && databaseName[0] != "" {
		dbName = databaseName[0]
	}

	// 连接数据库
	dbConn, err := connectToDatabase(&datasource, origPass, dbName)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	// 如果提供了数据库名，执行 USE database 语句（对于MySQL等需要显式选择数据库的情况）
	if dbName != "" && (strings.ToUpper(datasource.Type) == "MYSQL" || strings.ToUpper(datasource.Type) == "MARIADB" || strings.ToUpper(datasource.Type) == "GREATSQL" || strings.ToUpper(datasource.Type) == "TIDB" || strings.ToUpper(datasource.Type) == "DORIS" || strings.ToUpper(datasource.Type) == "OCEANBASE") {
		useSQL := fmt.Sprintf("USE `%s`", dbName)
		_, err = dbConn.Exec(useSQL)
		if err != nil {
			return nil, fmt.Errorf("选择数据库失败: %v", err)
		}
	}

	// 添加LIMIT限制（如果SQL中没有LIMIT）
	sqlQuery = addLimitIfNeeded(sqlQuery, MaxQueryRows)

	// 执行查询
	dataList, err := database.QueryRemote(dbConn, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("执行SQL查询失败: %v", err)
	}

	// 限制返回行数
	if len(dataList) > MaxQueryRows {
		dataList = dataList[:MaxQueryRows]
	}

	return dataList, nil
}

// ExecuteLocalQuery 执行本地MySQL查询
func ExecuteLocalQuery(sqlQuery string) ([]map[string]interface{}, error) {
	// 验证SQL（使用0作为datasourceId，因为本地查询不需要数据源ID）
	if err := ValidateSQL(sqlQuery); err != nil {
		return nil, err
	}

	// 检查本地MySQL连接是否可用
	if database.SQL == nil {
		return nil, fmt.Errorf("本地MySQL连接未初始化")
	}

	// 添加LIMIT限制（如果SQL中没有LIMIT）
	sqlQuery = addLimitIfNeeded(sqlQuery, MaxQueryRows)

	// 使用本地MySQL连接执行查询
	dataList, err := database.QueryAll(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("执行本地MySQL查询失败: %v", err)
	}

	// 限制返回行数
	if len(dataList) > MaxQueryRows {
		dataList = dataList[:MaxQueryRows]
	}

	return dataList, nil
}

// ExecuteQueryWithPagination 执行查询并支持分页
func ExecuteQueryWithPagination(sqlQuery string, datasourceId int, page, pageSize int, databaseName ...string) ([]map[string]interface{}, int, error) {
	// 验证SQL
	if err := ValidateSQL(sqlQuery); err != nil {
		return nil, 0, err
	}

	// 查询数据源信息
	var datasource model.Datasource
	result := database.DB.Where("id = ? AND enable = 1", datasourceId).First(&datasource)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, 0, fmt.Errorf("数据源不存在或已禁用")
		}
		return nil, 0, fmt.Errorf("查询数据源失败: %v", result.Error)
	}

	// 解密密码
	var origPass string
	if datasource.Pass != "" {
		var err error
		origPass, err = utils.AesPassDecode(datasource.Pass, setting.Setting.DbPassKey)
		if err != nil {
			return nil, 0, fmt.Errorf("密码解密失败: %v", err)
		}
	}

	// 获取数据库名（如果提供）
	var dbName string
	if len(databaseName) > 0 && databaseName[0] != "" {
		dbName = databaseName[0]
	}

	// 连接数据库
	dbConn, err := connectToDatabase(&datasource, origPass, dbName)
	if err != nil {
		return nil, 0, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer dbConn.Close()

	// 如果提供了数据库名，执行 USE database 语句（对于MySQL等需要显式选择数据库的情况）
	if dbName != "" && (strings.ToUpper(datasource.Type) == "MYSQL" || strings.ToUpper(datasource.Type) == "MARIADB" || strings.ToUpper(datasource.Type) == "GREATSQL" || strings.ToUpper(datasource.Type) == "TIDB" || strings.ToUpper(datasource.Type) == "DORIS" || strings.ToUpper(datasource.Type) == "OCEANBASE") {
		useSQL := fmt.Sprintf("USE `%s`", dbName)
		_, err = dbConn.Exec(useSQL)
		if err != nil {
			return nil, 0, fmt.Errorf("选择数据库失败: %v", err)
		}
	}

	// 获取总数
	total, err := getTotalCount(sqlQuery, dbConn, datasource.Type)
	if err != nil {
		return nil, 0, fmt.Errorf("获取总数失败: %v", err)
	}

	// 添加分页
	paginatedSQL := addPaginationToSQL(sqlQuery, datasource.Type, page, pageSize)

	// 执行查询
	dataList, err := database.QueryRemote(dbConn, paginatedSQL)
	if err != nil {
		return nil, 0, fmt.Errorf("执行SQL查询失败: %v", err)
	}

	return dataList, total, nil
}

// getTotalCount 获取查询结果总数
func getTotalCount(sqlQuery string, dbConn *sql.DB, datasourceType string) (int, error) {
	sqlLower := strings.ToLower(strings.TrimSpace(sqlQuery))

	// 对于SHOW语句，直接执行查询并返回结果行数
	if strings.HasPrefix(sqlLower, "show") {
		rows, err := database.QueryRemote(dbConn, sqlQuery)
		if err != nil {
			return 0, fmt.Errorf("执行SHOW语句失败: %v", err)
		}
		return len(rows), nil
	}

	// 对于DESC/DESCRIBE语句，也直接执行查询并返回结果行数
	if strings.HasPrefix(sqlLower, "desc") || strings.HasPrefix(sqlLower, "describe") {
		rows, err := database.QueryRemote(dbConn, sqlQuery)
		if err != nil {
			return 0, fmt.Errorf("执行DESC语句失败: %v", err)
		}
		return len(rows), nil
	}

	// 构建COUNT查询
	// 注意：某些数据库可能不支持子查询，需要根据数据库类型调整
	// 移除末尾的分号，避免子查询语法错误
	cleanSQL := strings.TrimRight(strings.TrimSpace(sqlQuery), ";")
	countSQL := fmt.Sprintf("SELECT COUNT(*) as total FROM (%s) AS subquery", cleanSQL)

	// 执行COUNT查询
	rows, err := database.QueryRemote(dbConn, countSQL)
	if err != nil {
		// 如果子查询失败，尝试简化查询（移除ORDER BY等）
		// 这是一个降级方案
		zap.L().Warn("COUNT子查询失败，尝试简化查询", zap.Error(err))
		// 可以尝试其他方法，但为了简单起见，先返回错误
		return 0, fmt.Errorf("获取总数失败: %v", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	// 提取总数（可能在不同的键名下）
	var totalValue interface{}
	var found bool

	// 尝试不同的键名
	for _, key := range []string{"total", "count(*)", "COUNT(*)"} {
		if val, ok := rows[0][key]; ok {
			totalValue = val
			found = true
			break
		}
	}

	// 如果没找到，使用第一个值
	if !found && len(rows[0]) > 0 {
		for _, val := range rows[0] {
			totalValue = val
			found = true
			break
		}
	}

	if !found {
		return 0, fmt.Errorf("无法从查询结果中提取总数")
	}

	// 转换总数
	switch v := totalValue.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case []byte:
		// 某些数据库可能返回[]byte，尝试转换为字符串再转数字
		return 0, fmt.Errorf("无法解析总数（类型为[]byte）: %v", totalValue)
	case string:
		// 某些数据库可能返回字符串
		var total int
		if _, err := fmt.Sscanf(v, "%d", &total); err == nil {
			return total, nil
		}
		return 0, fmt.Errorf("无法解析总数（类型为string）: %v", totalValue)
	default:
		return 0, fmt.Errorf("未知的总数类型: %T, 值: %v", totalValue, totalValue)
	}
}

// addPaginationToSQL 为SQL添加分页（LIMIT/OFFSET）
func addPaginationToSQL(sqlQuery string, datasourceType string, page, pageSize int) string {
	sqlLower := strings.ToLower(strings.TrimSpace(sqlQuery))

	// 对SHOW、DESC、DESCRIBE语句不添加分页
	if strings.HasPrefix(sqlLower, "show") || strings.HasPrefix(sqlLower, "desc") || strings.HasPrefix(sqlLower, "describe") {
		return sqlQuery
	}

	// 计算offset
	offset := (page - 1) * pageSize

	// 检查是否已有LIMIT/OFFSET
	if strings.Contains(sqlLower, " limit ") || strings.Contains(sqlLower, " offset ") {
		// 如果已有LIMIT/OFFSET，需要替换它们
		// 移除现有的LIMIT和OFFSET
		re := regexp.MustCompile(`(?i)\s+LIMIT\s+\d+(\s+OFFSET\s+\d+)?`)
		sqlQuery = re.ReplaceAllString(sqlQuery, "")
	}

	// 根据数据库类型添加分页语法
	upperType := strings.ToUpper(datasourceType)
	switch upperType {
	case "MYSQL", "MARIADB", "GREATSQL", "TIDB", "DORIS", "OCEANBASE", "POSTGRESQL":
		// MySQL和PostgreSQL使用 LIMIT ... OFFSET ... 语法
		// 移除末尾的分号
		sqlQuery = strings.TrimRight(sqlQuery, ";")
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlQuery, pageSize, offset)
	case "ORACLE":
		// Oracle使用 ROWNUM 或 FETCH FIRST ... ROWS ONLY
		// 这里使用FETCH FIRST语法（Oracle 12c+）
		sqlQuery = strings.TrimRight(sqlQuery, ";")
		return fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", sqlQuery, offset, pageSize)
	default:
		// 默认使用MySQL语法
		sqlQuery = strings.TrimRight(sqlQuery, ";")
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", sqlQuery, pageSize, offset)
	}
}

// SQLExecResult 执行结果封装
type SQLExecResult struct {
	SQL     string                   `json:"sql"`
	Rows    []map[string]interface{} `json:"rows"`
	Elapsed int64                    `json:"elapsed_ms,omitempty"`
	Err     string                   `json:"error,omitempty"`
}

// ExecuteSQLSet 执行多条SQL（顺序执行，后续可扩展并发与依赖）
// useLocal: true 使用本地MySQL；false 使用远程数据源（datasourceId必填）
func ExecuteSQLSet(sqlSet []model.SqlSetItem, datasourceId int, useLocal bool) ([]SQLExecResult, error) {
	results := make([]SQLExecResult, 0, len(sqlSet))
	for _, item := range sqlSet {
		sqlText := strings.TrimSpace(item.Sql)
		if sqlText == "" {
			continue
		}
		var (
			rows []map[string]interface{}
			err  error
		)
		if useLocal {
			rows, err = ExecuteLocalQuery(sqlText)
		} else {
			rows, err = ExecuteQuery(sqlText, datasourceId)
		}
		res := SQLExecResult{
			SQL:  sqlText,
			Rows: rows,
		}
		if err != nil {
			res.Err = err.Error()
			results = append(results, res)
			return results, fmt.Errorf("执行SQL失败: %v", err)
		}
		results = append(results, res)
	}
	return results, nil
}

// connectToDatabase 连接数据库
func connectToDatabase(datasource *model.Datasource, password string, databaseName ...string) (*sql.DB, error) {
	var dbName string
	if len(databaseName) > 0 {
		dbName = databaseName[0]
	}

	switch strings.ToUpper(datasource.Type) {
	case "MYSQL", "MARIADB", "GREATSQL", "TIDB", "DORIS", "OCEANBASE":
		opts := []db.Option{
			db.WithDriver("mysql"),
			db.WithHost(datasource.Host),
			db.WithPort(datasource.Port),
			db.WithUsername(datasource.User),
			db.WithPassword(password),
		}
		if dbName != "" {
			opts = append(opts, db.WithDatabase(dbName))
		}
		return db.Connect(opts...)
	case "POSTGRESQL":
		opts := []db.Option{
			db.WithDriver("postgres"),
			db.WithHost(datasource.Host),
			db.WithPort(datasource.Port),
			db.WithUsername(datasource.User),
			db.WithPassword(password),
		}
		if dbName != "" {
			opts = append(opts, db.WithDatabase(dbName))
		}
		return db.Connect(opts...)
	case "ORACLE":
		return db.Connect(
			db.WithDriver("godror"),
			db.WithHost(datasource.Host),
			db.WithPort(datasource.Port),
			db.WithUsername(datasource.User),
			db.WithPassword(password),
			db.WithSid(datasource.Dbid),
		)
	case "CLICKHOUSE":
		// ClickHouse连接需要特殊处理
		return nil, fmt.Errorf("ClickHouse连接暂未实现，请使用其他数据源")
	default:
		return nil, fmt.Errorf("不支持的数据源类型: %s", datasource.Type)
	}
}

// addLimitIfNeeded 如果SQL中没有LIMIT，添加LIMIT限制
func addLimitIfNeeded(sqlQuery string, maxRows int) string {
	sqlLower := strings.ToLower(sqlQuery)

	// 对SHOW语句不自动添加LIMIT
	if strings.HasPrefix(strings.TrimSpace(sqlLower), "show") {
		return sqlQuery
	}

	// 检查是否已有LIMIT
	if strings.Contains(sqlLower, " limit ") {
		return sqlQuery
	}

	// 检查是否有ORDER BY（如果有，LIMIT应该加在ORDER BY之后）
	if strings.Contains(sqlLower, " order by ") {
		// 在ORDER BY之后添加LIMIT
		re := regexp.MustCompile(`(?i)(order\s+by\s+[^;]+)`)
		if re.MatchString(sqlQuery) {
			return re.ReplaceAllString(sqlQuery, fmt.Sprintf("$1 LIMIT %d", maxRows))
		}
	}

	// 如果没有ORDER BY，在末尾添加LIMIT
	// 移除末尾的分号
	sqlQuery = strings.TrimRight(sqlQuery, ";")
	return fmt.Sprintf("%s LIMIT %d", sqlQuery, maxRows)
}

// FormatResult 格式化查询结果
func FormatResult(data []map[string]interface{}, queryType string) string {
	if len(data) == 0 {
		return "查询结果为空。"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("查询成功，共返回 %d 条记录。\n\n", len(data)))

	// 根据查询类型格式化
	switch queryType {
	case "status":
		result.WriteString(formatStatusResult(data))
	case "performance":
		result.WriteString(formatPerformanceResult(data))
	case "metadata":
		result.WriteString(formatMetadataResult(data))
	default:
		result.WriteString(formatDefaultResult(data))
	}

	return result.String()
}

// formatStatusResult 格式化状态查询结果
func formatStatusResult(data []map[string]interface{}) string {
	var result strings.Builder
	result.WriteString("**数据库状态信息：**\n\n")

	for i, row := range data {
		if i >= 10 { // 只显示前10条
			result.WriteString(fmt.Sprintf("\n... 还有 %d 条记录\n", len(data)-10))
			break
		}
		for key, value := range row {
			result.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// formatPerformanceResult 格式化性能查询结果
func formatPerformanceResult(data []map[string]interface{}) string {
	var result strings.Builder
	result.WriteString("**性能指标：**\n\n")

	for i, row := range data {
		if i >= 10 {
			result.WriteString(fmt.Sprintf("\n... 还有 %d 条记录\n", len(data)-10))
			break
		}
		for key, value := range row {
			result.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// formatMetadataResult 格式化元数据查询结果
func formatMetadataResult(data []map[string]interface{}) string {
	var result strings.Builder
	result.WriteString("**元数据信息：**\n\n")

	for i, row := range data {
		if i >= 20 {
			result.WriteString(fmt.Sprintf("\n... 还有 %d 条记录\n", len(data)-20))
			break
		}
		for key, value := range row {
			result.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// formatDefaultResult 格式化默认查询结果
func formatDefaultResult(data []map[string]interface{}) string {
	var result strings.Builder

	// 显示表格式结果（简化版）
	if len(data) > 0 {
		// 获取列名
		columns := make([]string, 0)
		for key := range data[0] {
			columns = append(columns, key)
		}

		// 显示列名
		result.WriteString("| ")
		for _, col := range columns {
			result.WriteString(fmt.Sprintf("%s | ", col))
		}
		result.WriteString("\n|")
		for range columns {
			result.WriteString(" --- |")
		}
		result.WriteString("\n")

		// 显示数据（最多10行）
		maxRows := 10
		if len(data) < maxRows {
			maxRows = len(data)
		}

		for i := 0; i < maxRows; i++ {
			result.WriteString("| ")
			for _, col := range columns {
				value := data[i][col]
				result.WriteString(fmt.Sprintf("%v | ", value))
			}
			result.WriteString("\n")
		}

		if len(data) > maxRows {
			result.WriteString(fmt.Sprintf("\n... 还有 %d 条记录\n", len(data)-maxRows))
		}
	}

	return result.String()
}
