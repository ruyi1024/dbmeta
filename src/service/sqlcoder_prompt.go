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
	"fmt"
	"strings"
)

// BuildSQLCoderPrompt 构建SQLCoder格式的Prompt（单表查询）
func BuildSQLCoderPrompt(userQuery string, context *QueryContext) string {
	var prompt strings.Builder

	// Instructions部分（中文）- 强化指令
	prompt.WriteString("### 指令说明:\n")
	prompt.WriteString("你的任务是将自然语言问题转换为完整可执行的SQL SELECT查询语句。\n")
	prompt.WriteString("【重要】必须严格遵守以下规则：\n")
	prompt.WriteString("1. 只能使用下方【数据库结构】部分中明确列出的表名，绝对不能自己编造或猜测表名。\n")
	prompt.WriteString("2.【强制要求】字段名必须严格从下方【数据库结构】中对应表的 column_name 字段获取，绝对不能自己编造、猜测或使用字段注释中的文字作为字段名。\n")
	prompt.WriteString("3. 表名必须完全匹配，包括大小写和下划线。\n")
	prompt.WriteString("4. SQL语句中不能包含任何中文字符，表名和字段名必须使用英文、数字和下划线。\n")
	prompt.WriteString("5. 输出必须以SELECT开头，SQL语句以英文分号结尾，不能包含任何其他字符。\n")
	prompt.WriteString("6. SQL中不需要出现数据库名，例如select * from alarm_rule，而不是select * from aidba.alarm_rule。\n")
	prompt.WriteString("7. SQL必须遵守SQL语法规范，不能出现语法错误。正确的写法参考：select * from alarm_rule where event_type='mysql' and enable=1 order by create_time desc limit 10;\n")
	prompt.WriteString("7. 必须只能输出一条完整且可以直接执行的SQL语句。\n")

	// Database Schema部分 - 动态注入元数据
	prompt.WriteString("### 数据库结构:\n")
	if context != nil && context.Metadata != nil {
		// 收集所有可用的表名（用于验证）
		availableTables := make(map[string]bool)

		// 如果指定了表名，优先显示该表的详细信息
		if context.TableName != "" {
			// 查找表注释
			var tableComment string
			var matchedTable map[string]interface{}
			var dbName string
			if len(context.Metadata.Tables) > 0 {
				for _, table := range context.Metadata.Tables {
					if tableName, ok := table["table_name"].(string); ok && tableName == context.TableName {
						matchedTable = table
						if comment, ok := table["table_comment"].(string); ok && comment != "" {
							tableComment = comment
						}
						if db, ok := table["database_name"].(string); ok && db != "" {
							dbName = db
						} else {
							dbName = context.DatabaseName
						}
						availableTables[tableName] = true
						break
					}
				}
			} else {
				dbName = context.DatabaseName
			}

			// 输出表信息
			if matchedTable != nil || context.TableName != "" {
				if tableComment != "" {
					prompt.WriteString(fmt.Sprintf("表: %s.%s -- %s\n", dbName, context.TableName, tableComment))
				} else {
					prompt.WriteString(fmt.Sprintf("表: %s.%s\n", dbName, context.TableName))
				}

				// 从新格式中获取字段信息（从 table["columns"] 中获取）
				var columns []map[string]interface{}
				if matchedTable != nil {
					if cols, ok := matchedTable["columns"].([]interface{}); ok {
						for _, colInterface := range cols {
							if col, ok := colInterface.(map[string]interface{}); ok {
								columns = append(columns, col)
							}
						}
					}
				}

				// 如果没有从新格式获取到，尝试从旧的 Columns 格式获取（向后兼容）
				if len(columns) == 0 {
					for _, col := range context.Metadata.Columns {
						colTableName, _ := col["table_name"].(string)
						colDbName, _ := col["database_name"].(string)
						if colTableName == context.TableName && (colDbName == "" || colDbName == dbName || dbName == "") {
							columns = append(columns, col)
						}
					}
				}

				// 输出字段信息
				if len(columns) > 0 {
					prompt.WriteString("字段列表（【必须使用】以下字段的 column_name，不能使用注释文字）:\n")
					// 收集所有字段名，用于后续强调
					var columnNames []string
					for _, col := range columns {
						schemaLine := FormatColumnForPrompt(col, dbName, context.TableName)
						prompt.WriteString("  " + schemaLine + "\n")
						if colName, ok := col["column_name"].(string); ok && colName != "" {
							columnNames = append(columnNames, colName)
						}
					}
					// 明确列出所有可用字段名
					if len(columnNames) > 0 {
						prompt.WriteString(fmt.Sprintf("\n【重要】表 %s.%s 的可用字段名列表（必须使用这些字段名）: %s\n", dbName, context.TableName, strings.Join(columnNames, ", ")))
						prompt.WriteString("生成的SQL中使用的字段名必须是上述列表中的字段名之一，不能使用其他字段名。\n")
					}
				} else {
					prompt.WriteString("字段列表: (暂无字段信息)\n")
				}
			}
		}

		// 列出所有可用表（完整显示所有表的字段信息）
		if len(context.Metadata.Tables) > 0 {
			// 无论是否指定了表名，都完整显示所有表的字段信息
			if context.TableName != "" {
				prompt.WriteString("\n所有可用表列表（包含完整字段信息）:\n")
			} else {
				prompt.WriteString("可用表列表（包含完整字段信息）:\n")
			}

			// 显示所有表的完整字段信息
			for _, table := range context.Metadata.Tables {
				tableName, _ := table["table_name"].(string)
				dbName := context.DatabaseName
				tableComment := ""
				if comment, ok := table["table_comment"].(string); ok {
					tableComment = comment
				}
				availableTables[tableName] = true

				// 显示表信息
				if tableComment != "" {
					prompt.WriteString(fmt.Sprintf("\n表: %s.%s -- %s\n", dbName, tableName, tableComment))
				} else {
					prompt.WriteString(fmt.Sprintf("\n表: %s.%s\n", dbName, tableName))
				}

				// 从新格式中获取字段信息（从 table["columns"] 中获取）
				var columns []map[string]interface{}
				if cols, ok := table["columns"].([]interface{}); ok {
					for _, colInterface := range cols {
						if col, ok := colInterface.(map[string]interface{}); ok {
							columns = append(columns, col)
						}
					}
				}

				// 如果没有从新格式获取到，尝试从旧的 Columns 格式获取（向后兼容）
				if len(columns) == 0 {
					for _, col := range context.Metadata.Columns {
						colTableName, _ := col["table_name"].(string)
						colDbName, _ := col["database_name"].(string)
						if colTableName == tableName && (colDbName == "" || colDbName == dbName || dbName == "") {
							columns = append(columns, col)
						}
					}
				}

				if len(columns) > 0 {
					prompt.WriteString("字段列表（【必须使用】以下字段的 column_name，不能使用注释文字）:\n")
					// 收集所有字段名，用于后续强调
					var columnNames []string
					for _, col := range columns {
						schemaLine := FormatColumnForPrompt(col, dbName, tableName)
						prompt.WriteString("  " + schemaLine + "\n")
						if colName, ok := col["column_name"].(string); ok && colName != "" {
							columnNames = append(columnNames, colName)
						}
					}
					// 明确列出所有可用字段名
					if len(columnNames) > 0 {
						prompt.WriteString(fmt.Sprintf("  可用字段名（column_name）: %s\n", strings.Join(columnNames, ", ")))
					}
				} else {
					prompt.WriteString("字段列表: (暂无字段信息)\n")
				}
			}
		} else if len(context.Metadata.Databases) > 0 {
			// 只显示数据库
			prompt.WriteString("可用数据库列表:\n")
			for _, db := range context.Metadata.Databases {
				if dbName, ok := db["database_name"].(string); ok {
					prompt.WriteString(fmt.Sprintf("  - %s\n", dbName))
				}
			}
		}

		// 强调表名和字段名限制
		if len(availableTables) > 0 {
			prompt.WriteString("\n【重要提醒】生成的SQL必须严格遵守以下规则：\n")
			prompt.WriteString("1. 表名必须是上述列表中的表名之一，不能使用其他表名。\n")
			// 明确列出所有可用表名
			var tableList []string
			for tableName := range availableTables {
				tableList = append(tableList, tableName)
			}
			if len(tableList) > 0 {
				prompt.WriteString(fmt.Sprintf("可用表名列表: %s\n", strings.Join(tableList, ", ")))
			}
			prompt.WriteString("2. 【强制要求】字段名必须严格从上述字段列表中的 column_name 字段获取，绝对不能自己编造、猜测字段名，也不能使用字段注释（column_comment）中的文字作为字段名。\n")
			prompt.WriteString("3. 如果查询条件中使用了字段（如WHERE、ORDER BY等），这些字段必须使用上述字段列表中对应表的 column_name，且必须在字段列表中存在。\n")
			prompt.WriteString("4. 生成的SQL必须且只能使用上述表名和字段名（column_name），不能使用中文表名/字段名或不在列表中的表名/字段名。\n")
			prompt.WriteString("5. 示例：如果字段列表显示 'alarm_rule.id (bigint) -- 主键ID'，则SQL中必须使用 'id' 作为字段名，不能使用 '主键ID' 或其他名称。\n")
		}

	}

	prompt.WriteString("\n### 问题:\n")
	prompt.WriteString(userQuery + "\n\n")
	prompt.WriteString("### 答案:\n")

	return prompt.String()
}

// extractDatasourceInfo 从元数据中提取数据源信息，只返回1条涉及SQL库和表的元数据
func extractDatasourceInfo(context *QueryContext) map[string]interface{} {
	if context == nil || context.Metadata == nil {
		return nil
	}

	// 优先从指定的表中提取数据源信息
	if context.TableName != "" && len(context.Metadata.Tables) > 0 {
		for _, table := range context.Metadata.Tables {
			tableName, _ := table["table_name"].(string)
			if tableName == context.TableName {
				// 找到匹配的表，提取数据源信息
				datasourceInfo := make(map[string]interface{})
				if dsType, ok := table["datasource_type"].(string); ok && dsType != "" {
					datasourceInfo["datasource_type"] = dsType
				}
				if host, ok := table["host"].(string); ok && host != "" {
					datasourceInfo["host"] = host
				}
				if port, ok := table["port"].(string); ok && port != "" {
					datasourceInfo["port"] = port
				}
				// 如果所有字段都有值，返回
				if len(datasourceInfo) > 0 {
					return datasourceInfo
				}
			}
		}
	}

	// 如果从表中没有找到，尝试从数据库中提取
	if context.DatabaseName != "" && len(context.Metadata.Databases) > 0 {
		for _, db := range context.Metadata.Databases {
			dbName, _ := db["database_name"].(string)
			if dbName == context.DatabaseName {
				// 找到匹配的数据库，提取数据源信息
				datasourceInfo := make(map[string]interface{})
				if dsType, ok := db["datasource_type"].(string); ok && dsType != "" {
					datasourceInfo["datasource_type"] = dsType
				}
				if host, ok := db["host"].(string); ok && host != "" {
					datasourceInfo["host"] = host
				}
				if port, ok := db["port"].(string); ok && port != "" {
					datasourceInfo["port"] = port
				}
				// 如果所有字段都有值，返回
				if len(datasourceInfo) > 0 {
					return datasourceInfo
				}
			}
		}
	}

	// 如果都没有找到，返回第一个可用的数据源信息
	if len(context.Metadata.Tables) > 0 {
		table := context.Metadata.Tables[0]
		datasourceInfo := make(map[string]interface{})
		if dsType, ok := table["datasource_type"].(string); ok && dsType != "" {
			datasourceInfo["datasource_type"] = dsType
		}
		if host, ok := table["host"].(string); ok && host != "" {
			datasourceInfo["host"] = host
		}
		if port, ok := table["port"].(string); ok && port != "" {
			datasourceInfo["port"] = port
		}
		if len(datasourceInfo) > 0 {
			return datasourceInfo
		}
	}

	// 最后尝试从数据库中获取
	if len(context.Metadata.Databases) > 0 {
		db := context.Metadata.Databases[0]
		datasourceInfo := make(map[string]interface{})
		if dsType, ok := db["datasource_type"].(string); ok && dsType != "" {
			datasourceInfo["datasource_type"] = dsType
		}
		if host, ok := db["host"].(string); ok && host != "" {
			datasourceInfo["host"] = host
		}
		if port, ok := db["port"].(string); ok && port != "" {
			datasourceInfo["port"] = port
		}
		if len(datasourceInfo) > 0 {
			return datasourceInfo
		}
	}

	return nil
}

// BuildSQLCoderPromptWithTables 为多表查询构建Prompt
func BuildSQLCoderPromptWithTables(userQuery string, context *QueryContext, tableNames []string) string {
	var prompt strings.Builder

	// Instructions部分
	prompt.WriteString("### Instructions:\n")
	prompt.WriteString("Your task is to convert a natural language question into a SQL query.\n")
	prompt.WriteString("Only return the SQL query, no explanations.\n")
	prompt.WriteString("Use the exact table and column names from the schema below.\n")
	prompt.WriteString("When joining tables, use appropriate JOIN syntax.\n\n")

	// Database Schema部分 - 按表分组注入
	prompt.WriteString("### Database Schema:\n")
	if context != nil && context.Metadata != nil {
		for _, tableName := range tableNames {
			// 查找表注释
			var tableComment string
			if len(context.Metadata.Tables) > 0 {
				for _, table := range context.Metadata.Tables {
					if tName, ok := table["table_name"].(string); ok && tName == tableName {
						if comment, ok := table["table_comment"].(string); ok && comment != "" {
							tableComment = comment
						}
						break
					}
				}
			}

			// 输出表名和注释
			if tableComment != "" {
				prompt.WriteString(fmt.Sprintf("\nTable: %s.%s -- %s\n", context.DatabaseName, tableName, tableComment))
			} else {
				prompt.WriteString(fmt.Sprintf("\nTable: %s.%s\n", context.DatabaseName, tableName))
			}

			// 获取该表的字段（从新格式中获取）
			var columns []map[string]interface{}
			for _, table := range context.Metadata.Tables {
				if tName, ok := table["table_name"].(string); ok && tName == tableName {
					if cols, ok := table["columns"].([]interface{}); ok {
						for _, colInterface := range cols {
							if col, ok := colInterface.(map[string]interface{}); ok {
								columns = append(columns, col)
							}
						}
					}
					break
				}
			}

			// 如果没有从新格式获取到，尝试从旧的 Columns 格式获取（向后兼容）
			if len(columns) == 0 {
				for _, col := range context.Metadata.Columns {
					if colTableName, ok := col["table_name"].(string); ok && colTableName == tableName {
						columns = append(columns, col)
					}
				}
			}

			// 输出字段信息
			for _, col := range columns {
				schemaLine := FormatColumnForPrompt(col, context.DatabaseName, tableName)
				prompt.WriteString("  " + schemaLine + "\n")
			}
		}
	}

	prompt.WriteString("\n### Question:\n")
	prompt.WriteString(userQuery + "\n\n")
	prompt.WriteString("### Answer:\n")

	return prompt.String()
}

// FormatColumnForPrompt 格式化字段信息用于Prompt
func FormatColumnForPrompt(col map[string]interface{}, databaseName, tableName string) string {
	colName := ""
	if name, ok := col["column_name"].(string); ok {
		colName = name
	}

	dataType := ""
	if dt, ok := col["data_type"].(string); ok {
		dataType = dt
	}

	colComment := ""
	if comment, ok := col["column_comment"].(string); ok && comment != "" {
		colComment = comment
	}

	// 格式：database.table.column (type) -- comment
	line := fmt.Sprintf("%s.%s.%s (%s)", databaseName, tableName, colName, dataType)
	if colComment != "" {
		line += fmt.Sprintf(" -- %s", colComment)
	}

	return line
}
