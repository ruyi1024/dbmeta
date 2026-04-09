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

package ai

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// LogAnalysisResult AI分析结果
type LogAnalysisResult struct {
	ID             int64     `json:"id"`
	SessionID      string    `json:"session_id"`
	AnalysisType   string    `json:"analysis_type"`  // 分析类型：error, performance, security, etc.
	Severity       string    `json:"severity"`       // 严重程度：critical, high, medium, low
	Summary        string    `json:"summary"`        // 分析摘要
	RootCause      string    `json:"root_cause"`     // 根本原因
	Recommendation string    `json:"recommendation"` // 建议措施
	Keywords       string    `json:"keywords"`       // 关键词
	Pattern        string    `json:"pattern"`        // 匹配模式
	Confidence     float64   `json:"confidence"`     // 置信度
	CreatedAt      time.Time `json:"created_at"`
}

// LogAnalyzer 日志分析器
type LogAnalyzer struct {
	ErrorPatterns       []ErrorPattern
	SecurityPatterns    []SecurityPattern
	PerformancePatterns []PerformancePattern
}

// ErrorPattern 错误模式
type ErrorPattern struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
}

// SecurityPattern 安全模式
type SecurityPattern struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
}

// PerformancePattern 性能模式
type PerformancePattern struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
}

// NewLogAnalyzer 创建日志分析器
func NewLogAnalyzer() *LogAnalyzer {
	return &LogAnalyzer{
		ErrorPatterns:       getDefaultErrorPatterns(),
		SecurityPatterns:    getDefaultSecurityPatterns(),
		PerformancePatterns: getDefaultPerformancePatterns(),
	}
}

// AnalyzeLog 分析日志
func (la *LogAnalyzer) AnalyzeLog(sessionID, message string) (*LogAnalysisResult, error) {
	// 1. 错误分析
	errorResult := la.analyzeErrors(message)
	if errorResult != nil {
		errorResult.SessionID = sessionID
		return errorResult, nil
	}

	// 2. 安全分析
	securityResult := la.analyzeSecurity(message)
	if securityResult != nil {
		securityResult.SessionID = sessionID
		return securityResult, nil
	}

	// 3. 性能分析
	performanceResult := la.analyzePerformance(message)
	if performanceResult != nil {
		performanceResult.SessionID = sessionID
		return performanceResult, nil
	}

	// 4. 通用分析
	generalResult := la.analyzeGeneral(message)
	generalResult.SessionID = sessionID
	return generalResult, nil
}

// analyzeErrors 分析错误
func (la *LogAnalyzer) analyzeErrors(message string) *LogAnalysisResult {
	for _, pattern := range la.ErrorPatterns {
		matched, _ := regexp.MatchString(pattern.Pattern, message)
		if matched {
			return &LogAnalysisResult{
				AnalysisType:   "error",
				Severity:       pattern.Severity,
				Summary:        pattern.Description,
				RootCause:      la.extractRootCause(message, pattern.Pattern),
				Recommendation: pattern.Solution,
				Keywords:       la.extractKeywords(message),
				Pattern:        pattern.Name,
				Confidence:     0.85,
				CreatedAt:      time.Now(),
			}
		}
	}
	return nil
}

// analyzeSecurity 分析安全问题
func (la *LogAnalyzer) analyzeSecurity(message string) *LogAnalysisResult {
	for _, pattern := range la.SecurityPatterns {
		matched, _ := regexp.MatchString(pattern.Pattern, message)
		if matched {
			return &LogAnalysisResult{
				AnalysisType:   "security",
				Severity:       pattern.Severity,
				Summary:        pattern.Description,
				RootCause:      la.extractRootCause(message, pattern.Pattern),
				Recommendation: pattern.Solution,
				Keywords:       la.extractKeywords(message),
				Pattern:        pattern.Name,
				Confidence:     0.90,
				CreatedAt:      time.Now(),
			}
		}
	}
	return nil
}

// analyzePerformance 分析性能问题
func (la *LogAnalyzer) analyzePerformance(message string) *LogAnalysisResult {
	for _, pattern := range la.PerformancePatterns {
		matched, _ := regexp.MatchString(pattern.Pattern, message)
		if matched {
			return &LogAnalysisResult{
				AnalysisType:   "performance",
				Severity:       pattern.Severity,
				Summary:        pattern.Description,
				RootCause:      la.extractRootCause(message, pattern.Pattern),
				Recommendation: pattern.Solution,
				Keywords:       la.extractKeywords(message),
				Pattern:        pattern.Name,
				Confidence:     0.80,
				CreatedAt:      time.Now(),
			}
		}
	}
	return nil
}

// analyzeGeneral 通用分析
func (la *LogAnalyzer) analyzeGeneral(message string) *LogAnalysisResult {
	// 基于关键词的通用分析
	keywords := la.extractKeywords(message)
	severity := la.determineSeverity(message)

	return &LogAnalysisResult{
		AnalysisType:   "general",
		Severity:       severity,
		Summary:        fmt.Sprintf("检测到日志活动，包含关键词: %s", keywords),
		RootCause:      "需要进一步分析",
		Recommendation: "建议查看详细日志信息",
		Keywords:       keywords,
		Pattern:        "general",
		Confidence:     0.60,
		CreatedAt:      time.Now(),
	}
}

// extractRootCause 提取根本原因
func (la *LogAnalyzer) extractRootCause(message, pattern string) string {
	// 简单的根本原因提取逻辑
	if strings.Contains(strings.ToLower(message), "timeout") {
		return "网络或服务超时"
	}
	if strings.Contains(strings.ToLower(message), "connection") {
		return "连接问题"
	}
	if strings.Contains(strings.ToLower(message), "memory") {
		return "内存不足"
	}
	if strings.Contains(strings.ToLower(message), "disk") {
		return "磁盘空间不足"
	}
	return "需要进一步分析"
}

// extractKeywords 提取关键词
func (la *LogAnalyzer) extractKeywords(message string) string {
	keywords := []string{}

	// 提取常见关键词
	commonKeywords := []string{
		"error", "failed", "exception", "timeout", "connection",
		"memory", "disk", "database", "network", "security",
		"authentication", "authorization", "permission",
	}

	messageLower := strings.ToLower(message)
	for _, keyword := range commonKeywords {
		if strings.Contains(messageLower, keyword) {
			keywords = append(keywords, keyword)
		}
	}

	return strings.Join(keywords, ",")
}

// determineSeverity 确定严重程度
func (la *LogAnalyzer) determineSeverity(message string) string {
	messageLower := strings.ToLower(message)

	if strings.Contains(messageLower, "critical") || strings.Contains(messageLower, "fatal") {
		return "critical"
	}
	if strings.Contains(messageLower, "error") || strings.Contains(messageLower, "exception") {
		return "high"
	}
	if strings.Contains(messageLower, "warning") || strings.Contains(messageLower, "warn") {
		return "medium"
	}
	return "low"
}

// getDefaultErrorPatterns 获取默认错误模式
func getDefaultErrorPatterns() []ErrorPattern {
	return []ErrorPattern{
		{
			Name:        "Database Connection Error",
			Pattern:     `(?i)(database|db).*(connection|connect).*(failed|error|timeout)`,
			Severity:    "high",
			Description: "数据库连接错误",
			Solution:    "检查数据库服务状态和连接配置",
		},
		{
			Name:        "Memory Out of Error",
			Pattern:     `(?i)(out of memory|memory.*error|oom)`,
			Severity:    "critical",
			Description: "内存不足错误",
			Solution:    "增加内存或优化内存使用",
		},
		{
			Name:        "Network Timeout",
			Pattern:     `(?i)(timeout|timed out|connection.*timeout)`,
			Severity:    "medium",
			Description: "网络超时",
			Solution:    "检查网络连接和超时配置",
		},
	}
}

// getDefaultSecurityPatterns 获取默认安全模式
func getDefaultSecurityPatterns() []SecurityPattern {
	return []SecurityPattern{
		{
			Name:        "Authentication Failed",
			Pattern:     `(?i)(authentication|auth).*(failed|denied|invalid)`,
			Severity:    "high",
			Description: "认证失败",
			Solution:    "检查用户凭据和认证配置",
		},
		{
			Name:        "Unauthorized Access",
			Pattern:     `(?i)(unauthorized|access.*denied|permission.*denied)`,
			Severity:    "high",
			Description: "未授权访问",
			Solution:    "检查用户权限和访问控制",
		},
		{
			Name:        "Security Violation",
			Pattern:     `(?i)(security|violation|breach|attack)`,
			Severity:    "critical",
			Description: "安全违规",
			Solution:    "立即检查安全事件并采取相应措施",
		},
	}
}

// getDefaultPerformancePatterns 获取默认性能模式
func getDefaultPerformancePatterns() []PerformancePattern {
	return []PerformancePattern{
		{
			Name:        "Slow Query",
			Pattern:     `(?i)(slow.*query|query.*slow|performance.*issue)`,
			Severity:    "medium",
			Description: "慢查询检测",
			Solution:    "优化查询语句和数据库索引",
		},
		{
			Name:        "High CPU Usage",
			Pattern:     `(?i)(cpu.*high|high.*cpu|performance.*degradation)`,
			Severity:    "medium",
			Description: "CPU使用率过高",
			Solution:    "检查CPU密集型任务和优化代码",
		},
		{
			Name:        "Disk Space Issue",
			Pattern:     `(?i)(disk.*full|space.*low|storage.*issue)`,
			Severity:    "high",
			Description: "磁盘空间不足",
			Solution:    "清理磁盘空间或增加存储容量",
		},
	}
}
