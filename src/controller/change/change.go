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
*/

package change

import (
	"dbmcloud/log"
	"dbmcloud/src/database"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 通用查询请求结构
type QueryRequest struct {
	StartTime     string `json:"startTime"`     // 开始时间
	EndTime       string `json:"endTime"`       // 结束时间
	Status        string `json:"status"`        // 状态
	Environment   string `json:"environment"`   // 环境
	Page          int    `json:"page"`          // 页码
	PageSize      int    `json:"pageSize"`      // 每页大小
	SearchKeyword string `json:"searchKeyword"` // 搜索关键词
}

// 发布清单记录结构
type ReleaseRecord struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Version         string    `json:"version"`
	Status          string    `json:"status"`
	Environment     string    `json:"environment"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	Creator         string    `json:"creator"`
	Description     string    `json:"description"`
	Priority        string    `json:"priority"`
	Risk            string    `json:"risk"`
	ReleaseType     string    `json:"releaseType"`
	AffectedSystems string    `json:"affectedSystems"`
	RollbackPlan    string    `json:"rollbackPlan"`
}

// 运维变更工单记录结构
type WorkOrderRecord struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	WorkOrderType    string    `json:"workOrderType"`
	Status           string    `json:"status"`
	Environment      string    `json:"environment"`
	StartTime        time.Time `json:"startTime"`
	EndTime          time.Time `json:"endTime"`
	Creator          string    `json:"creator"`
	Assignee         string    `json:"assignee"`
	Description      string    `json:"description"`
	Priority         string    `json:"priority"`
	Risk             string    `json:"risk"`
	Category         string    `json:"category"`
	AffectedServices string    `json:"affectedServices"`
	Solution         string    `json:"solution"`
}

// 自动化变更记录结构
type AutoChangeRecord struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	AutomationType string    `json:"automationType"`
	Status         string    `json:"status"`
	Environment    string    `json:"environment"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	Creator        string    `json:"creator"`
	Description    string    `json:"description"`
	Priority       string    `json:"priority"`
	Risk           string    `json:"risk"`
	ScriptName     string    `json:"scriptName"`
	ExecutionMode  string    `json:"executionMode"`
	TargetServers  string    `json:"targetServers"`
	SuccessRate    float64   `json:"successRate"`
}

// 通用响应结构
type QueryResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	Today struct {
		ReleaseCount    int `json:"releaseCount"`
		WorkOrderCount  int `json:"workOrderCount"`
		AutoChangeCount int `json:"autoChangeCount"`
		FaultCount      int `json:"faultCount"`
	} `json:"today"`
	Annual struct {
		ReleaseCount    int `json:"releaseCount"`
		WorkOrderCount  int `json:"workOrderCount"`
		AutoChangeCount int `json:"autoChangeCount"`
		FaultCount      int `json:"faultCount"`
	} `json:"annual"`
	Trends struct {
		ReleaseTrend    []TrendData `json:"releaseTrend"`
		WorkOrderTrend  []TrendData `json:"workOrderTrend"`
		AutoChangeTrend []TrendData `json:"autoChangeTrend"`
		FaultTrend      []TrendData `json:"faultTrend"`
	} `json:"trends"`
}

// TrendData 趋势数据
type TrendData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// QueryReleaseList 发布清单查询
func QueryReleaseList(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Error("绑定发布清单查询请求失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	// 时间范围过滤
	if req.StartTime != "" {
		whereConditions = append(whereConditions, "start_time >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		whereConditions = append(whereConditions, "end_time <= ?")
		args = append(args, req.EndTime)
	}

	// 状态过滤
	if req.Status != "" {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, req.Status)
	}

	// 环境过滤
	if req.Environment != "" {
		whereConditions = append(whereConditions, "environment = ?")
		args = append(args, req.Environment)
	}

	// 关键词搜索
	if req.SearchKeyword != "" {
		whereConditions = append(whereConditions, "(title LIKE ? OR description LIKE ? OR creator LIKE ?)")
		keyword := "%" + req.SearchKeyword + "%"
		args = append(args, keyword, keyword, keyword)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取数据库连接
	db := database.DB

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM release_records WHERE %s", whereClause)
	err := db.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		log.Logger.Error("查询发布记录总数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	// 查询数据
	offset := (req.Page - 1) * req.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, title, version, status, environment, 
			start_time, end_time, creator, description, 
			priority, risk, release_type, affected_systems, rollback_plan
		FROM release_records 
		WHERE %s 
		ORDER BY start_time DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	var records []ReleaseRecord
	err = db.Raw(dataQuery, args...).Scan(&records).Error
	if err != nil {
		log.Logger.Error("查询发布记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	// 添加调试日志
	log.Logger.Info("发布清单查询结果", zap.Int("total", total), zap.Int("records_count", len(records)))

	response := QueryResponse{
		Success:  true,
		Message:  "查询成功",
		Data:     records,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// QueryWorkOrderList 运维变更查询
func QueryWorkOrderList(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Error("绑定运维变更查询请求失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	// 时间范围过滤
	if req.StartTime != "" {
		whereConditions = append(whereConditions, "start_time >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		whereConditions = append(whereConditions, "end_time <= ?")
		args = append(args, req.EndTime)
	}

	// 状态过滤
	if req.Status != "" {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, req.Status)
	}

	// 环境过滤
	if req.Environment != "" {
		whereConditions = append(whereConditions, "environment = ?")
		args = append(args, req.Environment)
	}

	// 关键词搜索
	if req.SearchKeyword != "" {
		whereConditions = append(whereConditions, "(title LIKE ? OR description LIKE ? OR creator LIKE ?)")
		keyword := "%" + req.SearchKeyword + "%"
		args = append(args, keyword, keyword, keyword)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取数据库连接
	db := database.DB

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workorder_records WHERE %s", whereClause)
	err := db.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		log.Logger.Error("查询工单记录总数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	// 查询数据
	offset := (req.Page - 1) * req.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, title, workorder_type, status, environment, 
			start_time, end_time, creator, assignee, description, 
			priority, risk, category, affected_services, solution
		FROM workorder_records 
		WHERE %s 
		ORDER BY start_time DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	var records []WorkOrderRecord
	err = db.Raw(dataQuery, args...).Scan(&records).Error
	if err != nil {
		log.Logger.Error("查询工单记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	// 添加调试日志
	log.Logger.Info("运维变更查询结果", zap.Int("total", total), zap.Int("records_count", len(records)))

	response := QueryResponse{
		Success:  true,
		Message:  "查询成功",
		Data:     records,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// QueryAutoChangeList 自动化变更查询
func QueryAutoChangeList(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Logger.Error("绑定自动化变更查询请求失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	// 时间范围过滤
	if req.StartTime != "" {
		whereConditions = append(whereConditions, "start_time >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		whereConditions = append(whereConditions, "end_time <= ?")
		args = append(args, req.EndTime)
	}

	// 状态过滤
	if req.Status != "" {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, req.Status)
	}

	// 环境过滤
	if req.Environment != "" {
		whereConditions = append(whereConditions, "environment = ?")
		args = append(args, req.Environment)
	}

	// 关键词搜索
	if req.SearchKeyword != "" {
		whereConditions = append(whereConditions, "(title LIKE ? OR description LIKE ? OR creator LIKE ?)")
		keyword := "%" + req.SearchKeyword + "%"
		args = append(args, keyword, keyword, keyword)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取数据库连接
	db := database.DB

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM autochange_records WHERE %s", whereClause)
	err := db.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		log.Logger.Error("查询自动化变更记录总数失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	// 查询数据
	offset := (req.Page - 1) * req.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, title, automation_type, status, environment, 
			start_time, end_time, creator, description, 
			priority, risk, script_name, execution_mode, target_servers, success_rate
		FROM autochange_records 
		WHERE %s 
		ORDER BY start_time DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	var records []AutoChangeRecord
	err = db.Raw(dataQuery, args...).Scan(&records).Error
	if err != nil {
		log.Logger.Error("查询自动化变更记录失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询失败",
		})
		return
	}

	// 添加调试日志
	log.Logger.Info("自动化变更查询结果", zap.Int("total", total), zap.Int("records_count", len(records)))

	response := QueryResponse{
		Success:  true,
		Message:  "查询成功",
		Data:     records,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetChangeDashboard 获取变更发布统计仪表板数据
func GetChangeDashboard(c *gin.Context) {
	// 获取数据库连接
	db := database.DB

	// 获取今日数据
	today := time.Now().Format("2006-01-02")
	todayStats := getTodayStats(db, today)

	// 获取年度数据
	currentYear := time.Now().Format("2006")
	annualStats := getAnnualStats(db, currentYear)

	// 获取趋势数据
	trends := getTrendData(db)

	// 构建响应数据
	dashboard := DashboardStats{
		Today:  todayStats,
		Annual: annualStats,
		Trends: trends,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取成功",
		"data":    dashboard,
	})
}

// getTodayStats 获取今日统计数据
func getTodayStats(db *gorm.DB, today string) struct {
	ReleaseCount    int `json:"releaseCount"`
	WorkOrderCount  int `json:"workOrderCount"`
	AutoChangeCount int `json:"autoChangeCount"`
	FaultCount      int `json:"faultCount"`
} {
	// 查询今日发布数
	var releaseCount int
	db.Raw("SELECT COUNT(*) FROM release_records WHERE DATE(start_time) = ?", today).Scan(&releaseCount)

	// 查询今日工单数
	var workOrderCount int
	db.Raw("SELECT COUNT(*) FROM workorder_records WHERE DATE(start_time) = ?", today).Scan(&workOrderCount)

	// 查询今日自动化变更数
	var autoChangeCount int
	db.Raw("SELECT COUNT(*) FROM autochange_records WHERE DATE(start_time) = ?", today).Scan(&autoChangeCount)

	// 查询今日故障数（从工单中统计故障处理类型）
	var faultCount int
	db.Raw("SELECT COUNT(*) FROM workorder_records WHERE workorder_type = '故障处理' AND DATE(start_time) = ?", today).Scan(&faultCount)

	log.Logger.Info("今日统计数据",
		zap.String("date", today),
		zap.Int("releaseCount", releaseCount),
		zap.Int("workOrderCount", workOrderCount),
		zap.Int("autoChangeCount", autoChangeCount),
		zap.Int("faultCount", faultCount))

	return struct {
		ReleaseCount    int `json:"releaseCount"`
		WorkOrderCount  int `json:"workOrderCount"`
		AutoChangeCount int `json:"autoChangeCount"`
		FaultCount      int `json:"faultCount"`
	}{
		ReleaseCount:    releaseCount,
		WorkOrderCount:  workOrderCount,
		AutoChangeCount: autoChangeCount,
		FaultCount:      faultCount,
	}
}

// getAnnualStats 获取年度统计数据
func getAnnualStats(db *gorm.DB, year string) struct {
	ReleaseCount    int `json:"releaseCount"`
	WorkOrderCount  int `json:"workOrderCount"`
	AutoChangeCount int `json:"autoChangeCount"`
	FaultCount      int `json:"faultCount"`
} {
	// 查询年度发布数
	var releaseCount int
	db.Raw("SELECT COUNT(*) FROM release_records WHERE YEAR(start_time) = ?", year).Scan(&releaseCount)

	// 查询年度工单数
	var workOrderCount int
	db.Raw("SELECT COUNT(*) FROM workorder_records WHERE YEAR(start_time) = ?", year).Scan(&workOrderCount)

	// 查询年度自动化变更数
	var autoChangeCount int
	db.Raw("SELECT COUNT(*) FROM autochange_records WHERE YEAR(start_time) = ?", year).Scan(&autoChangeCount)

	// 查询年度故障数
	var faultCount int
	db.Raw("SELECT COUNT(*) FROM workorder_records WHERE workorder_type = '故障处理' AND YEAR(start_time) = ?", year).Scan(&faultCount)

	log.Logger.Info("年度统计数据",
		zap.String("year", year),
		zap.Int("releaseCount", releaseCount),
		zap.Int("workOrderCount", workOrderCount),
		zap.Int("autoChangeCount", autoChangeCount),
		zap.Int("faultCount", faultCount))

	return struct {
		ReleaseCount    int `json:"releaseCount"`
		WorkOrderCount  int `json:"workOrderCount"`
		AutoChangeCount int `json:"autoChangeCount"`
		FaultCount      int `json:"faultCount"`
	}{
		ReleaseCount:    releaseCount,
		WorkOrderCount:  workOrderCount,
		AutoChangeCount: autoChangeCount,
		FaultCount:      faultCount,
	}
}

// getTrendData 获取趋势数据
func getTrendData(db *gorm.DB) struct {
	ReleaseTrend    []TrendData `json:"releaseTrend"`
	WorkOrderTrend  []TrendData `json:"workOrderTrend"`
	AutoChangeTrend []TrendData `json:"autoChangeTrend"`
	FaultTrend      []TrendData `json:"faultTrend"`
} {
	// 查询最近7天的数据
	startDate := time.Now().AddDate(0, 0, -6).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	log.Logger.Info("获取趋势数据",
		zap.String("startDate", startDate),
		zap.String("endDate", endDate))

	// 发布趋势
	releaseTrend := getTypeTrend(db, "release_records", startDate, endDate)
	log.Logger.Info("发布趋势数据", zap.Int("count", len(releaseTrend)))

	// 工单趋势
	workOrderTrend := getTypeTrend(db, "workorder_records", startDate, endDate)
	log.Logger.Info("工单趋势数据", zap.Int("count", len(workOrderTrend)))

	// 自动化变更趋势
	autoChangeTrend := getTypeTrend(db, "autochange_records", startDate, endDate)
	log.Logger.Info("自动化变更趋势数据", zap.Int("count", len(autoChangeTrend)))

	// 故障趋势（从工单中统计故障处理类型）
	faultTrend := getFaultTrend(db, startDate, endDate)
	log.Logger.Info("故障趋势数据", zap.Int("count", len(faultTrend)))

	return struct {
		ReleaseTrend    []TrendData `json:"releaseTrend"`
		WorkOrderTrend  []TrendData `json:"workOrderTrend"`
		AutoChangeTrend []TrendData `json:"autoChangeTrend"`
		FaultTrend      []TrendData `json:"faultTrend"`
	}{
		ReleaseTrend:    releaseTrend,
		WorkOrderTrend:  workOrderTrend,
		AutoChangeTrend: autoChangeTrend,
		FaultTrend:      faultTrend,
	}
}

// getTypeTrend 获取指定表的趋势数据
func getTypeTrend(db *gorm.DB, tableName, startDate, endDate string) []TrendData {
	var trends []TrendData

	// 生成最近7天的日期列表
	dateList := generateDateList(startDate, endDate)

	// 查询每天的数据
	for _, date := range dateList {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(start_time) = ?", tableName)
		db.Raw(query, date).Scan(&count)

		trends = append(trends, TrendData{
			Date:  date,
			Count: count,
		})
	}

	return trends
}

// getFaultTrend 获取故障趋势数据
func getFaultTrend(db *gorm.DB, startDate, endDate string) []TrendData {
	var trends []TrendData

	// 生成最近7天的日期列表
	dateList := generateDateList(startDate, endDate)

	// 查询每天的故障数据
	for _, date := range dateList {
		var count int
		query := "SELECT COUNT(*) FROM workorder_records WHERE workorder_type = '故障处理' AND DATE(start_time) = ?"
		db.Raw(query, date).Scan(&count)

		trends = append(trends, TrendData{
			Date:  date,
			Count: count,
		})
	}

	return trends
}

// generateDateList 生成日期列表
func generateDateList(startDate, endDate string) []string {
	var dates []string

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format("2006-01-02"))
	}

	return dates
}

// TestChangeAPI 测试变更API是否正常工作
func TestChangeAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "变更API服务正常",
		"timestamp": time.Now().Unix(),
	})
}
