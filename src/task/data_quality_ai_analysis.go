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
	"dbmcloud/log"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/service"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go dataQualityAiAnalysisCrontabTask()
}

// dataQualityAiAnalysisCrontabTask 启动数据质量AI分析定时任务
func dataQualityAiAnalysisCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))

	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "data_quality_ai_analysis").Take(&record)

	// 如果任务配置不存在，使用默认的cron表达式（每小时执行一次）
	if record.Crontab == "" {
		record.Crontab = "0 * * * *" // 每小时执行一次
	}

	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "data_quality_ai_analysis").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='data_quality_ai_analysis'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doDataQualityAiAnalysis()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='data_quality_ai_analysis'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

// doDataQualityAiAnalysis 执行数据质量AI分析任务
func doDataQualityAiAnalysis() {
	logger := log.Logger
	logger.Info("开始执行数据质量AI分析任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("data_quality_ai_analysis")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	// 获取最新的评估记录（未进行AI分析的）
	var latestAssessment model.DataQualityAssessment
	result := database.DB.Where("status = 1").
		Order("assessment_time DESC").
		First(&latestAssessment)

	if result.Error != nil || latestAssessment.Id == 0 {
		noDataMsg := "没有找到可分析的数据质量评估记录"
		logger.Info(noDataMsg)
		taskLogger.Success(noDataMsg)
		return
	}

	// 检查是否已经进行过AI分析
	var existingAnalysis model.DataQualityAiAnalysis
	database.DB.Where("assessment_id = ?", latestAssessment.Id).First(&existingAnalysis)
	if existingAnalysis.Id > 0 {
		skipMsg := fmt.Sprintf("评估记录 ID %d 已经进行过AI分析，跳过", latestAssessment.Id)
		logger.Info(skipMsg)
		taskLogger.Success(skipMsg)
		return
	}

	logger.Info("找到待分析的评估记录",
		zap.Int64("assessment_id", latestAssessment.Id),
		zap.String("database", latestAssessment.DatabaseName),
		zap.Float64("overall_score", latestAssessment.OverallScore))

	taskLogger.UpdateResult(fmt.Sprintf("开始分析评估记录 ID %d (数据库: %s, 总体评分: %.2f)",
		latestAssessment.Id, latestAssessment.DatabaseName, latestAssessment.OverallScore))

	// 获取评估相关的数据
	analysisData, err := prepareAnalysisData(latestAssessment)
	if err != nil {
		errorMsg := fmt.Sprintf("准备分析数据失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 获取启用的AI模型（优先使用优先级最高的）
	aiModels, err := service.GetEnabledModels()
	if err != nil || len(aiModels) == 0 {
		errorMsg := "没有找到启用的AI模型"
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 使用第一个启用的模型（优先级最高的）
	aiModel := aiModels[0]
	logger.Info("使用AI模型进行分析",
		zap.String("model_name", aiModel.Name),
		zap.String("provider", aiModel.Provider))

	// 调用AI进行分析
	aiAnalysisResult, err := performAiAnalysis(aiModel, latestAssessment, analysisData, taskLogger)
	if err != nil {
		errorMsg := fmt.Sprintf("AI分析失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 保存AI分析结果
	if err := saveAiAnalysisResult(latestAssessment.Id, aiModel, aiAnalysisResult); err != nil {
		errorMsg := fmt.Sprintf("保存AI分析结果失败: %v", err)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	successMsg := fmt.Sprintf("AI分析完成，已保存分析结果 (评估ID: %d, 模型: %s)", latestAssessment.Id, aiModel.Name)
	logger.Info(successMsg)
	taskLogger.Success(successMsg)
}

// prepareAnalysisData 准备分析数据
func prepareAnalysisData(assessment model.DataQualityAssessment) (map[string]interface{}, error) {
	// 获取质量问题列表
	var issues []model.DataQualityIssue
	database.DB.Where("assessment_id = ?", assessment.Id).
		Order("issue_level DESC, issue_count DESC").
		Limit(100). // 限制最多100个问题，避免数据过大
		Find(&issues)

	// 按问题类型分组统计
	issueStats := make(map[string]map[string]interface{})
	for _, issue := range issues {
		if issueStats[issue.IssueType] == nil {
			issueStats[issue.IssueType] = map[string]interface{}{
				"count":      0,
				"high":       0,
				"medium":     0,
				"low":        0,
				"totalCount": 0,
			}
		}
		stats := issueStats[issue.IssueType]
		stats["count"] = stats["count"].(int) + 1
		stats["totalCount"] = stats["totalCount"].(int) + issue.IssueCount
		if issue.IssueLevel == "high" {
			stats["high"] = stats["high"].(int) + 1
		} else if issue.IssueLevel == "medium" {
			stats["medium"] = stats["medium"].(int) + 1
		} else {
			stats["low"] = stats["low"].(int) + 1
		}
	}

	// 获取最近的历史记录用于趋势分析
	var recentHistories []model.DataQualityHistory
	database.DB.Where("database_name = ?", assessment.DatabaseName).
		Order("assessment_date DESC").
		Limit(10).
		Find(&recentHistories)

	// 构建分析数据
	analysisData := map[string]interface{}{
		"assessment": map[string]interface{}{
			"id":                 assessment.Id,
			"database_name":      assessment.DatabaseName,
			"assessment_time":    assessment.AssessmentTime.Format("2006-01-02 15:04:05"),
			"total_tables":       assessment.TotalTables,
			"total_columns":      assessment.TotalColumns,
			"total_issues":       assessment.TotalIssues,
			"overall_score":      assessment.OverallScore,
			"overall_level":      assessment.OverallLevel,
			"field_completeness": assessment.FieldCompleteness,
			"field_accuracy":     assessment.FieldAccuracy,
			"table_completeness": assessment.TableCompleteness,
			"data_consistency":   assessment.DataConsistency,
			"data_uniqueness":    assessment.DataUniqueness,
			"data_timeliness":    assessment.DataTimeliness,
		},
		"issue_statistics": issueStats,
		"recent_histories": recentHistories,
		"top_issues":       getTopIssues(issues, 20), // 取前20个最重要的问题
	}

	return analysisData, nil
}

// performAiAnalysis 执行AI分析
func performAiAnalysis(aiModel model.AIModel, assessment model.DataQualityAssessment, analysisData map[string]interface{}, taskLogger *TaskLogger) (*AiAnalysisResult, error) {
	logger := log.Logger

	// 创建AI客户端
	client, err := service.NewAIClient(&aiModel)
	if err != nil {
		return nil, fmt.Errorf("创建AI客户端失败: %v", err)
	}

	// 构建分析提示词
	prompt := buildDataQualityAnalysisPrompt(assessment, analysisData)

	// 构造消息
	messages := []service.Message{
		{
			Role:    "system",
			Content: "你是一个专业的数据质量分析专家，擅长分析数据库质量评估数据，识别质量问题，提供优化建议。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// 调用AI模型
	taskLogger.UpdateResult("正在调用AI模型进行分析...")
	response, err := client.Chat(messages, &service.ChatOptions{
		Temperature: aiModel.Temperature,
		MaxTokens:   aiModel.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("调用AI模型失败: %v", err)
	}

	logger.Info("AI分析完成", zap.String("model", aiModel.Name))
	taskLogger.UpdateResult("AI分析完成，正在解析结果...")

	// 解析AI返回的结果
	aiResult, err := parseAiAnalysisResponse(response.Content, assessment)
	if err != nil {
		logger.Warn("解析AI响应失败，使用默认结果", zap.Error(err))
		// 如果解析失败，使用默认结果
		aiResult = createDefaultAnalysisResult(assessment, response.Content)
	}

	return aiResult, nil
}

// buildDataQualityAnalysisPrompt 构建数据质量分析提示词
func buildDataQualityAnalysisPrompt(assessment model.DataQualityAssessment, analysisData map[string]interface{}) string {
	// 将分析数据转换为JSON
	dataJson, _ := json.MarshalIndent(analysisData, "", "  ")

	prompt := fmt.Sprintf(`请对以下数据质量评估结果进行深度分析，并提供专业的分析报告。

## 评估概览
- 数据库名称: %s
- 评估时间: %s
- 总体评分: %.2f/100 (%s)
- 评估表数: %d
- 评估字段数: %d
- 发现问题数: %d

## 质量指标
- 字段完整性: %.2f%%
- 字段准确性: %.2f%%
- 表完整性: %.2f%%
- 数据一致性: %.2f%%
- 数据唯一性: %.2f%%
- 数据及时性: %.2f%%

## 详细数据
%s

## 分析要求
请基于以上数据，提供以下分析内容（请以JSON格式返回）：

1. **总体评分和等级** (ai_score, ai_level)
   - 基于各项指标综合评估，给出0-100的评分
   - 等级：优秀(>=90)、良好(80-89)、一般(70-79)、较差(60-69)、差(<60)

2. **趋势分析** (trend_analysis, trend_direction, trend_percentage)
   - 如果有历史数据，分析质量趋势（上升/下降/稳定）
   - 计算变化百分比

3. **智能洞察** (insights)
   - 识别关键问题和风险点
   - 分析问题根源
   - 每个洞察包含：insight_content, insight_type, priority (1-高, 2-中, 3-低)

4. **优化建议** (recommendations)
   - 针对发现的问题提供具体可行的优化建议
   - 每个建议包含：title, description, recommendation_type (high/medium/low), priority, expected_improvement

请返回JSON格式，结构如下：
{
  "ai_score": 85.5,
  "ai_level": "良好",
  "trend_analysis": "数据质量较上次评估有所提升...",
  "trend_direction": "上升",
  "trend_percentage": 5.2,
  "insights": [
    {
      "insight_content": "完整性指标较低，主要原因是...",
      "insight_type": "完整性",
      "priority": 1
    }
  ],
  "recommendations": [
    {
      "title": "优化空值处理",
      "description": "建议对空值率超过20%%的字段进行数据补全...",
      "recommendation_type": "high",
      "priority": "high",
      "expected_improvement": 10.5
    }
  ]
}`,
		assessment.DatabaseName,
		assessment.AssessmentTime.Format("2006-01-02 15:04:05"),
		assessment.OverallScore,
		assessment.OverallLevel,
		assessment.TotalTables,
		assessment.TotalColumns,
		assessment.TotalIssues,
		assessment.FieldCompleteness,
		assessment.FieldAccuracy,
		assessment.TableCompleteness,
		assessment.DataConsistency,
		assessment.DataUniqueness,
		assessment.DataTimeliness,
		string(dataJson))

	return prompt
}

// AiAnalysisResult AI分析结果结构
type AiAnalysisResult struct {
	AiScore         float64
	AiLevel         string
	TrendAnalysis   string
	TrendDirection  string
	TrendPercentage float64
	Insights        []AiInsight
	Recommendations []AiRecommendation
}

// AiInsight AI洞察
type AiInsight struct {
	InsightContent string
	InsightType    string
	Priority       int
}

// AiRecommendation AI建议
type AiRecommendation struct {
	Title               string
	Description         string
	RecommendationType  string
	Priority            string
	ExpectedImprovement float64
}

// parseAiAnalysisResponse 解析AI响应
func parseAiAnalysisResponse(responseContent string, assessment model.DataQualityAssessment) (*AiAnalysisResult, error) {
	// 尝试提取JSON部分（AI可能返回Markdown格式的JSON）
	jsonStart := strings.Index(responseContent, "{")
	jsonEnd := strings.LastIndex(responseContent, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("未找到有效的JSON内容")
	}

	jsonContent := responseContent[jsonStart : jsonEnd+1]

	var result struct {
		AiScore         float64 `json:"ai_score"`
		AiLevel         string  `json:"ai_level"`
		TrendAnalysis   string  `json:"trend_analysis"`
		TrendDirection  string  `json:"trend_direction"`
		TrendPercentage float64 `json:"trend_percentage"`
		Insights        []struct {
			InsightContent string `json:"insight_content"`
			InsightType    string `json:"insight_type"`
			Priority       int    `json:"priority"`
		} `json:"insights"`
		Recommendations []struct {
			Title               string  `json:"title"`
			Description         string  `json:"description"`
			RecommendationType  string  `json:"recommendation_type"`
			Priority            string  `json:"priority"`
			ExpectedImprovement float64 `json:"expected_improvement"`
		} `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 转换为内部结构
	insights := make([]AiInsight, len(result.Insights))
	for i, insight := range result.Insights {
		insights[i] = AiInsight{
			InsightContent: insight.InsightContent,
			InsightType:    insight.InsightType,
			Priority:       insight.Priority,
		}
	}

	recommendations := make([]AiRecommendation, len(result.Recommendations))
	for i, rec := range result.Recommendations {
		recommendations[i] = AiRecommendation{
			Title:               rec.Title,
			Description:         rec.Description,
			RecommendationType:  rec.RecommendationType,
			Priority:            rec.Priority,
			ExpectedImprovement: rec.ExpectedImprovement,
		}
	}

	return &AiAnalysisResult{
		AiScore:         result.AiScore,
		AiLevel:         result.AiLevel,
		TrendAnalysis:   result.TrendAnalysis,
		TrendDirection:  result.TrendDirection,
		TrendPercentage: result.TrendPercentage,
		Insights:        insights,
		Recommendations: recommendations,
	}, nil
}

// createDefaultAnalysisResult 创建默认分析结果（当AI解析失败时使用）
func createDefaultAnalysisResult(assessment model.DataQualityAssessment, rawResponse string) *AiAnalysisResult {
	// 基于评估数据生成默认结果
	aiScore := assessment.OverallScore
	aiLevel := assessment.OverallLevel

	// 生成默认洞察
	insights := []AiInsight{
		{
			InsightContent: fmt.Sprintf("数据质量总体评分为%.2f，处于%s水平", aiScore, aiLevel),
			InsightType:    "总体评估",
			Priority:       2,
		},
	}

	if assessment.TotalIssues > 0 {
		insights = append(insights, AiInsight{
			InsightContent: fmt.Sprintf("发现%d个质量问题，需要重点关注", assessment.TotalIssues),
			InsightType:    "问题统计",
			Priority:       1,
		})
	}

	// 生成默认建议
	recommendations := []AiRecommendation{}
	if assessment.FieldCompleteness < 80 {
		recommendations = append(recommendations, AiRecommendation{
			Title:               "提升数据完整性",
			Description:         "字段完整性较低，建议检查并补全缺失数据",
			RecommendationType:  "high",
			Priority:            "high",
			ExpectedImprovement: 5.0,
		})
	}

	return &AiAnalysisResult{
		AiScore:         aiScore,
		AiLevel:         aiLevel,
		TrendAnalysis:   "暂无历史数据，无法进行趋势分析",
		TrendDirection:  "未知",
		TrendPercentage: 0,
		Insights:        insights,
		Recommendations: recommendations,
	}
}

// saveAiAnalysisResult 保存AI分析结果
func saveAiAnalysisResult(assessmentId int64, aiModel model.AIModel, result *AiAnalysisResult) error {
	// 保存AI分析主记录
	aiAnalysis := model.DataQualityAiAnalysis{
		AssessmentId:    assessmentId,
		AiScore:         result.AiScore,
		AiLevel:         result.AiLevel,
		TrendAnalysis:   result.TrendAnalysis,
		TrendDirection:  result.TrendDirection,
		TrendPercentage: result.TrendPercentage,
		AnalysisTime:    time.Now(),
		ModelVersion:    fmt.Sprintf("%s-%s", aiModel.Name, aiModel.Provider),
	}

	if err := database.DB.Create(&aiAnalysis).Error; err != nil {
		return fmt.Errorf("保存AI分析记录失败: %v", err)
	}

	// 保存AI洞察
	for _, insight := range result.Insights {
		aiInsight := model.DataQualityAiInsight{
			AiAnalysisId:   aiAnalysis.Id,
			InsightContent: insight.InsightContent,
			InsightType:    insight.InsightType,
			Priority:       insight.Priority,
		}
		if err := database.DB.Create(&aiInsight).Error; err != nil {
			log.Logger.Warn("保存AI洞察失败", zap.Error(err))
		}
	}

	// 保存AI建议
	for _, rec := range result.Recommendations {
		aiRec := model.DataQualityAiRecommendation{
			AiAnalysisId:        aiAnalysis.Id,
			RecommendationType:  rec.RecommendationType,
			Title:               rec.Title,
			Description:         rec.Description,
			Priority:            rec.Priority,
			ExpectedImprovement: rec.ExpectedImprovement,
			Status:              0, // 未采纳
		}
		if err := database.DB.Create(&aiRec).Error; err != nil {
			log.Logger.Warn("保存AI建议失败", zap.Error(err))
		}
	}

	return nil
}

// getTopIssues 获取前N个最重要的问题
func getTopIssues(issues []model.DataQualityIssue, limit int) []model.DataQualityIssue {
	if len(issues) == 0 {
		return []model.DataQualityIssue{}
	}
	if len(issues) <= limit {
		return issues
	}
	return issues[:limit]
}

// ExecuteDataQualityAiAnalysis 导出函数，用于手动执行任务
func ExecuteDataQualityAiAnalysis() {
	doDataQualityAiAnalysis()
}
