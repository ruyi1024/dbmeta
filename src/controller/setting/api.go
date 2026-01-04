package setting

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ApiConfigList API配置接口 - 支持增删改查
func ApiConfigList(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		// 查询API配置列表
		var dataList []model.ApiConfig
		var total int64

		// 获取查询参数
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
		apiName := c.Query("api_name")
		protocol := c.Query("protocol")
		enable := c.Query("enable")

		// 构建查询条件
		db := database.DB.Model(&model.ApiConfig{})
		if apiName != "" {
			db = db.Where("api_name LIKE ?", "%"+apiName+"%")
		}
		if protocol != "" {
			db = db.Where("protocol = ?", protocol)
		}
		if enable != "" {
			db = db.Where("enable = ?", enable)
		}

		// 获取总数
		db.Count(&total)

		// 分页查询
		offset := (page - 1) * pageSize
		result := db.Order("gmt_created DESC").Offset(offset).Limit(pageSize).Find(&dataList)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "查询API配置失败: " + result.Error.Error(),
			})
			return
		}

		// 返回结果
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"list":     dataList,
				"total":    total,
				"page":     page,
				"pageSize": pageSize,
			},
			"message": "查询成功",
		})

	case "POST":
		// 创建API配置
		var apiConfig model.ApiConfig
		if err := c.ShouldBindJSON(&apiConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "参数错误: " + err.Error(),
			})
			return
		}

		// 验证必填字段
		if apiConfig.ApiName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "API名称不能为空",
			})
			return
		}
		if apiConfig.ApiUrl == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "API URL不能为空",
			})
			return
		}

		// 设置默认值
		if apiConfig.Protocol == "" {
			apiConfig.Protocol = "HTTP"
		}
		if apiConfig.Method == "" {
			apiConfig.Method = "GET"
		}
		if apiConfig.AuthType == "" {
			apiConfig.AuthType = "NONE"
		}
		if apiConfig.ExpectedCodes == "" {
			apiConfig.ExpectedCodes = "200"
		}
		if apiConfig.Timeout == 0 {
			apiConfig.Timeout = 30
		}

		// 创建记录
		result := database.DB.Create(&apiConfig)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "创建API配置失败: " + result.Error.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    apiConfig,
			"message": "创建成功",
		})

	case "PUT":
		// 更新API配置
		var apiConfig model.ApiConfig
		if err := c.ShouldBindJSON(&apiConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "参数错误: " + err.Error(),
			})
			return
		}

		// 验证ID
		if apiConfig.Id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "ID不能为空",
			})
			return
		}

		// 验证必填字段
		if apiConfig.ApiName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "API名称不能为空",
			})
			return
		}
		if apiConfig.ApiUrl == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "API URL不能为空",
			})
			return
		}

		// 更新记录
		result := database.DB.Model(&model.ApiConfig{}).Where("id = ?", apiConfig.Id).Updates(&apiConfig)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "更新API配置失败: " + result.Error.Error(),
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "API配置不存在",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "更新成功",
		})

	case "DELETE":
		// 删除API配置
		var requestData struct {
			Id int64 `json:"id"`
		}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "参数错误: " + err.Error(),
			})
			return
		}

		if requestData.Id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "ID不能为空",
			})
			return
		}

		// 删除记录
		result := database.DB.Delete(&model.ApiConfig{}, requestData.Id)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "删除API配置失败: " + result.Error.Error(),
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "API配置不存在",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "删除成功",
		})

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"message": "不支持的请求方法",
		})
	}
}
