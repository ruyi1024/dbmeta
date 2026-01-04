package task

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// TaskLogList 获取任务日志列表
func TaskLogList(c *gin.Context) {
	method := c.Request.Method
	if method == "GET" {
		var db = database.DB

		// 查询条件
		if c.Query("task_key") != "" {
			db = db.Where("task_key = ?", c.Query("task_key"))
		}
		if c.Query("status") != "" {
			db = db.Where("status = ?", c.Query("status"))
		}

		// 日期范围查询
		if c.Query("start_date") != "" {
			startDate, err := time.Parse("2006-01-02", c.Query("start_date"))
			if err == nil {
				db = db.Where("gmt_created >= ?", startDate)
			}
		}
		if c.Query("end_date") != "" {
			endDate, err := time.Parse("2006-01-02", c.Query("end_date"))
			if err == nil {
				// 结束日期加一天，包含当天
				endDate = endDate.Add(24 * time.Hour)
				db = db.Where("gmt_created < ?", endDate)
			}
		}

		// 排序
		sorterMap := make(map[string]string)
		sorterData := c.Query("sorter")
		if sorterData != "" {
			json.Unmarshal([]byte(sorterData), &sorterMap)
		}

		// 默认按创建时间倒序
		if len(sorterMap) == 0 {
			db = db.Order("gmt_created DESC")
		} else {
			for sortField, sortOrder := range sorterMap {
				if sortField != "" && sortOrder != "" {
					order := "ASC"
					if sortOrder == "descend" {
						order = "DESC"
					}
					db = db.Order(fmt.Sprintf("%s %s", sortField, order))
				}
			}
		}

		// 分页
		pageSize := 10
		currentPage := 1
		if c.Query("pageSize") != "" {
			if size, err := strconv.Atoi(c.Query("pageSize")); err == nil && size > 0 {
				pageSize = size
			}
		}
		if c.Query("currentPage") != "" {
			if page, err := strconv.Atoi(c.Query("currentPage")); err == nil && page > 0 {
				currentPage = page
			}
		}

		offset := (currentPage - 1) * pageSize

		var dataList []model.TaskLog
		var total int64

		// 获取总数
		db.Model(&model.TaskLog{}).Count(&total)

		// 获取分页数据
		result := db.Offset(offset).Limit(pageSize).Find(&dataList)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"msg":         "OK",
			"data":        dataList,
			"total":       total,
			"pageSize":    pageSize,
			"currentPage": currentPage,
		})
		return
	}
}

// TaskLogDetail 获取任务日志详情
func TaskLogDetail(c *gin.Context) {
	method := c.Request.Method
	if method == "GET" {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "缺少日志ID"})
			return
		}

		var taskLog model.TaskLog
		result := database.DB.Where("id = ?", id).First(&taskLog)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "日志不存在: " + result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "OK",
			"data":    taskLog,
		})
		return
	}
}

// TaskLogStats 获取任务执行统计
func TaskLogStats(c *gin.Context) {
	method := c.Request.Method
	if method == "GET" {
		taskKey := c.Query("task_key")
		if taskKey == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "缺少任务键"})
			return
		}

		// 获取最近7天的统计
		now := time.Now()
		sevenDaysAgo := now.AddDate(0, 0, -7)

		var stats []struct {
			Date    string `json:"date"`
			Total   int64  `json:"total"`
			Success int64  `json:"success"`
			Failed  int64  `json:"failed"`
		}

		// 按日期统计
		rows, err := database.DB.Raw(`
			SELECT 
				DATE(gmt_created) as date,
				COUNT(*) as total,
				SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
			FROM task_log 
			WHERE task_key = ? AND gmt_created >= ?
			GROUP BY DATE(gmt_created)
			ORDER BY date DESC
		`, taskKey, sevenDaysAgo).Rows()

		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询统计失败: " + err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var stat struct {
				Date    string `json:"date"`
				Total   int64  `json:"total"`
				Success int64  `json:"success"`
				Failed  int64  `json:"failed"`
			}
			rows.Scan(&stat.Date, &stat.Total, &stat.Success, &stat.Failed)
			stats = append(stats, stat)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "OK",
			"data":    stats,
		})
		return
	}
}
