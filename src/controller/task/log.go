package task

import (
	"dbmeta-core/src/database"
	"dbmeta-core/src/model"
	"encoding/json"
	"fmt"
	"math"
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

// TaskTodayStats 获取今日任务执行统计（按小时）
func TaskTodayStats(c *gin.Context) {
	method := c.Request.Method
	if method == "GET" {
		var db = database.DB

		// 获取今天的开始时间
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		type HourStat struct {
			Bucket  string
			Total   int64
			Failed  int64
			Success int64
		}

		// 计算起止时间
		hourNow := now.Truncate(time.Hour)
		start24h := hourNow.Add(-23 * time.Hour) // 连续24个小时

		// 今日按小时统计
		rowsToday, err := db.Raw(`
			SELECT 
				DATE_FORMAT(gmt_created, '%H:00') as bucket,
				COUNT(*) as total,
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
				SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success
			FROM task_log 
			WHERE gmt_created >= ?
			GROUP BY DATE_FORMAT(gmt_created, '%H:00')
			ORDER BY bucket ASC
		`, todayStart).Rows()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询统计失败: " + err.Error()})
			return
		}
		defer rowsToday.Close()

		todayMap := make(map[string]HourStat)
		for rowsToday.Next() {
			var stat HourStat
			rowsToday.Scan(&stat.Bucket, &stat.Total, &stat.Failed, &stat.Success)
			todayMap[stat.Bucket] = stat
		}

		// 最近24小时（滚动）统计
		rows24h, err := db.Raw(`
			SELECT 
				DATE_FORMAT(gmt_created, '%m-%d %H:00') as bucket,
				COUNT(*) as total,
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed,
				SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success
			FROM task_log 
			WHERE gmt_created >= ?
			GROUP BY DATE_FORMAT(gmt_created, '%m-%d %H:00')
			ORDER BY bucket ASC
		`, start24h).Rows()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询24小时统计失败: " + err.Error()})
			return
		}
		defer rows24h.Close()

		h24Map := make(map[string]HourStat)
		for rows24h.Next() {
			var stat HourStat
			rows24h.Scan(&stat.Bucket, &stat.Total, &stat.Failed, &stat.Success)
			h24Map[stat.Bucket] = stat
		}

		// 今日总计、失败总计
		var todayTotal, todayFailedTotal, todaySuccessTotal int64
		for _, stat := range todayMap {
			todayTotal += stat.Total
			todayFailedTotal += stat.Failed
			todaySuccessTotal += stat.Success
		}

		// 最新执行时间（今日）
		var lastTaskLog model.TaskLog
		var lastExecuteTime string
		result := db.Where("gmt_created >= ?", todayStart).Order("gmt_created DESC").First(&lastTaskLog)
		if result.Error == nil {
			lastExecuteTime = lastTaskLog.CreatedAt.Format("2006-01-02 15:04:05")
		}

		// 构建今日趋势（0-23点）
		todayTrend := make([]map[string]interface{}, 0, 24)
		todayFailedTrend := make([]map[string]interface{}, 0, 24)
		successRateTrend := make([]map[string]interface{}, 0, 24)
		for i := 0; i < 24; i++ {
			bucket := fmt.Sprintf("%02d:00", i)
			stat := todayMap[bucket]
			todayTrend = append(todayTrend, map[string]interface{}{
				"x": bucket,
				"y": stat.Total,
			})
			todayFailedTrend = append(todayFailedTrend, map[string]interface{}{
				"x": bucket,
				"y": stat.Failed,
			})
			var rate float64
			if stat.Total > 0 {
				rate = math.Round(float64(stat.Success) / float64(stat.Total) * 100)
			}
			successRateTrend = append(successRateTrend, map[string]interface{}{
				"x": bucket,
				"y": rate,
			})
		}

		// 滚动24小时趋势
		h24Trend := make([]map[string]interface{}, 0, 24)
		slotTime := start24h
		for i := 0; i < 24; i++ {
			bucket := slotTime.Format("01-02 15:00")
			stat := h24Map[bucket]
			h24Trend = append(h24Trend, map[string]interface{}{
				"x": bucket,
				"y": stat.Total,
			})
			slotTime = slotTime.Add(time.Hour)
		}

		// 24小时总执行次数
		var h24Total int64
		for _, stat := range h24Map {
			h24Total += stat.Total
		}

		// 成功率
		var successRate float64
		if todayTotal > 0 {
			successRate = math.Round(float64(todaySuccessTotal) / float64(todayTotal) * 100)
		}
		successRateStr := fmt.Sprintf("%.0f%%", successRate)

		c.JSON(http.StatusOK, gin.H{
			"success":          true,
			"msg":              "OK",
			"todayTotal":       todayTotal,
			"todayFailedTotal": todayFailedTotal,
			"todayTrend":       todayTrend,
			"todayFailedTrend": todayFailedTrend,
			"hour24Total":      h24Total,
			"hour24Trend":      h24Trend,
			"successRate":      successRate,
			"successRateStr":   successRateStr,
			"successRateTrend": successRateTrend,
			"lastExecuteTime":  lastExecuteTime,
		})
		return
	}
}
