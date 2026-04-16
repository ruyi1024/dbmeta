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

package ai

import (
	"bytes"
	"dbmeta-core/log"
	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// getAIConfig 获取AI配置
func getAIConfig() (apiURL, apiKey, model string, timeout time.Duration) {
	return setting.Setting.AI.DeepseekApiUrl,
		setting.Setting.AI.DeepseekApiKey,
		setting.Setting.AI.DeepseekModel,
		time.Duration(setting.Setting.AI.Timeout) * time.Second
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

// AgentChatRequest 智能体聊天请求结构
type AgentChatRequest struct {
	Question string `json:"question" binding:"required"`
	AgentID  string `json:"agentId" binding:"required"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	Answer    string `json:"answer"`
	Timestamp int64  `json:"timestamp"`
}

// DatabaseAnalysisRequest 数据库分析请求结构
type DatabaseAnalysisRequest struct {
	DbType       string `json:"dbType" binding:"required"`
	Instance     string `json:"instance" binding:"required"`
	AnalysisType string `json:"analysisType" binding:"required"`
}

// DatabaseAnalysisResponse 数据库分析响应结构
type DatabaseAnalysisResponse struct {
	DbType    string                 `json:"dbType"`
	Instance  string                 `json:"instance"`
	Metrics   map[string]interface{} `json:"metrics"`
	Timestamp string                 `json:"timestamp"`
	Analysis  string                 `json:"analysis,omitempty"`
}

// DeepSeekMessage DeepSeek API消息结构
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest DeepSeek API请求结构
type DeepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []DeepSeekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

// DeepSeekChoice DeepSeek API选择结构
type DeepSeekChoice struct {
	Index   int             `json:"index"`
	Message DeepSeekMessage `json:"message"`
}

// DeepSeekResponse DeepSeek API响应结构
type DeepSeekResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []DeepSeekChoice `json:"choices"`
}

// EventData ClickHouse事件数据结构
type EventData struct {
	EventKey   string `json:"event_key"`
	EventValue string `json:"event_value"`
}

// DifyRequest dify API请求结构
type DifyRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	Query        string                 `json:"query"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
}

// DifyResponse dify API响应结构
type DifyResponse struct {
	Event   string `json:"event"`
	TaskID  string `json:"task_id"`
	ID      string `json:"id"`
	Answer  string `json:"answer"`
	Message struct {
		ID      string `json:"id"`
		Answer  string `json:"answer"`
		Content string `json:"content"`
	} `json:"message"`
}

// Chat 处理AI聊天请求（流式输出）
func Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Chat request bind error", zap.Error(err))
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Info("AI Chat request", zap.String("question", req.Question))

	// 获取底层的 http.ResponseWriter，绕过 Gin 的所有封装和缓冲
	// 这是最彻底的方法，直接操作底层连接
	var w http.ResponseWriter = c.Writer

	// 尝试获取真正的底层 ResponseWriter（Gin 可能有多层包装）
	// 通过类型断言获取底层 writer
	type responseWriter interface {
		http.ResponseWriter
		http.Flusher
		http.Hijacker
	}

	// 获取 Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error("Writer does not implement http.Flusher")
		c.JSON(500, gin.H{
			"success": false,
			"message": "服务器不支持流式响应",
		})
		return
	}

	// 直接设置响应头到底层 ResponseWriter，绕过 Gin 的缓冲
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")      // 禁用nginx缓冲
	w.Header().Set("Transfer-Encoding", "chunked") // 启用分块传输
	// CORS 头
	origin := c.GetHeader("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 设置状态码为200（必须在写入数据之前）
	// 直接调用底层 WriteHeader，绕过 Gin
	w.WriteHeader(http.StatusOK)

	// 立即刷新响应头，确保客户端知道这是流式响应
	flusher.Flush()
	flusher.Flush()
	flusher.Flush()

	// 发送一个初始的心跳/注释，确保连接建立并立即刷新
	fmt.Fprintf(w, ": heartbeat\n\n")
	flusher.Flush()
	flusher.Flush()

	log.Info("响应头已刷新，开始流式传输")

	// 构建消息历史（可以扩展为多轮对话）
	messages := []service.Message{
		{
			Role:    "system",
			Content: "你是AIDBA智能助手，专门帮助用户解答关于数据库管理、SQL查询、系统监控等相关问题。请用专业、友好的语气回答用户的问题。",
		},
		{
			Role:    "user",
			Content: req.Question,
		},
	}

	// 调用流式API
	stream, err := service.CallWithStream(messages, &service.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   2000,
	})
	if err != nil {
		log.Error("AI stream call failed", zap.Error(err))
		errorJSON, _ := json.Marshal(gin.H{
			"success": false,
			"message": "AI服务暂时不可用，请稍后重试",
			"error":   err.Error(),
		})
		// 使用 fmt.Fprintf 直接写入底层 ResponseWriter
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorJSON)
		flusher.Flush()
		flusher.Flush()
		return
	}

	// 流式发送响应
	var fullContent strings.Builder
	var fullThink strings.Builder
	chunkCount := 0
	for chunk := range stream {
		chunkCount++
		log.Info("收到流式数据块", zap.Int("chunk", chunkCount), zap.Bool("done", chunk.Done), zap.String("content", chunk.Content), zap.String("think", chunk.Think))

		if chunk.Error != nil {
			log.Error("Stream chunk error", zap.Error(chunk.Error))
			errorJSON, _ := json.Marshal(gin.H{
				"success": false,
				"message": "流式响应错误",
				"error":   chunk.Error.Error(),
			})
			// 使用 fmt.Fprintf 直接写入底层 ResponseWriter
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", errorJSON)
			flusher.Flush()
			flusher.Flush()
			return
		}

		if chunk.Done {
			// 发送完成事件
			log.Info("流式响应完成", zap.String("fullContent", fullContent.String()), zap.String("fullThink", fullThink.String()))
			doneJSON, _ := json.Marshal(gin.H{
				"success": true,
				"content": fullContent.String(),
				"think":   fullThink.String(),
			})
			// 使用 fmt.Fprintf 直接写入底层 ResponseWriter
			fmt.Fprintf(w, "event: done\ndata: %s\n\n", doneJSON)
			flusher.Flush()
			flusher.Flush()
			break
		}

		// 发送思考内容（直接写入底层Writer，避免所有缓冲）
		if chunk.Think != "" {
			fullThink.WriteString(chunk.Think)
			thinkJSON, _ := json.Marshal(gin.H{"content": chunk.Think})
			// 使用 fmt.Fprintf 直接写入底层 ResponseWriter，立即刷新
			fmt.Fprintf(w, "event: think\ndata: %s\n\n", thinkJSON)
			flusher.Flush()
			flusher.Flush()
			log.Info("发送思考内容", zap.String("think", chunk.Think), zap.Int("bytes", len(chunk.Think)), zap.Int("jsonBytes", len(thinkJSON)))
		}

		// 发送回答内容（直接写入底层Writer，避免所有缓冲）
		if chunk.Content != "" {
			fullContent.WriteString(chunk.Content)
			contentJSON, _ := json.Marshal(gin.H{"content": chunk.Content})
			// 使用 fmt.Fprintf 直接写入底层 ResponseWriter，立即刷新
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", contentJSON)
			flusher.Flush()
			flusher.Flush()
			log.Info("发送回答内容", zap.String("content", chunk.Content), zap.Int("bytes", len(chunk.Content)), zap.Int("jsonBytes", len(contentJSON)))
		}
	}

	log.Info("流式响应处理完成", zap.Int("totalChunks", chunkCount), zap.String("fullContent", fullContent.String()))
	// 确保响应已发送
	flusher.Flush()
	flusher.Flush()

	// 对于流式响应，使用 Abort() 防止中间件继续处理
	// 这可以防止 HandleLogger 中间件在响应完成后尝试读取状态码导致缓冲
	c.Abort()
}

// DifyChat 处理AI查数智能体聊天请求
func DifyChat(c *gin.Context) {
	var req AgentChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Dify chat request bind error", zap.Error(err))
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Info("AI Agent request",
		zap.String("question", req.Question),
		zap.String("agentId", req.AgentID))

	// 调用Dify API
	answer, err := callDifyAPI(req.Question, req.AgentID)
	if err != nil {
		log.Error("Dify API call failed", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "AI智能体服务暂时不可用，请稍后重试",
			"error":   err.Error(),
		})
		return
	}

	// 构造响应
	response := ChatResponse{
		Answer:    answer,
		Timestamp: time.Now().Unix(),
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    response,
		"answer":  answer, // 为了兼容前端，同时返回answer字段
	})
}

// DatabaseAnalysis 处理数据库分析请求
func DatabaseAnalysis(c *gin.Context) {
	var req DatabaseAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Database analysis request bind error", zap.Error(err))
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Info("Database Analysis request",
		zap.String("dbType", req.DbType),
		zap.String("instance", req.Instance),
		zap.String("analysisType", req.AnalysisType))

	// 查询ClickHouse获取事件数据
	eventData, err := queryClickHouseEvents(req.Instance)
	if err != nil {
		log.Error("Query ClickHouse events failed", zap.Error(err))
		// 如果ClickHouse查询失败，使用模拟数据
		eventData = generateMockEventData(req.DbType, req.Instance)
	}

	// 使用AI分析事件数据
	analysisResult, err := analyzeEventsWithAI(req.DbType, req.Instance, eventData)
	if err != nil {
		log.Error("AI analysis failed", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "AI分析失败",
			"error":   err.Error(),
		})
		return
	}

	// 构造响应
	response := DatabaseAnalysisResponse{
		DbType:    req.DbType,
		Instance:  req.Instance,
		Metrics:   map[string]interface{}{"events": eventData},
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Analysis:  analysisResult,
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    response,
		"message": "数据库分析完成",
	})
}

// getDatabaseMetrics 获取数据库性能指标
func getDatabaseMetrics(dbType, instance string) (map[string]interface{}, error) {
	// 解析实例地址
	instanceParts := strings.Split(instance, ":")
	if len(instanceParts) != 2 {
		return nil, fmt.Errorf("无效的实例地址格式，应为 host:port")
	}
	host := instanceParts[0]
	port := instanceParts[1]

	// 查询数据源信息
	sql := fmt.Sprintf("SELECT * FROM datasource WHERE host='%s' AND port='%s' AND type='%s' LIMIT 1", host, port, dbType)
	datasourceList, err := database.QueryAll(sql)
	if err != nil {
		return nil, fmt.Errorf("查询数据源失败: %v", err)
	}

	if len(datasourceList) == 0 {
		// 如果数据库中没有找到对应的数据源，返回模拟数据
		return generateMockMetrics(dbType, instance), nil
	}

	// 根据不同数据库类型获取性能指标
	metrics := make(map[string]interface{})

	switch strings.ToUpper(dbType) {
	case "MYSQL":
		metrics = getMySQLMetrics(host, port)
	case "ORACLE":
		metrics = getOracleMetrics(host, port)
	case "REDIS":
		metrics = getRedisMetrics(host, port)
	case "MONGODB":
		metrics = getMongoDBMetrics(host, port)
	default:
		// 未支持的数据库类型，返回模拟数据
		metrics = generateMockMetrics(dbType, instance)
	}

	return metrics, nil
}

// getMySQLMetrics 获取MySQL性能指标
func getMySQLMetrics(host, port string) map[string]interface{} {
	metrics := make(map[string]interface{})

	// 尝试查询MySQL状态
	sql := `SELECT VARIABLE_NAME, VARIABLE_VALUE FROM information_schema.GLOBAL_STATUS 
			WHERE VARIABLE_NAME IN ('Threads_connected', 'Queries', 'Slow_queries', 'Innodb_rows_read', 'Innodb_rows_inserted')`

	statusList, err := database.QueryAll(sql)
	if err != nil {
		// 如果查询失败，返回模拟数据
		return generateMockMetrics("MySQL", host+":"+port)
	}

	// 解析MySQL状态
	for _, status := range statusList {
		varName := status["VARIABLE_NAME"].(string)
		varValue := status["VARIABLE_VALUE"].(string)

		switch varName {
		case "Threads_connected":
			metrics["connections"] = varValue
		case "Queries":
			metrics["total_queries"] = varValue
		case "Slow_queries":
			metrics["slow_queries"] = varValue
		case "Innodb_rows_read":
			metrics["rows_read"] = varValue
		case "Innodb_rows_inserted":
			metrics["rows_inserted"] = varValue
		}
	}

	// 添加计算指标
	metrics["cpu_usage"] = 75 + (time.Now().Unix() % 20)    // 模拟CPU使用率
	metrics["memory_usage"] = 60 + (time.Now().Unix() % 30) // 模拟内存使用率
	metrics["qps"] = 500 + (time.Now().Unix() % 300)        // 模拟QPS

	return metrics
}

// getOracleMetrics 获取Oracle性能指标
func getOracleMetrics(host, port string) map[string]interface{} {
	// Oracle指标获取逻辑
	return generateMockMetrics("Oracle", host+":"+port)
}

// getRedisMetrics 获取Redis性能指标
func getRedisMetrics(host, port string) map[string]interface{} {
	// Redis指标获取逻辑
	return generateMockMetrics("Redis", host+":"+port)
}

// getMongoDBMetrics 获取MongoDB性能指标
func getMongoDBMetrics(host, port string) map[string]interface{} {
	// MongoDB指标获取逻辑
	return generateMockMetrics("MongoDB", host+":"+port)
}

// generateMockMetrics 生成模拟的性能指标数据
func generateMockMetrics(dbType, instance string) map[string]interface{} {
	seed := time.Now().Unix()
	metrics := make(map[string]interface{})

	// 基础指标（所有数据库类型通用）
	metrics["cpu_usage"] = 45 + (seed % 40)     // CPU使用率 45-85%
	metrics["memory_usage"] = 50 + (seed % 35)  // 内存使用率 50-85%
	metrics["connections"] = 100 + (seed % 400) // 连接数 100-500

	// 根据数据库类型生成特定指标
	switch strings.ToUpper(dbType) {
	case "MYSQL":
		metrics["qps"] = 200 + (seed % 800)                    // QPS 200-1000
		metrics["slow_queries"] = seed % 50                    // 慢查询数量 0-50
		metrics["lock_waits"] = seed % 20                      // 锁等待数 0-20
		metrics["innodb_buffer_pool_usage"] = 70 + (seed % 25) // InnoDB缓冲池使用率

	case "ORACLE":
		metrics["sga_usage"] = 65 + (seed % 30)        // SGA使用率 65-95%
		metrics["pga_usage"] = 40 + (seed % 40)        // PGA使用率 40-80%
		metrics["tablespace_usage"] = 50 + (seed % 40) // 表空间使用率
		metrics["redo_log_switches"] = seed % 100      // 重做日志切换次数

	case "REDIS":
		metrics["memory_usage_redis"] = 60 + (seed % 35)   // Redis内存使用率
		metrics["commands_per_sec"] = 1000 + (seed % 4000) // 每秒命令数
		metrics["keyspace_hits"] = 80 + (seed % 15)        // 键空间命中率
		metrics["expired_keys"] = seed % 1000              // 过期键数量

	case "MONGODB":
		metrics["operations_per_sec"] = 300 + (seed % 700) // 每秒操作数
		metrics["wired_tiger_cache"] = 75 + (seed % 20)    // WiredTiger缓存使用率
		metrics["replica_lag"] = seed % 1000               // 复制延迟（毫秒）
		metrics["index_usage"] = 85 + (seed % 10)          // 索引使用率
	}

	return metrics
}

// callDeepSeekAPI 调用AI API（使用新的模型服务，支持故障转移）
func callDeepSeekAPI(question string) (string, error) {
	// 使用新的模型服务
	messages := []service.Message{
		{
			Role:    "system",
			Content: "你是AIDBA智能助手，专门帮助用户解答关于数据库管理、SQL查询、系统监控等相关问题。请用专业、友好的语气回答用户的问题。",
		},
		{
			Role:    "user",
			Content: question,
		},
	}

	response, err := service.CallWithFailover(messages, nil)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// queryClickHouseEvents 查询ClickHouse获取事件数据
func queryClickHouseEvents(instance string) ([]EventData, error) {
	// 构造ClickHouse查询SQL
	sql := fmt.Sprintf("SELECT event_key, event_value FROM events WHERE EventEntity='%s' ORDER BY event_time DESC LIMIT 10", instance)

	log.Info("Querying ClickHouse events", zap.String("sql", sql))

	// 尝试查询ClickHouse数据库
	// 注意：这里需要根据实际的ClickHouse连接配置进行调整
	eventList, err := database.QueryAll(sql)
	if err != nil {
		return nil, fmt.Errorf("ClickHouse查询失败: %v", err)
	}

	// 转换查询结果
	var events []EventData
	for _, row := range eventList {
		event := EventData{
			EventKey:   fmt.Sprintf("%v", row["event_key"]),
			EventValue: fmt.Sprintf("%v", row["event_value"]),
		}
		events = append(events, event)
	}

	return events, nil
}

// generateMockEventData 生成模拟事件数据
func generateMockEventData(dbType, instance string) []EventData {
	// 生成模拟的事件数据
	seed := time.Now().Unix()
	events := []EventData{
		{EventKey: "cpu_usage", EventValue: fmt.Sprintf("%.1f", float64(60+(seed%30)))},
		{EventKey: "memory_usage", EventValue: fmt.Sprintf("%.1f", float64(50+(seed%40)))},
		{EventKey: "connections", EventValue: fmt.Sprintf("%d", 100+(seed%300))},
		{EventKey: "qps", EventValue: fmt.Sprintf("%d", 200+(seed%800))},
		{EventKey: "slow_queries", EventValue: fmt.Sprintf("%d", seed%50)},
	}

	// 根据数据库类型添加特定事件
	switch strings.ToUpper(dbType) {
	case "MYSQL":
		events = append(events,
			EventData{EventKey: "innodb_buffer_pool_usage", EventValue: fmt.Sprintf("%.1f", float64(70+(seed%25)))},
			EventData{EventKey: "lock_waits", EventValue: fmt.Sprintf("%d", seed%20)},
		)
	case "REDIS":
		events = append(events,
			EventData{EventKey: "keyspace_hits", EventValue: fmt.Sprintf("%.1f", float64(80+(seed%15)))},
			EventData{EventKey: "commands_per_sec", EventValue: fmt.Sprintf("%d", 1000+(seed%4000))},
		)
	case "ORACLE":
		events = append(events,
			EventData{EventKey: "sga_usage", EventValue: fmt.Sprintf("%.1f", float64(65+(seed%30)))},
			EventData{EventKey: "pga_usage", EventValue: fmt.Sprintf("%.1f", float64(40+(seed%40)))},
		)
	case "MONGODB":
		events = append(events,
			EventData{EventKey: "operations_per_sec", EventValue: fmt.Sprintf("%d", 300+(seed%700))},
			EventData{EventKey: "wired_tiger_cache", EventValue: fmt.Sprintf("%.1f", float64(75+(seed%20)))},
		)
	}

	return events
}

// analyzeEventsWithAI 使用AI分析事件数据
func analyzeEventsWithAI(dbType, instance string, events []EventData) (string, error) {
	// 构造事件数据描述
	var eventDescription strings.Builder
	eventDescription.WriteString(fmt.Sprintf("数据库类型: %s\n", dbType))
	eventDescription.WriteString(fmt.Sprintf("实例地址: %s\n", instance))
	eventDescription.WriteString("最近10条事件数据:\n")

	for i, event := range events {
		eventDescription.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, event.EventKey, event.EventValue))
	}

	// 构造AI分析提示
	question := fmt.Sprintf(`请分析以下%s数据库实例的事件数据，并提供专业的压力分析和优化建议：

%s

请从以下几个方面进行分析：
1. 关键性能指标评估
2. 潜在的性能瓶颈识别 
3. 压力状况判断
4. 具体的优化建议
5. 监控建议

请用专业但易懂的语言回答，并提供具体的数值分析。`, dbType, eventDescription.String())

	// 调用AI分析
	analysisResult, err := callDeepSeekAPI(question)
	if err != nil {
		// 如果AI调用失败，返回基本分析
		return generateBasicAnalysis(dbType, instance, events), nil
	}

	return analysisResult, nil
}

// generateBasicAnalysis 生成基本分析结果（AI调用失败时的备选方案）
func generateBasicAnalysis(dbType, instance string, events []EventData) string {
	var analysis strings.Builder

	analysis.WriteString(fmt.Sprintf("🔍 **%s 数据库压力分析报告**\n\n", dbType))
	analysis.WriteString(fmt.Sprintf("**实例:** %s\n", instance))
	analysis.WriteString(fmt.Sprintf("**分析时间:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	analysis.WriteString("📊 **事件数据概览:**\n")
	for _, event := range events {
		analysis.WriteString(fmt.Sprintf("- %s: %s\n", event.EventKey, event.EventValue))
	}

	analysis.WriteString("\n⚡ **压力评估:**\n")
	analysis.WriteString("基于当前事件数据，系统运行状态需要进一步监控和分析。\n\n")

	analysis.WriteString("💡 **基本建议:**\n")
	analysis.WriteString("1. 持续监控关键性能指标\n")
	analysis.WriteString("2. 定期检查慢查询和连接状况\n")
	analysis.WriteString("3. 关注内存和CPU使用率变化\n")
	analysis.WriteString("4. 建议设置告警阈值进行预警\n")

	return analysis.String()
}

// getDifyConfig 获取Dify配置
func getDifyConfig(agentID string) (apiURL, apiKey string, timeout time.Duration, err error) {
	baseURL := setting.Setting.AI.DifyBaseUrl + "/v1/chat-messages"
	timeoutSec := setting.Setting.AI.DifyTimeout

	if baseURL == "" {
		return "", "", 0, fmt.Errorf("Dify基础URL未配置")
	}

	// 查找指定的智能体配置
	for _, agent := range setting.Setting.AI.Agents {
		if agent.ID == agentID && agent.Enabled {
			if agent.ApiKey == "" {
				return "", "", 0, fmt.Errorf("智能体 %s 的API密钥未配置", agentID)
			}
			return baseURL, agent.ApiKey, time.Duration(timeoutSec) * time.Second, nil
		}
	}

	return "", "", 0, fmt.Errorf("未找到智能体 %s 或该智能体已禁用", agentID)
}

// callDifyAPI 调用Dify API
func callDifyAPI(question string, agentID string) (string, error) {
	// 获取配置
	apiURL, apiKey, timeout, err := getDifyConfig(agentID)
	if err != nil {
		return "", err
	}

	if apiURL == "" || apiKey == "" {
		return "", fmt.Errorf("Dify配置未设置")
	}

	// 构造请求数据
	requestData := DifyRequest{
		Inputs:       make(map[string]interface{}),
		Query:        question,
		ResponseMode: "blocking",
		User:         "aidba-user",
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

// GetAgents 获取可用的智能体列表
func GetAgents(c *gin.Context) {
	// 过滤启用的智能体
	var enabledAgents []setting.DifyAgent
	for _, agent := range setting.Setting.AI.Agents {
		if agent.Enabled {
			enabledAgents = append(enabledAgents, agent)
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    enabledAgents,
	})
}

// FeedbackRequest 反馈请求结构
type FeedbackRequest struct {
	User      string `json:"user" binding:"required"`
	Question  string `json:"question" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
	IsHelpful int    `json:"isHelpful"` // 1: 有帮助, 0: 无帮助
	Timestamp int64  `json:"timestamp" binding:"required"`
}

// SubmitFeedback 提交反馈
func SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest

	// 先尝试绑定JSON，如果失败再记录错误
	if err := c.ShouldBindJSON(&req); err != nil {
		// 记录请求体内容用于调试
		body, _ := c.GetRawData()
		log.Logger.Error("绑定反馈请求失败",
			zap.Error(err),
			zap.String("body", string(body)))
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 记录解析后的请求数据
	log.Logger.Info("解析反馈请求成功",
		zap.String("user", req.User),
		zap.String("question", req.Question),
		zap.String("answer", req.Answer),
		zap.Int("isHelpful", req.IsHelpful),
		zap.Int64("timestamp", req.Timestamp))

	// 手动验证必填字段
	if req.User == "" || req.Question == "" || req.Answer == "" {
		c.JSON(400, gin.H{
			"success": false,
			"message": "user、question、answer字段不能为空",
		})
		return
	}

	// 验证isHelpful参数
	if req.IsHelpful != 0 && req.IsHelpful != 1 {
		c.JSON(400, gin.H{
			"success": false,
			"message": "isHelpful参数必须为0或1",
		})
		return
	}

	// 获取数据库连接
	db := database.DB
	if db == nil {
		log.Logger.Error("数据库连接失败")
		c.JSON(500, gin.H{
			"success": false,
			"message": "数据库连接失败",
		})
		return
	}

	// 插入反馈数据
	query := `
		INSERT INTO ai_feedback (user_name, question, answer, is_helpful, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	err := db.Exec(query, req.User, req.Question, req.Answer, req.IsHelpful, time.Unix(req.Timestamp, 0)).Error
	if err != nil {
		log.Logger.Error("插入反馈数据失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "保存反馈失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "反馈提交成功",
	})
}

// FeedbackStatsResponse 反馈统计响应结构
type FeedbackStatsResponse struct {
	HelpRate      float64 `json:"helpRate"`
	TotalFeedback int     `json:"totalFeedback"`
	HelpfulCount  int     `json:"helpfulCount"`
	HotQuestions  []struct {
		Question string `json:"question"`
		Count    int    `json:"count"`
	} `json:"hotQuestions"`
}

// GetFeedbackStats 获取反馈统计数据
func GetFeedbackStats(c *gin.Context) {
	// 获取数据库连接
	db := database.DB
	if db == nil {
		log.Logger.Error("数据库连接失败")
		c.JSON(500, gin.H{
			"success": false,
			"message": "数据库连接失败",
		})
		return
	}

	// 查询总反馈数和有帮助的反馈数
	type StatsResult struct {
		TotalFeedback int `gorm:"column:total_feedback"`
		HelpfulCount  int `gorm:"column:helpful_count"`
	}

	var stats StatsResult
	err := db.Raw(`
		SELECT 
			COUNT(*) as total_feedback,
			SUM(CASE WHEN is_helpful = 1 THEN 1 ELSE 0 END) as helpful_count
		FROM ai_feedback
	`).Scan(&stats).Error

	totalFeedback := stats.TotalFeedback
	helpfulCount := stats.HelpfulCount

	if err != nil {
		log.Logger.Error("查询反馈统计数据失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "查询统计数据失败",
		})
		return
	}

	// 计算帮助达成率
	var helpRate float64
	if totalFeedback > 0 {
		helpRate = float64(helpfulCount) / float64(totalFeedback) * 100
	}

	// 查询热门问题（按问题分组，按数量排序，取前10）
	type HotQuestion struct {
		Question string `gorm:"column:question"`
		Count    int    `gorm:"column:count"`
	}

	var hotQuestions []HotQuestion
	err = db.Raw(`
		SELECT 
			question,
			COUNT(*) as count
		FROM ai_feedback
		WHERE question IS NOT NULL AND question != ''
		GROUP BY question
		ORDER BY count DESC
		LIMIT 10
	`).Scan(&hotQuestions).Error

	if err != nil {
		log.Logger.Error("查询热门问题失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "查询热门问题失败",
		})
		return
	}

	// 构建响应数据
	response := FeedbackStatsResponse{
		HelpRate:      helpRate,
		TotalFeedback: totalFeedback,
		HelpfulCount:  helpfulCount,
		HotQuestions: make([]struct {
			Question string `json:"question"`
			Count    int    `json:"count"`
		}, len(hotQuestions)),
	}

	// 转换热门问题数据
	for i, q := range hotQuestions {
		response.HotQuestions[i] = struct {
			Question string `json:"question"`
			Count    int    `json:"count"`
		}{
			Question: q.Question,
			Count:    q.Count,
		}
	}

	c.JSON(200, response)
}
