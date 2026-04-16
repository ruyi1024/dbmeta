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

import "time"

// 场景常量：与 ai_model_default.scenario 对应
const (
	AIModelScenarioGrading             = "grading"               // 数据分级（AI 批处理等）
	AIModelScenarioTableColumnComment  = "table_column_comment"  // AI 生成表/字段备注（元数据注释任务）
	AIModelScenarioTableColumnAccuracy = "table_column_accuracy" // 表字段与注释准确度评估
)

// AIModelDefault 各场景默认使用的 AI 模型（model_id 指向 ai_models.id）
type AIModelDefault struct {
	Scenario   string    `gorm:"primaryKey;size:64" json:"scenario"`
	ModelId    *int      `gorm:"column:model_id;index" json:"model_id"`
	GmtUpdated time.Time `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (AIModelDefault) TableName() string {
	return "ai_model_default"
}
