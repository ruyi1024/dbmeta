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

package ai

import (
	"dbmcloud/log"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/service"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ChatQueryRequest 聊天查询请求
type ChatQueryRequest struct {
	SessionId    string `json:"session_id" binding:"required"`
	Question     string `json:"question" binding:"required"`
	DatasourceId int    `json:"datasource_id"`
	DatabaseName string `json:"database_name"`
	TableName    string `json:"table_name"`
	ResetContext bool   `json:"reset_context"` // 是否重置上下文（新问题时设置为true，不使用之前的多轮对话缓存）
}

// ChatQueryResponse 聊天查询响应
type ChatQueryResponse struct {
	Answer      string                   `json:"answer"`
	SqlQuery    string                   `json:"sql_query"`
	QueryResult []map[string]interface{} `json:"query_result"`
	Timestamp   int64                    `json:"timestamp"`
	Options     []string                 `json:"options,omitempty"` // 多轮对话的选择选项（当问题类型为select时）
}

// ChatQuery 处理AI聊天查询请求
func ChatQuery(c *gin.Context) {
	var req ChatQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Chat query request bind error", zap.Error(err))
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	username, _ := c.Get("username")
	userName := username.(string)

	log.Info("AI Chat query request",
		zap.String("session_id", req.SessionId),
		zap.String("question", req.Question),
		zap.Int("datasource_id", req.DatasourceId))

	// 验证会话是否存在
	session, err := service.GetSession(req.SessionId)
	if err != nil {
		log.Warn("会话不存在", zap.String("session_id", req.SessionId), zap.Error(err))
		c.JSON(404, gin.H{
			"success": false,
			"message": "会话不存在",
			"error":   err.Error(),
		})
		return
	}

	// 验证会话是否属于当前用户
	if session.UserName != userName {
		log.Warn("无权限访问会话", zap.String("session_id", req.SessionId), zap.String("session_user", session.UserName), zap.String("current_user", userName))
		c.JSON(403, gin.H{
			"success": false,
			"message": "无权限访问此会话",
		})
		return
	}

	log.Info("会话验证通过", zap.String("session_id", req.SessionId), zap.String("user", userName))

	// 保存用户消息
	err = service.SaveMessage(req.SessionId, "user", req.Question, "", nil)
	if err != nil {
		log.Error("保存用户消息失败", zap.Error(err))
	} else {
		log.Debug("用户消息已保存", zap.String("session_id", req.SessionId))
	}

	var sqlQuery string
	var queryResult []map[string]interface{}
	var answer string
	var datasourceId int
	var databaseName, tableName string
	var ruleMatch *service.RuleMatchResult

	// 确定数据库名和表名
	if req.DatabaseName != "" {
		databaseName = req.DatabaseName
	}
	if req.TableName != "" {
		tableName = req.TableName
	}

	// 如果 reset_context=false，先查找多轮对话上下文
	// 如果找到了匹配的上下文，直接使用该上下文对应的规则，避免用户输入（如选择值）被误匹配到其他规则
	var multiRoundContext *service.MultiRoundContext
	if !req.ResetContext {
		messages, msgErr := service.GetSessionMessages(req.SessionId, 20)
		if msgErr == nil {
			// 查找最近的多轮对话上下文
			for i := len(messages) - 1; i >= 0; i-- {
				msg := messages[i]
				if msg.Role == "assistant" {
					context, extractErr := service.ExtractMultiRoundContext(msg.Content)
					if extractErr == nil && context != nil {
						multiRoundContext = context
						log.Info("检测到正在进行多轮对话",
							zap.Int64("rule_id", context.RuleId),
							zap.String("rule_name", context.RuleName),
							zap.Int("current_step", context.CurrentStep),
							zap.Any("collected_params", context.Collected))
						break
					}
				}
			}
		}
	} else {
		log.Info("请求中设置了 reset_context=true，跳过多轮对话上下文检查，开始新流程",
			zap.String("question", req.Question))
	}

	// 如果找到了多轮对话上下文，直接加载对应的规则，不进行规则匹配
	// 这样可以避免用户输入（如选择值"MySQL"）被误匹配到其他规则
	if multiRoundContext != nil {
		var rule model.SemanticSqlRule
		result := database.DB.Where("id = ?", multiRoundContext.RuleId).First(&rule)
		if result.Error == nil {
			log.Info("从多轮对话上下文中加载规则，跳过了规则匹配",
				zap.Int64("rule_id", rule.Id),
				zap.String("rule_name", rule.RuleName))
			ruleMatch = &service.RuleMatchResult{
				Rule:   &rule,
				Params: make(map[string]string), // 参数将在后续步骤中从collectedParams中获取
			}
		} else {
			log.Warn("无法加载多轮对话中的规则，将重新匹配", zap.Error(result.Error), zap.Int64("rule_id", multiRoundContext.RuleId))
			// 如果加载失败，继续执行规则匹配
		}
	}

	// 如果没有多轮对话上下文或规则加载失败，进行规则匹配
	if ruleMatch == nil || ruleMatch.Rule == nil {
		log.Info("开始规则匹配", zap.String("question", req.Question))
		ruleMatch, err = service.MatchRule(req.Question)
		if err != nil {
			log.Error("规则匹配失败", zap.Error(err))
		} else if ruleMatch != nil && ruleMatch.Rule != nil {
			log.Info("规则匹配成功",
				zap.String("rule_name", ruleMatch.Rule.RuleName),
				zap.Int64("rule_id", ruleMatch.Rule.Id),
				zap.Int8("use_local_db", ruleMatch.Rule.UseLocalDB),
				zap.Int8("multi_round_enabled", ruleMatch.Rule.MultiRoundEnabled),
				zap.Any("extracted_params", ruleMatch.Params))
		} else {
			log.Info("未匹配到规则，将使用AI生成SQL")
		}
	}

	// 如果规则匹配成功
	if ruleMatch != nil && ruleMatch.Rule != nil {
		// 检查是否需要多轮信息收集
		log.Debug("检查多轮信息收集状态", zap.String("rule_name", ruleMatch.Rule.RuleName), zap.Bool("reset_context", req.ResetContext))
		needMore, nextQuestion, collectedParams, options, multiRoundErr := service.CheckMultiRoundInfo(req.SessionId, ruleMatch.Rule, req.Question, req.ResetContext)
		if multiRoundErr != nil {
			log.Error("检查多轮信息收集失败", zap.Error(multiRoundErr))
		}

		// 如果需要继续收集信息，返回提示问题
		if needMore {
			log.Info("需要继续收集多轮信息",
				zap.String("rule_name", ruleMatch.Rule.RuleName),
				zap.String("next_question", nextQuestion),
				zap.Any("collected_params", collectedParams),
				zap.Int("options_count", len(options)))
			answer = nextQuestion
			// 保存系统提示消息
			err = service.SaveMessage(req.SessionId, "assistant", answer, "", nil)
			if err != nil {
				log.Error("保存系统提示消息失败", zap.Error(err))
			}
			// 统一返回格式，使用data字段包装
			response := ChatQueryResponse{
				Answer:      answer,
				SqlQuery:    "",
				QueryResult: nil,
				Timestamp:   time.Now().Unix(),
				Options:     options, // 如果是select类型，包含选项列表
			}
			c.JSON(200, gin.H{
				"success": true,
				"data":    response,
			})
			return
		}

		// 如果已收集参数，应用到SQL模板
		// 注意：如果 reset_context=true，collectedParams 应该是空的（因为 CheckMultiRoundInfo 会忽略历史上下文）
		// 只有在多轮对话继续时（reset_context=false），才会使用之前收集的参数
		if len(collectedParams) > 0 && ruleMatch.Rule.MultiRoundEnabled == 1 {
			log.Info("应用多轮收集的参数到SQL模板",
				zap.String("rule_name", ruleMatch.Rule.RuleName),
				zap.Bool("reset_context", req.ResetContext),
				zap.Any("collected_params", collectedParams))
			// 将收集的参数合并到规则参数中
			if ruleMatch.Params == nil {
				ruleMatch.Params = make(map[string]string)
			}
			for k, v := range collectedParams {
				ruleMatch.Params[k] = v
			}
			// 应用参数到SQL模板（无论是否有参数映射都要应用）
			ruleMatch.Rule.SqlTemplate = service.ApplyCollectedParams(ruleMatch.Rule.SqlTemplate, collectedParams, ruleMatch.Rule.ParameterMapping)
			log.Debug("已应用参数到SQL模板", zap.String("sql_template", ruleMatch.Rule.SqlTemplate))
		} else if req.ResetContext {
			// 如果 reset_context=true 但没有收集到参数，说明这是新对话，不应该使用之前的参数
			log.Info("reset_context=true，不使用之前收集的参数，开始新对话",
				zap.String("rule_name", ruleMatch.Rule.RuleName))
		}

		// 直接使用规则SQL（支持单SQL和多SQL）
		var (
			context  *service.QueryContext
			sqlSet   []model.SqlSetItem
			useLocal = ruleMatch.Rule.UseLocalDB == 1
			isMulti  bool
		)

		// 如果是非本地执行，且多轮对话中收集了db_name和table_name参数，则从参数中获取并查找数据源
		if !useLocal && len(collectedParams) > 0 && ruleMatch.Rule.MultiRoundEnabled == 1 {
			// 检查收集的参数中是否有db_name和table_name
			paramDbName := ""
			paramTableName := ""

			// 尝试从收集的参数中获取数据库名和表名（支持多种可能的参数名）
			for key, value := range collectedParams {
				keyLower := strings.ToLower(key)
				if keyLower == "db_name" || keyLower == "database_name" || keyLower == "database" || keyLower == "db" {
					paramDbName = value
					log.Info("从多轮对话参数中获取到数据库名", zap.String("key", key), zap.String("value", value))
				}
				if keyLower == "table_name" || keyLower == "table" {
					paramTableName = value
					log.Info("从多轮对话参数中获取到表名", zap.String("key", key), zap.String("value", value))
				}
			}

			// 如果从参数中获取到了数据库名或表名，且当前没有数据源ID，则查找数据源
			if (paramDbName != "" || paramTableName != "") && req.DatasourceId <= 0 {
				log.Info("多轮对话中收集到数据库/表信息，尝试查找数据源",
					zap.String("db_name", paramDbName),
					zap.String("table_name", paramTableName))

				foundDatasourceId, findErr := service.FindDatasourceByDatabaseAndTable(paramDbName, paramTableName)
				if findErr == nil {
					datasourceId = foundDatasourceId
					databaseName = paramDbName
					tableName = paramTableName
					log.Info("通过多轮对话参数成功找到数据源",
						zap.Int("datasource_id", datasourceId),
						zap.String("database", databaseName),
						zap.String("table", tableName))
				} else {
					log.Warn("通过多轮对话参数查找数据源失败",
						zap.Error(findErr),
						zap.String("database", paramDbName),
						zap.String("table", paramTableName))
				}
			}
		}

		// 解析sql_set（多SQL）
		sqlSet, parseErr := service.ParseSqlSet(ruleMatch.Rule)
		if parseErr != nil {
			log.Warn("解析SQL集合失败", zap.Error(parseErr))
		}
		if len(sqlSet) > 0 {
			isMulti = true
			log.Info("检测到多SQL集合", zap.Int("sql_count", len(sqlSet)))
			// 如果已收集参数，应用到多SQL集合中的每个SQL
			if len(collectedParams) > 0 && ruleMatch.Rule.MultiRoundEnabled == 1 {
				log.Debug("应用多轮参数到SQL集合")
				for i := range sqlSet {
					sqlSet[i].Sql = service.ApplyCollectedParams(sqlSet[i].Sql, collectedParams, ruleMatch.Rule.ParameterMapping)
				}
			}
		}

		// 如果需要上下文参数（如数据库名、表名），尝试从语义中提取
		if req.DatasourceId > 0 {
			log.Info("已提供数据源ID，尝试提取数据库和表名", zap.Int("datasource_id", req.DatasourceId))
			// 如果提供了数据源ID，尝试构建简单上下文以获取数据库名和表名
			// 但不需要获取完整的表结构信息
			if databaseName == "" && tableName == "" {
				extractedDb, extractedTable := service.ExtractDatabaseAndTableFromQuery(req.Question)
				if extractedDb != "" {
					databaseName = extractedDb
					log.Info("从问题中提取到数据库名", zap.String("database", databaseName))
				}
				if extractedTable != "" {
					tableName = extractedTable
					log.Info("从问题中提取到表名", zap.String("table", tableName))
				}
			}

			// 构建简单上下文（不获取表结构）
			if databaseName != "" || tableName != "" {
				// 查询数据源信息以获取host和port
				var datasource model.Datasource
				result := database.DB.Where("id = ?", req.DatasourceId).First(&datasource)
				if result.Error == nil {
					context = &service.QueryContext{
						DatasourceId:   req.DatasourceId,
						DatasourceType: datasource.Type,
						Host:           datasource.Host,
						Port:           datasource.Port,
						DatabaseName:   databaseName,
						TableName:      tableName,
					}
				}
			}
		} else if ruleMatch.Rule.UseLocalDB == 1 {
			// 本地MySQL规则，从规则SQL中提取数据库和表名（如果需要）
			log.Info("使用本地MySQL规则，从SQL模板中提取数据库和表名")
			ruleSQLTemplate := service.GenerateSQLFromRule(ruleMatch.Rule, ruleMatch.Params, nil)
			extractedDb, extractedTable := service.ExtractDatabaseAndTableFromSQL(ruleSQLTemplate)
			if extractedTable != "" {
				tableName = extractedTable
				log.Info("从SQL模板中提取到表名", zap.String("table", tableName))
			}
			if extractedDb == "" {
				databaseName = service.GetLocalDatabaseName()
				log.Info("使用本地数据库名", zap.String("database", databaseName))
			} else {
				databaseName = extractedDb
				log.Info("从SQL模板中提取到数据库名", zap.String("database", databaseName))
			}
		} else if !useLocal && datasourceId <= 0 {
			// 远程且未指定数据源，尝试从语义提取
			// 如果请求中提供了数据源ID，直接使用
			if req.DatasourceId > 0 {
				datasourceId = req.DatasourceId
				log.Info("使用请求中提供的数据源ID", zap.Int("datasource_id", datasourceId))
			} else {
				// 如果没有提供数据源ID，尝试从语义提取
				// 注意：如果已经从多轮对话参数中设置了databaseName和tableName，这里不会重复提取
				log.Info("远程规则且未指定数据源，尝试从语义提取数据库和表名")
				if databaseName == "" && tableName == "" {
					extractedDb, extractedTable := service.ExtractDatabaseAndTableFromQuery(req.Question)
					if extractedDb != "" {
						databaseName = extractedDb
						log.Info("从问题中提取到数据库名", zap.String("database", databaseName))
					}
					if extractedTable != "" {
						tableName = extractedTable
						log.Info("从问题中提取到表名", zap.String("table", tableName))
					}
				}
				if databaseName != "" || tableName != "" {
					log.Info("根据数据库和表名查找数据源", zap.String("database", databaseName), zap.String("table", tableName))
					foundDatasourceId, findErr := service.FindDatasourceByDatabaseAndTable(databaseName, tableName)
					if findErr == nil {
						datasourceId = foundDatasourceId
						log.Info("成功找到数据源", zap.Int("datasource_id", datasourceId))
					} else {
						log.Warn("查找数据源失败", zap.Error(findErr), zap.String("database", databaseName), zap.String("table", tableName))
					}
				}
			}
		}

		// 直接使用规则生成SQL（单SQL）或使用SQL集（多SQL）
		if !isMulti {
			sqlQuery = service.GenerateSQLFromRule(ruleMatch.Rule, ruleMatch.Params, context)
			log.Info("规则匹配成功，直接使用规则SQL", zap.String("rule", ruleMatch.Rule.RuleName), zap.String("sql", sqlQuery), zap.Int8("use_local_db", ruleMatch.Rule.UseLocalDB))
		} else {
			log.Info("规则匹配成功，使用多SQL集合", zap.String("rule", ruleMatch.Rule.RuleName), zap.Int("sql_count", len(sqlSet)), zap.Int8("use_local_db", ruleMatch.Rule.UseLocalDB))
		}

		// 执行SQL查询
		if isMulti {
			// 多SQL执行
			log.Info("开始执行多SQL集合",
				zap.Int("sql_count", len(sqlSet)),
				zap.Int("datasource_id", datasourceId),
				zap.Bool("use_local", useLocal))
			results, execErr := service.ExecuteSQLSet(sqlSet, datasourceId, useLocal)
			if execErr != nil {
				log.Error("执行多SQL集合失败", zap.Error(execErr))
				err = execErr
			} else {
				log.Info("多SQL集合执行成功", zap.Int("result_count", len(results)))
				// 聚合指标
				metrics := make(map[string]interface{})
				for idx, item := range sqlSet {
					if len(item.Outputs) > 0 && len(results) > idx && len(results[idx].Rows) > 0 {
						row := results[idx].Rows[0]
						for k, v := range item.Outputs {
							if val, ok := row[v]; ok {
								metrics[k] = val
							}
						}
					}
				}
				log.Info("聚合指标完成", zap.Any("metrics", metrics), zap.Int("metric_count", len(metrics)))
				// 生成报告（使用AI基于查询数据和report_template生成）
				sqlList := make([]string, 0, len(results))
				allQueryResults := make([][]map[string]interface{}, 0, len(results))
				for _, r := range results {
					sqlList = append(sqlList, r.SQL)
					allQueryResults = append(allQueryResults, r.Rows)
				}
				log.Info("开始生成报告", zap.Int("sql_count", len(sqlList)), zap.Int("result_set_count", len(allQueryResults)))
				// 传入所有查询结果数据，让AI生成完整报告
				if len(allQueryResults) > 0 {
					answer = service.RenderReport(ruleMatch.Rule.ReportTemplate, metrics, sqlList, allQueryResults...)
				} else {
					answer = service.RenderReport(ruleMatch.Rule.ReportTemplate, metrics, sqlList)
				}
				log.Info("报告生成完成", zap.Int("answer_length", len(answer)))
				// 返回第一个结果作为表格数据（如果有）
				if len(results) > 0 {
					queryResult = results[0].Rows
					sqlQuery = results[0].SQL
				}
			}
		} else {
			if ruleMatch.Rule.UseLocalDB == 1 {
				log.Info("执行本地MySQL查询", zap.String("sql", sqlQuery))
				queryResult, err = service.ExecuteLocalQuery(sqlQuery)
				if err != nil {
					log.Error("本地MySQL查询失败", zap.Error(err), zap.String("sql", sqlQuery))
				} else {
					log.Info("本地MySQL查询成功", zap.Int("row_count", len(queryResult)))
				}
			} else {
				// 使用远程数据源执行，需要查找数据源
				// 如果还没有数据源ID，尝试查找
				if datasourceId <= 0 {
					// 如果请求中提供了数据源ID，直接使用
					if req.DatasourceId > 0 {
						datasourceId = req.DatasourceId
						log.Info("使用请求中提供的数据源ID", zap.Int("datasource_id", datasourceId))
					} else {
						// 如果没有提供数据源ID，尝试从语义中提取数据库和表名，然后查找数据源
						// 注意：如果已经从多轮对话参数中设置了databaseName和tableName，这里不会重复提取
						if databaseName == "" && tableName == "" {
							extractedDb, extractedTable := service.ExtractDatabaseAndTableFromQuery(req.Question)
							if extractedDb != "" {
								databaseName = extractedDb
								log.Info("从问题中提取到数据库名", zap.String("database", databaseName))
							}
							if extractedTable != "" {
								tableName = extractedTable
								log.Info("从问题中提取到表名", zap.String("table", tableName))
							}
						}

						// 如果提取到了数据库名或表名，尝试查找数据源
						if databaseName != "" || tableName != "" {
							foundDatasourceId, findErr := service.FindDatasourceByDatabaseAndTable(databaseName, tableName)
							if findErr != nil {
								log.Warn("根据数据库和表名查找数据源失败", zap.Error(findErr), zap.String("database", databaseName), zap.String("table", tableName))
								answer = fmt.Sprintf("无法找到对应的数据源。请在界面选择数据源，或在提问中明确数据库/表信息。详细错误: %s", findErr.Error())
							} else {
								datasourceId = foundDatasourceId
								log.Info("通过元数据表找到数据源", zap.Int("datasource_id", datasourceId), zap.String("database", databaseName), zap.String("table", tableName))
							}
						} else {
							answer = "无法从问题中识别数据库或表信息。请在界面选择数据源，或在提问中明确数据库名/表名。"
						}
					}
				} else {
					// 已经有数据源ID（可能是从多轮对话参数中获取的），直接使用
					log.Info("使用已找到的数据源ID", zap.Int("datasource_id", datasourceId), zap.String("database", databaseName), zap.String("table", tableName))
				}

				// 如果找到了数据源ID，执行查询
				if datasourceId > 0 && answer == "" {
					log.Info("执行远程数据源查询", zap.Int("datasource_id", datasourceId), zap.String("sql", sqlQuery))
					queryResult, err = service.ExecuteQuery(sqlQuery, datasourceId)
					if err != nil {
						log.Error("远程数据源查询失败", zap.Error(err), zap.Int("datasource_id", datasourceId), zap.String("sql", sqlQuery))
					} else {
						log.Info("远程数据源查询成功", zap.Int("datasource_id", datasourceId), zap.Int("row_count", len(queryResult)))
					}
				}
			}
		}

		// 处理查询结果
		if answer == "" {
			if err != nil {
				log.Error("执行SQL查询失败", zap.Error(err), zap.String("sql", sqlQuery))
				answer = fmt.Sprintf("执行查询失败: %v\n\n生成的SQL: %s", err, sqlQuery)
			} else {
				// 生成报告：从首行构造metrics（简单取首行字段），或空
				metrics := make(map[string]interface{})
				if len(queryResult) > 0 {
					for k, v := range queryResult[0] {
						metrics[k] = v
					}
				}
				log.Info("开始生成单SQL报告", zap.Int("row_count", len(queryResult)), zap.Int("metric_count", len(metrics)))
				// 传入完整的查询结果数据，让AI生成完整报告
				answer = service.RenderReport(ruleMatch.Rule.ReportTemplate, metrics, []string{sqlQuery}, queryResult)
				log.Info("单SQL报告生成完成", zap.Int("answer_length", len(answer)))
			}
		}
	} else {
		// 规则匹配失败，使用AI生成SQL（Agent模式必须执行SQL，不能返回纯文本）
		log.Info("规则匹配失败，使用AI生成SQL")
		// 如果没有提供数据源ID，尝试从语义中提取数据库和表名，然后查找数据源
		if req.DatasourceId <= 0 {
			// 如果请求中没有提供数据库名和表名，尝试从语义中提取
			if databaseName == "" && tableName == "" {
				log.Debug("尝试从问题中提取数据库和表名")
				extractedDb, extractedTable := service.ExtractDatabaseAndTableFromQuery(req.Question)
				if extractedDb != "" {
					databaseName = extractedDb
					log.Info("从问题中提取到数据库名", zap.String("database", databaseName))
				}
				if extractedTable != "" {
					tableName = extractedTable
					log.Info("从问题中提取到表名", zap.String("table", tableName))
				}
			}

			// 如果提取到了数据库名或表名，尝试查找数据源
			if databaseName != "" || tableName != "" {
				log.Info("根据数据库和表名查找数据源", zap.String("database", databaseName), zap.String("table", tableName))
				foundDatasourceId, findErr := service.FindDatasourceByDatabaseAndTable(databaseName, tableName)
				if findErr != nil {
					log.Warn("根据数据库和表名查找数据源失败", zap.Error(findErr), zap.String("database", databaseName), zap.String("table", tableName))
					// Agent模式：即使找不到数据源，也要尝试AI生成SQL并使用本地MySQL执行
					log.Info("尝试使用AI生成SQL并使用本地MySQL执行")
					// 构建一个空的上下文用于AI生成SQL
					context := &service.QueryContext{}
					generator := &service.AISQLGenerator{}
					sqlQuery, err = generator.GenerateSQL(req.Question, context)
					if err != nil {
						log.Error("AI生成SQL失败", zap.Error(err))
						answer = fmt.Sprintf("无法找到对应的数据源，且AI生成SQL失败。%s", findErr.Error())
					} else {
						log.Info("AI生成SQL成功，尝试使用本地MySQL执行", zap.String("sql", sqlQuery))
						// 尝试使用本地MySQL执行
						queryResult, err = service.ExecuteLocalQuery(sqlQuery)
						if err != nil {
							log.Error("本地MySQL查询失败", zap.Error(err), zap.String("sql", sqlQuery))
							answer = fmt.Sprintf("无法找到对应的数据源，且本地MySQL查询失败。%s\n\n生成的SQL: %s\n\n错误: %v", findErr.Error(), sqlQuery, err)
						} else {
							log.Info("本地MySQL查询成功", zap.Int("row_count", len(queryResult)))
							// 使用AI生成报告，而不是直接格式化结果
							metrics := make(map[string]interface{})
							if len(queryResult) > 0 {
								for k, v := range queryResult[0] {
									metrics[k] = v
								}
							}
							log.Info("开始生成AI报告（规则匹配失败，使用本地MySQL）", zap.Int("row_count", len(queryResult)))
							answer = service.RenderReport("", metrics, []string{sqlQuery}, queryResult)
							log.Info("AI报告生成完成", zap.Int("answer_length", len(answer)))
						}
					}
				} else {
					datasourceId = foundDatasourceId
					log.Info("通过元数据表找到数据源", zap.Int("datasource_id", datasourceId), zap.String("database", databaseName), zap.String("table", tableName))
				}
			} else {
				// 无法提取数据库和表名，尝试使用AI生成SQL并使用本地MySQL执行
				log.Info("无法提取数据库和表名，尝试使用AI生成SQL并使用本地MySQL执行")
				context := &service.QueryContext{}
				generator := &service.AISQLGenerator{}
				sqlQuery, err = generator.GenerateSQL(req.Question, context)
				if err != nil {
					log.Error("AI生成SQL失败", zap.Error(err))
					answer = "无法从问题中识别数据库或表信息，且AI生成SQL失败。请明确指定数据库名和表名。"
				} else {
					log.Info("AI生成SQL成功", zap.String("sql", sqlQuery))
					// 尝试使用本地MySQL执行
					queryResult, err = service.ExecuteLocalQuery(sqlQuery)
					if err != nil {
						log.Error("本地MySQL查询失败", zap.Error(err), zap.String("sql", sqlQuery))
						answer = fmt.Sprintf("无法从问题中识别数据库或表信息，且本地MySQL查询失败。\n\n生成的SQL: %s\n\n错误: %v", sqlQuery, err)
					} else {
						log.Info("本地MySQL查询成功", zap.Int("row_count", len(queryResult)))
						// 使用AI生成报告，而不是直接格式化结果
						metrics := make(map[string]interface{})
						if len(queryResult) > 0 {
							for k, v := range queryResult[0] {
								metrics[k] = v
							}
						}
						log.Info("开始生成AI报告（规则匹配失败，使用本地MySQL）", zap.Int("row_count", len(queryResult)))
						answer = service.RenderReport("", metrics, []string{sqlQuery}, queryResult)
						log.Info("AI报告生成完成", zap.Int("answer_length", len(answer)))
					}
				}
			}
		} else {
			datasourceId = req.DatasourceId
		}

		// 如果找到了数据源ID，尝试使用AI生成SQL并执行查询
		if datasourceId > 0 && answer == "" {
			log.Info("开始构建查询上下文", zap.Int("datasource_id", datasourceId), zap.String("database", databaseName), zap.String("table", tableName))
			// 构建查询上下文
			context, err := service.BuildQueryContext(req.SessionId, databaseName, tableName)
			if err != nil {
				log.Error("构建查询上下文失败", zap.Error(err))
				answer = fmt.Sprintf("构建查询上下文失败: %v", err)
			} else {
				log.Info("查询上下文构建成功",
					zap.Int("datasource_id", datasourceId),
					zap.String("database", databaseName),
					zap.String("table", tableName),
					zap.Int("column_count", len(context.Metadata.Columns)))
				// 使用AI生成SQL
				generator := &service.AISQLGenerator{}
				sqlQuery, err = generator.GenerateSQL(req.Question, context)
				if err != nil {
					log.Error("AI生成SQL失败", zap.Error(err))
					answer = fmt.Sprintf("生成SQL失败: %v", err)
				} else {
					log.Info("AI生成SQL成功", zap.String("sql", sqlQuery))

					// 如果之前没有找到数据源，尝试从生成的SQL中提取数据库和表名
					if datasourceId <= 0 {
						extractedDb, extractedTable := service.ExtractDatabaseAndTableFromSQL(sqlQuery)
						if extractedDb != "" {
							databaseName = extractedDb
						}
						if extractedTable != "" {
							tableName = extractedTable
						}

						// 如果从SQL中提取到了数据库或表名，再次尝试查找数据源
						if databaseName != "" || tableName != "" {
							foundDatasourceId, findErr := service.FindDatasourceByDatabaseAndTable(databaseName, tableName)
							if findErr == nil {
								datasourceId = foundDatasourceId
								log.Info("从生成的SQL中找到数据源", zap.Int("datasource_id", datasourceId), zap.String("database", databaseName), zap.String("table", tableName))
								// 重新构建上下文以获取完整的表结构信息
								context, err = service.BuildQueryContext(req.SessionId, databaseName, tableName)
								if err != nil {
									log.Warn("重新构建查询上下文失败", zap.Error(err))
								} else {
									log.Info("已重新获取表结构信息",
										zap.String("database", databaseName),
										zap.String("table", tableName),
										zap.Int("column_count", len(context.Metadata.Columns)))
									// 如果表结构信息已更新，使用新的上下文重新生成SQL以提高准确性
									if tableName != "" && len(context.Metadata.Columns) > 0 {
										log.Info("使用更新后的表结构信息重新生成SQL",
											zap.String("database", databaseName),
											zap.String("table", tableName))
										// 重新生成SQL（使用包含完整表结构的上下文）
										newSqlQuery, genErr := generator.GenerateSQL(req.Question, context)
										if genErr == nil && newSqlQuery != "" {
											sqlQuery = newSqlQuery
											log.Info("使用表结构信息重新生成SQL成功", zap.String("new_sql", sqlQuery))
										} else {
											log.Warn("使用表结构信息重新生成SQL失败，使用原始SQL", zap.Error(genErr))
										}
									}
								}
							} else {
								log.Warn("从SQL中提取数据库和表名后仍无法找到数据源", zap.Error(findErr), zap.String("database", databaseName), zap.String("table", tableName))
							}
						}
					} else {
						// 如果已经有数据源ID，但表结构信息可能不完整，检查并更新
						if tableName != "" && (context == nil || len(context.Metadata.Columns) == 0) {
							log.Info("表结构信息不完整，重新获取",
								zap.String("database", databaseName),
								zap.String("table", tableName))
							context, err = service.BuildQueryContext(req.SessionId, databaseName, tableName)
							if err == nil && len(context.Metadata.Columns) > 0 {
								log.Info("已获取完整表结构信息，使用更新后的上下文重新生成SQL",
									zap.String("database", databaseName),
									zap.String("table", tableName),
									zap.Int("column_count", len(context.Metadata.Columns)))
								// 使用包含完整表结构的上下文重新生成SQL
								newSqlQuery, genErr := generator.GenerateSQL(req.Question, context)
								if genErr == nil && newSqlQuery != "" {
									sqlQuery = newSqlQuery
									log.Info("使用完整表结构信息重新生成SQL成功", zap.String("new_sql", sqlQuery))
								} else {
									log.Warn("使用完整表结构信息重新生成SQL失败，使用原始SQL", zap.Error(genErr))
								}
							}
						}
					}
				}

				// 执行SQL查询（只有在找到数据源时才执行）
				if sqlQuery != "" && datasourceId > 0 {
					log.Info("执行远程数据源查询", zap.Int("datasource_id", datasourceId), zap.String("sql", sqlQuery))
					queryResult, err = service.ExecuteQuery(sqlQuery, datasourceId)
					if err != nil {
						log.Error("执行SQL查询失败", zap.Error(err), zap.Int("datasource_id", datasourceId), zap.String("sql", sqlQuery))
						answer = fmt.Sprintf("执行查询失败: %v\n\n生成的SQL: %s", err, sqlQuery)
					} else {
						log.Info("远程数据源查询成功", zap.Int("datasource_id", datasourceId), zap.Int("row_count", len(queryResult)))
						// 使用AI生成报告，而不是直接格式化结果
						metrics := make(map[string]interface{})
						if len(queryResult) > 0 {
							for k, v := range queryResult[0] {
								metrics[k] = v
							}
						}
						log.Info("开始生成AI报告（规则匹配失败，使用远程数据源）", zap.Int("row_count", len(queryResult)))
						answer = service.RenderReport("", metrics, []string{sqlQuery}, queryResult)
						log.Info("AI报告生成完成", zap.Int("answer_length", len(answer)))
					}
				} else if sqlQuery != "" && datasourceId <= 0 {
					// 生成了SQL但没有找到数据源，尝试使用本地MySQL
					log.Info("生成了SQL但未找到数据源，尝试使用本地MySQL执行", zap.String("sql", sqlQuery))
					queryResult, err = service.ExecuteLocalQuery(sqlQuery)
					if err != nil {
						log.Error("本地MySQL查询失败", zap.Error(err), zap.String("sql", sqlQuery))
						answer = fmt.Sprintf("已生成SQL，但无法找到对应的数据源，且本地MySQL查询也失败。\n\n**生成的SQL:**\n```sql\n%s\n```\n\n错误: %v\n\n请确保数据库和表名正确，或联系管理员检查元数据配置。", sqlQuery, err)
					} else {
						log.Info("本地MySQL查询成功", zap.Int("row_count", len(queryResult)))
						// 使用AI生成报告，而不是直接格式化结果
						metrics := make(map[string]interface{})
						if len(queryResult) > 0 {
							for k, v := range queryResult[0] {
								metrics[k] = v
							}
						}
						log.Info("开始生成AI报告（规则匹配失败，使用本地MySQL）", zap.Int("row_count", len(queryResult)))
						answer = service.RenderReport("", metrics, []string{sqlQuery}, queryResult)
						log.Info("AI报告生成完成", zap.Int("answer_length", len(answer)))
					}
				}
			}
		} else if answer == "" {
			// Agent模式：必须通过SQL执行获取答案，不能返回纯文本
			answer = "无法识别数据库信息，请明确指定数据库名和表名，或确保已配置相应的语义规则。"
		}
	}

	// 保存助手回复
	err = service.SaveMessage(req.SessionId, "assistant", answer, sqlQuery, queryResult)
	if err != nil {
		log.Error("保存助手消息失败", zap.Error(err))
	} else {
		log.Debug("助手消息已保存", zap.String("session_id", req.SessionId))
	}

	// 构造响应
	response := ChatQueryResponse{
		Answer:      answer,
		SqlQuery:    sqlQuery,
		QueryResult: queryResult,
		Timestamp:   time.Now().Unix(),
	}

	log.Info("ChatQuery处理完成",
		zap.String("session_id", req.SessionId),
		zap.Bool("has_sql", sqlQuery != ""),
		zap.Int("result_row_count", len(queryResult)),
		zap.Int("answer_length", len(answer)))

	c.JSON(200, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetSessions 获取会话列表
func GetSessions(c *gin.Context) {
	username, _ := c.Get("username")
	userName := username.(string)

	sessions, err := service.ListSessions(userName)
	if err != nil {
		log.Error("获取会话列表失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "获取会话列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    sessions,
	})
}

// CreateSession 创建新会话
func CreateSession(c *gin.Context) {
	username, _ := c.Get("username")
	userName := username.(string)

	session, err := service.CreateSession(userName)
	if err != nil {
		log.Error("创建会话失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "创建会话失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    session,
	})
}

// DeleteSession 删除会话
func DeleteSession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	username, _ := c.Get("username")
	userName := username.(string)

	err := service.DeleteSession(sessionId, userName)
	if err != nil {
		log.Error("删除会话失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "删除会话失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// GetSessionMessages 获取会话消息历史
func GetSessionMessages(c *gin.Context) {
	sessionId := c.Param("sessionId")
	username, _ := c.Get("username")
	userName := username.(string)

	// 验证会话是否属于当前用户
	session, err := service.GetSession(sessionId)
	if err != nil {
		c.JSON(404, gin.H{
			"success": false,
			"message": "会话不存在",
		})
		return
	}

	if session.UserName != userName {
		c.JSON(403, gin.H{
			"success": false,
			"message": "无权限访问此会话",
		})
		return
	}

	messages, err := service.GetSessionMessages(sessionId, 0)
	if err != nil {
		log.Error("获取消息历史失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "获取消息历史失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    messages,
	})
}

// UpdateSessionTitle 更新会话标题
func UpdateSessionTitle(c *gin.Context) {
	sessionId := c.Param("sessionId")
	username, _ := c.Get("username")
	userName := username.(string)

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 验证会话是否属于当前用户
	session, err := service.GetSession(sessionId)
	if err != nil {
		c.JSON(404, gin.H{
			"success": false,
			"message": "会话不存在",
		})
		return
	}

	if session.UserName != userName {
		c.JSON(403, gin.H{
			"success": false,
			"message": "无权限访问此会话",
		})
		return
	}

	err = service.UpdateSessionTitle(sessionId, req.Title)
	if err != nil {
		log.Error("更新会话标题失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "更新会话标题失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新成功",
	})
}

// GetRules 获取语义规则列表
func GetRules(c *gin.Context) {
	var rules []model.SemanticSqlRule
	result := database.DB.Order("priority DESC, id DESC").Find(&rules)
	if result.Error != nil {
		log.Error("获取规则列表失败", zap.Error(result.Error))
		c.JSON(500, gin.H{
			"success": false,
			"message": "获取规则列表失败",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    rules,
	})
}

// CreateRule 创建规则
func CreateRule(c *gin.Context) {
	var rule model.SemanticSqlRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	result := database.DB.Create(&rule)
	if result.Error != nil {
		log.Error("创建规则失败", zap.Error(result.Error))
		c.JSON(500, gin.H{
			"success": false,
			"message": "创建规则失败",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    rule,
	})
}

// UpdateRule 更新规则
func UpdateRule(c *gin.Context) {
	id := c.Param("id")
	var rule model.SemanticSqlRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	result := database.DB.Model(&model.SemanticSqlRule{}).Where("id = ?", id).Updates(&rule)
	if result.Error != nil {
		log.Error("更新规则失败", zap.Error(result.Error))
		c.JSON(500, gin.H{
			"success": false,
			"message": "更新规则失败",
			"error":   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{
			"success": false,
			"message": "规则不存在",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新成功",
	})
}

// DeleteRule 删除规则
func DeleteRule(c *gin.Context) {
	id := c.Param("id")

	result := database.DB.Delete(&model.SemanticSqlRule{}, id)
	if result.Error != nil {
		log.Error("删除规则失败", zap.Error(result.Error))
		c.JSON(500, gin.H{
			"success": false,
			"message": "删除规则失败",
			"error":   result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{
			"success": false,
			"message": "规则不存在",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// GetRecommendedRules 获取推荐的规则列表（从semantic_sql_rules表获取，按权重排序）
func GetRecommendedRules(c *gin.Context) {
	var rules []model.SemanticSqlRule

	// 查询启用的规则，按Priority降序排序（权重高的排在前面）
	result := database.DB.Where("enabled = ?", 1).
		Order("priority DESC").
		Limit(10).
		Find(&rules)

	if result.Error != nil {
		log.Error("查询推荐规则失败", zap.Error(result.Error))
		c.JSON(500, gin.H{
			"success": false,
			"message": "查询推荐规则失败",
			"error":   result.Error.Error(),
		})
		return
	}

	// 构造响应数据
	type RuleItem struct {
		RuleName string `json:"rule_name"`
		Priority int    `json:"priority"`
	}

	ruleList := make([]RuleItem, 0, len(rules))
	for _, rule := range rules {
		ruleList = append(ruleList, RuleItem{
			RuleName: rule.RuleName,
			Priority: rule.Priority,
		})
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    ruleList,
	})
}
