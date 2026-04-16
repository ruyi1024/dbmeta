/*
Copyright 2024 The Lepus Team Group, website: https://www.lepus.cc
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
	"dbmeta-core/src/model"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// MultiRoundContext 多轮对话上下文
type MultiRoundContext struct {
	RuleId      int64             `json:"rule_id"`      // 规则ID
	RuleName    string            `json:"rule_name"`    // 规则名称
	Collected   map[string]string `json:"collected"`    // 已收集的参数，key为参数名，value为用户输入的值
	CurrentStep int               `json:"current_step"` // 当前步骤索引（从0开始）
}

// CheckMultiRoundInfo 检查多轮信息收集状态
// 返回：是否需要继续收集信息，下一个问题，已收集的参数，当前步骤的选项（如果是select类型）
// resetContext: 如果为true，忽略历史上下文，从头开始新的多轮对话
func CheckMultiRoundInfo(sessionId string, rule *model.SemanticSqlRule, userInput string, resetContext bool) (needMore bool, nextQuestion string, collectedParams map[string]string, options []string, err error) {
	// 如果规则未启用多轮对话，直接返回
	if rule.MultiRoundEnabled != 1 || len(rule.QuestionFlow) == 0 {
		return false, "", nil, nil, nil
	}

	// 从历史消息中提取已收集的参数
	collectedParams = make(map[string]string)
	var context *MultiRoundContext

	// 如果 resetContext=true，跳过上下文查找，从头开始
	if !resetContext {
		// 获取会话历史消息，查找多轮对话上下文
		messages, err := GetSessionMessages(sessionId, 20) // 获取最近20条消息
		if err != nil {
			return false, "", nil, nil, fmt.Errorf("获取会话历史失败: %v", err)
		}

		// 查找最近的多轮对话上下文（从assistant消息中查找）
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]
			if msg.Role == "assistant" {
				// 尝试从消息内容中解析多轮对话上下文
				// 格式：<!--MULTI_ROUND_CONTEXT:{"rule_id":1,"rule_name":"查询用户详细数据",...}-->
				re := regexp.MustCompile(`<!--MULTI_ROUND_CONTEXT:(.+?)-->`)
				matches := re.FindStringSubmatch(msg.Content)
				if len(matches) > 1 {
					if err := json.Unmarshal([]byte(matches[1]), &context); err == nil {
						// 验证上下文是否属于当前规则，避免不同规则的多轮对话相互串
						if context.RuleId == int64(rule.Id) {
							// 找到了匹配的上下文，使用已收集的参数
							collectedParams = context.Collected
							break
						} else {
							// 上下文属于其他规则，忽略它，继续查找或当作第一次匹配
							context = nil
						}
					}
				}
			}
		}
	}

	// 如果没有找到上下文，说明这是第一次匹配到多轮规则
	isFirstMatch := context == nil
	if isFirstMatch {
		context = &MultiRoundContext{
			RuleId:      int64(rule.Id),
			RuleName:    rule.RuleName,
			Collected:   make(map[string]string),
			CurrentStep: 0,
		}
	}

	// 处理当前用户输入
	// 如果是第一次匹配，用户输入是查询意图，不需要处理
	// 如果不是第一次匹配，用户输入应该是回答上一个问题的答案
	if !isFirstMatch && userInput != "" && context.CurrentStep > 0 {
		// 当前步骤应该是收集上一个问题的答案
		prevStepIndex := context.CurrentStep - 1
		if prevStepIndex >= 0 && prevStepIndex < len(rule.QuestionFlow) {
			prevStep := rule.QuestionFlow[prevStepIndex]
			// 验证用户输入（支持通过SQL获取的选项）
			if validateInput(userInput, prevStep, collectedParams, rule.ParameterMapping) {
				collectedParams[prevStep.Key] = userInput
				context.Collected[prevStep.Key] = userInput
				// 不立即增加CurrentStep，让下面的检查循环来决定下一个要收集的参数
				// 这样可以确保所有必填参数都被检查到
			} else {
				// 输入验证失败，返回错误提示
				contextJSON, _ := json.Marshal(context)
				contextMark := fmt.Sprintf("<!--MULTI_ROUND_CONTEXT:%s-->", string(contextJSON))
				// 返回当前步骤的选项（如果是select类型）
				var stepOptions []string
				if prevStep.Type == "select" {
					// 如果配置了SQL，通过SQL获取选项
					if prevStep.OptionsSQL != "" {
						options, sqlErr := getOptionsFromSQL(prevStep.OptionsSQL, collectedParams, rule.ParameterMapping)
						if sqlErr == nil && len(options) > 0 {
							stepOptions = options
						} else if len(prevStep.Options) > 0 {
							// SQL获取失败，使用静态选项
							stepOptions = prevStep.Options
						}
					} else if len(prevStep.Options) > 0 {
						// 使用静态选项列表
						stepOptions = prevStep.Options
					}
				}
				return true, fmt.Sprintf("输入格式不正确，%s%s", prevStep.Question, contextMark), collectedParams, stepOptions, nil
			}
		}
	}

	// 检查是否还有未收集的必填参数
	// 从0开始检查所有参数，确保不会跳过任何必填参数
	for i := 0; i < len(rule.QuestionFlow); i++ {
		step := rule.QuestionFlow[i]
		if step.Required {
			if _, ok := collectedParams[step.Key]; !ok {
				// 找到第一个未收集的必填参数，返回对应的问题
				question := step.Question

				// 如果是select类型，获取选项列表
				var stepOptions []string
				if step.Type == "select" {
					// 如果配置了SQL，通过SQL获取选项
					if step.OptionsSQL != "" {
						options, sqlErr := getOptionsFromSQL(step.OptionsSQL, collectedParams, rule.ParameterMapping)
						if sqlErr == nil && len(options) > 0 {
							stepOptions = options
							question += fmt.Sprintf("（可选：%s）", strings.Join(stepOptions, "、"))
						} else {
							// SQL获取失败，使用静态选项
							if len(step.Options) > 0 {
								stepOptions = step.Options
								question += fmt.Sprintf("（可选：%s）", strings.Join(stepOptions, "、"))
							}
						}
					} else if len(step.Options) > 0 {
						// 使用静态选项列表
						stepOptions = step.Options
						question += fmt.Sprintf("（可选：%s）", strings.Join(stepOptions, "、"))
					}
				}

				// 更新上下文，标记当前步骤（下一步将收集这个参数）
				context.CurrentStep = i + 1
				// 保存上下文到消息中（通过特殊标记）
				contextJSON, _ := json.Marshal(context)
				contextMark := fmt.Sprintf("<!--MULTI_ROUND_CONTEXT:%s-->", string(contextJSON))
				return true, question + contextMark, collectedParams, stepOptions, nil
			}
		}
	}

	// 所有必填参数已收集完成
	return false, "", collectedParams, nil, nil
}

// validateInput 验证用户输入
// collectedParams: 已收集的参数（用于SQL选项的动态获取）
// parameterMapping: 参数映射配置
func validateInput(input string, step model.QuestionFlowItem, collectedParams map[string]string, parameterMapping model.ParameterMapping) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return !step.Required
	}

	// 根据类型验证
	switch step.Type {
	case "select":
		// 获取选项列表（优先使用SQL，失败则使用静态选项）
		var options []string
		if step.OptionsSQL != "" {
			// 尝试通过SQL获取选项
			sqlOptions, sqlErr := getOptionsFromSQL(step.OptionsSQL, collectedParams, parameterMapping)
			if sqlErr == nil && len(sqlOptions) > 0 {
				options = sqlOptions
			} else if len(step.Options) > 0 {
				// SQL获取失败，使用静态选项
				options = step.Options
			}
		} else if len(step.Options) > 0 {
			// 使用静态选项列表
			options = step.Options
		}

		// 检查输入是否在选项中（不区分大小写）
		for _, opt := range options {
			if strings.EqualFold(input, opt) {
				return true
			}
		}
		return false
	case "number":
		// 验证是否为数字
		matched, _ := regexp.MatchString(`^-?\d+(\.\d+)?$`, input)
		return matched
	case "email":
		// 验证邮箱格式
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, input)
		return matched
	default:
		// text类型，基本验证通过
		return true
	}
}

// ApplyCollectedParams 将收集的参数应用到SQL模板
func ApplyCollectedParams(sqlTemplate string, collectedParams map[string]string, parameterMapping model.ParameterMapping) string {
	result := sqlTemplate

	// 如果提供了参数映射，先进行映射转换
	mappedParams := make(map[string]string)
	if len(parameterMapping) > 0 {
		for key, value := range collectedParams {
			if mappedKey, ok := parameterMapping[key]; ok {
				mappedParams[mappedKey] = value
			} else {
				mappedParams[key] = value
			}
		}
	} else {
		mappedParams = collectedParams
	}

	// 替换SQL模板中的占位符
	// 支持 {{key}} 和 {key} 两种格式
	for key, value := range mappedParams {
		// 检查是否是数据库名或表名参数（这些参数不应该加引号，因为它们是标识符）
		keyLower := strings.ToLower(key)
		isIdentifier := keyLower == "db_name" || keyLower == "database_name" || keyLower == "database" || keyLower == "db" ||
			keyLower == "table_name" || keyLower == "table" || keyLower == "schema_name" || keyLower == "schema"

		var finalValue string
		if isIdentifier {
			// 数据库名和表名是标识符，不加引号，直接使用
			finalValue = value
		} else {
			// 判断值是否需要加引号
			// 如果值已经是数字、NULL、或者包含SQL函数/表达式，不加引号
			needsQuotes := needsSQLQuotes(value)

			if needsQuotes {
				// 转义SQL注入风险字符，并添加单引号
				escapedValue := strings.ReplaceAll(value, "'", "''")
				finalValue = fmt.Sprintf("'%s'", escapedValue)
			} else {
				// 数字或SQL表达式，直接使用
				finalValue = value
			}
		}

		// 替换 {{key}} 格式
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), finalValue)
		// 替换 {key} 格式
		result = strings.ReplaceAll(result, fmt.Sprintf("{%s}", key), finalValue)
	}

	return result
}

// needsSQLQuotes 判断SQL值是否需要加引号
// 返回true表示需要加引号（字符串值），false表示不需要（数字、NULL、SQL表达式等）
func needsSQLQuotes(value string) bool {
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

// ExtractMultiRoundContext 从消息内容中提取多轮对话上下文
func ExtractMultiRoundContext(content string) (*MultiRoundContext, error) {
	re := regexp.MustCompile(`<!--MULTI_ROUND_CONTEXT:(.+?)-->`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil, fmt.Errorf("未找到多轮对话上下文")
	}

	var context MultiRoundContext
	if err := json.Unmarshal([]byte(matches[1]), &context); err != nil {
		return nil, fmt.Errorf("解析多轮对话上下文失败: %v", err)
	}

	return &context, nil
}

// getOptionsFromSQL 通过SQL获取选项列表
func getOptionsFromSQL(optionsSQL string, collectedParams map[string]string, parameterMapping model.ParameterMapping) ([]string, error) {
	// 如果SQL中包含占位符，先应用已收集的参数
	sql := ApplyCollectedParams(optionsSQL, collectedParams, parameterMapping)

	// 执行SQL查询（使用本地MySQL）
	results, err := ExecuteLocalQuery(sql)
	if err != nil {
		return nil, fmt.Errorf("执行选项SQL失败: %v", err)
	}

	// 从查询结果中提取选项
	// 假设SQL返回的第一列是选项值
	options := make([]string, 0, len(results))
	for _, row := range results {
		// 获取第一列的值
		for _, value := range row {
			if value != nil {
				options = append(options, fmt.Sprintf("%v", value))
				break // 只取第一列
			}
		}
	}

	return options, nil
}
