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
	"database/sql"
	"dbmcloud/log"
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/libary/mail"
	"dbmcloud/src/model"
	"dbmcloud/src/service"
	"dbmcloud/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// AnalysisTaskLogger 分析任务日志记录器
type AnalysisTaskLogger struct {
	TaskId    int
	TaskName  string
	LogId     int64
	StartTime time.Time
}

// NewAnalysisTaskLogger 创建新的分析任务日志记录器
func NewAnalysisTaskLogger(taskId int, taskName string) *AnalysisTaskLogger {
	return &AnalysisTaskLogger{
		TaskId:    taskId,
		TaskName:  taskName,
		StartTime: time.Now(),
	}
}

// Start 开始记录任务
func (atl *AnalysisTaskLogger) Start() error {
	taskLog := model.AnalysisTaskLog{
		TaskId:    atl.TaskId,
		TaskName:  atl.TaskName,
		StartTime: atl.StartTime,
		Status:    "running",
		Result:    "任务开始执行",
	}

	result := database.DB.Create(&taskLog)
	if result.Error != nil {
		return fmt.Errorf("创建分析任务日志失败: %v", result.Error)
	}

	atl.LogId = taskLog.Id
	return nil
}

// Success 记录任务成功
func (atl *AnalysisTaskLogger) Success(result string, dataCount int, reportContent string) error {
	if atl.LogId == 0 {
		return fmt.Errorf("任务日志ID未初始化")
	}

	completeTime := time.Now()
	duration := completeTime.Sub(atl.StartTime)

	updateData := map[string]interface{}{
		"complete_time":  &completeTime,
		"status":         "success",
		"result":         fmt.Sprintf("%s (执行时长: %v)", result, duration),
		"data_count":     dataCount,
		"report_content": reportContent,
	}

	dbResult := database.DB.Model(&model.AnalysisTaskLog{}).Where("id = ?", atl.LogId).Updates(updateData)
	if dbResult.Error != nil {
		return fmt.Errorf("更新分析任务日志失败: %v", dbResult.Error)
	}

	return nil
}

// Failed 记录任务失败
func (atl *AnalysisTaskLogger) Failed(errorMessage string) error {
	if atl.LogId == 0 {
		return fmt.Errorf("任务日志ID未初始化")
	}

	completeTime := time.Now()
	duration := completeTime.Sub(atl.StartTime)

	updateData := map[string]interface{}{
		"complete_time": &completeTime,
		"status":        "failed",
		"result":        fmt.Sprintf("任务执行失败 (执行时长: %v)", duration),
		"error_message": errorMessage,
	}

	dbResult := database.DB.Model(&model.AnalysisTaskLog{}).Where("id = ?", atl.LogId).Updates(updateData)
	if dbResult.Error != nil {
		return fmt.Errorf("更新分析任务日志失败: %v", dbResult.Error)
	}

	return nil
}

// UpdateResult 更新任务结果（不改变状态）
func (atl *AnalysisTaskLogger) UpdateResult(result string) error {
	if atl.LogId == 0 {
		return fmt.Errorf("任务日志ID未初始化")
	}

	dbResult := database.DB.Model(&model.AnalysisTaskLog{}).Where("id = ?", atl.LogId).Update("result", result)
	if dbResult.Error != nil {
		return fmt.Errorf("更新任务结果失败: %v", dbResult.Error)
	}

	return nil
}

func init() {
	go analysisTaskCrontabTask()
}

// analysisTaskCrontabTask 启动分析任务定时器
func analysisTaskCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))

	// 创建定时器
	c := cron.New()

	// 每分钟检查一次是否有需要执行的任务
	c.AddFunc("*/1 * * * *", func() {
		checkAndExecuteAnalysisTasks()
	})

	c.Start()

	// 保持程序运行
	select {}
}

// checkAndExecuteAnalysisTasks 检查并执行分析任务
func checkAndExecuteAnalysisTasks() {
	logger := log.Logger

	var tasks []model.AnalysisTask
	result := database.DB.Where("status = ?", 1).Find(&tasks)
	if result.Error != nil {
		logger.Error("查询分析任务失败", zap.Error(result.Error))
		return
	}

	if len(tasks) == 0 {
		return
	}

	now := time.Now()

	for _, task := range tasks {
		// 检查是否需要执行任务
		if shouldExecuteTask(task, now) {
			logger.Info("开始执行分析任务", zap.String("task_name", task.TaskName))

			// 异步执行任务
			go executeAnalysisTask(task)
		}
	}
}

// shouldExecuteTask 判断任务是否应该执行
func shouldExecuteTask(task model.AnalysisTask, now time.Time) bool {
	// 如果从未执行过，直接执行
	if task.LastRunTime == nil {
		return true
	}

	// 解析cron表达式
	schedule, err := cron.ParseStandard(task.CronExpression)
	if err != nil {
		log.Logger.Error("解析cron表达式失败", zap.String("cron", task.CronExpression), zap.Error(err))
		return false
	}

	// 计算下次执行时间
	nextRun := schedule.Next(*task.LastRunTime)

	// 如果当前时间已经超过下次执行时间，则执行任务
	return now.After(nextRun)
}

// executeAnalysisTask 执行分析任务
func executeAnalysisTask(task model.AnalysisTask) {
	logger := log.Logger

	// 创建任务日志记录器
	taskLogger := NewAnalysisTaskLogger(task.Id, task.TaskName)
	if err := taskLogger.Start(); err != nil {
		logger.Error(fmt.Sprintf("创建分析任务日志失败: %v", err))
		return
	}

	defer func() {
		// 更新任务的最后执行时间
		now := time.Now()
		database.DB.Model(&task).Update("last_run_time", &now)
	}()

	logger.Info("开始执行分析任务",
		zap.String("task_name", task.TaskName),
		zap.Int("task_id", task.Id),
		zap.Int("ai_model_id", task.AiModelId))
	taskLogger.UpdateResult("开始执行分析任务")

	// 1. 执行SQL查询获取数据
	data, err := executeSqlQueries(task.SqlQueries, task.DatasourceId, taskLogger)
	if err != nil {
		errorMsg := fmt.Sprintf("执行SQL查询失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	taskLogger.UpdateResult(fmt.Sprintf("SQL查询执行成功，获取到 %d 条数据", len(data)))

	// 检查是否有数据
	if len(data) == 0 {
		noDataMsg := "SQL查询未获取到任何数据，跳过后续分析和报告发送"
		logger.Info(noDataMsg)
		taskLogger.Success(noDataMsg, 0, "无数据可分析")
		return
	}

	// 2. 调用AI模型进行分析
	report, modelInfo, err := callAIForAnalysis(task.Prompt, data, task.AiModelId, taskLogger)
	if err != nil {
		errorMsg := fmt.Sprintf("调用AI模型失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 记录使用的模型信息
	modelDesc := fmt.Sprintf("使用模型: %s (%s)", modelInfo.Name, modelInfo.Provider)
	if modelInfo.Type == "dify" {
		modelDesc = "使用模型: Dify API"
	}
	taskLogger.UpdateResult(fmt.Sprintf("AI分析完成，生成分析报告。%s", modelDesc))
	logger.Info("AI分析完成",
		zap.String("model_name", modelInfo.Name),
		zap.String("model_provider", modelInfo.Provider),
		zap.String("model_type", modelInfo.Type))

	// 3. 发送邮件报告
	if err := sendAnalysisReport(task.ReportEmail, task.TaskName, report, data, modelInfo, taskLogger); err != nil {
		errorMsg := fmt.Sprintf("发送邮件报告失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	taskLogger.UpdateResult("邮件报告发送成功")

	// 4. 任务执行成功
	successMsg := fmt.Sprintf("分析任务执行成功，处理了 %d 条数据。%s", len(data), modelDesc)
	logger.Info(successMsg)
	taskLogger.Success(successMsg, len(data), report)
}

// executeSqlQueries 执行SQL查询
func executeSqlQueries(sqlQueries []string, datasourceId int, taskLogger *AnalysisTaskLogger) ([]map[string]interface{}, error) {
	var allData []map[string]interface{}

	for i, sql := range sqlQueries {
		taskLogger.UpdateResult(fmt.Sprintf("执行第 %d 个SQL查询: %s", i+1, sql))

		// 执行SQL查询
		data, err := executeSingleSql(sql, datasourceId)
		if err != nil {
			return nil, fmt.Errorf("执行SQL %d 失败: %v", i+1, err)
		}

		allData = append(allData, data...)
	}

	return allData, nil
}

// executeSingleSql 执行单个SQL查询
func executeSingleSql(sql string, datasourceId int) ([]map[string]interface{}, error) {
	// 查询数据源信息
	var datasource model.Datasource
	result := database.DB.Where("id = ? AND enable = 1", datasourceId).First(&datasource)
	if result.Error != nil {
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

	// 连接数据库
	db, err := connectToDatabase(datasource, origPass)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 执行SQL查询
	rows, err := db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("执行SQL查询失败: %v", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %v", err)
	}

	// 准备结果集
	var data []map[string]interface{}
	for rows.Next() {
		// 创建值的切片
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行数据
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %v", err)
		}

		// 构造行数据
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// 处理不同类型的数据
				switch v := val.(type) {
				case []byte:
					row[col] = string(v)
				case time.Time:
					row[col] = v.Format("2006-01-02 15:04:05")
				default:
					row[col] = v
				}
			} else {
				row[col] = nil
			}
		}
		data = append(data, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果集失败: %v", err)
	}

	return data, nil
}

// connectToDatabase 连接到数据库
func connectToDatabase(datasource model.Datasource, password string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch datasource.Type {
	case "MySQL", "TiDB", "Doris", "MariaDB", "GreatSQL", "OceanBase":
		db, err = database.Connect(
			database.WithDriver("mysql"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	case "PostgreSQL":
		db, err = database.Connect(
			database.WithDriver("postgres"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	case "ClickHouse":
		db, err = database.Connect(
			database.WithDriver("clickhouse"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	case "Oracle":
		db, err = database.Connect(
			database.WithDriver("godror"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithSid(datasource.Dbid))
	case "SQLServer":
		db, err = database.Connect(
			database.WithDriver("mssql"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", datasource.Type)
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

// AIModelInfo AI模型信息
type AIModelInfo struct {
	Id       int
	Name     string
	Provider string
	Type     string // "ai_model" 或 "dify"
}

// callAIForAnalysis 调用AI模型进行分析
// 如果aiModelId为0，使用Dify（向后兼容）；否则使用指定的AI模型
// 返回报告内容和模型信息
func callAIForAnalysis(prompt string, data []map[string]interface{}, aiModelId int, taskLogger *AnalysisTaskLogger) (string, *AIModelInfo, error) {
	log.Logger.Info("准备调用AI模型进行分析", zap.Int("ai_model_id", aiModelId))

	// 如果指定了AI模型ID，使用指定的模型
	if aiModelId > 0 {
		log.Logger.Info("使用指定的AI模型", zap.Int("ai_model_id", aiModelId))
		report, modelInfo, err := callAIModelForAnalysis(prompt, data, aiModelId, taskLogger)
		return report, modelInfo, err
	}

	// 否则使用Dify（向后兼容）
	log.Logger.Warn("AI模型ID为0，回退到Dify API（向后兼容）")
	report, err := callDifyForAnalysis(prompt, data, taskLogger)
	if err != nil {
		return "", nil, err
	}
	modelInfo := &AIModelInfo{
		Id:       0,
		Name:     "Dify API",
		Provider: "Dify",
		Type:     "dify",
	}
	return report, modelInfo, nil
}

// callAIModelForAnalysis 使用指定的AI模型进行分析
func callAIModelForAnalysis(prompt string, data []map[string]interface{}, aiModelId int, taskLogger *AnalysisTaskLogger) (string, *AIModelInfo, error) {
	log.Logger.Info("开始使用AI模型进行分析", zap.Int("ai_model_id", aiModelId))

	// 获取模型配置
	aiModel, err := service.GetModelById(aiModelId)
	if err != nil {
		log.Logger.Error("获取AI模型配置失败", zap.Int("ai_model_id", aiModelId), zap.Error(err))
		return "", nil, fmt.Errorf("获取AI模型配置失败: %v", err)
	}

	log.Logger.Info("获取到AI模型配置",
		zap.String("model_name", aiModel.Name),
		zap.String("provider", aiModel.Provider),
		zap.Int("enabled", int(aiModel.Enabled)))

	// 检查模型是否启用
	if aiModel.Enabled != 1 {
		log.Logger.Warn("AI模型未启用", zap.String("model_name", aiModel.Name))
		return "", nil, fmt.Errorf("AI模型 %s 未启用", aiModel.Name)
	}

	// 创建AI客户端
	client, err := service.NewAIClient(aiModel)
	if err != nil {
		return "", nil, fmt.Errorf("创建AI客户端失败: %v", err)
	}

	// 构造分析请求
	analysisRequest := buildAnalysisRequestForAIModel(prompt, data)

	// 调用AI模型
	response, err := client.Chat(analysisRequest.Messages, &service.ChatOptions{
		Temperature: aiModel.Temperature,
		MaxTokens:   aiModel.MaxTokens,
	})
	if err != nil {
		return "", nil, fmt.Errorf("调用AI模型失败: %v", err)
	}

	// 构造模型信息
	modelInfo := &AIModelInfo{
		Id:       aiModel.Id,
		Name:     aiModel.Name,
		Provider: aiModel.Provider,
		Type:     "ai_model",
	}

	return response.Content, modelInfo, nil
}

// buildAnalysisRequestForAIModel 为AI模型构造分析请求
func buildAnalysisRequestForAIModel(prompt string, data []map[string]interface{}) struct {
	Messages []service.Message
} {
	// 将数据转换为JSON字符串
	dataJson, _ := json.Marshal(data)

	// 构造完整的提示词
	fullPrompt := fmt.Sprintf(`你是一个专业的数据库分析报告生成助手。请根据以下数据进行分析，生成详细的分析报告。

提示词要求：
%s

数据内容：
%s

请基于以上数据进行分析，生成详细的分析报告。报告应使用Markdown格式，包括：
1. 数据概览和总体评估
2. 关键指标分析
3. 数据趋势和异常情况（如果有）
4. 结论和建议

请确保报告结构清晰，内容详实，便于决策参考。`, prompt, string(dataJson))

	return struct {
		Messages []service.Message
	}{
		Messages: []service.Message{
			{
				Role:    "system",
				Content: "你是一个专业的数据库分析报告生成助手，擅长将查询数据转化为清晰、专业的分析报告。",
			},
			{
				Role:    "user",
				Content: fullPrompt,
			},
		},
	}
}

// callDifyForAnalysis 调用Dify API进行分析
func callDifyForAnalysis(prompt string, data []map[string]interface{}, taskLogger *AnalysisTaskLogger) (string, error) {
	// 获取Dify配置
	apiURL, apiKey, timeout, err := getDifyConfigForAnalysis()
	if err != nil {
		return "", fmt.Errorf("获取Dify配置失败: %v", err)
	}

	// 构造分析请求
	analysisRequest := buildAnalysisRequest(prompt, data)

	// 调用Dify API
	response, err := callDifyAPI(apiURL, apiKey, timeout, analysisRequest)
	if err != nil {
		return "", fmt.Errorf("调用Dify API失败: %v", err)
	}

	return response, nil
}

// getDifyConfigForAnalysis 获取Dify配置
func getDifyConfigForAnalysis() (apiURL, apiKey string, timeout time.Duration, err error) {
	baseURL := setting.Setting.AI.DifyBaseUrl + "/v1/chat-messages"
	timeoutSec := setting.Setting.AI.DifyTimeout

	if baseURL == "" {
		return "", "", 0, fmt.Errorf("Dify基础URL未配置")
	}

	// 使用第一个启用的智能体
	for _, agent := range setting.Setting.AI.Agents {
		if agent.Enabled {
			if agent.ApiKey == "" {
				return "", "", 0, fmt.Errorf("智能体 %s 的API密钥未配置", agent.ID)
			}
			return baseURL, agent.ApiKey, time.Duration(timeoutSec) * time.Second, nil
		}
	}

	return "", "", 0, fmt.Errorf("未找到启用的智能体")
}

// buildAnalysisRequest 构造分析请求
func buildAnalysisRequest(prompt string, data []map[string]interface{}) map[string]interface{} {
	// 将数据转换为JSON字符串
	dataJson, _ := json.Marshal(data)

	// 构造完整的提示词
	fullPrompt := fmt.Sprintf("%s\n\n数据内容：\n%s\n\n请基于以上数据进行分析，生成详细的分析报告。", prompt, string(dataJson))

	return map[string]interface{}{
		"inputs":        make(map[string]interface{}),
		"query":         fullPrompt,
		"response_mode": "blocking",
		"user":          "aidba-analysis",
	}
}

// callDifyAPI 调用Dify API
func callDifyAPI(apiURL, apiKey string, timeout time.Duration, requestData map[string]interface{}) (string, error) {
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
	var difyResp struct {
		Answer  string `json:"answer"`
		Message struct {
			Answer  string `json:"answer"`
			Content string `json:"content"`
		} `json:"message"`
	}

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
		return "", fmt.Errorf("Dify API返回的回答为空")
	}

	return answer, nil
}

// sendAnalysisReport 发送分析报告邮件
func sendAnalysisReport(email, taskName, report string, data []map[string]interface{}, modelInfo *AIModelInfo, taskLogger *AnalysisTaskLogger) error {
	// 构造邮件内容
	subject := fmt.Sprintf("AIDBA智能任务 - %s", taskName)

	// 构造HTML邮件内容
	body := buildEmailBody(taskName, report, data, modelInfo)

	// 解析邮箱列表
	emailList := strings.Split(email, ";")
	var cleanEmails []string
	for _, e := range emailList {
		e = strings.TrimSpace(e)
		if e != "" {
			cleanEmails = append(cleanEmails, e)
		}
	}

	if len(cleanEmails) == 0 {
		return fmt.Errorf("没有有效的邮箱地址")
	}

	// 发送邮件
	if err := mail.Send(cleanEmails, subject, body); err != nil {
		return fmt.Errorf("发送邮件失败: %v", err)
	}

	return nil
}

// buildEmailBody 构造邮件内容
func buildEmailBody(taskName, report string, data []map[string]interface{}, modelInfo *AIModelInfo) string {
	var body strings.Builder

	// 邮件头部样式
	body.WriteString(`<html>
<head>
<meta charset="UTF-8">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Microsoft YaHei', sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 800px; margin: 0 auto; background: #ffffff; border: 1px solid #e0e0e0; overflow: hidden; }
.header { background: #4a5568; color: #ffffff; padding: 24px 30px; text-align: left; border-bottom: 2px solid #2d3748; }
.header h1 { margin: 0; font-size: 24px; font-weight: 500; }
.content { padding: 30px; }
.section { margin-bottom: 30px; }
.section h2 { color: #2d3748; font-size: 20px; font-weight: 500; margin: 0 0 15px 0; padding-bottom: 10px; border-bottom: 1px solid #e0e0e0; }
.section p { margin: 8px 0; color: #4a5568; }
.report-content { background: #f7fafc; padding: 20px; border-left: 3px solid #4a5568; color: #2d3748; white-space: pre-wrap; font-size: 14px; }
.footer { background: #f7fafc; padding: 20px 30px; text-align: center; color: #718096; font-size: 12px; border-top: 1px solid #e0e0e0; }
.meta-info { background: #f7fafc; padding: 15px; margin-bottom: 20px; border-left: 3px solid #cbd5e0; }
.meta-info p { margin: 5px 0; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>AIDBA智能任务</h1>
</div>
<div class="content">`)

	// 任务信息
	body.WriteString(fmt.Sprintf(`<div class="section">
<h2>任务信息</h2>
<div class="meta-info">
<p><strong>任务名称:</strong> %s</p>
<p><strong>执行时间:</strong> %s</p>`, taskName, time.Now().Format("2006-01-02 15:04:05")))

	// 添加模型信息
	if modelInfo != nil {
		modelDesc := fmt.Sprintf("%s (%s)", modelInfo.Name, modelInfo.Provider)
		if modelInfo.Type == "dify" {
			modelDesc = "Dify API"
		}
		body.WriteString(fmt.Sprintf(`<p><strong>使用模型:</strong> %s</p>`, modelDesc))
	}

	body.WriteString(fmt.Sprintf(`<p><strong>数据条数:</strong> %d</p>`, len(data)))
	body.WriteString(`</div>
</div>`)

	// 分析报告
	body.WriteString(`<div class="section">
<h2>分析报告</h2>
<div class="report-content">`)
	// 将Markdown转换为HTML（简单处理）
	reportHTML := convertMarkdownToHTML(report)
	body.WriteString(reportHTML)
	body.WriteString(`</div>
</div>`)

	body.WriteString(`</div>
<div class="footer">
<p>此邮件由AIDBA系统自动发送，请勿回复。</p>
</div>
</div>
</body>
</html>`)

	return body.String()
}

// convertMarkdownToHTML 简单的Markdown转HTML转换
func convertMarkdownToHTML(markdown string) string {
	html := markdown
	// 转换标题
	html = strings.ReplaceAll(html, "\n# ", "\n<h1>")
	html = strings.ReplaceAll(html, "\n## ", "\n<h2>")
	html = strings.ReplaceAll(html, "\n### ", "\n<h3>")
	html = strings.ReplaceAll(html, "\n#### ", "\n<h4>")

	// 转换换行
	html = strings.ReplaceAll(html, "\n", "<br>")

	// 转换粗体（简单处理，避免复杂正则）
	html = strings.ReplaceAll(html, "**", "")

	return html
}
