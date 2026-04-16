package task

import (
	"bytes"
	"dbmeta-core/log"
	"dbmeta-core/src/database"

	"dbmeta-core/src/libary/tool"
	"dbmeta-core/src/model"
	"dbmeta-core/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go collectorApiEventTask()
}

func collectorApiEventTask() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "collector_api_event").Take(&record)
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "collector_api_event").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='collector_api_event'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doCollectorApiEventTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='collector_api_event'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func doCollectorApiEventTask() {
	logger := log.Logger
	logger.Info("开始执行API接口监控任务")

	var db = database.DB
	var dataList []model.ApiConfig
	result := db.Where("enable=1").Find(&dataList)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("查询启用的API配置失败: %v", result.Error)
		logger.Error(errorMsg)
		return
	}

	if len(dataList) == 0 {
		successMsg := "没有启用的API配置需要监控"
		logger.Info(successMsg)
		return
	}

	logger.Info("找到需要监控的API配置", zap.Int("count", len(dataList)))

	successCount := 0
	failedCount := 0
	errorDetails := []string{}

	for i, apiConfig := range dataList {
		logger.Info("开始监控API",
			zap.String("api_name", apiConfig.ApiName),
			zap.String("api_url", apiConfig.ApiUrl),
			zap.String("method", apiConfig.Method),
			zap.Int("index", i+1),
			zap.Int("total", len(dataList)))

		err := startCollectApiEvent(apiConfig)
		if err != nil {
			failedCount++
			errorMsg := fmt.Sprintf("监控API %s 失败: %v", apiConfig.ApiName, err)
			errorDetails = append(errorDetails, errorMsg)
			logger.Error(errorMsg)
		} else {
			successCount++
			logger.Info("API监控成功", zap.String("api_name", apiConfig.ApiName))
		}

		if (i+1)%5 == 0 || i == len(dataList)-1 {
			progressMsg := fmt.Sprintf("已处理 %d/%d 个API (成功: %d, 失败: %d)",
				i+1, len(dataList), successCount, failedCount)
			logger.Info(progressMsg)
		}
	}

	var finalResult string
	if failedCount == 0 {
		finalResult = fmt.Sprintf("API监控任务完成，共监控 %d 个API，全部成功", successCount)
		logger.Info(finalResult)
	} else {
		finalResult = fmt.Sprintf("API监控任务完成，共监控 %d 个API，成功: %d, 失败: %d",
			len(dataList), successCount, failedCount)
		if len(errorDetails) > 0 {
			finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0])
			if len(errorDetails) > 1 {
				finalResult += fmt.Sprintf(" 等%d个错误", len(errorDetails))
			}
		}
		logger.Warn(finalResult)
	}
}

func startCollectApiEvent(apiConfig model.ApiConfig) error {
	eventEntity := apiConfig.ApiUrl
	eventType := "ApiMonitor"
	eventGroup := "API_Monitor"

	// 解析URL
	_, err := url.Parse(apiConfig.ApiUrl)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid API URL %s: %s", apiConfig.ApiUrl, err.Error())
		log.Logger.Error(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	// 构建请求
	response, responseTime, err := makeApiRequest(apiConfig)
	responseTimeMs := float64(responseTime.Nanoseconds()) / 1000000.0

	if response != nil {
		defer response.Body.Close()
	}

	events := make([]map[string]interface{}, 0)

	// 记录HTTP状态码事件
	statusCode := 0
	if response != nil {
		statusCode = response.StatusCode
	}

	statusDetail := make([]map[string]interface{}, 0)
	statusDetail = append(statusDetail, map[string]interface{}{
		"ApiName":     apiConfig.ApiName,
		"ApiUrl":      apiConfig.ApiUrl,
		"Method":      apiConfig.Method,
		"Status":      statusCode,
		"Description": apiConfig.ApiDescription,
	})

	statusEvent := map[string]interface{}{
		"event_uuid":   tool.GetUUID(),
		"event_time":   tool.GetNowTime(),
		"event_type":   eventType,
		"event_group":  eventGroup,
		"event_entity": eventEntity,
		"event_key":    "http_status",
		"event_value":  utils.IntToDecimal(statusCode),
		"event_tag":    apiConfig.ApiName,
		"event_unit":   "",
		"event_detail": utils.MapToStr(statusDetail),
	}
	events = append(events, statusEvent)

	// 记录响应时间事件
	responseTimeDetail := make([]map[string]interface{}, 0)
	responseTimeDetail = append(responseTimeDetail, map[string]interface{}{
		"ApiName":      apiConfig.ApiName,
		"ApiUrl":       apiConfig.ApiUrl,
		"Method":       apiConfig.Method,
		"ResponseTime": fmt.Sprintf("%.2fms", responseTimeMs),
		"Description":  apiConfig.ApiDescription,
	})

	responseTimeEvent := map[string]interface{}{
		"event_uuid":   tool.GetUUID(),
		"event_time":   tool.GetNowTime(),
		"event_type":   eventType,
		"event_group":  eventGroup,
		"event_entity": eventEntity,
		"event_key":    "response_time",
		"event_value":  utils.FloatToDecimal(responseTimeMs),
		"event_tag":    apiConfig.ApiName,
		"event_unit":   "ms",
		"event_detail": utils.MapToStr(responseTimeDetail),
	}
	events = append(events, responseTimeEvent)

	// 记录连接状态事件
	connectStatus := 1
	if err != nil {
		connectStatus = 0
	}

	connectDetail := make([]map[string]interface{}, 0)
	if err != nil {
		connectDetail = append(connectDetail, map[string]interface{}{
			"ApiName":      apiConfig.ApiName,
			"ApiUrl":       apiConfig.ApiUrl,
			"Method":       apiConfig.Method,
			"Error":        err.Error(),
			"ResponseTime": fmt.Sprintf("%.2fms", responseTimeMs),
			"Description":  apiConfig.ApiDescription,
		})
	} else {
		connectDetail = append(connectDetail, map[string]interface{}{
			"ApiName":      apiConfig.ApiName,
			"ApiUrl":       apiConfig.ApiUrl,
			"Method":       apiConfig.Method,
			"Status":       statusCode,
			"ResponseTime": fmt.Sprintf("%.2fms", responseTimeMs),
			"Description":  apiConfig.ApiDescription,
		})
	}

	connectEvent := map[string]interface{}{
		"event_uuid":   tool.GetUUID(),
		"event_time":   tool.GetNowTime(),
		"event_type":   eventType,
		"event_group":  eventGroup,
		"event_entity": eventEntity,
		"event_key":    "connect_status",
		"event_value":  utils.IntToDecimal(connectStatus),
		"event_tag":    apiConfig.ApiName,
		"event_unit":   "",
		"event_detail": utils.MapToStr(connectDetail),
	}
	events = append(events, connectEvent)

	// 记录健康状态事件（基于期望返回码）
	healthStatus := checkHealthStatus(statusCode, apiConfig.ExpectedCodes)
	healthDetail := make([]map[string]interface{}, 0)
	healthDetail = append(healthDetail, map[string]interface{}{
		"ApiName":       apiConfig.ApiName,
		"ApiUrl":        apiConfig.ApiUrl,
		"Method":        apiConfig.Method,
		"ActualStatus":  statusCode,
		"ExpectedCodes": apiConfig.ExpectedCodes,
		"HealthStatus":  healthStatus,
		"ResponseTime":  fmt.Sprintf("%.2fms", responseTimeMs),
		"Description":   apiConfig.ApiDescription,
	})

	healthEvent := map[string]interface{}{
		"event_uuid":   tool.GetUUID(),
		"event_time":   tool.GetNowTime(),
		"event_type":   eventType,
		"event_group":  eventGroup,
		"event_entity": eventEntity,
		"event_key":    "health_status",
		"event_value":  utils.IntToDecimal(healthStatus),
		"event_tag":    apiConfig.ApiName,
		"event_unit":   "",
		"event_detail": utils.MapToStr(healthDetail),
	}
	events = append(events, healthEvent)

	// 写入ClickHouse
	result := database.CK.Model(&model.Event{}).Create(events)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("Can't add API events data to clickhouse: %s", result.Error.Error())
		fmt.Println("Insert API Event To Clickhouse Error: " + result.Error.Error())
		log.Logger.Error(errorMsg)
		return fmt.Errorf(errorMsg)
	}

	if err != nil {
		log.Logger.Warn(fmt.Sprintf("API monitoring failed for %s: %s, Response Time: %.2fms",
			apiConfig.ApiName, err.Error(), responseTimeMs))
	} else {
		log.Logger.Info(fmt.Sprintf("API monitoring completed for %s: Status: %d, Response Time: %.2fms, Health: %d",
			apiConfig.ApiName, statusCode, responseTimeMs, healthStatus))
	}

	return nil
}

// makeApiRequest 执行API请求
func makeApiRequest(apiConfig model.ApiConfig) (*http.Response, time.Duration, error) {
	startTime := time.Now()

	// 构建完整URL（包含查询参数）
	fullURL := apiConfig.ApiUrl
	if apiConfig.Params != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(apiConfig.Params), &params); err == nil {
			query := url.Values{}
			for key, value := range params {
				query.Add(key, fmt.Sprintf("%v", value))
			}
			if len(query) > 0 {
				separator := "?"
				if strings.Contains(fullURL, "?") {
					separator = "&"
				}
				fullURL += separator + query.Encode()
			}
		}
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(apiConfig.Timeout) * time.Second,
		Transport: &http.Transport{
			Proxy: nil, // 禁用代理
		},
	}

	// 创建请求
	var req *http.Request
	var err error

	switch strings.ToUpper(apiConfig.Method) {
	case "GET":
		req, err = http.NewRequest("GET", fullURL, nil)
	case "POST":
		var body io.Reader
		if apiConfig.Body != "" {
			body = bytes.NewBufferString(apiConfig.Body)
		}
		req, err = http.NewRequest("POST", fullURL, body)
	case "PUT":
		var body io.Reader
		if apiConfig.Body != "" {
			body = bytes.NewBufferString(apiConfig.Body)
		}
		req, err = http.NewRequest("PUT", fullURL, body)
	case "DELETE":
		req, err = http.NewRequest("DELETE", fullURL, nil)
	default:
		req, err = http.NewRequest("GET", fullURL, nil)
	}

	if err != nil {
		return nil, time.Since(startTime), err
	}

	// 设置请求头
	if apiConfig.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(apiConfig.Headers), &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}
	}

	// 设置认证信息
	if apiConfig.Token != "" {
		switch strings.ToUpper(apiConfig.AuthType) {
		case "BEARER":
			req.Header.Set("Authorization", "Bearer "+apiConfig.Token)
		case "BASIC":
			req.Header.Set("Authorization", "Basic "+apiConfig.Token)
		case "API_KEY":
			req.Header.Set("X-API-Key", apiConfig.Token)
		}
	}

	// 执行请求
	response, err := client.Do(req)
	responseTime := time.Since(startTime)

	return response, responseTime, err
}

// checkHealthStatus 检查健康状态
func checkHealthStatus(actualStatus int, expectedCodes string) int {
	if expectedCodes == "" {
		expectedCodes = "200"
	}

	expectedList := strings.Split(expectedCodes, ",")
	for _, expected := range expectedList {
		expected = strings.TrimSpace(expected)
		if expectedCode, err := strconv.Atoi(expected); err == nil {
			if actualStatus == expectedCode {
				return 1 // 健康
			}
		}
	}

	return 0 // 不健康
}
