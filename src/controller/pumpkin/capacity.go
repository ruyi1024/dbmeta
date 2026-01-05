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
*/

package pumpkin

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
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
	// 获取近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// 查询近1小时的数据库容量数据，按数据库分组汇总（只计算总容量）
	querySQL := `
		SELECT 
			database_name,
			datasource_type,
			host,
			port,
			SUM(data_size + index_size + free_size) as total_data_size
		FROM pumpkin_table_size
		WHERE gmt_created >= ?
		GROUP BY database_name, datasource_type, host, port
		ORDER BY total_data_size DESC
		LIMIT 10
	`

	var results []struct {
		DatabaseName   string `gorm:"column:database_name"`
		DatasourceType string `gorm:"column:datasource_type"`
		Host           string `gorm:"column:host"`
		Port           string `gorm:"column:port"`
		TotalDataSize  int64  `gorm:"column:total_data_size"`
	}

	result := database.DB.Raw(querySQL, oneHourAgo).Scan(&results)
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
	// 获取近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// 查询近1小时的数据表容量数据，按数据大小排序
	var tableSizes []model.PumpkinTableSize
	result := database.DB.Where("gmt_created >= ?", oneHourAgo).
		Order("data_size DESC").
		Limit(10).
		Find(&tableSizes)

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
			"tableName":      item.TableNameField,
			"databaseName":   item.DatabaseName,
			"datasourceType": item.DatasourceType,
			"host":           item.Host,
			"port":           item.Port,
			"dataSize":       formatSize(item.DataSize),  // 格式化后的显示值
			"dataSizeBytes":  item.DataSize,              // 原始字节数，用于排序
			"indexSize":      formatSize(item.IndexSize), // 格式化后的显示值
			"indexSizeBytes": item.IndexSize,             // 原始字节数
			"rowCount":       item.TableRows,
			"avgRowLength":   formatAvgRowLength(item.AvgRowLength),
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
	// 获取近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// 总数据库数（去重）- 使用子查询
	var totalDatabases int64
	database.DB.Raw(`
		SELECT COUNT(DISTINCT database_name) 
		FROM pumpkin_table_size 
		WHERE gmt_created >= ?
	`, oneHourAgo).Scan(&totalDatabases)

	// 总数据表数（去重）- 使用子查询
	var totalTables int64
	database.DB.Raw(`
		SELECT COUNT(DISTINCT CONCAT(database_name, '.', table_name)) 
		FROM pumpkin_table_size 
		WHERE gmt_created >= ?
	`, oneHourAgo).Scan(&totalTables)

	// 总数据量
	var totalDataSize int64
	database.DB.Raw(`
		SELECT COALESCE(SUM(data_size), 0) 
		FROM pumpkin_table_size 
		WHERE gmt_created >= ?
	`, oneHourAgo).Scan(&totalDataSize)

	// 天增长数据量（获取24小时前的数据量，计算差值）
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	var oneDayAgoDataSize int64
	database.DB.Raw(`
		SELECT COALESCE(SUM(data_size), 0) 
		FROM pumpkin_table_size 
		WHERE gmt_created >= ? AND gmt_created < ?
	`, oneDayAgo, oneHourAgo).Scan(&oneDayAgoDataSize)

	dailyGrowth := totalDataSize - oneDayAgoDataSize
	if dailyGrowth < 0 {
		dailyGrowth = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"totalDatabases": totalDatabases,
			"totalTables":    totalTables,
			"totalDataSize":  formatSize(totalDataSize),
			"dailyGrowth":    formatSize(dailyGrowth),
		},
	})
}

// GetTableFragmentationTop10 获取表碎片率TOP10
func GetTableFragmentationTop10(c *gin.Context) {
	// 获取近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// 查询近1小时的数据表碎片率数据，计算碎片率并排序
	querySQL := `
		SELECT 
			table_name,
			database_name,
			datasource_type,
			host,
			port,
			data_size,
			index_size,
			free_size,
			CASE 
				WHEN (data_size + index_size + free_size) > 0 
				THEN (free_size * 100.0 / (data_size + index_size + free_size))
				ELSE 0 
			END as fragmentation_rate
		FROM pumpkin_table_size
		WHERE gmt_created >= ?
			AND (data_size + index_size + free_size) > 0
		GROUP BY table_name, database_name, datasource_type, host, port
		ORDER BY fragmentation_rate DESC
		LIMIT 10
	`

	var results []struct {
		TableName         string  `gorm:"column:table_name"`
		DatabaseName      string  `gorm:"column:database_name"`
		DatasourceType    string  `gorm:"column:datasource_type"`
		Host              string  `gorm:"column:host"`
		Port              string  `gorm:"column:port"`
		DataSize          int64   `gorm:"column:data_size"`
		IndexSize         int64   `gorm:"column:index_size"`
		FreeSize          int64   `gorm:"column:free_size"`
		FragmentationRate float64 `gorm:"column:fragmentation_rate"`
	}

	result := database.DB.Raw(querySQL, oneHourAgo).Scan(&results)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询表碎片率失败: " + result.Error.Error(),
		})
		return
	}

	// 转换为前端需要的格式
	dataList := make([]map[string]interface{}, 0)
	for i, item := range results {
		dataList = append(dataList, map[string]interface{}{
			"id":                     i + 1,
			"tableName":              item.TableName,
			"databaseName":           item.DatabaseName,
			"datasourceType":         item.DatasourceType,
			"host":                   item.Host,
			"port":                   item.Port,
			"fragmentationRate":      fmt.Sprintf("%.2f%%", item.FragmentationRate), // 格式化后的显示值（百分比）
			"fragmentationRateValue": item.FragmentationRate,                        // 原始数值，用于排序
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
	// 获取近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// 查询近1小时的数据表记录数数据，按记录数排序
	var tableSizes []model.PumpkinTableSize
	result := database.DB.Where("gmt_created >= ?", oneHourAgo).
		Order("table_rows DESC").
		Limit(10).
		Find(&tableSizes)

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
