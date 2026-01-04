package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// AnalysisTask 大模型分析任务
type AnalysisTask struct {
	Id              int        `gorm:"primarykey" json:"id"`
	TaskName        string     `gorm:"size:100;not null" json:"task_name"`
	TaskDescription string     `gorm:"size:500" json:"task_description"`
	DatasourceType  string     `gorm:"size:50;not null" json:"datasource_type"` // 数据源类型
	DatasourceId    int        `gorm:"not null" json:"datasource_id"`           // 数据源ID
	AiModelId       int        `gorm:"default:0" json:"ai_model_id"`            // AI模型ID
	SqlQueries      JsonArray  `gorm:"type:text" json:"sql_queries"`            // JSON数组存储多个SQL
	Prompt          string     `gorm:"type:text" json:"prompt"`
	CronExpression  string     `gorm:"size:100;not null" json:"cron_expression"`
	ReportEmail     string     `gorm:"size:200;not null" json:"report_email"`
	Status          int8       `gorm:"default:1" json:"status"` // 0: 禁用, 1: 启用
	LastRunTime     *time.Time `gorm:"column:last_run_time" json:"-"`
	NextRunTime     *time.Time `gorm:"column:next_run_time" json:"-"`
	CreatedAt       time.Time  `gorm:"column:gmt_created;index" json:"-"`
	UpdatedAt       time.Time  `gorm:"column:gmt_updated" json:"-"`
}

func (AnalysisTask) TableName() string {
	return "analysis_task"
}

// MarshalJSON 自定义JSON序列化，格式化时间字段
func (a AnalysisTask) MarshalJSON() ([]byte, error) {
	type Alias AnalysisTask

	// 创建临时结构体用于序列化
	aux := &struct {
		*Alias
		LastRunTime string `json:"last_run_time,omitempty"`
		NextRunTime string `json:"next_run_time,omitempty"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}{
		Alias: (*Alias)(&a),
	}

	// 格式化LastRunTime
	if a.LastRunTime != nil {
		aux.LastRunTime = a.LastRunTime.Format("2006-01-02 15:04:05")
	}

	// 格式化NextRunTime
	if a.NextRunTime != nil {
		aux.NextRunTime = a.NextRunTime.Format("2006-01-02 15:04:05")
	}

	// 格式化CreatedAt
	aux.CreatedAt = a.CreatedAt.Format("2006-01-02 15:04:05")

	// 格式化UpdatedAt
	aux.UpdatedAt = a.UpdatedAt.Format("2006-01-02 15:04:05")

	return json.Marshal(aux)
}

// AnalysisTaskLog 分析任务执行日志
type AnalysisTaskLog struct {
	Id            int64      `gorm:"primarykey" json:"id"`
	TaskId        int        `gorm:"index" json:"task_id"`
	TaskName      string     `gorm:"size:100" json:"task_name"`
	StartTime     time.Time  `gorm:"column:start_time" json:"-"`
	CompleteTime  *time.Time `gorm:"column:complete_time" json:"-"`
	Status        string     `gorm:"size:20;default:'running'" json:"status"` // running, success, failed
	Result        string     `gorm:"type:text" json:"result"`
	DataCount     int        `gorm:"default:0" json:"data_count"`
	ReportContent string     `gorm:"type:longtext" json:"report_content"`
	ErrorMessage  string     `gorm:"type:text" json:"error_message"`
	CreatedAt     time.Time  `gorm:"column:gmt_created;index" json:"-"`
}

func (AnalysisTaskLog) TableName() string {
	return "analysis_task_log"
}

// MarshalJSON 自定义JSON序列化，格式化时间字段
func (a AnalysisTaskLog) MarshalJSON() ([]byte, error) {
	type Alias AnalysisTaskLog

	// 创建临时结构体用于序列化
	aux := &struct {
		*Alias
		StartTime    string `json:"start_time"`
		CompleteTime string `json:"complete_time,omitempty"`
		CreatedAt    string `json:"created_at"`
	}{
		Alias: (*Alias)(&a),
	}

	// 格式化StartTime
	aux.StartTime = a.StartTime.Format("2006-01-02 15:04:05")

	// 格式化CompleteTime
	if a.CompleteTime != nil {
		aux.CompleteTime = a.CompleteTime.Format("2006-01-02 15:04:05")
	}

	// 格式化CreatedAt
	aux.CreatedAt = a.CreatedAt.Format("2006-01-02 15:04:05")

	return json.Marshal(aux)
}

// SqlQueryWithDatasource SQL查询与数据源映射结构
type SqlQueryWithDatasource struct {
	Sql          string `json:"sql"`
	DatasourceId int    `json:"datasource_id"`
}

// SqlQueryArray 用于处理SQL查询数组的GORM类型
type SqlQueryArray []SqlQueryWithDatasource

// Value 实现driver.Valuer接口
func (s SqlQueryArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现sql.Scanner接口
func (s *SqlQueryArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-string value into SqlQueryArray")
	}

	return json.Unmarshal(bytes, s)
}

// JsonArray 用于处理JSON数组的GORM类型
type JsonArray []string

// Value 实现driver.Valuer接口
func (j JsonArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JsonArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-string value into JsonArray")
	}

	return json.Unmarshal(bytes, j)
}

// MarshalJSON 实现json.Marshaler接口
func (j JsonArray) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(j))
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (j *JsonArray) UnmarshalJSON(data []byte) error {
	if j == nil {
		*j = make(JsonArray, 0)
	}
	return json.Unmarshal(data, (*[]string)(j))
}
