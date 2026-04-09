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

package dataquality

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DashboardInfo 获取数据质量大盘信息
func DashboardInfo(c *gin.Context) {
	var d = make(map[string]interface{})

	// 获取最新的评估记录
	var latestAssessment model.DataQualityAssessment
	result := database.DB.Order("assessment_time DESC").First(&latestAssessment)
	if result.Error != nil || latestAssessment.Id == 0 {
		// 如果没有数据，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "success",
			"data": d,
		})
		return
	}

	// 获取统计数据
	var totalTables, totalColumns, totalIssues int64
	database.DB.Model(&model.DataQualityAssessment{}).
		Select("COALESCE(SUM(total_tables), 0)").
		Where("status = 1").
		Scan(&totalTables)
	database.DB.Model(&model.DataQualityAssessment{}).
		Select("COALESCE(SUM(total_columns), 0)").
		Where("status = 1").
		Scan(&totalColumns)
	database.DB.Model(&model.DataQualityIssue{}).
		Where("status IN (1, 2)").
		Count(&totalIssues)

	// 获取AI分析结果
	var aiAnalysis model.DataQualityAiAnalysis
	database.DB.Where("assessment_id = ?", latestAssessment.Id).First(&aiAnalysis)

	// 获取AI洞察
	var aiInsights []model.DataQualityAiInsight
	if aiAnalysis.Id > 0 {
		database.DB.Where("ai_analysis_id = ?", aiAnalysis.Id).
			Order("priority DESC").
			Find(&aiInsights)

		// 获取AI建议
		var aiRecommendations []model.DataQualityAiRecommendation
		database.DB.Where("ai_analysis_id = ?", aiAnalysis.Id).
			Order("CASE recommendation_type WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END").
			Find(&aiRecommendations)
	}

	// 获取分布数据
	var distributions []model.DataQualityDistribution
	database.DB.Where("assessment_id = ?", latestAssessment.Id).Find(&distributions)

	// 构建分布数据
	completenessData := make([]map[string]interface{}, 0)
	accuracyData := make([]map[string]interface{}, 0)
	consistencyData := make([]map[string]interface{}, 0)
	uniquenessData := make([]map[string]interface{}, 0)

	for _, dist := range distributions {
		item := map[string]interface{}{
			"type":  dist.Category,
			"value": dist.Value,
		}
		switch dist.DistributionType {
		case "completeness":
			completenessData = append(completenessData, item)
		case "accuracy":
			accuracyData = append(accuracyData, item)
		case "consistency":
			consistencyData = append(consistencyData, item)
		case "uniqueness":
			uniquenessData = append(uniquenessData, item)
		}
	}

	// 获取问题列表
	var issues []model.DataQualityIssue
	database.DB.Where("assessment_id = ?", latestAssessment.Id).
		Order("CASE issue_level WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END, issue_count DESC").
		Limit(100).
		Find(&issues)

	// 构建AI洞察列表
	insightsList := make([]string, 0)
	for _, insight := range aiInsights {
		insightsList = append(insightsList, insight.InsightContent)
	}

	// 获取AI建议
	var aiRecommendations []model.DataQualityAiRecommendation
	if aiAnalysis.Id > 0 {
		database.DB.Where("ai_analysis_id = ?", aiAnalysis.Id).
			Order("CASE recommendation_type WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 END").
			Find(&aiRecommendations)
	}

	// 构建AI建议列表
	recommendationsList := make([]map[string]interface{}, 0)
	for _, rec := range aiRecommendations {
		recommendationsList = append(recommendationsList, map[string]interface{}{
			"type":        rec.RecommendationType,
			"title":       rec.Title,
			"desc":        rec.Description,
			"priority":    rec.Priority,
			"improvement": rec.ExpectedImprovement,
		})
	}

	// 构建问题列表
	issueList := make([]map[string]interface{}, 0)
	for _, issue := range issues {
		issueList = append(issueList, map[string]interface{}{
			"key":           issue.Id,
			"tableName":     issue.TableNameX,
			"columnName":    issue.ColumnName,
			"issueType":     issue.IssueType,
			"issueLevel":    issue.IssueLevel,
			"issueDesc":     issue.IssueDesc,
			"issueCount":    issue.IssueCount,
			"lastCheckTime": issue.CheckTime.Format("2006-01-02 15:04:05"),
		})
	}

	d["totalTables"] = totalTables
	d["totalColumns"] = totalColumns
	d["totalIssues"] = totalIssues
	d["fieldCompleteness"] = latestAssessment.FieldCompleteness
	d["fieldAccuracy"] = latestAssessment.FieldAccuracy
	d["tableCompleteness"] = latestAssessment.TableCompleteness
	d["dataConsistency"] = latestAssessment.DataConsistency
	d["dataUniqueness"] = latestAssessment.DataUniqueness
	d["dataTimeliness"] = latestAssessment.DataTimeliness
	d["completenessData"] = completenessData
	d["accuracyData"] = accuracyData
	d["consistencyData"] = consistencyData
	d["uniquenessData"] = uniquenessData
	d["issueList"] = issueList

	// AI分析数据
	aiData := make(map[string]interface{})
	aiData["overallScore"] = aiAnalysis.AiScore
	aiData["overallLevel"] = aiAnalysis.AiLevel
	aiData["analysisTime"] = aiAnalysis.AnalysisTime.Format("2006-01-02 15:04:05")
	aiData["insights"] = insightsList
	aiData["recommendations"] = recommendationsList
	aiData["trendAnalysis"] = aiAnalysis.TrendAnalysis
	d["aiAnalysis"] = aiData

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": d,
	})
}

// GetIssues 获取质量问题列表
func GetIssues(c *gin.Context) {
	var issues []model.DataQualityIssue

	// ProTable使用current作为页码参数，同时兼容page参数
	current := c.Query("current")
	page := c.Query("page")
	if current != "" {
		page = current
	}
	if page == "" {
		page = "1"
	}
	pageNum := utils.StrToInt(page)

	// ProTable使用pageSize作为每页大小参数
	pageSizeStr := c.Query("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize := utils.StrToInt(pageSizeStr)

	issueType := c.Query("issueType")
	issueLevel := c.Query("issueLevel")
	status := c.Query("status")
	tableName := c.Query("tableName")
	columnName := c.Query("columnName")

	query := database.DB.Model(&model.DataQualityIssue{})

	if issueType != "" {
		query = query.Where("issue_type = ?", issueType)
	}
	if issueLevel != "" {
		query = query.Where("issue_level = ?", issueLevel)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if tableName != "" {
		query = query.Where("table_name LIKE ?", "%"+tableName+"%")
	}
	if columnName != "" {
		query = query.Where("column_name LIKE ?", "%"+columnName+"%")
	}

	var total int64
	query.Count(&total)

	offset := (pageNum - 1) * pageSize
	query.Order("FIELD(issue_level, 'high', 'medium', 'low'), issue_count DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&issues)

	issueList := make([]map[string]interface{}, 0)
	for _, issue := range issues {
		issueList = append(issueList, map[string]interface{}{
			"key":           issue.Id,
			"databaseName":  issue.DatabaseName,
			"tableName":     issue.TableNameX,
			"columnName":    issue.ColumnName,
			"issueType":     issue.IssueType,
			"issueLevel":    issue.IssueLevel,
			"issueDesc":     issue.IssueDesc,
			"issueCount":    issue.IssueCount,
			"status":        issue.Status,
			"handler":       issue.Handler,
			"handleTime":    issue.HandleTime,
			"handleRemark":  issue.HandleRemark,
			"lastCheckTime": issue.CheckTime.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"list":     issueList,
			"total":    total,
			"page":     pageNum,
			"pageSize": pageSize,
		},
	})
}

// UpdateIssueStatus 更新问题状态
func UpdateIssueStatus(c *gin.Context) {
	var req struct {
		Id           int64  `json:"id" binding:"required"`
		Status       int8   `json:"status" binding:"required"`
		Handler      string `json:"handler"`
		HandleRemark string `json:"handleRemark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	var issue model.DataQualityIssue
	if err := database.DB.First(&issue, req.Id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "问题不存在",
		})
		return
	}

	now := time.Now()
	issue.Status = req.Status
	if req.Handler != "" {
		issue.Handler = req.Handler
	}
	if req.HandleRemark != "" {
		issue.HandleRemark = req.HandleRemark
	}
	if req.Status == 3 { // 已处理
		issue.HandleTime = &now
	}

	if err := database.DB.Save(&issue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}
