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
	"bytes"
	"dbmeta-core/log"
	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const aiGeneralColumnCommentBatchSize = 20

type aiColumnCommentBatchItem struct {
	ID        int    `json:"id"`
	AiComment string `json:"ai_comment"`
}

func init() {
	go aiGeneralColumnCommentCrontabTask()
}

func aiGeneralColumnCommentCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "ai_general_column_comment").Take(&record)
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "ai_general_column_comment").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='ai_general_column_comment'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doAiGeneralColumnCommentTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='ai_general_column_comment'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func doAiGeneralColumnCommentTask() {
	logger := log.Logger
	logger.Info("开始执行AI生成字段注释任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("ai_general_column_comment")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	invoker, err := newTableColumnCommentLLMInvoker(logger)
	if err != nil {
		errorMsg := fmt.Sprintf("初始化模型调用失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 查询需要生成注释的字段
	var columns []model.MetaColumn
	result := database.DB.Where("(ai_comment IS NULL OR ai_comment = '') AND is_deleted = 0").Find(&columns)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("查询字段数据失败: %v", result.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	if len(columns) == 0 {
		successMsg := "没有需要生成AI注释的字段"
		logger.Info(successMsg)
		taskLogger.Success(successMsg)
		return
	}

	logger.Info("找到需要生成AI注释的字段", zap.Int("count", len(columns)))
	taskLogger.UpdateResult(fmt.Sprintf("找到 %d 个需要生成AI注释的字段", len(columns)))

	successCount := 0
	failedCount := 0
	errorDetails := []string{}

	for start := 0; start < len(columns); start += aiGeneralColumnCommentBatchSize {
		end := start + aiGeneralColumnCommentBatchSize
		if end > len(columns) {
			end = len(columns)
		}
		chunk := columns[start:end]

		commentMap, batchErr := generateColumnCommentBatch(chunk, invoker)
		if batchErr != nil {
			msg := fmt.Sprintf("批次 %d-%d 调用模型失败: %v", start+1, end, batchErr)
			logger.Error(msg)
			errorDetails = append(errorDetails, msg)
			failedCount += len(chunk)
			taskLogger.UpdateResult(fmt.Sprintf("已处理 %d/%d 个字段 (成功: %d, 失败: %d)", successCount+failedCount, len(columns), successCount, failedCount))
			continue
		}

		for _, column := range chunk {
			comment := cleanColumnComment(commentMap[column.Id])
			if comment == "" {
				errorMsg := fmt.Sprintf("处理字段 %s.%s 失败: 模型未返回有效注释", column.TableNameX, column.ColumnName)
				logger.Error(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				failedCount++
				continue
			}
			updateResult := database.DB.Model(&model.MetaColumn{}).Where("id = ?", column.Id).Update("ai_comment", comment)
			if updateResult.Error != nil {
				errorMsg := fmt.Sprintf("处理字段 %s.%s 失败: 更新数据库失败: %v", column.TableNameX, column.ColumnName, updateResult.Error)
				logger.Error(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				failedCount++
				continue
			}
			logger.Info("成功为字段生成AI注释",
				zap.String("table_name", column.TableNameX),
				zap.String("column_name", column.ColumnName),
				zap.String("comment", comment))
			successCount++
		}
		taskLogger.UpdateResult(fmt.Sprintf("已处理 %d/%d 个字段 (成功: %d, 失败: %d)", successCount+failedCount, len(columns), successCount, failedCount))
	}

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 总计: %d, 成功: %d, 失败: %d", len(columns), successCount, failedCount)
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0]) // 只记录第一个错误详情
		if len(errorDetails) > 1 {
			finalResult += fmt.Sprintf(" 等 %d 个错误", len(errorDetails))
		}
	}

	if failedCount == 0 {
		taskLogger.Success(finalResult)
	} else {
		taskLogger.Failed(finalResult)
	}

	logger.Info(finalResult)
}

func generateColumnCommentBatch(columns []model.MetaColumn, invoker *gradingLLMInvoker) (map[int]string, error) {
	lines := make([]string, 0, len(columns))
	for _, column := range columns {
		lines = append(lines, fmt.Sprintf(`{"id":%d,"table_name":"%s","column_name":"%s","data_type":"%s","is_nullable":"%s","default_value":"%s"}`,
			column.Id,
			escapeJSONString(column.TableNameX),
			escapeJSONString(column.ColumnName),
			escapeJSONString(column.DataType),
			escapeJSONString(column.IsNullable),
			escapeJSONString(column.DefaultValue),
		))
	}

	prompt := fmt.Sprintf(`你是数据库字段注释助手。请为每个字段生成简洁中文注释。
要求：
1. 每条注释不超过20个中文字符；
2. 基于字段名、表名、数据类型、是否可空、默认值综合推断；
3. 只输出 JSON 数组，不要 Markdown，不要额外说明；
4. 返回格式严格为：
[{"id":1,"ai_comment":"用户手机号"}]
5. 必须保留输入 id，且每个输入 id 都要返回。

待生成数据：
[%s]`, strings.Join(lines, ","))

	answer, err := invoker.complete(prompt)
	if err != nil {
		return nil, err
	}
	items, err := parseColumnCommentBatch(answer)
	if err != nil {
		return nil, err
	}
	out := make(map[int]string, len(items))
	for _, item := range items {
		if item.ID <= 0 {
			continue
		}
		out[item.ID] = item.AiComment
	}
	return out, nil
}

func parseColumnCommentBatch(answer string) ([]aiColumnCommentBatchItem, error) {
	s := strings.TrimSpace(answer)
	if i := strings.Index(s, "["); i >= 0 {
		if j := strings.LastIndex(s, "]"); j > i {
			s = s[i : j+1]
		}
	}
	var payload []aiColumnCommentBatchItem
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return nil, fmt.Errorf("解析批量字段注释响应失败: %v", err)
	}
	return payload, nil
}

func getDifyConfigForColumnComment() (apiURL, apiKey string, timeout time.Duration, err error) {
	baseURL := setting.Setting.AI.DifyBaseUrl
	timeoutSec := setting.Setting.AI.DifyTimeout

	if baseURL == "" {
		return "", "", 0, fmt.Errorf("Dify基础URL未配置")
	}

	targetAgentID := "common_chat_agent"

	for _, agent := range setting.Setting.AI.Agents {
		if agent.ID == targetAgentID && agent.Enabled {
			if agent.ApiKey == "" {
				return "", "", 0, fmt.Errorf("智能体 %s 的API密钥未配置", agent.ID)
			}
			// 构造完整的API URL
			fullURL := fmt.Sprintf("%s/v1/chat-messages", baseURL)
			return fullURL, agent.ApiKey, time.Duration(timeoutSec) * time.Second, nil
		}
	}

	return "", "", 0, fmt.Errorf("未找到智能体 %s 或该智能体已禁用", targetAgentID)
}

func callDifyAPIForColumnComment(question string, apiURL, apiKey string, timeout time.Duration) (string, error) {
	// 构造请求数据
	requestData := DifyRequest{
		Inputs:       make(map[string]interface{}),
		Query:        question,
		ResponseMode: "blocking",
		User:         "aidba-column-comment-task",
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var difyResp DifyResponse
	if err := json.Unmarshal(body, &difyResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取回答内容
	answer := ""
	if difyResp.Answer != "" {
		answer = difyResp.Answer
	} else if difyResp.Message.Answer != "" {
		answer = difyResp.Message.Answer
	} else if difyResp.Message.Content != "" {
		answer = difyResp.Message.Content
	}

	if answer == "" {
		return "", fmt.Errorf("API返回的回答为空")
	}

	return answer, nil
}

func cleanColumnComment(comment string) string {
	// 移除多余的空白字符
	comment = strings.TrimSpace(comment)

	// 移除引号
	comment = strings.Trim(comment, `"'`)

	// 限制长度
	if len(comment) > 50 {
		comment = comment[:50]
	}

	// 移除换行符和特殊字符
	comment = strings.ReplaceAll(comment, "\n", "")
	comment = strings.ReplaceAll(comment, "\r", "")

	return comment
}

// ExecuteAiGeneralColumnCommentTask 手动触发，与定时任务逻辑一致（计划任务平台「手工运行」）
func ExecuteAiGeneralColumnCommentTask() {
	doAiGeneralColumnCommentTask()
}
