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

package service

import (
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetEnabledModels 获取所有启用的模型，按优先级排序
func GetEnabledModels() ([]model.AIModel, error) {
	var models []model.AIModel
	result := database.DB.Where("enabled = 1").
		Order("priority DESC, id ASC").
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("查询启用的模型失败: %v", result.Error)
	}
	return models, nil
}

// GetAllModels 获取所有模型
func GetAllModels() ([]model.AIModel, error) {
	var models []model.AIModel
	result := database.DB.Order("priority DESC, id ASC").Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("查询模型列表失败: %v", result.Error)
	}
	return models, nil
}

// GetModelById 根据ID获取模型配置
func GetModelById(id int) (*model.AIModel, error) {
	var aiModel model.AIModel
	result := database.DB.Where("id = ?", id).First(&aiModel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("模型不存在")
		}
		return nil, fmt.Errorf("查询模型失败: %v", result.Error)
	}
	return &aiModel, nil
}

// CreateModel 创建模型配置
func CreateModel(aiModel *model.AIModel) error {
	// 如果提供了API密钥，需要加密存储
	if aiModel.ApiKey != "" {
		encryptedKey, err := utils.AesPassEncode(aiModel.ApiKey, setting.Setting.DbPassKey)
		if err != nil {
			return fmt.Errorf("加密API密钥失败: %v", err)
		}
		aiModel.ApiKey = encryptedKey
	}

	result := database.DB.Create(aiModel)
	if result.Error != nil {
		return fmt.Errorf("创建模型失败: %v", result.Error)
	}
	return nil
}

// UpdateModel 更新模型配置
func UpdateModel(id int, aiModel *model.AIModel) error {
	// 如果提供了新的API密钥，需要加密存储
	if aiModel.ApiKey != "" {
		// 检查是否是已加密的密钥（如果以特定格式开头，可能是已加密的）
		// 这里简单判断：如果长度较短且不包含特殊字符，可能是未加密的
		encryptedKey, err := utils.AesPassEncode(aiModel.ApiKey, setting.Setting.DbPassKey)
		if err != nil {
			return fmt.Errorf("加密API密钥失败: %v", err)
		}
		aiModel.ApiKey = encryptedKey
	} else {
		// 如果没有提供新密钥，保持原有密钥不变
		var existingModel model.AIModel
		if err := database.DB.Where("id = ?", id).First(&existingModel).Error; err == nil {
			aiModel.ApiKey = existingModel.ApiKey
		}
	}

	result := database.DB.Model(&model.AIModel{}).Where("id = ?", id).Updates(aiModel)
	if result.Error != nil {
		return fmt.Errorf("更新模型失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("模型不存在")
	}
	return nil
}

// DeleteModel 删除模型配置
func DeleteModel(id int) error {
	result := database.DB.Delete(&model.AIModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除模型失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("模型不存在")
	}
	return nil
}

// ToggleModel 启用/禁用模型
func ToggleModel(id int, enabled int8) error {
	result := database.DB.Model(&model.AIModel{}).Where("id = ?", id).Update("enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("更新模型状态失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("模型不存在")
	}
	return nil
}

// GetDecryptedApiKey 获取解密后的API密钥
// 如果解密失败，会尝试判断是否为未加密的原始密钥（比如在测试配置时）
func GetDecryptedApiKey(aiModel *model.AIModel) (string, error) {
	if aiModel.ApiKey == "" {
		return "", nil
	}

	// 先检查是否是有效的十六进制字符串（加密后的格式）
	// 加密后的字符串是十六进制编码，只包含 0-9, a-f, A-F
	isHexString := true
	for _, char := range aiModel.ApiKey {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			isHexString = false
			break
		}
	}

	// 如果看起来不像十六进制字符串（比如包含 sk- 前缀或其他字符），可能是未加密的原始密钥
	if !isHexString {
		return aiModel.ApiKey, nil
	}

	// 尝试解密
	decryptedKey, err := utils.AesPassDecode(aiModel.ApiKey, setting.Setting.DbPassKey)
	if err != nil {
		// 解密失败，可能是未加密的原始密钥（虽然看起来像十六进制，但实际不是加密数据）
		// 或者加密格式不正确
		// 对于测试配置场景，如果解密失败，尝试返回原始值
		// 但需要判断：如果长度很短（<32字符），可能是未加密的；如果很长，可能是加密格式错误
		keyLen := len(aiModel.ApiKey)
		if keyLen < 32 {
			// 短字符串，可能是未加密的原始密钥
			return aiModel.ApiKey, nil
		}
		// 长字符串但解密失败，返回错误
		return "", fmt.Errorf("解密API密钥失败: %v", err)
	}
	return decryptedKey, nil
}

// TestModelConnection 测试模型连接
func TestModelConnection(aiModel *model.AIModel) error {
	// 这里先简单实现，后续会在ai_client中实现具体的测试逻辑
	// 暂时只验证配置是否完整
	if aiModel.ApiUrl == "" {
		return fmt.Errorf("API地址不能为空")
	}
	if aiModel.ModelName == "" {
		return fmt.Errorf("模型名称不能为空")
	}

	// 对于需要API密钥的提供商，检查密钥是否存在
	needApiKey := map[string]bool{
		model.ProviderOpenAI:    true,
		model.ProviderDeepSeek:  true,
		model.ProviderQwen:      true,
		model.ProviderDifyLocal: false, // Dify本地可能不需要密钥
		model.ProviderOllama:    false,
		model.ProviderLMStudio:  false,
		model.ProviderVLLM:      false,
	}

	if needApiKey[aiModel.Provider] && aiModel.ApiKey == "" {
		return fmt.Errorf("该提供商需要API密钥")
	}

	zap.L().Info("模型配置验证通过", zap.String("provider", aiModel.Provider), zap.String("model", aiModel.ModelName))
	return nil
}
