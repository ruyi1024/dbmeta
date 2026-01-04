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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/nsqio/go-nsq"

	"dbmcloud/log"
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/libary/ai"
)

// LogEntry 日志条目结构（与agent中的LogEntry保持一致）
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields"`
	Raw       string            `json:"raw"`
	HostIP    string            `json:"host_ip"`  // 主机IP地址
	Hostname  string            `json:"hostname"` // 主机名
	LogFile   string            `json:"log_file"` // 日志文件名
	LogPath   string            `json:"log_path"` // 日志文件完整路径
}

// CollectedLog 收集到的日志记录
type CollectedLog struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Source      string    `gorm:"size:50" json:"source"`                   // 日志源
	HostIP      string    `gorm:"size:45" json:"host_ip"`                  // 主机IP
	Hostname    string    `gorm:"size:100" json:"hostname"`                // 主机名
	Level       string    `gorm:"size:20" json:"level"`                    // 日志级别
	Message     string    `gorm:"type:text" json:"message"`                // 日志消息
	Raw         string    `gorm:"type:longtext" json:"raw"`                // 原始日志
	Fields      string    `gorm:"type:text" json:"fields"`                 // 额外字段(JSON格式)
	LogTime     time.Time `json:"log_time"`                                // 日志时间
	CollectedAt time.Time `gorm:"column:collected_at" json:"collected_at"` // 收集时间

	// 文件信息字段
	LogFile string `gorm:"size:255" json:"log_file"` // 日志文件名
	LogPath string `gorm:"size:500" json:"log_path"` // 日志文件完整路径

	// AI分析相关字段
	SessionID         string `gorm:"size:100" json:"session_id"`              // 会话ID，用于关联相关日志
	GroupID           string `gorm:"size:100" json:"group_id"`                // 分组ID，用于日志聚合
	IsComplete        bool   `gorm:"default:false" json:"is_complete"`        // 是否为完整日志
	AggregatedMessage string `gorm:"type:longtext" json:"aggregated_message"` // 聚合后的完整消息
}

func (CollectedLog) TableName() string {
	return "log_collected"
}

func init() {
	go logConsumer()
}

func logConsumer() {
	time.Sleep(time.Second * time.Duration(60))
	start := time.Now()
	fmt.Printf("Log consumer start at %s \n", start)
	log.Logger.Info(fmt.Sprintf("Log consumer start at %s", start))

	runtime.GOMAXPROCS(runtime.NumCPU())

	// 创建NSQ消费者，消费log-collector主题
	consumer, err := nsq.NewConsumer("log-collector", "log-processor", nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	consumer.AddHandler(&LogConsumerT{})

	// 连接到NSQ服务器
	if err := consumer.ConnectToNSQD(setting.Setting.NsqServer); err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	select {
	case <-signals:
	}
}

// LogConsumerT NSQ消息处理器
type LogConsumerT struct{}

func (*LogConsumerT) HandleMessage(msg *nsq.Message) error {
	// 处理接收到的日志消息
	processLogMessage(string(msg.Body))
	return nil
}

// processLogMessage 处理日志消息
func processLogMessage(messageBody string) {
	// 添加消息ID用于去重检测
	messageHash := fmt.Sprintf("%x", md5.Sum([]byte(messageBody)))

	// 解析JSON消息
	var logEntry LogEntry
	err := json.Unmarshal([]byte(messageBody), &logEntry)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("解析日志消息失败: %v, 消息内容: %s", err, messageBody))
		return
	}

	// 调试信息：显示接收到的消息
	log.Logger.Info(fmt.Sprintf("接收到日志消息: 来源=%s, 主机=%s, 级别=%s, 消息哈希=%s",
		logEntry.Source, logEntry.Hostname, logEntry.Level, messageHash))

	// 创建数据库记录
	collectedLog := CollectedLog{
		Source:      logEntry.Source,
		HostIP:      logEntry.HostIP,
		Hostname:    logEntry.Hostname,
		Level:       logEntry.Level,
		Message:     logEntry.Message,
		Raw:         logEntry.Raw,
		LogTime:     logEntry.Timestamp,
		CollectedAt: time.Now(),
		LogFile:     logEntry.LogFile,
		LogPath:     logEntry.LogPath,
	}

	// 序列化Fields字段
	if len(logEntry.Fields) > 0 {
		fieldsJSON, err := json.Marshal(logEntry.Fields)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("序列化Fields字段失败: %v", err))
		} else {
			collectedLog.Fields = string(fieldsJSON)
		}
	}

	// 保存到数据库
	if err := database.DB.Create(&collectedLog).Error; err != nil {
		log.Logger.Error(fmt.Sprintf("保存日志到数据库失败: %v", err))
		return
	}

	// 记录成功日志
	log.Logger.Info(fmt.Sprintf("成功保存日志: 来源=%s, 主机=%s(%s), 文件=%s, 路径=%s, 级别=%s, 消息=%s",
		collectedLog.Source, collectedLog.Hostname, collectedLog.HostIP,
		collectedLog.LogFile, collectedLog.LogPath, collectedLog.Level, truncateString(collectedLog.Message, 100)))

	// 尝试聚合相关日志
	aggregatedLog := tryAggregateLogs(&collectedLog)
	if aggregatedLog != nil {
		// 处理聚合后的完整日志
		processAggregatedLog(aggregatedLog)
	} else {
		// 根据日志级别进行进一步处理
		processLogByLevel(&collectedLog)
	}
}

// processLogByLevel 根据日志级别进行进一步处理
func processLogByLevel(log *CollectedLog) {
	switch strings.ToUpper(log.Level) {
	case "ERROR", "FATAL", "CRITICAL":
		// 处理错误日志
		processErrorLog(log)
	case "WARNING", "WARN":
		// 处理警告日志
		processWarningLog(log)
	case "INFO":
		// 处理信息日志
		processInfoLog(log)
	default:
		// 处理其他级别日志
		processOtherLog(log)
	}
}

// processErrorLog 处理错误日志
func processErrorLog(logEntry *CollectedLog) {
	log.Logger.Warn(fmt.Sprintf("检测到错误日志: 主机=%s, 级别=%s, 消息=%s",
		logEntry.Hostname, logEntry.Level, truncateString(logEntry.Message, 200)))

	// 这里可以添加错误日志的特殊处理逻辑
	// 例如：发送告警、统计错误数量等
}

// processWarningLog 处理警告日志
func processWarningLog(logEntry *CollectedLog) {
	log.Logger.Info(fmt.Sprintf("检测到警告日志: 主机=%s, 级别=%s, 消息=%s",
		logEntry.Hostname, logEntry.Level, truncateString(logEntry.Message, 200)))

	// 这里可以添加警告日志的特殊处理逻辑
}

// processInfoLog 处理信息日志
func processInfoLog(logEntry *CollectedLog) {
	// 信息日志通常不需要特殊处理，只记录即可
	log.Logger.Debug(fmt.Sprintf("处理信息日志: 主机=%s, 消息=%s",
		logEntry.Hostname, truncateString(logEntry.Message, 100)))
}

// processOtherLog 处理其他级别日志
func processOtherLog(logEntry *CollectedLog) {
	log.Logger.Debug(fmt.Sprintf("处理其他级别日志: 主机=%s, 级别=%s, 消息=%s",
		logEntry.Hostname, logEntry.Level, truncateString(logEntry.Message, 100)))
}

// tryAggregateLogs 尝试聚合相关日志
func tryAggregateLogs(currentLog *CollectedLog) *CollectedLog {
	// 生成会话ID和分组ID
	sessionID := generateSessionID(currentLog)
	groupID := generateGroupID(currentLog)

	// 更新当前日志的分组信息
	currentLog.SessionID = sessionID
	currentLog.GroupID = groupID

	// 查找相关的日志记录（最近5分钟内的相同主机和来源）
	var relatedLogs []CollectedLog
	timeWindow := currentLog.LogTime.Add(-5 * time.Minute)

	database.DB.Where(
		"source = ? AND host_ip = ? AND hostname = ? AND log_time >= ? AND log_time <= ? AND id != ?",
		currentLog.Source, currentLog.HostIP, currentLog.Hostname,
		timeWindow, currentLog.LogTime, currentLog.ID,
	).Order("log_time ASC").Find(&relatedLogs)

	if len(relatedLogs) == 0 {
		// 没有相关日志，检查当前日志是否完整
		if isCompleteLog(currentLog.Message) {
			currentLog.IsComplete = true
			currentLog.AggregatedMessage = currentLog.Message
			database.DB.Save(currentLog)
			return currentLog
		}
		return nil
	}

	// 聚合相关日志
	aggregatedMessage := aggregateLogMessages(append(relatedLogs, *currentLog))

	// 检查是否形成完整日志
	if isCompleteLog(aggregatedMessage) {
		// 更新所有相关日志的聚合信息
		updateLogsAggregation(append(relatedLogs, *currentLog), sessionID, groupID, aggregatedMessage, true)

		// 返回聚合后的完整日志
		completeLog := &CollectedLog{
			Source:            currentLog.Source,
			HostIP:            currentLog.HostIP,
			Hostname:          currentLog.Hostname,
			Level:             getHighestLevel(append(relatedLogs, *currentLog)),
			Message:           aggregatedMessage,
			Raw:               aggregatedMessage,
			LogTime:           currentLog.LogTime,
			CollectedAt:       time.Now(),
			SessionID:         sessionID,
			GroupID:           groupID,
			IsComplete:        true,
			AggregatedMessage: aggregatedMessage,
		}

		// 保存聚合后的完整日志
		database.DB.Create(completeLog)

		log.Logger.Info(fmt.Sprintf("成功聚合日志: 会话ID=%s, 分组ID=%s, 聚合消息长度=%d",
			sessionID, groupID, len(aggregatedMessage)))

		return completeLog
	}

	return nil
}

// generateSessionID 生成会话ID
func generateSessionID(log *CollectedLog) string {
	// 基于主机、来源和时间窗口生成会话ID
	timeWindow := log.LogTime.Truncate(5 * time.Minute) // 5分钟时间窗口
	return fmt.Sprintf("%s_%s_%s_%d", log.Source, log.HostIP, log.Hostname, timeWindow.Unix())
}

// generateGroupID 生成分组ID
func generateGroupID(log *CollectedLog) string {
	// 基于消息特征生成分组ID
	messageHash := fmt.Sprintf("%x", md5.Sum([]byte(log.Message)))
	return fmt.Sprintf("%s_%s_%s", log.Source, log.HostIP, messageHash[:8])
}

// isCompleteLog 判断日志是否完整
func isCompleteLog(message string) bool {
	// 检查日志是否包含完整的错误信息
	completeIndicators := []string{
		"Exception", "Error", "Failed", "Success", "Complete",
		"finished", "completed", "ended", "started",
		"stack trace", "traceback", "caused by",
	}

	messageLower := strings.ToLower(message)
	for _, indicator := range completeIndicators {
		if strings.Contains(messageLower, strings.ToLower(indicator)) {
			return true
		}
	}

	// 检查消息长度（过短的消息可能不完整）
	if len(strings.TrimSpace(message)) < 50 {
		return false
	}

	// 检查是否包含时间戳和级别信息
	hasTimestamp := strings.Contains(message, "2025-") || strings.Contains(message, "2024-")
	hasLevel := strings.Contains(strings.ToUpper(message), "ERROR") ||
		strings.Contains(strings.ToUpper(message), "INFO") ||
		strings.Contains(strings.ToUpper(message), "WARN")

	return hasTimestamp && hasLevel
}

// aggregateLogMessages 聚合日志消息
func aggregateLogMessages(logs []CollectedLog) string {
	// 按时间排序
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].LogTime.Before(logs[j].LogTime)
	})

	var messages []string
	for _, logEntry := range logs {
		messages = append(messages, logEntry.Message)
	}

	// 合并消息，用换行符分隔
	return strings.Join(messages, "\n")
}

// getHighestLevel 获取最高日志级别
func getHighestLevel(logs []CollectedLog) string {
	levelPriority := map[string]int{
		"CRITICAL": 5,
		"FATAL":    4,
		"ERROR":    3,
		"WARNING":  2,
		"WARN":     2,
		"INFO":     1,
		"DEBUG":    0,
	}

	highestLevel := "INFO"
	maxPriority := 0

	for _, logEntry := range logs {
		if priority, exists := levelPriority[strings.ToUpper(logEntry.Level)]; exists {
			if priority > maxPriority {
				maxPriority = priority
				highestLevel = logEntry.Level
			}
		}
	}

	return highestLevel
}

// updateLogsAggregation 更新日志聚合信息
func updateLogsAggregation(logs []CollectedLog, sessionID, groupID, aggregatedMessage string, isComplete bool) {
	for _, logEntry := range logs {
		logEntry.SessionID = sessionID
		logEntry.GroupID = groupID
		logEntry.AggregatedMessage = aggregatedMessage
		logEntry.IsComplete = isComplete
		database.DB.Save(&logEntry)
	}
}

// processAggregatedLog 处理聚合后的完整日志
func processAggregatedLog(logEntry *CollectedLog) {
	log.Logger.Info(fmt.Sprintf("处理聚合后的完整日志: 会话ID=%s, 级别=%s, 消息长度=%d",
		logEntry.SessionID, logEntry.Level, len(logEntry.AggregatedMessage)))

	// 这里可以添加AI分析逻辑
	// 例如：调用AI服务进行日志分析
	performAIAnalysis(logEntry)
}

// performAIAnalysis 执行AI分析
func performAIAnalysis(logEntry *CollectedLog) {
	// 创建AI分析器
	analyzer := ai.NewLogAnalyzer()

	// 分析日志内容
	result, err := analyzer.AnalyzeLog(logEntry.SessionID, logEntry.AggregatedMessage)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("AI分析失败: %v", err))
		return
	}

	// 保存分析结果到数据库
	if err := database.DB.Create(result).Error; err != nil {
		log.Logger.Error(fmt.Sprintf("保存AI分析结果失败: %v", err))
		return
	}

	log.Logger.Info(fmt.Sprintf("AI分析完成: 会话ID=%s, 分析类型=%s, 严重程度=%s, 置信度=%.2f",
		result.SessionID, result.AnalysisType, result.Severity, result.Confidence))
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
