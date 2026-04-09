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
	"encoding/json"
	"time"
)

// AIModel AI模型配置
type AIModel struct {
	Id            int       `gorm:"primarykey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`                    // 模型名称
	Provider      string    `gorm:"size:50;not null;index" json:"provider"`           // 提供商类型
	ApiUrl        string    `gorm:"size:500;not null" json:"api_url"`                 // API地址
	ApiKey        string    `gorm:"size:500" json:"api_key"`                          // API密钥（加密存储）
	ModelName     string    `gorm:"size:100;not null" json:"model_name"`              // 模型标识
	Priority      int       `gorm:"default:0;index" json:"priority"`                  // 优先级
	Enabled       int8      `gorm:"default:0;index" json:"enabled"`                   // 是否启用
	Timeout       int       `gorm:"default:30" json:"timeout"`                        // 超时时间（秒）
	MaxTokens     int       `gorm:"default:2000" json:"max_tokens"`                   // 最大token数
	Temperature   float64   `gorm:"type:decimal(3,2);default:0.7" json:"temperature"` // 温度参数
	StreamEnabled int8      `gorm:"default:0" json:"stream_enabled"`                  // 是否支持流式响应
	Description   string    `gorm:"type:text" json:"description"`                     // 模型描述
	CreatedAt     time.Time `gorm:"column:gmt_created;index" json:"gmt_created"`
	UpdatedAt     time.Time `gorm:"column:gmt_updated" json:"gmt_updated"`
}

func (AIModel) TableName() string {
	return "ai_models"
}

// UnmarshalJSON 自定义 JSON 反序列化，支持 enabled 和 stream_enabled 字段的 bool 和 int8 类型
func (m *AIModel) UnmarshalJSON(data []byte) error {
	// 使用临时结构体来处理灵活的字段类型
	type Alias AIModel
	aux := &struct {
		Enabled       interface{} `json:"enabled,omitempty"`
		StreamEnabled interface{} `json:"stream_enabled,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理 enabled 字段：支持 bool、int8 和数字类型
	if aux.Enabled != nil {
		switch v := aux.Enabled.(type) {
		case bool:
			if v {
				m.Enabled = 1
			} else {
				m.Enabled = 0
			}
		case float64:
			m.Enabled = int8(v)
		case int:
			m.Enabled = int8(v)
		case int8:
			m.Enabled = v
		case int32:
			m.Enabled = int8(v)
		case int64:
			m.Enabled = int8(v)
		}
	}

	// 处理 stream_enabled 字段：支持 bool、int8 和数字类型
	if aux.StreamEnabled != nil {
		switch v := aux.StreamEnabled.(type) {
		case bool:
			if v {
				m.StreamEnabled = 1
			} else {
				m.StreamEnabled = 0
			}
		case float64:
			m.StreamEnabled = int8(v)
		case int:
			m.StreamEnabled = int8(v)
		case int8:
			m.StreamEnabled = v
		case int32:
			m.StreamEnabled = int8(v)
		case int64:
			m.StreamEnabled = int8(v)
		}
	}

	return nil
}

// Provider types
const (
	ProviderOllama    = "ollama"
	ProviderLMStudio  = "lm_studio"
	ProviderVLLM      = "vllm"
	ProviderDifyLocal = "dify_local"
	ProviderOpenAI    = "openai"
	ProviderDeepSeek  = "deepseek"
	ProviderQwen      = "qwen"
)
