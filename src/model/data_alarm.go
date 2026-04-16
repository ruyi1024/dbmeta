package model

import (
	"encoding/json"
	"time"
)

// DataAlarm 数据告警任务
type DataAlarm struct {
	Id               int        `gorm:"primarykey" json:"id"`
	AlarmName        string     `gorm:"size:100;not null" json:"alarm_name"`
	AlarmDescription string     `gorm:"size:500" json:"alarm_description"`
	DatasourceType   string     `gorm:"size:50;not null" json:"datasource_type"`  // 数据源类型
	DatasourceId     int        `gorm:"not null" json:"datasource_id"`            // 数据源ID
	DatabaseName     string     `gorm:"size:100" json:"database_name"`            // 数据库名（可选）
	SqlQuery         string     `gorm:"type:text;not null" json:"sql_query"`      // SQL查询（只支持一条）
	RuleOperator     string     `gorm:"size:10;not null" json:"rule_operator"`    // 规则操作符: >, <, =, >=, <=, !=
	RuleValue        int        `gorm:"not null" json:"rule_value"`               // 规则值（数据量）
	EmailContent     string     `gorm:"type:text" json:"email_content"`           // 自定义邮件内容描述
	EmailTo          string     `gorm:"size:500;not null" json:"email_to"`        // 接收邮箱（多个用逗号分隔）
	CronExpression   string     `gorm:"size:100;not null" json:"cron_expression"` // Cron表达式
	Status           int8       `gorm:"default:1" json:"status"`                  // 0: 禁用, 1: 启用
	LastRunTime      *time.Time `gorm:"column:last_run_time" json:"-"`
	NextRunTime      *time.Time `gorm:"column:next_run_time" json:"-"`
	CreatedAt        time.Time  `gorm:"column:gmt_created;index" json:"-"`
	UpdatedAt        time.Time  `gorm:"column:gmt_updated" json:"-"`
}

func (DataAlarm) TableName() string {
	return "data_alarm"
}

// MarshalJSON 自定义JSON序列化，格式化时间字段
func (d DataAlarm) MarshalJSON() ([]byte, error) {
	type Alias DataAlarm

	aux := &struct {
		*Alias
		LastRunTime string `json:"last_run_time,omitempty"`
		NextRunTime string `json:"next_run_time,omitempty"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}{
		Alias: (*Alias)(&d),
	}

	if d.LastRunTime != nil {
		aux.LastRunTime = d.LastRunTime.Format("2006-01-02 15:04:05")
	}

	if d.NextRunTime != nil {
		aux.NextRunTime = d.NextRunTime.Format("2006-01-02 15:04:05")
	}

	aux.CreatedAt = d.CreatedAt.Format("2006-01-02 15:04:05")
	aux.UpdatedAt = d.UpdatedAt.Format("2006-01-02 15:04:05")

	return json.Marshal(aux)
}

// DataAlarmLog 数据告警执行日志
type DataAlarmLog struct {
	Id           int64      `gorm:"primarykey" json:"id"`
	AlarmId      int        `gorm:"index" json:"alarm_id"`
	AlarmName    string     `gorm:"size:100" json:"alarm_name"`
	StartTime    time.Time  `gorm:"column:start_time" json:"-"`
	CompleteTime *time.Time `gorm:"column:complete_time" json:"-"`
	Status       string     `gorm:"size:20;default:'running'" json:"status"` // running, success, failed, triggered
	DataCount    int        `gorm:"default:0" json:"data_count"`             // 查询结果数据量
	RuleMatched  bool       `gorm:"default:0" json:"rule_matched"`           // 规则是否匹配
	EmailSent    bool       `gorm:"default:0" json:"email_sent"`             // 邮件是否已发送
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	ReportHTML   string     `gorm:"column:report_html;type:longtext" json:"-"`
	CreatedAt    time.Time  `gorm:"column:gmt_created;index" json:"-"`
}

func (DataAlarmLog) TableName() string {
	return "data_alarm_log"
}

// MarshalJSON 自定义JSON序列化，格式化时间字段
func (d DataAlarmLog) MarshalJSON() ([]byte, error) {
	type Alias DataAlarmLog

	aux := &struct {
		*Alias
		StartTime    string `json:"start_time"`
		CompleteTime string `json:"complete_time,omitempty"`
		CreatedAt    string `json:"created_at"`
	}{
		Alias: (*Alias)(&d),
	}

	aux.StartTime = d.StartTime.Format("2006-01-02 15:04:05")

	if d.CompleteTime != nil {
		aux.CompleteTime = d.CompleteTime.Format("2006-01-02 15:04:05")
	}

	aux.CreatedAt = d.CreatedAt.Format("2006-01-02 15:04:05")

	return json.Marshal(aux)
}
