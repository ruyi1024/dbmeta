/*
Copyright 2014-2022 The Lepus Team Group, website: https://www.lepus.cc
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
*/

package model

import "time"

// 场景常量：与 ai_model_default.scenario 对应
const (
	AIModelScenarioGrading = "grading" // 数据分级（AI 批处理等）
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
