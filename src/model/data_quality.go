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
	"time"
)

// DataQualityAssessment 数据质量评估主表
type DataQualityAssessment struct {
	Id                int64              `gorm:"primaryKey" json:"id"`
	AssessmentTime    time.Time          `gorm:"column:assessment_time;index" json:"assessment_time"`
	DatasourceId      *int64             `gorm:"column:datasource_id;index" json:"datasource_id"`
	DatabaseName      string             `gorm:"column:database_name;size:255;index" json:"database_name"`
	TotalTables       int                `gorm:"column:total_tables;default:0" json:"total_tables"`
	TotalColumns      int                `gorm:"column:total_columns;default:0" json:"total_columns"`
	TotalIssues       int                `gorm:"column:total_issues;default:0" json:"total_issues"`
	OverallScore      float64            `gorm:"column:overall_score;type:decimal(5,2);default:0.00" json:"overall_score"`
	OverallLevel      string             `gorm:"column:overall_level;size:20;default:'未知'" json:"overall_level"`
	FieldCompleteness float64            `gorm:"column:field_completeness;type:decimal(5,2);default:0.00" json:"field_completeness"`
	FieldAccuracy     float64            `gorm:"column:field_accuracy;type:decimal(5,2);default:0.00" json:"field_accuracy"`
	TableCompleteness float64            `gorm:"column:table_completeness;type:decimal(5,2);default:0.00" json:"table_completeness"`
	DataConsistency   float64            `gorm:"column:data_consistency;type:decimal(5,2);default:0.00" json:"data_consistency"`
	DataUniqueness    float64            `gorm:"column:data_uniqueness;type:decimal(5,2);default:0.00" json:"data_uniqueness"`
	DataTimeliness    float64            `gorm:"column:data_timeliness;type:decimal(5,2);default:0.00" json:"data_timeliness"`
	Status            int8               `gorm:"column:status;default:1" json:"status"`
	Issues            []DataQualityIssue `gorm:"-" json:"-"` // 不持久化到数据库，仅用于内存处理
	CreatedAt         time.Time          `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time          `gorm:"column:updated_at" json:"updated_at"`
}

func (DataQualityAssessment) TableName() string {
	return "data_quality_assessment"
}

// DataQualityIssue 质量问题详情表
type DataQualityIssue struct {
	Id           int64      `gorm:"primaryKey" json:"id"`
	AssessmentId int64      `gorm:"column:assessment_id;index;not null" json:"assessment_id"`
	DatabaseName string     `gorm:"column:database_name;size:255;index" json:"database_name"`
	TableNameX   string     `gorm:"column:table_name;size:255;not null;index" json:"table_name"`
	ColumnName   string     `gorm:"column:column_name;size:255;index" json:"column_name"`
	IssueType    string     `gorm:"column:issue_type;size:50;not null;index" json:"issue_type"`
	IssueLevel   string     `gorm:"column:issue_level;size:20;not null;index" json:"issue_level"`
	IssueDesc    string     `gorm:"column:issue_desc;type:text" json:"issue_desc"`
	IssueCount   int        `gorm:"column:issue_count;default:0" json:"issue_count"`
	IssueRate    float64    `gorm:"column:issue_rate;type:decimal(5,2);default:0.00" json:"issue_rate"`
	SampleData   string     `gorm:"column:sample_data;type:text" json:"sample_data"`
	CheckTime    time.Time  `gorm:"column:check_time;not null;index" json:"check_time"`
	Status       int8       `gorm:"column:status;default:1" json:"status"` // 1:待处理 2:处理中 3:已处理 0:已忽略
	Handler      string     `gorm:"column:handler;size:100" json:"handler"`
	HandleTime   *time.Time `gorm:"column:handle_time" json:"handle_time"`
	HandleRemark string     `gorm:"column:handle_remark;type:text" json:"handle_remark"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (DataQualityIssue) TableName() string {
	return "data_quality_issue"
}

// DataQualityAiAnalysis AI分析结果表
type DataQualityAiAnalysis struct {
	Id              int64     `gorm:"primaryKey" json:"id"`
	AssessmentId    int64     `gorm:"column:assessment_id;index;not null" json:"assessment_id"`
	AiScore         float64   `gorm:"column:ai_score;type:decimal(5,2);default:0.00" json:"ai_score"`
	AiLevel         string    `gorm:"column:ai_level;size:20;default:'未知'" json:"ai_level"`
	TrendAnalysis   string    `gorm:"column:trend_analysis;type:text" json:"trend_analysis"`
	TrendDirection  string    `gorm:"column:trend_direction;size:20" json:"trend_direction"`
	TrendPercentage float64   `gorm:"column:trend_percentage;type:decimal(5,2);default:0.00" json:"trend_percentage"`
	AnalysisTime    time.Time `gorm:"column:analysis_time;not null;index" json:"analysis_time"`
	ModelVersion    string    `gorm:"column:model_version;size:50" json:"model_version"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DataQualityAiAnalysis) TableName() string {
	return "data_quality_ai_analysis"
}

// DataQualityAiInsight AI智能洞察表
type DataQualityAiInsight struct {
	Id             int64     `gorm:"primaryKey" json:"id"`
	AiAnalysisId   int64     `gorm:"column:ai_analysis_id;index;not null" json:"ai_analysis_id"`
	InsightContent string    `gorm:"column:insight_content;type:text;not null" json:"insight_content"`
	InsightType    string    `gorm:"column:insight_type;size:50" json:"insight_type"`
	Priority       int       `gorm:"column:priority;default:0;index" json:"priority"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (DataQualityAiInsight) TableName() string {
	return "data_quality_ai_insight"
}

// DataQualityAiRecommendation AI优化建议表
type DataQualityAiRecommendation struct {
	Id                  int64     `gorm:"primaryKey" json:"id"`
	AiAnalysisId        int64     `gorm:"column:ai_analysis_id;index;not null" json:"ai_analysis_id"`
	RecommendationType  string    `gorm:"column:recommendation_type;size:20;not null;index" json:"recommendation_type"`
	Title               string    `gorm:"column:title;size:255;not null" json:"title"`
	Description         string    `gorm:"column:description;type:text;not null" json:"description"`
	Priority            string    `gorm:"column:priority;size:20;not null" json:"priority"`
	ExpectedImprovement float64   `gorm:"column:expected_improvement;type:decimal(5,2);default:0.00" json:"expected_improvement"`
	RelatedIssues       string    `gorm:"column:related_issues;type:text" json:"related_issues"`
	Status              int8      `gorm:"column:status;default:0;index" json:"status"` // 0:未采纳 1:已采纳 2:已实施
	CreatedAt           time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DataQualityAiRecommendation) TableName() string {
	return "data_quality_ai_recommendation"
}

// DataQualityDistribution 质量指标分布数据表
type DataQualityDistribution struct {
	Id               int64     `gorm:"primaryKey" json:"id"`
	AssessmentId     int64     `gorm:"column:assessment_id;index;not null" json:"assessment_id"`
	DistributionType string    `gorm:"column:distribution_type;size:50;not null;index" json:"distribution_type"`
	Category         string    `gorm:"column:category;size:100;not null" json:"category"`
	Value            int64     `gorm:"column:value;default:0" json:"value"`
	Percentage       float64   `gorm:"column:percentage;type:decimal(5,2);default:0.00" json:"percentage"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
}

func (DataQualityDistribution) TableName() string {
	return "data_quality_distribution"
}

// DataQualityRule 质量规则配置表
type DataQualityRule struct {
	Id         int64     `gorm:"primaryKey" json:"id"`
	RuleName   string    `gorm:"column:rule_name;size:255;not null" json:"rule_name"`
	RuleType   string    `gorm:"column:rule_type;size:50;not null;index" json:"rule_type"`
	RuleDesc   string    `gorm:"column:rule_desc;type:text" json:"rule_desc"`
	RuleConfig string    `gorm:"column:rule_config;type:text" json:"rule_config"`
	Threshold  float64   `gorm:"column:threshold;type:decimal(5,2);default:0.00" json:"threshold"`
	Severity   string    `gorm:"column:severity;size:20;default:'medium';index" json:"severity"`
	Enabled    int8      `gorm:"column:enabled;default:1;index" json:"enabled"`
	CreatedBy  string    `gorm:"column:created_by;size:100" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DataQualityRule) TableName() string {
	return "data_quality_rule"
}

// DataQualityTask 质量评估任务表
type DataQualityTask struct {
	Id             int64      `gorm:"primaryKey" json:"id"`
	TaskName       string     `gorm:"column:task_name;size:255;not null" json:"task_name"`
	TaskType       string     `gorm:"column:task_type;size:50;not null;index" json:"task_type"`
	DatasourceId   *int64     `gorm:"column:datasource_id;index" json:"datasource_id"`
	DatabaseName   string     `gorm:"column:database_name;size:255" json:"database_name"`
	TableFilter    string     `gorm:"column:table_filter;type:text" json:"table_filter"`
	ScheduleConfig string     `gorm:"column:schedule_config;type:text" json:"schedule_config"`
	Status         string     `gorm:"column:status;size:20;default:'pending';index" json:"status"`
	StartTime      *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime        *time.Time `gorm:"column:end_time" json:"end_time"`
	Duration       int        `gorm:"column:duration;default:0" json:"duration"`
	ResultSummary  string     `gorm:"column:result_summary;type:text" json:"result_summary"`
	ErrorMessage   string     `gorm:"column:error_message;type:text" json:"error_message"`
	CreatedBy      string     `gorm:"column:created_by;size:100" json:"created_by"`
	CreatedAt      time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (DataQualityTask) TableName() string {
	return "data_quality_task"
}

// DataQualityHistory 质量评估历史表
type DataQualityHistory struct {
	Id                int64     `gorm:"primaryKey" json:"id"`
	AssessmentId      int64     `gorm:"column:assessment_id;index;not null" json:"assessment_id"`
	DatasourceId      *int64    `gorm:"column:datasource_id;index" json:"datasource_id"`
	DatabaseName      string    `gorm:"column:database_name;size:255;index" json:"database_name"`
	AssessmentDate    time.Time `gorm:"column:assessment_date;type:date;not null;index" json:"assessment_date"`
	OverallScore      float64   `gorm:"column:overall_score;type:decimal(5,2);default:0.00" json:"overall_score"`
	FieldCompleteness float64   `gorm:"column:field_completeness;type:decimal(5,2);default:0.00" json:"field_completeness"`
	FieldAccuracy     float64   `gorm:"column:field_accuracy;type:decimal(5,2);default:0.00" json:"field_accuracy"`
	TableCompleteness float64   `gorm:"column:table_completeness;type:decimal(5,2);default:0.00" json:"table_completeness"`
	DataConsistency   float64   `gorm:"column:data_consistency;type:decimal(5,2);default:0.00" json:"data_consistency"`
	DataUniqueness    float64   `gorm:"column:data_uniqueness;type:decimal(5,2);default:0.00" json:"data_uniqueness"`
	DataTimeliness    float64   `gorm:"column:data_timeliness;type:decimal(5,2);default:0.00" json:"data_timeliness"`
	TotalIssues       int       `gorm:"column:total_issues;default:0" json:"total_issues"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
}

func (DataQualityHistory) TableName() string {
	return "data_quality_history"
}
