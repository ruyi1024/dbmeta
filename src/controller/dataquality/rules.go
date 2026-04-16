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

package dataquality

import (
	"dbmeta-core/src/database"
	"dbmeta-core/src/model"
	"dbmeta-core/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRules 获取质量规则列表
func GetRules(c *gin.Context) {
	var rules []model.DataQualityRule

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

	ruleType := c.Query("ruleType")
	enabled := c.Query("enabled")

	query := database.DB.Model(&model.DataQualityRule{})

	if ruleType != "" {
		query = query.Where("rule_type = ?", ruleType)
	}
	if enabled != "" {
		query = query.Where("enabled = ?", enabled)
	}

	var total int64
	query.Count(&total)

	offset := (pageNum - 1) * pageSize
	query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rules)

	ruleList := make([]map[string]interface{}, 0)
	for _, rule := range rules {
		ruleList = append(ruleList, map[string]interface{}{
			"id":         rule.Id,
			"ruleName":   rule.RuleName,
			"ruleType":   rule.RuleType,
			"ruleDesc":   rule.RuleDesc,
			"ruleConfig": rule.RuleConfig,
			"threshold":  rule.Threshold,
			"severity":   rule.Severity,
			"enabled":    rule.Enabled,
			"createdBy":  rule.CreatedBy,
			"createdAt":  rule.CreatedAt,
			"updatedAt":  rule.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"list":     ruleList,
			"total":    total,
			"page":     pageNum,
			"pageSize": pageSize,
		},
	})
}

// CreateRule 创建质量规则
func CreateRule(c *gin.Context) {
	var rule model.DataQualityRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	if err := database.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": rule,
	})
}

// UpdateRule 更新质量规则
func UpdateRule(c *gin.Context) {
	var req struct {
		Id         int64   `json:"id" binding:"required"`
		RuleName   string  `json:"ruleName"`
		RuleType   string  `json:"ruleType"`
		RuleDesc   string  `json:"ruleDesc"`
		RuleConfig string  `json:"ruleConfig"`
		Threshold  float64 `json:"threshold"`
		Severity   string  `json:"severity"`
		Enabled    int8    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	var rule model.DataQualityRule
	if err := database.DB.First(&rule, req.Id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "规则不存在",
		})
		return
	}

	if req.RuleName != "" {
		rule.RuleName = req.RuleName
	}
	if req.RuleType != "" {
		rule.RuleType = req.RuleType
	}
	if req.RuleDesc != "" {
		rule.RuleDesc = req.RuleDesc
	}
	if req.RuleConfig != "" {
		rule.RuleConfig = req.RuleConfig
	}
	if req.Threshold > 0 {
		rule.Threshold = req.Threshold
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	rule.Enabled = req.Enabled

	if err := database.DB.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": rule,
	})
}

// DeleteRule 删除质量规则
func DeleteRule(c *gin.Context) {
	id := utils.StrToInt64(c.Param("id"))

	if err := database.DB.Delete(&model.DataQualityRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
