package data

import (
	"dbmcloud/src/database"
	"dbmcloud/src/libary/mail"
	"dbmcloud/src/model"
	"dbmcloud/src/service"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DataAlarmList 获取数据告警列表
func DataAlarmList(c *gin.Context) {
	method := c.Request.Method
	if method == "GET" {
		var db = database.DB

		// 查询条件
		if c.Query("alarm_name") != "" {
			db = db.Where("alarm_name LIKE ?", "%"+c.Query("alarm_name")+"%")
		}
		if c.Query("datasource_type") != "" {
			db = db.Where("datasource_type = ?", c.Query("datasource_type"))
		}
		if c.Query("status") != "" {
			db = db.Where("status = ?", c.Query("status"))
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

		var dataList []model.DataAlarm
		var total int64

		// 获取总数
		db.Model(&model.DataAlarm{}).Count(&total)

		// 获取数据
		result := db.Offset(offset).Limit(pageSize).Find(&dataList)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"msg":     "查询失败: " + result.Error.Error(),
			})
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

	if method == "POST" {
		var alarm model.DataAlarm
		if err := c.BindJSON(&alarm); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
			return
		}

		// 验证必填字段
		if alarm.AlarmName == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警名称不能为空"})
			return
		}
		if alarm.DatasourceType == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源类型不能为空"})
			return
		}
		if alarm.DatasourceId == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源不能为空"})
			return
		}
		if alarm.SqlQuery == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "SQL查询不能为空"})
			return
		}
		if alarm.RuleOperator == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "规则操作符不能为空"})
			return
		}
		if alarm.CronExpression == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Cron表达式不能为空"})
			return
		}
		if alarm.EmailTo == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "接收邮箱不能为空"})
			return
		}

		// 创建告警
		result := database.DB.Create(&alarm)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "创建失败: " + result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "创建成功"})
		return
	}

	if method == "PUT" {
		var alarm model.DataAlarm
		if err := c.BindJSON(&alarm); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
			return
		}

		if alarm.Id == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警ID不能为空"})
			return
		}

		// 更新告警
		updateData := map[string]interface{}{
			"alarm_name":        alarm.AlarmName,
			"alarm_description": alarm.AlarmDescription,
			"datasource_type":   alarm.DatasourceType,
			"datasource_id":     alarm.DatasourceId,
			"sql_query":         alarm.SqlQuery,
			"rule_operator":     alarm.RuleOperator,
			"rule_value":        alarm.RuleValue,
			"email_content":     alarm.EmailContent,
			"email_to":          alarm.EmailTo,
			"cron_expression":   alarm.CronExpression,
			"status":            alarm.Status,
		}
		result := database.DB.Model(&model.DataAlarm{}).Where("id = ?", alarm.Id).Updates(updateData)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "更新失败: " + result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "更新成功"})
		return
	}

	if method == "DELETE" {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警ID不能为空"})
			return
		}

		alarmId, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "无效的告警ID"})
			return
		}

		// 删除告警
		result := database.DB.Delete(&model.DataAlarm{}, alarmId)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "删除失败: " + result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "删除成功"})
		return
	}
}

// ToggleDataAlarmStatus 启用/禁用数据告警
func ToggleDataAlarmStatus(c *gin.Context) {
	var request struct {
		Id     int `json:"id"`
		Status int `json:"status"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
		return
	}

	if request.Id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警ID不能为空"})
		return
	}

	result := database.DB.Model(&model.DataAlarm{}).Where("id = ?", request.Id).Update("status", request.Status)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "更新失败: " + result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "更新成功"})
}

// ExecuteDataAlarm 手动执行数据告警
func ExecuteDataAlarm(c *gin.Context) {
	var request struct {
		Id int `json:"id" binding:"required"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
		return
	}

	var alarm model.DataAlarm
	result := database.DB.Where("id = ?", request.Id).First(&alarm)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警不存在"})
		return
	}

	// 异步执行告警任务
	go func() {
		executeAlarmTaskInternal(&alarm)
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "告警任务已开始执行"})
}

// DataAlarmLogs 获取数据告警执行日志
func DataAlarmLogs(c *gin.Context) {
	var db = database.DB

	// 查询条件
	if c.Query("alarm_id") != "" {
		if alarmId, err := strconv.Atoi(c.Query("alarm_id")); err == nil {
			db = db.Where("alarm_id = ?", alarmId)
		}
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
			endDate = endDate.Add(24 * time.Hour)
			db = db.Where("gmt_created < ?", endDate)
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

	var dataList []model.DataAlarmLog
	var total int64

	// 获取总数
	db.Model(&model.DataAlarmLog{}).Count(&total)

	// 获取数据
	result := db.Order("gmt_created DESC").Offset(offset).Limit(pageSize).Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"msg":     "查询失败: " + result.Error.Error(),
		})
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
}

// GetDataAlarmDetail 获取数据告警详情
func GetDataAlarmDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警ID不能为空"})
		return
	}

	alarmId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "无效的告警ID"})
		return
	}

	var alarm model.DataAlarm
	result := database.DB.Where("id = ?", alarmId).First(&alarm)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "告警不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": alarm})
}

// TestSqlQuery 测试SQL查询
func TestSqlQuery(c *gin.Context) {
	var request struct {
		Sql            string `json:"sql" binding:"required"`
		DatasourceType string `json:"datasource_type" binding:"required"`
		DatasourceId   int    `json:"datasource_id" binding:"required"`
		DatabaseName   string `json:"database_name"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "参数解析失败: " + err.Error()})
		return
	}

	// 获取数据源
	var datasource model.Datasource
	result := database.DB.Where("id = ?", request.DatasourceId).First(&datasource)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源不存在"})
		return
	}

	// 执行SQL查询（如果指定了数据库名，传递数据库名参数）
	var data []map[string]interface{}
	var err error
	if request.DatabaseName != "" {
		data, err = service.ExecuteQuery(request.Sql, request.DatasourceId, request.DatabaseName)
	} else {
		data, err = service.ExecuteQuery(request.Sql, request.DatasourceId)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "SQL执行失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// GetDatasourceTypeList 获取数据源类型列表
func GetDatasourceTypeList(c *gin.Context) {
	var dataList []model.DatasourceType
	var total int64

	// 查询条件
	query := database.DB.Where("enable = 1")

	// 获取总数
	query.Model(&model.DatasourceType{}).Count(&total)

	// 获取数据
	result := query.Select("id, name").Order("sort ASC").Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    dataList,
		"total":   total,
	})
}

// GetDatasourceList 获取数据源列表
func GetDatasourceList(c *gin.Context) {
	var dataList []model.Datasource
	var total int64

	// 查询条件
	query := database.DB.Where("enable = 1")

	if c.Query("type") != "" {
		query = query.Where("type = ?", c.Query("type"))
	}
	if c.Query("env") != "" {
		query = query.Where("env = ?", c.Query("env"))
	}

	// 获取总数
	query.Model(&model.Datasource{}).Count(&total)

	// 获取数据
	result := query.Select("id, name, type, host, port, dbid, env").Order("name ASC").Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    dataList,
		"total":   total,
	})
}

// GetDatabaseList 获取数据库列表
func GetDatabaseList(c *gin.Context) {
	datasourceId := c.Query("datasource_id")
	if datasourceId == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源ID不能为空"})
		return
	}

	// 获取数据源信息
	var datasource model.Datasource
	result := database.DB.Where("id = ? AND enable = 1", datasourceId).First(&datasource)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据源不存在或已禁用"})
		return
	}

	// 根据数据源类型查询数据库列表
	var databases []model.MetaDatabase
	query := database.DB.Where("is_deleted = 0").
		Where("datasource_type = ?", datasource.Type).
		Where("host = ?", datasource.Host).
		Where("port = ?", datasource.Port)

	result = query.Select("database_name").Group("database_name").Order("database_name ASC").Find(&databases)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "查询数据库列表失败: " + result.Error.Error()})
		return
	}

	// 转换为简单的数据库名列表
	dbNames := make([]string, 0, len(databases))
	for _, db := range databases {
		if db.DatabaseName != "" {
			dbNames = append(dbNames, db.DatabaseName)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    dbNames,
	})
}

// ExecuteAlarmTask 执行告警任务（导出函数，供定时任务调用）
func ExecuteAlarmTask(alarm *model.DataAlarm) {
	executeAlarmTaskInternal(alarm)
}

// executeAlarmTaskInternal 执行告警任务（内部函数）
func executeAlarmTaskInternal(alarm *model.DataAlarm) {
	// 创建日志记录
	log := model.DataAlarmLog{
		AlarmId:   alarm.Id,
		AlarmName: alarm.AlarmName,
		StartTime: time.Now(),
		Status:    "running",
	}
	database.DB.Create(&log)

	// 获取数据源
	var datasource model.Datasource
	result := database.DB.Where("id = ?", alarm.DatasourceId).First(&datasource)
	if result.Error != nil {
		log.Status = "failed"
		log.ErrorMessage = "数据源不存在: " + result.Error.Error()
		completeTime := time.Now()
		log.CompleteTime = &completeTime
		database.DB.Save(&log)
		return
	}

	// 执行SQL查询（如果指定了数据库名，传递数据库名参数）
	var data []map[string]interface{}
	var err error
	if alarm.DatabaseName != "" {
		data, err = service.ExecuteQuery(alarm.SqlQuery, alarm.DatasourceId, alarm.DatabaseName)
	} else {
		data, err = service.ExecuteQuery(alarm.SqlQuery, alarm.DatasourceId)
	}
	if err != nil {
		log.Status = "failed"
		log.ErrorMessage = "SQL执行失败: " + err.Error()
		completeTime := time.Now()
		log.CompleteTime = &completeTime
		database.DB.Save(&log)
		return
	}

	// 获取数据量
	dataCount := len(data)
	log.DataCount = dataCount

	// 判断规则是否匹配
	ruleMatched := checkRule(dataCount, alarm.RuleOperator, alarm.RuleValue)
	log.RuleMatched = ruleMatched

	// 如果规则匹配，发送邮件
	if ruleMatched {
		log.Status = "triggered"
		emailSent := sendAlarmEmail(alarm, dataCount, data)
		log.EmailSent = emailSent
	} else {
		log.Status = "success"
	}

	// 更新日志
	completeTime := time.Now()
	log.CompleteTime = &completeTime
	database.DB.Save(&log)

	// 更新告警的最后运行时间
	database.DB.Model(&model.DataAlarm{}).Where("id = ?", alarm.Id).Update("last_run_time", time.Now())
}

// checkRule 检查规则是否匹配
func checkRule(dataCount int, operator string, ruleValue int) bool {
	switch operator {
	case ">":
		return dataCount > ruleValue
	case "<":
		return dataCount < ruleValue
	case "=":
		return dataCount == ruleValue
	case ">=":
		return dataCount >= ruleValue
	case "<=":
		return dataCount <= ruleValue
	case "!=":
		return dataCount != ruleValue
	default:
		return false
	}
}

// sendAlarmEmail 发送告警邮件
func sendAlarmEmail(alarm *model.DataAlarm, dataCount int, data []map[string]interface{}) bool {
	// 解析邮箱列表（支持英文分号分隔）
	emailList := strings.Split(alarm.EmailTo, ";")
	var cleanEmails []string
	for _, e := range emailList {
		e = strings.TrimSpace(e)
		if e != "" {
			cleanEmails = append(cleanEmails, e)
		}
	}

	if len(cleanEmails) == 0 {
		return false
	}

	// 构造邮件主题
	subject := fmt.Sprintf("数据告警 - %s", alarm.AlarmName)

	// 构造邮件内容
	body := buildAlarmEmailBody(alarm, dataCount, data)

	// 发送邮件
	if err := mail.Send(cleanEmails, subject, body); err != nil {
		zap.L().Error("发送告警邮件失败", zap.Error(err))
		return false
	}

	return true
}

// buildAlarmEmailBody 构造告警邮件内容
func buildAlarmEmailBody(alarm *model.DataAlarm, dataCount int, data []map[string]interface{}) string {
	var html strings.Builder

	html.WriteString("<html><body>")
	html.WriteString("<h2>数据告警通知</h2>")
	html.WriteString(fmt.Sprintf("<p><strong>告警名称：</strong>%s</p>", alarm.AlarmName))
	html.WriteString(fmt.Sprintf("<p><strong>告警描述：</strong>%s</p>", alarm.AlarmDescription))
	html.WriteString(fmt.Sprintf("<p><strong>触发时间：</strong>%s</p>", time.Now().Format("2006-01-02 15:04:05")))
	html.WriteString(fmt.Sprintf("<p><strong>数据量：</strong>%d 条</p>", dataCount))
	html.WriteString(fmt.Sprintf("<p><strong>规则：</strong>数据量 %s %d</p>", getOperatorText(alarm.RuleOperator), alarm.RuleValue))

	// 如果有自定义邮件内容描述，显示在表格上方
	if alarm.EmailContent != "" {
		html.WriteString("<div style='margin: 20px 0; padding: 15px; background-color: #f5f5f5; border-left: 4px solid #1890ff;'>")
		html.WriteString("<h3>告警说明：</h3>")
		html.WriteString("<p>" + strings.ReplaceAll(alarm.EmailContent, "\n", "<br>") + "</p>")
		html.WriteString("</div>")
	}

	// 显示数据表格
	if len(data) > 0 {
		html.WriteString("<h3>查询结果：</h3>")
		html.WriteString("<table border='1' cellpadding='5' cellspacing='0' style='border-collapse: collapse; width: 100%;'>")

		// 表头
		html.WriteString("<tr style='background-color: #f0f0f0;'>")
		for key := range data[0] {
			html.WriteString(fmt.Sprintf("<th>%s</th>", key))
		}
		html.WriteString("</tr>")

		// 数据行（最多显示100行）
		maxRows := 100
		if len(data) < maxRows {
			maxRows = len(data)
		}
		for i := 0; i < maxRows; i++ {
			html.WriteString("<tr>")
			for _, value := range data[i] {
				html.WriteString(fmt.Sprintf("<td>%v</td>", value))
			}
			html.WriteString("</tr>")
		}
		html.WriteString("</table>")

		if len(data) > maxRows {
			html.WriteString(fmt.Sprintf("<p><em>（仅显示前 %d 条，共 %d 条）</em></p>", maxRows, len(data)))
		}
	}

	html.WriteString("</body></html>")
	return html.String()
}

// getOperatorText 获取操作符文本
func getOperatorText(operator string) string {
	operatorMap := map[string]string{
		">":  "大于",
		"<":  "小于",
		"=":  "等于",
		">=": "大于等于",
		"<=": "小于等于",
		"!=": "不等于",
	}
	if text, ok := operatorMap[operator]; ok {
		return text
	}
	return operator
}
