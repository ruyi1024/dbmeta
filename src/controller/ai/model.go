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
	"dbmeta-core/log"
	"dbmeta-core/src/model"
	"dbmeta-core/src/service"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetAIModelDefaults 获取各场景默认模型 ID
func GetAIModelDefaults(c *gin.Context) {
	m, err := service.GetAIModelDefaults()
	if err != nil {
		log.Error("获取默认模型配置失败", zap.Error(err))
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	gradingID := m[model.AIModelScenarioGrading]
	tableColAccuracyID := m[model.AIModelScenarioTableColumnAccuracy]
	tableColCommentID := m[model.AIModelScenarioTableColumnComment]
	sqlGenerationID := m[model.AIModelScenarioSQLGeneration]
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"grading_model_id":               gradingID,
			"table_column_accuracy_model_id": tableColAccuracyID,
			"table_column_comment_model_id":  tableColCommentID,
			"sql_generation_model_id":        sqlGenerationID,
		},
	})
}

type aiModelDefaultsUpdateReq struct {
	GradingModelId             *int `json:"grading_model_id"`
	TableColumnAccuracyModelId *int `json:"table_column_accuracy_model_id"`
	TableColumnCommentModelId  *int `json:"table_column_comment_model_id"`
	SQLGenerationModelId       *int `json:"sql_generation_model_id"`
}

// UpdateAIModelDefaults 更新各场景默认模型
func UpdateAIModelDefaults(c *gin.Context) {
	var req aiModelDefaultsUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": "请求参数错误", "error": err.Error()})
		return
	}
	if err := service.SetAIModelDefaultForScenario(model.AIModelScenarioGrading, req.GradingModelId); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := service.SetAIModelDefaultForScenario(model.AIModelScenarioTableColumnAccuracy, req.TableColumnAccuracyModelId); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := service.SetAIModelDefaultForScenario(model.AIModelScenarioTableColumnComment, req.TableColumnCommentModelId); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := service.SetAIModelDefaultForScenario(model.AIModelScenarioSQLGeneration, req.SQLGenerationModelId); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "message": "保存成功"})
}

// GetModels 获取模型列表
func GetModels(c *gin.Context) {
	models, err := service.GetAllModels()
	if err != nil {
		log.Error("获取模型列表失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "获取模型列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    models,
	})
}

// GetEnabledModels 获取启用的模型列表
func GetEnabledModels(c *gin.Context) {
	models, err := service.GetEnabledModels()
	if err != nil {
		log.Error("获取启用的模型列表失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "获取启用的模型列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    models,
	})
}

// CreateModel 创建模型配置
func CreateModel(c *gin.Context) {
	var aiModel model.AIModel
	if err := c.ShouldBindJSON(&aiModel); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证必填字段
	if aiModel.Name == "" || aiModel.Provider == "" || aiModel.ApiUrl == "" || aiModel.ModelName == "" {
		c.JSON(400, gin.H{
			"success": false,
			"message": "名称、提供商、API地址和模型名称不能为空",
		})
		return
	}

	err := service.CreateModel(&aiModel)
	if err != nil {
		log.Error("创建模型失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "创建模型失败",
			"error":   err.Error(),
		})
		return
	}

	// 刷新缓存
	service.RefreshModelsCache()

	c.JSON(200, gin.H{
		"success": true,
		"data":    aiModel,
		"message": "创建成功",
	})
}

// UpdateModel 更新模型配置
func UpdateModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "无效的模型ID",
			"error":   err.Error(),
		})
		return
	}

	var aiModel model.AIModel
	if err := c.ShouldBindJSON(&aiModel); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err = service.UpdateModel(id, &aiModel)
	if err != nil {
		log.Error("更新模型失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "更新模型失败",
			"error":   err.Error(),
		})
		return
	}

	// 刷新缓存
	service.RefreshModelsCache()

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新成功",
	})
}

// DeleteModel 删除模型配置
func DeleteModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	err = service.DeleteModel(id)
	if err != nil {
		log.Error("删除模型失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "删除模型失败",
			"error":   err.Error(),
		})
		return
	}

	// 刷新缓存
	service.RefreshModelsCache()

	c.JSON(200, gin.H{
		"success": true,
		"message": "删除成功",
	})
}

// ToggleModel 启用/禁用模型
func ToggleModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	var req struct {
		Enabled int8 `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	if req.Enabled != 0 && req.Enabled != 1 {
		c.JSON(400, gin.H{
			"success": false,
			"message": "enabled参数必须为0或1",
		})
		return
	}

	err = service.ToggleModel(id, req.Enabled)
	if err != nil {
		log.Error("更新模型状态失败", zap.Error(err))
		c.JSON(500, gin.H{
			"success": false,
			"message": "更新模型状态失败",
			"error":   err.Error(),
		})
		return
	}

	// 刷新缓存
	service.RefreshModelsCache()

	c.JSON(200, gin.H{
		"success": true,
		"message": "更新成功",
	})
}

// TestModel 测试模型连接
func TestModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "无效的模型ID",
		})
		return
	}

	aiModel, err := service.GetModelById(id)
	if err != nil {
		c.JSON(404, gin.H{
			"success": false,
			"message": "模型不存在",
			"error":   err.Error(),
		})
		return
	}

	// 验证配置
	if err := service.TestModelConnection(aiModel); err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "配置验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 尝试创建客户端并测试连接
	client, err := service.NewAIClient(aiModel)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "创建客户端失败",
			"error":   err.Error(),
		})
		return
	}

	// 测试连接
	if err := client.TestConnection(); err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "连接测试失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "连接测试成功",
	})
}

// TestModelConfig 测试模型配置（不需要id，用于创建前测试）
func TestModelConfig(c *gin.Context) {
	var aiModel model.AIModel
	if err := c.ShouldBindJSON(&aiModel); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证配置
	if err := service.TestModelConnection(&aiModel); err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "配置验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 尝试创建客户端并测试连接
	// 注意：在测试配置时，API Key 可能是未加密的，GetDecryptedApiKey 会自动处理
	client, err := service.NewAIClient(&aiModel)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "创建客户端失败",
			"error":   err.Error(),
		})
		return
	}

	// 测试连接
	if err := client.TestConnection(); err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "连接测试失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "连接测试成功",
	})
}
