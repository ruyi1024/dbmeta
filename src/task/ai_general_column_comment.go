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
	"bytes"
	"dbmcloud/log"
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

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

	// 获取Dify配置
	apiURL, apiKey, timeout, err := getDifyConfigForColumnComment()
	if err != nil {
		errorMsg := fmt.Sprintf("获取Dify配置失败: %v", err)
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

	for i, column := range columns {
		logger.Info("处理字段", zap.Int("index", i+1), zap.Int("total", len(columns)), zap.String("table_name", column.TableNameX), zap.String("column_name", column.ColumnName))

		err := processColumnComment(column, apiURL, apiKey, timeout)
		if err != nil {
			errorMsg := fmt.Sprintf("处理字段 %s.%s 失败: %v", column.TableNameX, column.ColumnName, err)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount++
		} else {
			successCount++
		}

		// 更新进度
		progressMsg := fmt.Sprintf("已处理 %d/%d 个字段 (成功: %d, 失败: %d)", i+1, len(columns), successCount, failedCount)
		taskLogger.UpdateResult(progressMsg)
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

func processColumnComment(column model.MetaColumn, apiURL, apiKey string, timeout time.Duration) error {
	logger := log.Logger

	// 构造提示词
	prompt := fmt.Sprintf(`请为数据库字段 '%s' 生成一个简洁的中文注释，只返回生成的注释即可，不要返回任何其他内容，要求：
1. 注释要简洁明了，不超过20个字符
2. 使用中文描述字段的用途和含义
3. 根据字段名、数据类型、是否可空等信息推测字段含义
4. 如果是用户相关字段，说明是用户什么信息
5. 如果是业务相关字段，说明是什么业务数据
6. 如果是时间相关字段，说明是什么时间
7. 如果是状态相关字段，说明是什么状态

表名: %s
字段名: %s
数据类型: %s
是否可空: %s
默认值: %s`, column.ColumnName, column.TableNameX, column.ColumnName, column.DataType, column.IsNullable, column.DefaultValue)

	// 调用Dify API
	response, err := callDifyAPIForColumnComment(prompt, apiURL, apiKey, timeout)
	if err != nil {
		return fmt.Errorf("调用Dify API失败: %v", err)
	}

	// 清理和验证响应
	cleanedComment := cleanColumnComment(response)
	if cleanedComment == "" {
		return fmt.Errorf("生成的注释为空")
	}

	// 更新数据库
	updateResult := database.DB.Model(&model.MetaColumn{}).Where("id = ?", column.Id).Update("ai_comment", cleanedComment)
	if updateResult.Error != nil {
		return fmt.Errorf("更新数据库失败: %v", updateResult.Error)
	}

	logger.Info("成功为字段生成AI注释", zap.String("table_name", column.TableNameX), zap.String("column_name", column.ColumnName), zap.String("comment", cleanedComment))

	// 添加延迟，避免API调用过于频繁
	time.Sleep(2 * time.Second)

	return nil
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
