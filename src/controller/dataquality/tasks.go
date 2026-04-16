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
	qualityTask "dbmeta-core/src/task"
	"dbmeta-core/src/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTasks 获取评估任务列表
func GetTasks(c *gin.Context) {
	var tasks []model.DataQualityTask

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

	taskType := c.Query("taskType")
	status := c.Query("status")

	query := database.DB.Model(&model.DataQualityTask{})

	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	offset := (pageNum - 1) * pageSize
	query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks)

	taskList := make([]map[string]interface{}, 0)
	for _, task := range tasks {
		taskList = append(taskList, map[string]interface{}{
			"id":             task.Id,
			"taskName":       task.TaskName,
			"taskType":       task.TaskType,
			"datasourceId":   task.DatasourceId,
			"databaseName":   task.DatabaseName,
			"tableFilter":    task.TableFilter,
			"scheduleConfig": task.ScheduleConfig,
			"status":         task.Status,
			"startTime":      task.StartTime,
			"endTime":        task.EndTime,
			"duration":       task.Duration,
			"resultSummary":  task.ResultSummary,
			"errorMessage":   task.ErrorMessage,
			"createdBy":      task.CreatedBy,
			"createdAt":      task.CreatedAt,
			"updatedAt":      task.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"list":     taskList,
			"total":    total,
			"page":     pageNum,
			"pageSize": pageSize,
		},
	})
}

// CreateTask 创建评估任务
func CreateTask(c *gin.Context) {
	var req struct {
		TaskName       string `json:"taskName" binding:"required"`
		TaskType       string `json:"taskType" binding:"required"`
		DatasourceId   *int64 `json:"datasourceId"`
		DatabaseName   string `json:"databaseName" binding:"required"`
		TableFilter    string `json:"tableFilter"`
		ScheduleConfig string `json:"scheduleConfig"`
		CreatedBy      string `json:"createdBy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	// 如果提供了datasourceId，验证数据库是否存在
	if req.DatasourceId != nil {
		var dbInfo model.MetaDatabase
		if err := database.DB.Where("id = ? AND is_deleted = 0", *req.DatasourceId).First(&dbInfo).Error; err != nil {
			// 如果通过ID找不到，尝试通过数据库名称查找
			if err := database.DB.Where("database_name = ? AND is_deleted = 0", req.DatabaseName).First(&dbInfo).Error; err == nil {
				req.DatasourceId = &dbInfo.Id
			}
		}
	} else {
		// 如果没有提供datasourceId，通过数据库名称查找
		var dbInfo model.MetaDatabase
		if err := database.DB.Where("database_name = ? AND is_deleted = 0", req.DatabaseName).First(&dbInfo).Error; err == nil {
			req.DatasourceId = &dbInfo.Id
		}
	}

	task := model.DataQualityTask{
		TaskName:       req.TaskName,
		TaskType:       req.TaskType,
		DatasourceId:   req.DatasourceId,
		DatabaseName:   req.DatabaseName,
		TableFilter:    req.TableFilter,
		ScheduleConfig: req.ScheduleConfig,
		Status:         "pending",
		CreatedBy:      req.CreatedBy,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "创建成功",
		"data": task,
	})
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(c *gin.Context) {
	var req struct {
		Id     int64  `json:"id" binding:"required"`
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	var task model.DataQualityTask
	if err := database.DB.First(&task, req.Id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "任务不存在",
		})
		return
	}

	// 如果状态是running，启动异步执行
	if req.Status == "running" && task.Status != "running" {
		// 异步执行任务
		qualityTask.ExecuteDataQualityTask(req.Id)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "任务已开始执行",
		})
		return
	}

	now := time.Now()
	if req.Status == "running" && task.StartTime == nil {
		task.StartTime = &now
	} else if (req.Status == "success" || req.Status == "failed") && task.EndTime == nil {
		task.EndTime = &now
		if task.StartTime != nil {
			task.Duration = int(now.Sub(*task.StartTime).Seconds())
		}
	}

	task.Status = req.Status

	if err := database.DB.Save(&task).Error; err != nil {
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

// DeleteTask 删除任务
func DeleteTask(c *gin.Context) {
	id := utils.StrToInt64(c.Param("id"))

	if err := database.DB.Delete(&model.DataQualityTask{}, id).Error; err != nil {
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

// ExecuteTask 手动执行评估任务
func ExecuteTask(c *gin.Context) {
	var req struct {
		Id int64 `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数解析失败: " + err.Error(),
		})
		return
	}

	// 查询任务
	var task model.DataQualityTask
	if err := database.DB.First(&task, req.Id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "任务不存在",
		})
		return
	}

	// 检查任务状态
	if task.Status == "running" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "任务正在执行中，请勿重复执行",
		})
		return
	}

	// 异步执行任务
	qualityTask.ExecuteDataQualityTask(req.Id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "任务已开始执行",
	})
}
