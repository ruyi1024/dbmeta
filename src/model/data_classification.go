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

package model

import (
	"time"
)

// DataSecurityGrade 数据安全分级字典（对标 GB 通用三类：一般数据 / 重要数据 / 核心数据）
type DataSecurityGrade struct {
	Id          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	GradeCode   string    `gorm:"column:grade_code;size:32;uniqueIndex;not null" json:"grade_code"`
	GradeName   string    `gorm:"column:grade_name;size:64;not null" json:"grade_name"`
	LevelOrder  int8      `gorm:"column:level_order;not null;index" json:"level_order"`
	Description string    `gorm:"column:description;size:512" json:"description"`
	StandardRef string    `gorm:"column:standard_ref;size:256" json:"standard_ref"`
	Enable      int8      `gorm:"column:enable;default:1" json:"enable"`
	GmtCreated  time.Time `gorm:"column:gmt_created" json:"gmt_created"`
	GmtUpdated  time.Time `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (DataSecurityGrade) TableName() string {
	return "data_security_grade"
}

// DataAssetSecurityGrade 数据资产安全分级标注（库/表/列；column_name 为空串表示整表默认）
type DataAssetSecurityGrade struct {
	Id           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	DatasourceId int       `gorm:"column:datasource_id;not null;index:idx_ds_db_table;uniqueIndex:uk_asset_scope" json:"datasource_id"`
	DatabaseName string    `gorm:"column:database_name;size:255;default:'';index:idx_ds_db_table;uniqueIndex:uk_asset_scope" json:"database_name"`
	TableNameX   string    `gorm:"column:table_name;size:255;not null;index:idx_ds_db_table;uniqueIndex:uk_asset_scope" json:"table_name"`
	ColumnName   string    `gorm:"column:column_name;size:255;default:'';uniqueIndex:uk_asset_scope" json:"column_name"`
	GradeId      int64     `gorm:"column:grade_id;not null;index" json:"grade_id"`
	AssignSource string    `gorm:"column:assign_source;size:32;default:manual" json:"assign_source"`
	Confidence   *int8     `gorm:"column:confidence" json:"confidence"`
	Remark       string    `gorm:"column:remark;size:512" json:"remark"`
	CreatedBy    string    `gorm:"column:created_by;size:64" json:"created_by"`
	UpdatedBy    string    `gorm:"column:updated_by;size:64" json:"updated_by"`
	GmtCreated   time.Time `gorm:"column:gmt_created" json:"gmt_created"`
	GmtUpdated   time.Time `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (DataAssetSecurityGrade) TableName() string {
	return "data_asset_security_grade"
}

// DataAssetSecurityGradeLog 数据资产分级变更历史
type DataAssetSecurityGradeLog struct {
	Id         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AssetId    int64     `gorm:"column:asset_id;not null;index" json:"asset_id"`
	GradeIdOld *int64    `gorm:"column:grade_id_old" json:"grade_id_old"`
	GradeIdNew int64     `gorm:"column:grade_id_new;not null" json:"grade_id_new"`
	Action     string    `gorm:"column:action;size:32;not null" json:"action"`
	Reason     string    `gorm:"column:reason;size:512" json:"reason"`
	Operator   string    `gorm:"column:operator;size:64" json:"operator"`
	GmtCreated time.Time `gorm:"column:gmt_created" json:"gmt_created"`
}

func (DataAssetSecurityGradeLog) TableName() string {
	return "data_asset_security_grade_log"
}
