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

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ChatSession 聊天会话
type ChatSession struct {
	Id        int64     `gorm:"primarykey" json:"id"`
	SessionId string    `gorm:"size:100;uniqueIndex;not null" json:"session_id"`
	UserName  string    `gorm:"size:50;index;not null" json:"user_name"`
	Title     string    `gorm:"size:200" json:"title"`
	CreatedAt time.Time `gorm:"column:gmt_created;index" json:"gmt_created"`
	UpdatedAt time.Time `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Id          int64            `gorm:"primarykey" json:"id"`
	SessionId   string           `gorm:"size:100;index;not null" json:"session_id"`
	Role        string           `gorm:"size:20;not null" json:"role"` // user or assistant
	Content     string           `gorm:"type:text" json:"content"`
	SqlQuery    string           `gorm:"type:text" json:"sql_query"`    // 生成的SQL（如果是查询）
	QueryResult QueryResultArray `gorm:"type:text" json:"query_result"` // 查询结果（JSON格式）
	CreatedAt   time.Time        `gorm:"column:gmt_created;index" json:"gmt_created"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

// SqlSetItem SQL集合项（用于多SQL查询）
type SqlSetItem struct {
	Sql         string            `json:"sql"`         // SQL模板
	Description string            `json:"description"` // SQL描述
	DependsOn   string            `json:"depends_on"`  // 依赖的前一个SQL的索引（可选）
	Outputs     map[string]string `json:"outputs"`     // 输出字段映射，如 {"threads_running": "Threads_running"}
}

// QuestionFlowItem 问题流程项（用于多轮对话）
type QuestionFlowItem struct {
	Key         string   `json:"key"`         // 参数键名，如 "user_id", "db_type", "db_name"
	Question    string   `json:"question"`    // 提示问题，如 "请输入查询用户的ID"
	Type        string   `json:"type"`        // 输入类型：text/select/number，默认为text
	Options     []string `json:"options"`     // 选项列表（当type为select时使用），如 ["MySQL", "PostgreSQL", "Oracle"]
	OptionsSQL  string   `json:"options_sql"` // 获取选项的SQL（当type为select时，如果设置了此字段，会执行SQL获取选项列表）
	Required    bool     `json:"required"`    // 是否必填，默认为true
	Validation  string   `json:"validation"`  // 验证规则（可选），如 "number", "email"
	Description string   `json:"description"` // 参数描述（可选）
}

// QuestionFlowArray 问题流程数组
type QuestionFlowArray []QuestionFlowItem

// Value 实现 driver.Valuer 接口
func (j QuestionFlowArray) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *QuestionFlowArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// ParameterMapping 参数映射配置（定义如何将收集的信息映射到SQL参数）
type ParameterMapping map[string]string

// Value 实现 driver.Valuer 接口
func (j ParameterMapping) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *ParameterMapping) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// SemanticSqlRule 语义-SQL规则
type SemanticSqlRule struct {
	Id                int64             `gorm:"primarykey" json:"id"`
	RuleName          string            `gorm:"size:100;not null" json:"rule_name"`
	SemanticPattern   string            `gorm:"type:text;not null" json:"semantic_pattern"` // 语义模式（支持正则）
	SqlTemplate       string            `gorm:"type:text;not null" json:"sql_template"`     // SQL模板（支持参数占位符，单SQL时使用）
	SqlSet            string            `gorm:"type:text" json:"sql_set"`                   // SQL集合（JSON格式，多SQL时使用）
	QueryType         string            `gorm:"size:50;index" json:"query_type"`            // status/performance/metadata/custom/report
	Description       string            `gorm:"type:text" json:"description"`
	ReportTemplate    string            `gorm:"type:text" json:"report_template"`                              // 报告模板（支持占位符）
	Enabled           int8              `gorm:"default:1" json:"enabled"`                                      // 0: 禁用, 1: 启用
	Priority          int               `gorm:"default:0" json:"priority"`                                     // 优先级（数字越大优先级越高）
	UseLocalDB        int8              `gorm:"default:0;comment:'0:使用远程数据源,1:使用本地MySQL'" json:"use_local_db"` // 0: 使用远程数据源, 1: 使用本地MySQL
	MultiRoundEnabled int8              `gorm:"default:0;comment:'0:单轮对话,1:多轮对话'" json:"multi_round_enabled"`  // 0: 单轮对话, 1: 多轮对话
	QuestionFlow      QuestionFlowArray `gorm:"type:text;comment:'问题流程配置(JSON格式)'" json:"question_flow"`       // 问题流程配置
	ParameterMapping  ParameterMapping  `gorm:"type:text;comment:'参数映射配置(JSON格式)'" json:"parameter_mapping"`   // 参数映射配置
	CreatedAt         time.Time         `gorm:"column:gmt_created" json:"gmt_created"`
	UpdatedAt         time.Time         `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (SemanticSqlRule) TableName() string {
	return "semantic_sql_rules"
}

// QueryResultArray 用于存储查询结果的JSON数组
type QueryResultArray []map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j QueryResultArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *QueryResultArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}
