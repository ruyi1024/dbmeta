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
	"dbmcloud/src/model"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	enabledModelsCache []model.AIModel
	modelsCacheMutex   sync.RWMutex
)

// CallWithFailover 带故障转移的模型调用
func CallWithFailover(messages []Message, options *ChatOptions) (*ChatResponse, error) {
	models, err := GetEnabledModels()
	if err != nil {
		return nil, fmt.Errorf("获取启用的模型失败: %v", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("没有启用的模型")
	}

	var lastErr error
	for _, aiModel := range models {
		client, err := NewAIClient(&aiModel)
		if err != nil {
			zap.L().Warn("创建AI客户端失败", zap.String("model", aiModel.Name), zap.Error(err))
			lastErr = err
			continue
		}

		response, err := client.Chat(messages, options)
		if err != nil {
			zap.L().Warn("模型调用失败", zap.String("model", aiModel.Name), zap.Error(err))
			lastErr = err
			continue
		}

		zap.L().Info("模型调用成功", zap.String("model", aiModel.Name))
		return response, nil
	}

	return nil, fmt.Errorf("所有模型调用失败，最后一个错误: %v", lastErr)
}

// CallWithStream 流式调用（支持故障转移）
func CallWithStream(messages []Message, options *ChatOptions) (<-chan *StreamChunk, error) {
	models, err := GetEnabledModels()
	if err != nil {
		return nil, fmt.Errorf("获取启用的模型失败: %v", err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("没有启用的模型")
	}

	// 对于流式调用，只使用第一个启用的模型
	// 如果需要故障转移，可以在流式响应失败时切换到下一个模型
	aiModel := models[0]
	client, err := NewAIClient(&aiModel)
	if err != nil {
		return nil, fmt.Errorf("创建AI客户端失败: %v", err)
	}

	stream, err := client.ChatStream(messages, options)
	if err != nil {
		// 如果第一个模型失败，尝试下一个
		if len(models) > 1 {
			zap.L().Warn("第一个模型流式调用失败，尝试下一个", zap.String("model", aiModel.Name), zap.Error(err))
			aiModel = models[1]
			client, err = NewAIClient(&aiModel)
			if err != nil {
				return nil, fmt.Errorf("创建备用AI客户端失败: %v", err)
			}
			return client.ChatStream(messages, options)
		}
		return nil, err
	}

	return stream, nil
}

// SelectModel 根据优先级选择模型（返回第一个启用的模型）
func SelectModel() (*model.AIModel, error) {
	models, err := GetEnabledModels()
	if err != nil {
		return nil, err
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("没有启用的模型")
	}

	return &models[0], nil
}

// RefreshModelsCache 刷新模型缓存
func RefreshModelsCache() error {
	models, err := GetEnabledModels()
	if err != nil {
		return err
	}

	modelsCacheMutex.Lock()
	defer modelsCacheMutex.Unlock()
	enabledModelsCache = models
	return nil
}

// GetCachedEnabledModels 获取缓存的启用模型列表
func GetCachedEnabledModels() []model.AIModel {
	modelsCacheMutex.RLock()
	defer modelsCacheMutex.RUnlock()
	return enabledModelsCache
}
