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
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	taskRunner "dbmcloud/src/task"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func OptionList(c *gin.Context) {
	var db = database.DB
	method := c.Request.Method
	if method == "GET" {
		if c.Query("enable") != "" {
			db = db.Where("enable=?", c.Query("enable"))
		}
		if c.Query("task_key") != "" {
			db = db.Where("task_key like ? ", "%"+c.Query("task_key")+"%")
		}
		if c.Query("task_name") != "" {
			db = db.Where("task_name like ? ", "%"+c.Query("task_name")+"%")
		}
		sorterMap := make(map[string]string)
		sorterData := c.Query("sorter")
		json.Unmarshal([]byte(sorterData), &sorterMap)
		for sortField, sortOrder := range sorterMap {
			if sortField != "" && sortOrder != "" {
				db = db.Order(fmt.Sprintf("%s %s", sortField, strings.Replace(sortOrder, "end", "", 1)))
			}
		}
		var dataList []model.TaskOption
		result := db.Find(&dataList)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "OK",
			"data":    dataList,
			"total":   len(dataList),
		})
		return

	}
	if method == "POST" {
		var record model.TaskOption
		c.BindJSON(&record)
		result := db.Create(&record)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Insert Error: " + result.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
		return

	}

	if method == "PUT" {
		var record model.TaskOption
		c.BindJSON(&record)
		//gin里面如果更新为0则字段不会更新，可以使用select指定更新字段解决
		result := db.Model(&record).Select("task_name", "task_description", "crontab", "enable").Omit("task_key").Where("task_key = ?", record.TaskKey).Updates(record)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Update Error: " + result.Error.Error()})
			return
		}

		c.JSON(200, gin.H{"success": true})
		return
	}

	if method == "DELETE" {
		var record model.TaskOption
		c.BindJSON(&record)
		result := db.Model(&model.TaskOption{}).Where("task_key = ?", record.TaskKey).Delete(record)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Delete Error:" + result.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
		return
	}
}

// ExecuteTask 手动执行任务
func ExecuteTask(c *gin.Context) {
	var req struct {
		TaskKey string `json:"task_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"msg":     "参数解析失败: " + err.Error(),
		})
		return
	}

	var db = database.DB
	var taskOption model.TaskOption
	result := db.Where("task_key = ?", req.TaskKey).First(&taskOption)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"msg":     "任务不存在",
		})
		return
	}

	// 手工执行允许执行未启用的任务（用于测试），但给出提示
	if taskOption.Enable != 1 {
		// 允许执行，但记录日志提示任务未启用
		// 继续执行任务
	}

	// 根据 task_key 调用对应的任务函数
	go func() {
		// 更新心跳时间
		db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", req.TaskKey).Updates(map[string]interface{}{
			"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999"),
		})

		// 调用对应的任务函数
		switch req.TaskKey {
		case "gather_pumpkin":
			taskRunner.ExecutePumpkinTask()
		case "gather_pumpkin_growth":
			taskRunner.ExecutePumpkinGrowthTask()
		case "gather_dbmeta":
			taskRunner.ExecuteDbMetaTask()
		case "check_datasource":
			taskRunner.ExecuteDatasourceCheck()
		case "gather_sensitive":
			// 如果有 gather_sensitive 任务，在这里调用
			// taskRunner.ExecuteSensitiveTask()
		case "data_quality_ai_analysis":
			taskRunner.ExecuteDataQualityAiAnalysis()
		case "ai_grading_batch":
			taskRunner.ExecuteAiGradingBatchTask()
		case "ai_general_table_comment":
			taskRunner.ExecuteAiGeneralTableCommentTask()
		case "ai_general_column_comment":
			taskRunner.ExecuteAiGeneralColumnCommentTask()
		case "ai_apply_table_comment":
			taskRunner.ExecuteAiApplyTableCommentTask()
		case "ai_apply_column_comment":
			taskRunner.ExecuteAiApplyColumnCommentTask()
		case "ai_table_comment_accuracy":
			taskRunner.ExecuteAiTableCommentAccuracyTask()
		case "ai_column_comment_accuracy":
			taskRunner.ExecuteAiColumnCommentAccuracyTask()
		default:
			// 其他任务可以在这里添加
		}

		// 更新心跳结束时间
		db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", req.TaskKey).Updates(map[string]interface{}{
			"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999"),
		})
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "任务已开始执行",
	})
}
