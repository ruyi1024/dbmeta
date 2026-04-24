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

package pumpkin

import (
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 格式化字节大小为可读字符串
func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	} else if bytes < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2f TB", float64(bytes)/(1024*1024*1024*1024))
	}
}

// 格式化平均行长度
func formatAvgRowLength(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	}
}

// GetDatabaseCapacityTop10Chart 获取数据库容量TOP10（用于图表）
func GetDatabaseCapacityTop10Chart(c *gin.Context) {
	// 获取今天起始时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 查询今天每个数据库最新一条容量记录，并取TOP10
	querySQL := `
		SELECT 
			t1.database_name,
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_size as total_data_size
		FROM pumpkin_database_growth t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, MAX(gmt_created) as max_created
			FROM pumpkin_database_growth
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name
		) t2 ON t1.datasource_type = t2.datasource_type
			AND t1.host = t2.host
			AND t1.port = t2.port
			AND t1.database_name = t2.database_name
			AND t1.gmt_created = t2.max_created
		ORDER BY t1.database_size DESC
		LIMIT 10
	`

	var results []struct {
		DatabaseName   string `gorm:"column:database_name"`
		DatasourceType string `gorm:"column:datasource_type"`
		Host           string `gorm:"column:host"`
		Port           string `gorm:"column:port"`
		TotalDataSize  int64  `gorm:"column:total_data_size"`
	}

	result := database.DB.Raw(querySQL, todayStart).Scan(&results)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询数据库容量失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式（图表用，只需要容量数据）
	dataList := make([]map[string]interface{}, 0)
	for i, item := range results {
		dataList = append(dataList, map[string]interface{}{
			"id":             i + 1,
			"databaseName":   item.DatabaseName,
			"datasourceType": item.DatasourceType,
			"host":           item.Host,
			"port":           item.Port,
			"dataSize":       formatSize(item.TotalDataSize), // 格式化后的显示值
			"dataSizeBytes":  item.TotalDataSize,             // 原始字节数，用于排序
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   len(dataList),
	})
}

// GetDatabaseCapacityTop10 获取数据库容量信息（用于表格，支持分页、搜索、排序）
// 从 pumpkin_database_growth 表获取今天的数据
func GetDatabaseCapacityTop10(c *gin.Context) {
	// 获取查询参数
	current := c.DefaultQuery("current", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	databaseName := c.Query("databaseName")     // 数据库名搜索
	datasourceType := c.Query("datasourceType") // 数据库类型搜索
	host := c.Query("host")                     // 主机搜索
	port := c.Query("port")                     // 端口搜索

	// 解析分页参数
	var currentPage, pageSizeInt int
	fmt.Sscanf(current, "%d", &currentPage)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)
	if currentPage < 1 {
		currentPage = 1
	}
	if pageSizeInt < 1 {
		pageSizeInt = 10
	}
	offset := (currentPage - 1) * pageSizeInt

	// 构建WHERE条件 - 获取今天的数据
	whereClause := "WHERE DATE(gmt_created) = CURDATE()"
	args := []interface{}{}
	if databaseName != "" {
		whereClause += " AND database_name LIKE ?"
		args = append(args, "%"+databaseName+"%")
	}
	if datasourceType != "" {
		whereClause += " AND datasource_type LIKE ?"
		args = append(args, "%"+datasourceType+"%")
	}
	if host != "" {
		whereClause += " AND host LIKE ?"
		args = append(args, "%"+host+"%")
	}
	if port != "" {
		whereClause += " AND port LIKE ?"
		args = append(args, "%"+port+"%")
	}

	// 获取排序参数
	orderBy := "database_size DESC" // 默认按数据大小降序
	if sortField := c.Query("sortField"); sortField != "" {
		sortOrder := c.DefaultQuery("sortOrder", "desc")
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}
		switch sortField {
		case "databaseName":
			orderBy = fmt.Sprintf("database_name %s", strings.ToUpper(sortOrder))
		case "datasourceType":
			orderBy = fmt.Sprintf("datasource_type %s", strings.ToUpper(sortOrder))
		case "dataSize":
			orderBy = fmt.Sprintf("database_size %s", strings.ToUpper(sortOrder))
		case "rowCount":
			orderBy = fmt.Sprintf("database_rows %s", strings.ToUpper(sortOrder))
		case "tableCount":
			orderBy = fmt.Sprintf("table_count %s", strings.ToUpper(sortOrder))
		case "dataSizeIncr":
			orderBy = fmt.Sprintf("database_size_incr %s", strings.ToUpper(sortOrder))
		case "rowCountIncr":
			orderBy = fmt.Sprintf("database_rows_incr %s", strings.ToUpper(sortOrder))
		default:
			orderBy = "database_size DESC"
		}
	}

	// 查询总数
	var total int64
	countSQL := fmt.Sprintf(`
		SELECT COUNT(DISTINCT CONCAT(database_name, '-', datasource_type, '-', host, '-', port))
		FROM pumpkin_database_growth
		%s
	`, whereClause)
	database.DB.Raw(countSQL, args...).Scan(&total)

	// 查询数据 - 从 pumpkin_database_growth 表获取今天的数据
	querySQL := fmt.Sprintf(`
		SELECT 
			id,
			database_name,
			datasource_type,
			host,
			port,
			database_size,
			database_rows,
			table_count,
			database_size_incr,
			database_rows_incr
		FROM pumpkin_database_growth
		%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)
	args = append(args, pageSizeInt, offset)

	var results []struct {
		Id               int64  `gorm:"column:id"`
		DatabaseName     string `gorm:"column:database_name"`
		DatasourceType   string `gorm:"column:datasource_type"`
		Host             string `gorm:"column:host"`
		Port             string `gorm:"column:port"`
		DatabaseSize     int64  `gorm:"column:database_size"`
		DatabaseRows     int64  `gorm:"column:database_rows"`
		TableCount       int64  `gorm:"column:table_count"`
		DatabaseSizeIncr int64  `gorm:"column:database_size_incr"`
		DatabaseRowsIncr int64  `gorm:"column:database_rows_incr"`
	}

	result := database.DB.Raw(querySQL, args...).Scan(&results)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询数据库容量失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式（表格用，包含详细信息）
	dataList := make([]map[string]interface{}, 0)
	for _, item := range results {
		dataList = append(dataList, map[string]interface{}{
			"id":                item.Id,
			"databaseName":      item.DatabaseName,
			"datasourceType":    item.DatasourceType,
			"host":              item.Host,
			"port":              item.Port,
			"tableCount":        item.TableCount,                   // 表数量
			"dataSize":          formatSize(item.DatabaseSize),     // 格式化后的显示值
			"dataSizeBytes":     item.DatabaseSize,                 // 原始字节数，用于排序
			"rowCount":          item.DatabaseRows,                 // 数据记录条数
			"dataSizeIncr":      formatSize(item.DatabaseSizeIncr), // 数据存储日增长（格式化）
			"dataSizeIncrBytes": item.DatabaseSizeIncr,             // 数据存储日增长（原始字节数，用于排序）
			"rowCountIncr":      item.DatabaseRowsIncr,             // 数据记录日增长
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   total,
	})
}

// GetTableCapacityTop10 获取数据表容量TOP10（用于图表，从pumpkin_table_size获取）
func GetTableCapacityTop10(c *gin.Context) {
	// 获取今天起始时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 查询今天每个表最新一条容量记录，并按容量排序取TOP10
	var tableSizes []model.PumpkinTableGrowth
	result := database.DB.Raw(`
		SELECT 
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			t1.table_size,
			t1.table_rows
		FROM pumpkin_table_growth t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, table_name, MAX(gmt_created) as max_created
			FROM pumpkin_table_growth
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type
			AND t1.host = t2.host
			AND t1.port = t2.port
			AND t1.database_name = t2.database_name
			AND t1.table_name = t2.table_name
			AND t1.gmt_created = t2.max_created
		ORDER BY t1.table_size DESC
		LIMIT 10
	`, todayStart).Scan(&tableSizes)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询数据表容量失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	dataList := make([]map[string]interface{}, 0)
	for i, item := range tableSizes {
		dataList = append(dataList, map[string]interface{}{
			"id":             i + 1,
			"tableName":      item.TableNameX,
			"databaseName":   item.DatabaseName,
			"datasourceType": item.DatasourceType,
			"host":           item.Host,
			"port":           item.Port,
			"dataSize":       formatSize(item.TableSize), // 格式化后的显示值
			"dataSizeBytes":  item.TableSize,             // 原始字节数，用于排序
			"indexSize":      "-",                        // 图表暂不展示索引大小，保留字段避免前端兼容问题
			"indexSizeBytes": int64(0),
			"rowCount":       item.TableRows,
			"avgRowLength":   "-",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   len(dataList),
	})
}

// GetTableCapacityGrowth 获取数据表容量信息（用于表格，支持分页、搜索、排序）
// 从 pumpkin_table_growth 表获取今天的数据
func GetTableCapacityGrowth(c *gin.Context) {
	// 获取查询参数
	current := c.DefaultQuery("current", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	databaseName := c.Query("databaseName")     // 数据库名搜索
	tableName := c.Query("tableName")           // 表名搜索
	datasourceType := c.Query("datasourceType") // 数据库类型搜索
	host := c.Query("host")                     // 主机搜索
	port := c.Query("port")                     // 端口搜索

	// 解析分页参数
	var currentPage, pageSizeInt int
	fmt.Sscanf(current, "%d", &currentPage)
	fmt.Sscanf(pageSize, "%d", &pageSizeInt)
	if currentPage < 1 {
		currentPage = 1
	}
	if pageSizeInt < 1 {
		pageSizeInt = 10
	}
	offset := (currentPage - 1) * pageSizeInt

	// 构建WHERE条件 - 获取今天的数据
	whereClause := "WHERE DATE(gmt_created) = CURDATE()"
	args := []interface{}{}
	if databaseName != "" {
		whereClause += " AND database_name LIKE ?"
		args = append(args, "%"+databaseName+"%")
	}
	if tableName != "" {
		whereClause += " AND table_name LIKE ?"
		args = append(args, "%"+tableName+"%")
	}
	if datasourceType != "" {
		whereClause += " AND datasource_type LIKE ?"
		args = append(args, "%"+datasourceType+"%")
	}
	if host != "" {
		whereClause += " AND host LIKE ?"
		args = append(args, "%"+host+"%")
	}
	if port != "" {
		whereClause += " AND port LIKE ?"
		args = append(args, "%"+port+"%")
	}

	// 获取排序参数
	orderBy := "table_size DESC" // 默认按数据大小降序
	if sortField := c.Query("sortField"); sortField != "" {
		sortOrder := c.DefaultQuery("sortOrder", "desc")
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}
		switch sortField {
		case "databaseName":
			orderBy = fmt.Sprintf("database_name %s", strings.ToUpper(sortOrder))
		case "tableName":
			orderBy = fmt.Sprintf("table_name %s", strings.ToUpper(sortOrder))
		case "datasourceType":
			orderBy = fmt.Sprintf("datasource_type %s", strings.ToUpper(sortOrder))
		case "dataSize":
			orderBy = fmt.Sprintf("table_size %s", strings.ToUpper(sortOrder))
		case "rowCount":
			orderBy = fmt.Sprintf("table_rows %s", strings.ToUpper(sortOrder))
		case "dataSizeIncr":
			orderBy = fmt.Sprintf("table_size_incr %s", strings.ToUpper(sortOrder))
		case "rowCountIncr":
			orderBy = fmt.Sprintf("table_rows_incr %s", strings.ToUpper(sortOrder))
		default:
			orderBy = "table_size DESC"
		}
	}

	// 查询总数
	var total int64
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM pumpkin_table_growth
		%s
	`, whereClause)
	database.DB.Raw(countSQL, args...).Scan(&total)

	// 查询数据 - 从 pumpkin_table_growth 表获取今天的数据
	querySQL := fmt.Sprintf(`
		SELECT 
			id,
			database_name,
			table_name,
			datasource_type,
			host,
			port,
			table_size,
			table_rows,
			table_size_incr,
			COALESCE(table_rows_incr, 0) as table_rows_incr
		FROM pumpkin_table_growth
		%s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)
	args = append(args, pageSizeInt, offset)

	var results []struct {
		Id             int64  `gorm:"column:id"`
		DatabaseName   string `gorm:"column:database_name"`
		TableName      string `gorm:"column:table_name"`
		DatasourceType string `gorm:"column:datasource_type"`
		Host           string `gorm:"column:host"`
		Port           string `gorm:"column:port"`
		TableSize      int64  `gorm:"column:table_size"`
		TableRows      int64  `gorm:"column:table_rows"`
		TableSizeIncr  int64  `gorm:"column:table_size_incr"`
		TableRowsIncr  int64  `gorm:"column:table_rows_incr"`
	}

	result := database.DB.Raw(querySQL, args...).Scan(&results)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询数据表容量失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式（表格用，包含详细信息）
	dataList := make([]map[string]interface{}, 0)
	for _, item := range results {
		dataList = append(dataList, map[string]interface{}{
			"id":                item.Id,
			"databaseName":      item.DatabaseName,
			"tableName":         item.TableName,
			"datasourceType":    item.DatasourceType,
			"host":              item.Host,
			"port":              item.Port,
			"dataSize":          formatSize(item.TableSize),     // 格式化后的显示值
			"dataSizeBytes":     item.TableSize,                 // 原始字节数，用于排序
			"rowCount":          item.TableRows,                 // 数据记录条数
			"dataSizeIncr":      formatSize(item.TableSizeIncr), // 数据存储日增长（格式化）
			"dataSizeIncrBytes": item.TableSizeIncr,             // 数据存储日增长（原始字节数，用于排序）
			"rowCountIncr":      item.TableRowsIncr,             // 数据记录日增长
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   total,
	})
}

// GetCapacityStats 获取数据容量统计信息
func GetCapacityStats(c *gin.Context) {
	// 获取今天的时间范围
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// 总数据库数（去重）- 从pumpkin_database_growth表统计今天每个数据库的最新记录
	var totalDatabases int64
	if err := database.DB.Raw(`
		SELECT COUNT(DISTINCT t1.database_name) 
		FROM pumpkin_database_growth t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, MAX(gmt_created) as max_created
			FROM pumpkin_database_growth
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.gmt_created = t2.max_created
	`, todayStart).Scan(&totalDatabases).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询数据库数失败: " + err.Error()})
		return
	}

	// 总数据表数（去重）- 从pumpkin_table_growth表统计今天每个表的最新记录
	var totalTables int64
	if err := database.DB.Raw(`
		SELECT COUNT(DISTINCT CONCAT(t1.database_name, '.', t1.table_name)) 
		FROM pumpkin_table_growth t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, table_name, MAX(gmt_created) as max_created
			FROM pumpkin_table_growth
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.table_name = t2.table_name 
			AND t1.gmt_created = t2.max_created
	`, todayStart).Scan(&totalTables).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询数据表数失败: " + err.Error()})
		return
	}

	// 总数据量 - 从pumpkin_database_growth表统计今天每个数据库的最新记录，然后求和
	var totalDataSize int64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(t1.database_size), 0) 
		FROM pumpkin_database_growth t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, MAX(gmt_created) as max_created
			FROM pumpkin_database_growth
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.gmt_created = t2.max_created
	`, todayStart).Scan(&totalDataSize).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询总数据量失败: " + err.Error()})
		return
	}

	// 总数据记录数 - 从pumpkin_database_growth表统计今天每个数据库的最新记录，然后求和
	var totalRows int64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(t1.database_rows), 0) 
		FROM pumpkin_database_growth t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, MAX(gmt_created) as max_created
			FROM pumpkin_database_growth
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name
		) t2 ON t1.datasource_type = t2.datasource_type 
			AND t1.host = t2.host 
			AND t1.port = t2.port 
			AND t1.database_name = t2.database_name 
			AND t1.gmt_created = t2.max_created
	`, todayStart).Scan(&totalRows).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询总记录数失败: " + err.Error()})
		return
	}

	// 天增长数据量 - 从pumpkin_database_growth表统计今天的增量
	var dailyGrowth int64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(database_size_incr), 0) 
		FROM pumpkin_database_growth 
		WHERE gmt_created >= ?
	`, todayStart).Scan(&dailyGrowth).Error; err != nil {
		// 如果查询失败，设置为0
		dailyGrowth = 0
	}
	if dailyGrowth < 0 {
		dailyGrowth = 0
	}

	// 天增长记录数 - 从pumpkin_database_growth表统计今天的增量
	var dailyGrowthRows int64
	if err := database.DB.Raw(`
		SELECT COALESCE(SUM(database_rows_incr), 0) 
		FROM pumpkin_database_growth 
		WHERE gmt_created >= ?
	`, todayStart).Scan(&dailyGrowthRows).Error; err != nil {
		// 如果查询失败，设置为0
		dailyGrowthRows = 0
	}
	if dailyGrowthRows < 0 {
		dailyGrowthRows = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"totalDatabases":  totalDatabases,
			"totalTables":     totalTables,
			"totalDataSize":   formatSize(totalDataSize),
			"totalRows":       totalRows,
			"dailyGrowth":     formatSize(dailyGrowth),
			"dailyGrowthRows": dailyGrowthRows,
		},
	})
}

// GetDatabaseTypeDistribution 获取按数据库类型聚合的容量与记录数分布
func GetDatabaseTypeDistribution(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	querySQL := `
		SELECT
			t.datasource_type,
			COUNT(*) as database_count,
			CAST(COALESCE(SUM(t.database_size), 0) AS SIGNED) as total_data_size,
			CAST(COALESCE(SUM(t.database_rows), 0) AS SIGNED) as total_rows
		FROM (
			SELECT
				t1.datasource_type,
				t1.database_size,
				t1.database_rows
			FROM pumpkin_database_growth t1
			INNER JOIN (
				SELECT datasource_type, host, port, database_name, MAX(gmt_created) as max_created
				FROM pumpkin_database_growth
				WHERE gmt_created >= ?
				GROUP BY datasource_type, host, port, database_name
			) t2 ON t1.datasource_type = t2.datasource_type
				AND t1.host = t2.host
				AND t1.port = t2.port
				AND t1.database_name = t2.database_name
				AND t1.gmt_created = t2.max_created
		) t
		GROUP BY t.datasource_type
		ORDER BY total_data_size DESC
	`

	var results []struct {
		DatasourceType string `gorm:"column:datasource_type"`
		DatabaseCount  int64  `gorm:"column:database_count"`
		TotalDataSize  int64  `gorm:"column:total_data_size"`
		TotalRows      int64  `gorm:"column:total_rows"`
	}

	result := database.DB.Raw(querySQL, todayStart).Scan(&results)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询数据库类型分布失败: " + result.Error.Error(),
		})
		return
	}

	dataList := make([]map[string]interface{}, 0, len(results))
	for _, item := range results {
		dataList = append(dataList, map[string]interface{}{
			"datasourceType":   item.DatasourceType,
			"databaseCount":    item.DatabaseCount,
			"totalDataSize":    formatSize(item.TotalDataSize),
			"totalDataSizeBytes": item.TotalDataSize,
			"totalRows":        item.TotalRows,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   len(dataList),
	})
}

// GetTableFragmentationTop10 获取表碎片大小TOP10（保留原接口路径）
func GetTableFragmentationTop10(c *gin.Context) {
	// 获取今天起始时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 查询今天每个表最新一条碎片记录，并按碎片大小取TOP10
	var tableSizes []model.PumpkinTableSize
	result := database.DB.Raw(`
		SELECT
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			t1.free_size
		FROM pumpkin_table_size t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, table_name, MAX(gmt_created) as max_created
			FROM pumpkin_table_size
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type
			AND t1.host = t2.host
			AND t1.port = t2.port
			AND t1.database_name = t2.database_name
			AND t1.table_name = t2.table_name
			AND t1.gmt_created = t2.max_created
		WHERE t1.free_size > 0
		ORDER BY t1.free_size DESC
		LIMIT 10
	`, todayStart).Scan(&tableSizes)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询表碎片大小失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	dataList := make([]map[string]interface{}, 0)
	for i, item := range tableSizes {
		dataList = append(dataList, map[string]interface{}{
			"id":             i + 1,
			"tableName":      item.TableNameField,
			"databaseName":   item.DatabaseName,
			"datasourceType": item.DatasourceType,
			"host":           item.Host,
			"port":           item.Port,
			"freeSize":       formatSize(item.FreeSize), // 格式化后的显示值
			"freeSizeBytes":  item.FreeSize,             // 原始字节数，用于排序
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   len(dataList),
	})
}

// GetTableRowsTop10 获取表记录数TOP10
func GetTableRowsTop10(c *gin.Context) {
	// 获取今天起始时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 查询今天每个表最新一条记录数数据，并按记录数排序取TOP10
	var tableSizes []model.PumpkinTableSize
	result := database.DB.Raw(`
		SELECT
			t1.datasource_type,
			t1.host,
			t1.port,
			t1.database_name,
			t1.table_name,
			t1.table_rows
		FROM pumpkin_table_size t1
		INNER JOIN (
			SELECT datasource_type, host, port, database_name, table_name, MAX(gmt_created) as max_created
			FROM pumpkin_table_size
			WHERE gmt_created >= ?
			GROUP BY datasource_type, host, port, database_name, table_name
		) t2 ON t1.datasource_type = t2.datasource_type
			AND t1.host = t2.host
			AND t1.port = t2.port
			AND t1.database_name = t2.database_name
			AND t1.table_name = t2.table_name
			AND t1.gmt_created = t2.max_created
		ORDER BY t1.table_rows DESC
		LIMIT 10
	`, todayStart).Scan(&tableSizes)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询表记录数失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	dataList := make([]map[string]interface{}, 0)
	for i, item := range tableSizes {
		// 格式化行数显示
		var rowCountDisplay string
		if item.TableRows >= 1000000000 {
			rowCountDisplay = fmt.Sprintf("%.2fB", float64(item.TableRows)/1000000000.0)
		} else if item.TableRows >= 1000000 {
			rowCountDisplay = fmt.Sprintf("%.2fM", float64(item.TableRows)/1000000.0)
		} else if item.TableRows >= 1000 {
			rowCountDisplay = fmt.Sprintf("%.2fK", float64(item.TableRows)/1000.0)
		} else {
			rowCountDisplay = fmt.Sprintf("%d", item.TableRows)
		}

		dataList = append(dataList, map[string]interface{}{
			"id":             i + 1,
			"tableName":      item.TableNameField,
			"databaseName":   item.DatabaseName,
			"datasourceType": item.DatasourceType,
			"host":           item.Host,
			"port":           item.Port,
			"rowCount":       rowCountDisplay, // 格式化后的显示值
			"rowCountValue":  item.TableRows,  // 原始数值，用于排序
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataList,
		"total":   len(dataList),
	})
}
