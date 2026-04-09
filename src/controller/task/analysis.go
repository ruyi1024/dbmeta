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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// calculateNextRunTime 计算下次执行时间
func calculateNextRunTime(cronExpression string, lastRunTime *time.Time, now time.Time) *time.Time {
	// 解析cron表达式
	schedule, err := cron.ParseStandard(cronExpression)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("解析cron表达式失败: %s, 错误: %v", cronExpression, err))
		return nil
	}

	// 如果从未执行过，计算从当前时间开始的下次执行时间
	if lastRunTime == nil {
		nextRun := schedule.Next(now)
		return &nextRun
	}

	// 计算从上次执行时间开始的下次执行时间
	nextRun := schedule.Next(*lastRunTime)
	return &nextRun
}

// AnalysisTaskList 获取分析任务列表
func AnalysisTaskList(c *gin.Context) {
	method := c.Request.Method
	var db = database.DB

	if method == "GET" {
		// 查询条件
		if c.Query("task_name") != "" {
			db = db.Where("task_name LIKE ?", "%"+c.Query("task_name")+"%")
		}
		if c.Query("status") != "" {
			db = db.Where("status = ?", c.Query("status"))
		}

		// 分页
		pageSize := 10
		currentPage := 1
		if c.Query("pageSize") != "" {
			if size, err := strconv.Atoi(c.Query("pageSize")); err == nil && size > 0 {
				pageSize = size
			}
		}
		if c.Query("currentPage") != "" {
			if page, err := strconv.Atoi(c.Query("currentPage")); err == nil && page > 0 {
				currentPage = page
			}
		}

		offset := (currentPage - 1) * pageSize

		var dataList []model.AnalysisTask
		var total int64

		// 获取总数
		db.Model(&model.AnalysisTask{}).Count(&total)

		// 获取分页数据
		result := db.Order("gmt_created DESC").Offset(offset).Limit(pageSize).Find(&dataList)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
			return
		}

		// 计算每个任务的下次执行时间
		now := time.Now()
		for i := range dataList {
			nextRunTime := calculateNextRunTime(dataList[i].CronExpression, dataList[i].LastRunTime, now)
			dataList[i].NextRunTime = nextRunTime
		}

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"msg":         "OK",
			"data":        dataList,
			"total":       total,
			"pageSize":    pageSize,
			"currentPage": currentPage,
		})
		return
	}

	if method == "POST" {
		var task model.AnalysisTask
		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
			return
		}

		// 验证必填字段
		if task.TaskName == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务名称不能为空"})
			return
		}
		if task.DatasourceType == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源类型不能为空"})
			return
		}
		if task.DatasourceId == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源不能为空"})
			return
		}
		if task.CronExpression == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Cron表达式不能为空"})
			return
		}
		if task.ReportEmail == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "报告邮箱不能为空"})
			return
		}
		if task.AiModelId == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "AI模型不能为空"})
			return
		}

		// 创建任务
		result := database.DB.Create(&task)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "创建失败: " + result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "创建成功"})
		return
	}

	if method == "PUT" {
		var task model.AnalysisTask
		if err := c.BindJSON(&task); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
			return
		}

		if task.Id == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务ID不能为空"})
			return
		}
		if task.DatasourceType == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源类型不能为空"})
			return
		}
		if task.DatasourceId == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源不能为空"})
			return
		}
		if task.AiModelId == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "AI模型不能为空"})
			return
		}

		// 更新任务（使用map确保ai_model_id即使为0也能更新）
		updateData := map[string]interface{}{
			"task_name":        task.TaskName,
			"task_description": task.TaskDescription,
			"datasource_type":  task.DatasourceType,
			"datasource_id":    task.DatasourceId,
			"ai_model_id":      task.AiModelId,
			"sql_queries":      task.SqlQueries,
			"prompt":           task.Prompt,
			"cron_expression":  task.CronExpression,
			"report_email":     task.ReportEmail,
			"status":           task.Status,
		}
		result := database.DB.Model(&model.AnalysisTask{}).Where("id = ?", task.Id).Updates(updateData)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "更新失败: " + result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "更新成功"})
		return
	}

	if method == "DELETE" {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务ID不能为空"})
			return
		}

		taskId, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "无效的任务ID"})
			return
		}

		// 删除任务
		result := database.DB.Delete(&model.AnalysisTask{}, taskId)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "删除失败: " + result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "删除成功"})
		return
	}
}

// ToggleAnalysisTaskStatus 启用/禁用分析任务
func ToggleAnalysisTaskStatus(c *gin.Context) {
	var request struct {
		Id     int `json:"id"`
		Status int `json:"status"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
		return
	}

	if request.Id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务ID不能为空"})
		return
	}

	if request.Status != 0 && request.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "状态值无效"})
		return
	}

	// 更新状态
	result := database.DB.Model(&model.AnalysisTask{}).Where("id = ?", request.Id).Update("status", request.Status)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "更新失败: " + result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务不存在"})
		return
	}

	statusText := "启用"
	if request.Status == 0 {
		statusText = "禁用"
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": statusText + "成功"})
}

// ExecuteAnalysisTask 手动执行分析任务
func ExecuteAnalysisTask(c *gin.Context) {
	var request struct {
		Id int `json:"id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
		return
	}

	if request.Id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务ID不能为空"})
		return
	}

	// 查询任务
	var task model.AnalysisTask
	result := database.DB.Where("id = ?", request.Id).First(&task)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "任务不存在"})
		return
	}

	// 异步执行任务
	go executeAnalysisTaskAsync(task)

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "任务已开始执行"})
}

// AnalysisTaskLogs 获取分析任务执行日志
func AnalysisTaskLogs(c *gin.Context) {
	method := c.Request.Method
	if method != "GET" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "不支持的请求方法"})
		return
	}

	var db = database.DB

	// 查询条件
	if c.Query("task_id") != "" {
		db = db.Where("task_id = ?", c.Query("task_id"))
	}
	if c.Query("status") != "" {
		db = db.Where("status = ?", c.Query("status"))
	}

	// 日期范围查询
	if c.Query("start_date") != "" {
		startDate, err := time.Parse("2006-01-02", c.Query("start_date"))
		if err == nil {
			db = db.Where("gmt_created >= ?", startDate)
		}
	}
	if c.Query("end_date") != "" {
		endDate, err := time.Parse("2006-01-02", c.Query("end_date"))
		if err == nil {
			// 结束日期加一天，包含当天
			endDate = endDate.Add(24 * time.Hour)
			db = db.Where("gmt_created < ?", endDate)
		}
	}

	// 分页
	pageSize := 10
	currentPage := 1
	if c.Query("pageSize") != "" {
		if size, err := strconv.Atoi(c.Query("pageSize")); err == nil && size > 0 {
			pageSize = size
		}
	}
	if c.Query("currentPage") != "" {
		if page, err := strconv.Atoi(c.Query("currentPage")); err == nil && page > 0 {
			currentPage = page
		}
	}

	offset := (currentPage - 1) * pageSize

	var dataList []model.AnalysisTaskLog
	var total int64

	// 获取总数
	db.Model(&model.AnalysisTaskLog{}).Count(&total)

	// 获取分页数据
	result := db.Order("gmt_created DESC").Offset(offset).Limit(pageSize).Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"msg":         "OK",
		"data":        dataList,
		"total":       total,
		"pageSize":    pageSize,
		"currentPage": currentPage,
	})
}

// TestSqlQuery 测试SQL查询
func TestSqlQuery(c *gin.Context) {
	var request struct {
		Sql            string `json:"sql"`
		DatasourceType string `json:"datasource_type"`
		DatasourceId   int    `json:"datasource_id"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
		return
	}

	if request.Sql == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "SQL语句不能为空"})
		return
	}

	if request.DatasourceType == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源类型不能为空"})
		return
	}

	if request.DatasourceId == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源不能为空"})
		return
	}

	// 执行SQL测试
	data, err := executeSingleSql(request.Sql, request.DatasourceId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "SQL测试失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "SQL测试成功",
		"data":    data,
		"count":   len(data),
	})
}

// TestDifyConnection 测试Dify连接
func TestDifyConnection(c *gin.Context) {
	// 获取Dify配置
	apiURL, apiKey, timeout, err := getDifyConfigForAnalysis()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "获取Dify配置失败: " + err.Error()})
		return
	}

	// 构造测试请求
	testRequest := map[string]interface{}{
		"inputs":        make(map[string]interface{}),
		"query":         "这是一个连接测试，请回复'连接成功'",
		"response_mode": "blocking",
		"user":          "aidba-test",
	}

	// 调用Dify API
	response, err := callDifyAPI(apiURL, apiKey, timeout, testRequest)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Dify连接测试失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "Dify连接测试成功",
		"data":    response,
	})
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

// GetDatasourceTypeList 获取数据源类型列表
func GetDatasourceTypeList(c *gin.Context) {
	var dataList []model.DatasourceType
	var total int64

	// 查询条件
	query := database.DB.Where("enable = 1")

	// 获取总数
	query.Model(&model.DatasourceType{}).Count(&total)

	// 获取数据
	result := query.Select("id, name").Order("sort ASC").Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    dataList,
		"total":   total,
	})
}

// GetDatasourceList 获取数据源列表
func GetDatasourceList(c *gin.Context) {
	var dataList []model.Datasource
	var total int64

	// 查询条件
	query := database.DB.Where("enable = 1")

	if c.Query("type") != "" {
		query = query.Where("type = ?", c.Query("type"))
	}
	if c.Query("env") != "" {
		query = query.Where("env = ?", c.Query("env"))
	}

	// 获取总数
	query.Model(&model.Datasource{}).Count(&total)

	// 获取数据
	result := query.Select("id, name, type, host, port, dbid, env").Order("name ASC").Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    dataList,
		"total":   total,
	})
}

// executeAnalysisTaskAsync 异步执行分析任务
func executeAnalysisTaskAsync(task model.AnalysisTask) {
	// 创建任务日志记录器
	taskLogger := NewAnalysisTaskLogger(task.Id, task.TaskName)
	if err := taskLogger.Start(); err != nil {
		log.Logger.Error(fmt.Sprintf("创建分析任务日志失败: %v", err))
		return
	}

	defer func() {
		// 更新任务的最后执行时间
		now := time.Now()
		database.DB.Model(&task).Update("last_run_time", &now)
	}()

	log.Logger.Info("开始执行分析任务",
		zap.String("task_name", task.TaskName),
		zap.Int("task_id", task.Id),
		zap.Int("ai_model_id", task.AiModelId))
	taskLogger.UpdateResult("开始执行分析任务")

	// 1. 执行SQL查询获取数据
	data, err := executeSqlQueries(task.SqlQueries, task.DatasourceId, taskLogger)
	if err != nil {
		errorMsg := fmt.Sprintf("执行SQL查询失败: %v", err)
		log.Logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	taskLogger.UpdateResult(fmt.Sprintf("SQL查询执行成功，获取到 %d 条数据", len(data)))

	// 检查是否有数据
	if len(data) == 0 {
		noDataMsg := "SQL查询未获取到任何数据，跳过后续分析和报告发送"
		log.Logger.Info(noDataMsg)
		taskLogger.Success(noDataMsg, 0, "无数据可分析")
		return
	}

	// 2. 调用AI模型进行分析
	report, modelInfo, err := callAIForAnalysis(task.Prompt, data, task.AiModelId, taskLogger)
	if err != nil {
		errorMsg := fmt.Sprintf("调用AI模型失败: %v", err)
		log.Logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 记录使用的模型信息
	modelDesc := fmt.Sprintf("使用模型: %s (%s)", modelInfo.Name, modelInfo.Provider)
	if modelInfo.Type == "dify" {
		modelDesc = "使用模型: Dify API"
	}
	taskLogger.UpdateResult(fmt.Sprintf("AI分析完成，生成分析报告。%s", modelDesc))
	log.Logger.Info("AI分析完成",
		zap.String("model_name", modelInfo.Name),
		zap.String("model_provider", modelInfo.Provider),
		zap.String("model_type", modelInfo.Type))

	// 3. 发送邮件报告
	if err := sendAnalysisReport(task.ReportEmail, task.TaskName, report, data, modelInfo, taskLogger); err != nil {
		errorMsg := fmt.Sprintf("发送邮件报告失败: %v", err)
		log.Logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	taskLogger.UpdateResult("邮件报告发送成功")

	// 4. 任务执行成功
	successMsg := fmt.Sprintf("分析任务执行成功，处理了 %d 条数据。%s", len(data), modelDesc)
	log.Logger.Info(successMsg)
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

// callDifyForAnalysis 调用Dify API进行分析（向后兼容）
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

// buildAnalysisRequest 构造分析请求（用于Dify）
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
.header .subtitle { margin: 8px 0 0 0; opacity: 0.85; font-size: 14px; color: #e2e8f0; }
.content { padding: 30px; }
.section { margin-bottom: 30px; }
.section h2 { color: #2d3748; border-bottom: 2px solid #4a5568; padding-bottom: 8px; margin-bottom: 16px; font-size: 18px; font-weight: 500; }
.info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px; margin-bottom: 20px; }
.info-item { background: #f7fafc; padding: 12px 16px; border: 1px solid #e2e8f0; border-left: 3px solid #4a5568; }
.info-label { font-weight: 500; color: #718096; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
.info-value { font-size: 15px; color: #2d3748; margin-top: 4px; }
.report-box { background: #ffffff; border: 1px solid #e2e8f0; padding: 20px; margin: 20px 0; }
.report-content { line-height: 1.8; color: #2d3748; }
.report-content h3 { color: #2d3748; margin: 20px 0 12px 0; font-size: 16px; font-weight: 500; }
.report-content h4 { color: #4a5568; margin: 16px 0 8px 0; font-size: 15px; font-weight: 500; }
.report-content p { margin: 12px 0; color: #4a5568; }
.report-content ul, .report-content ol { margin: 12px 0; padding-left: 24px; color: #4a5568; }
.report-content li { margin: 6px 0; }
.report-content strong { color: #2d3748; font-weight: 600; }
.footer { background: #2d3748; color: #e2e8f0; padding: 16px 30px; text-align: center; font-size: 12px; border-top: 1px solid #4a5568; }
</style>
</head>
<body>`)

	// 邮件头部
	body.WriteString(`<div class="container">
<div class="header">
<h1>AIDBA智能任务</h1>
<div class="subtitle">智能数据分析与洞察</div>
</div>
<div class="content">`)

	// 任务信息
	modelName := modelInfo.Name
	if modelInfo.Type == "dify" {
		modelName = "Dify API"
	}
	modelProvider := modelInfo.Provider
	if modelInfo.Type == "dify" {
		modelProvider = "Dify"
	}

	body.WriteString(`<div class="section">
<h2>任务信息</h2>
<div class="info-grid">
<div class="info-item">
<div class="info-label">任务名称</div>
<div class="info-value">` + taskName + `</div>
</div>
<div class="info-item">
<div class="info-label">执行时间</div>
<div class="info-value">` + time.Now().Format("2006-01-02 15:04:05") + `</div>
</div>
<div class="info-item">
<div class="info-label">数据条数</div>
<div class="info-value">` + fmt.Sprintf("%d 条", len(data)) + `</div>
</div>
<div class="info-item">
<div class="info-label">AI模型</div>
<div class="info-value">` + modelName + ` (` + modelProvider + `)</div>
</div>
</div>
</div>`)

	// 分析报告
	body.WriteString(`<div class="section">
<h2>分析报告</h2>
<div class="report-box">
<div class="report-content">`)

	// 格式化报告内容
	formattedReport := formatReportForEmail(report)
	body.WriteString(formattedReport)

	body.WriteString(`</div>
</div>
</div>`)

	// 邮件底部
	body.WriteString(`</div>
<div class="footer">
<p>此邮件由AIDBA系统自动发送，请勿回复</p>
<p style="margin-top: 8px; opacity: 0.85;">如有问题，请联系系统管理员</p>
</div>
</div>
</body>
</html>`)

	return body.String()
}

// formatReportForEmail 格式化报告内容用于邮件，将Markdown转换为HTML
func formatReportForEmail(report string) string {
	if report == "" {
		return "<p style='color: #718096; font-style: italic;'>暂无分析报告</p>"
	}

	// 转义HTML特殊字符（但保留Markdown标记）
	report = strings.ReplaceAll(report, "&", "&amp;")
	report = strings.ReplaceAll(report, "<", "&lt;")
	report = strings.ReplaceAll(report, ">", "&gt;")

	// 处理Markdown表格（先处理表格，避免被其他规则干扰）
	tableRe := regexp.MustCompile(`(?s)\|(.+)\|\s*\n\|[-\s\|]+\|\s*\n((?:\|.+\|\s*\n?)+)`)
	report = tableRe.ReplaceAllStringFunc(report, func(match string) string {
		return formatMarkdownTable(match)
	})

	// 处理Markdown标题 # ## ###
	lines := strings.Split(report, "\n")
	var htmlLines []string
	inList := false
	listType := ""

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行
		if line == "" {
			if inList {
				htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				inList = false
				listType = ""
			}
			htmlLines = append(htmlLines, "")
			continue
		}

		// 检查是否是表格行（包含 | 分隔符且不是标题行）
		if strings.Contains(line, "|") && !strings.HasPrefix(strings.TrimSpace(line), "|") {
			// 可能是表格的一部分，但格式不标准，按普通文本处理
		} else if strings.Contains(line, "|") && len(strings.Split(line, "|")) > 2 {
			// 跳过表格行，因为已经在前面处理过了
			continue
		}

		// 处理标题
		if strings.HasPrefix(line, "### ") {
			if inList {
				htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				inList = false
				listType = ""
			}
			content := processInlineMarkdown(strings.TrimPrefix(line, "### "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h3>%s</h3>", content))
		} else if strings.HasPrefix(line, "## ") {
			if inList {
				htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				inList = false
				listType = ""
			}
			content := processInlineMarkdown(strings.TrimPrefix(line, "## "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h2>%s</h2>", content))
		} else if strings.HasPrefix(line, "# ") {
			if inList {
				htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				inList = false
				listType = ""
			}
			content := processInlineMarkdown(strings.TrimPrefix(line, "# "))
			htmlLines = append(htmlLines, fmt.Sprintf("<h2>%s</h2>", content))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// 无序列表
			if !inList || listType != "ul" {
				if inList {
					htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				}
				htmlLines = append(htmlLines, "<ul>")
				inList = true
				listType = "ul"
			}
			content := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			content = processInlineMarkdown(content)
			htmlLines = append(htmlLines, fmt.Sprintf("<li>%s</li>", content))
		} else if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
			// 有序列表
			if !inList || listType != "ol" {
				if inList {
					htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				}
				htmlLines = append(htmlLines, "<ol>")
				inList = true
				listType = "ol"
			}
			content := regexp.MustCompile(`^\d+\.\s`).ReplaceAllString(line, "")
			content = processInlineMarkdown(content)
			htmlLines = append(htmlLines, fmt.Sprintf("<li>%s</li>", content))
		} else if strings.HasPrefix(line, "**") && strings.Contains(line, "**：") {
			// 处理类似 **机房信息**：... 的格式，识别为列表项
			if !inList || listType != "ul" {
				if inList {
					htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				}
				htmlLines = append(htmlLines, "<ul>")
				inList = true
				listType = "ul"
			}
			content := processInlineMarkdown(line)
			htmlLines = append(htmlLines, fmt.Sprintf("<li>%s</li>", content))
		} else {
			// 普通段落
			if inList {
				htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
				inList = false
				listType = ""
			}
			content := processInlineMarkdown(line)
			if content != "" {
				htmlLines = append(htmlLines, fmt.Sprintf("<p>%s</p>", content))
			}
		}

		// 最后一行如果是列表，需要关闭
		if i == len(lines)-1 && inList {
			htmlLines = append(htmlLines, fmt.Sprintf("</%s>", listType))
		}
	}

	result := strings.Join(htmlLines, "\n")
	return result
}

// processInlineMarkdown 处理行内Markdown格式（加粗、斜体、代码等）
func processInlineMarkdown(text string) string {
	// 处理代码 `code` (先处理，避免与加粗冲突)
	codeRe := regexp.MustCompile("`([^`]+?)`")
	text = codeRe.ReplaceAllString(text, "<code style='background: #f7fafc; padding: 2px 4px; border: 1px solid #e2e8f0;'>$1</code>")
	// 处理加粗 **text**
	boldRe := regexp.MustCompile(`\*\*([^*]+?)\*\*`)
	text = boldRe.ReplaceAllString(text, "<strong>$1</strong>")
	// 处理斜体 *text* (避免与加粗冲突)
	italicRe := regexp.MustCompile(`([^*])\*([^*]+?)\*([^*])`)
	text = italicRe.ReplaceAllString(text, "$1<em>$2</em>$3")
	return text
}

// formatMarkdownTable 将Markdown表格转换为HTML表格
func formatMarkdownTable(match string) string {
	lines := strings.Split(strings.TrimSpace(match), "\n")
	if len(lines) < 2 {
		return match
	}

	var html strings.Builder
	html.WriteString("<table style='width: 100%; border-collapse: collapse; margin: 16px 0; border: 1px solid #e2e8f0;'>")

	// 处理表头
	headerLine := lines[0]
	if strings.HasPrefix(headerLine, "|") && strings.HasSuffix(headerLine, "|") {
		headerLine = strings.Trim(headerLine, "|")
		cells := strings.Split(headerLine, "|")
		html.WriteString("<thead><tr>")
		for _, cell := range cells {
			cell = strings.TrimSpace(cell)
			cell = processInlineMarkdown(cell)
			html.WriteString(fmt.Sprintf("<th style='padding: 10px 12px; text-align: left; background: #f7fafc; border-bottom: 2px solid #e2e8f0; border-right: 1px solid #e2e8f0; font-weight: 500; color: #2d3748;'>%s</th>", cell))
		}
		html.WriteString("</tr></thead>")
	}

	// 处理表格数据行（跳过分隔行）
	html.WriteString("<tbody>")
	for i := 2; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			line = strings.Trim(line, "|")
			cells := strings.Split(line, "|")
			html.WriteString("<tr>")
			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				cell = processInlineMarkdown(cell)
				html.WriteString(fmt.Sprintf("<td style='padding: 10px 12px; border-bottom: 1px solid #e2e8f0; border-right: 1px solid #e2e8f0; color: #4a5568;'>%s</td>", cell))
			}
			html.WriteString("</tr>")
		}
	}
	html.WriteString("</tbody></table>")

	return html.String()
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
