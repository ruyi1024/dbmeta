package model

import "time"

// SettingKV 系统配置KV（如 notice 通信配置）。
type SettingKV struct {
	Id          int64     `gorm:"primaryKey" json:"id"`
	Category    string    `gorm:"size:50;index:idx_setting_category_key,priority:1" json:"category"`
	ConfigKey   string    `gorm:"column:config_key;size:100;uniqueIndex" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;type:text" json:"config_value"`
	Remark      string    `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time `gorm:"column:gmt_created" json:"gmt_created"`
	UpdatedAt   time.Time `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (SettingKV) TableName() string {
	return "settings"
}

