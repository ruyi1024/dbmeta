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
	"bufio"
	"bytes"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message AI消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatOptions 聊天选项
type ChatOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content string `json:"content"`
	Model   string `json:"model,omitempty"`
	Usage   *Usage `json:"usage,omitempty"`
}

// Usage Token使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 流式响应块
type StreamChunk struct {
	Content string `json:"content"`
	Think   string `json:"think,omitempty"` // 思考/推理内容
	Done    bool   `json:"done"`
	Error   error  `json:"error,omitempty"`
}

// AIClient AI客户端接口
type AIClient interface {
	// Chat 非流式聊天
	Chat(messages []Message, options *ChatOptions) (*ChatResponse, error)
	// ChatStream 流式聊天
	ChatStream(messages []Message, options *ChatOptions) (<-chan *StreamChunk, error)
	// TestConnection 测试连接
	TestConnection() error
}

// BaseClient 基础客户端
type BaseClient struct {
	Model   *model.AIModel
	ApiKey  string
	Timeout time.Duration
}

// normalizeApiUrl 根据提供商类型规范化API URL
func normalizeApiUrl(provider, apiUrl string) string {
	// 如果URL已经包含路径，直接返回
	if strings.Contains(apiUrl, "/api/") || strings.Contains(apiUrl, "/v1/") {
		return apiUrl
	}

	// 移除末尾的斜杠
	apiUrl = strings.TrimSuffix(apiUrl, "/")

	// 根据提供商类型添加默认端点
	switch provider {
	case model.ProviderOllama:
		// Ollama 使用 /v1/chat/completions 端点（OpenAI兼容格式）
		return apiUrl + "/v1/chat/completions"
	case model.ProviderLMStudio:
		// LM Studio 使用 /v1/chat/completions
		return apiUrl + "/v1/chat/completions"
	case model.ProviderVLLM:
		// vLLM 使用 /v1/chat/completions
		return apiUrl + "/v1/chat/completions"
	case model.ProviderDifyLocal:
		// Dify本地部署可能使用不同的端点，保持原样或添加默认路径
		return apiUrl
	case model.ProviderOpenAI:
		// OpenAI 使用 /v1/chat/completions
		if !strings.Contains(apiUrl, "/v1/chat/completions") {
			return apiUrl + "/v1/chat/completions"
		}
		return apiUrl
	case model.ProviderDeepSeek:
		// DeepSeek 使用 /v1/chat/completions
		if !strings.Contains(apiUrl, "/v1/chat/completions") {
			return apiUrl + "/v1/chat/completions"
		}
		return apiUrl
	case model.ProviderQwen:
		// Qwen 使用 /v1/chat/completions
		if !strings.Contains(apiUrl, "/v1/chat/completions") {
			return apiUrl + "/v1/chat/completions"
		}
		return apiUrl
	default:
		return apiUrl
	}
}

// NewAIClient 根据模型配置创建对应的客户端
func NewAIClient(aiModel *model.AIModel) (AIClient, error) {
	// 获取解密后的API密钥
	apiKey, err := GetDecryptedApiKey(aiModel)
	if err != nil {
		return nil, fmt.Errorf("获取API密钥失败: %v", err)
	}

	// 规范化API URL
	normalizedUrl := normalizeApiUrl(aiModel.Provider, aiModel.ApiUrl)

	// 创建模型副本并更新URL
	modelCopy := *aiModel
	modelCopy.ApiUrl = normalizedUrl

	baseClient := &BaseClient{
		Model:   &modelCopy,
		ApiKey:  apiKey,
		Timeout: time.Duration(aiModel.Timeout) * time.Second,
	}

	switch aiModel.Provider {
	case model.ProviderOllama, model.ProviderLMStudio, model.ProviderVLLM, model.ProviderDifyLocal:
		return &OpenAICompatibleClient{BaseClient: baseClient}, nil
	case model.ProviderOpenAI:
		return &OpenAIClient{BaseClient: baseClient}, nil
	case model.ProviderDeepSeek:
		return &DeepSeekClient{BaseClient: baseClient}, nil
	case model.ProviderQwen:
		return &QwenClient{BaseClient: baseClient}, nil
	default:
		return nil, fmt.Errorf("不支持的提供商类型: %s", aiModel.Provider)
	}
}

// OpenAICompatibleClient OpenAI兼容接口客户端（Ollama, LM Studio, vLLM, Dify本地）
type OpenAICompatibleClient struct {
	*BaseClient
}

// Chat 实现非流式聊天
func (c *OpenAICompatibleClient) Chat(messages []Message, options *ChatOptions) (*ChatResponse, error) {
	requestBody := map[string]interface{}{
		"model":       c.Model.ModelName,
		"messages":    messages,
		"stream":      false,
		"temperature": c.Model.Temperature,
		"max_tokens":  c.Model.MaxTokens,
	}

	if options != nil {
		if options.Temperature > 0 {
			requestBody["temperature"] = options.Temperature
		}
		if options.MaxTokens > 0 {
			requestBody["max_tokens"] = options.MaxTokens
		}
	}

	return c.callOpenAICompatibleAPI(requestBody)
}

// ChatStream 实现流式聊天
func (c *OpenAICompatibleClient) ChatStream(messages []Message, options *ChatOptions) (<-chan *StreamChunk, error) {
	requestBody := map[string]interface{}{
		"model":       c.Model.ModelName,
		"messages":    messages,
		"stream":      true,
		"temperature": c.Model.Temperature,
		"max_tokens":  c.Model.MaxTokens,
	}

	if options != nil {
		if options.Temperature > 0 {
			requestBody["temperature"] = options.Temperature
		}
		if options.MaxTokens > 0 {
			requestBody["max_tokens"] = options.MaxTokens
		}
	}

	return c.callOpenAICompatibleStreamAPI(requestBody)
}

// TestConnection 测试连接
func (c *OpenAICompatibleClient) TestConnection() error {
	testMessages := []Message{
		{Role: "user", Content: "Hello"},
	}
	_, err := c.Chat(testMessages, nil)
	if err != nil {
		return fmt.Errorf("测试AI模型连接失败 (URL: %s, Model: %s): %v", c.Model.ApiUrl, c.Model.ModelName, err)
	}
	return nil
}

// callOpenAICompatibleAPI 调用OpenAI兼容API
func (c *OpenAICompatibleClient) callOpenAICompatibleAPI(requestBody map[string]interface{}) (*ChatResponse, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	req, err := http.NewRequest("POST", c.Model.ApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	}

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI模型API请求失败 (URL: %s)，状态码: %d, 响应: %s", c.Model.ApiUrl, resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
		Usage *Usage `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("API响应中没有有效的回答")
	}

	return &ChatResponse{
		Content: apiResp.Choices[0].Message.Content,
		Model:   apiResp.Model,
		Usage:   apiResp.Usage,
	}, nil
}

// callOpenAICompatibleStreamAPI 调用OpenAI兼容流式API
func (c *OpenAICompatibleClient) callOpenAICompatibleStreamAPI(requestBody map[string]interface{}) (<-chan *StreamChunk, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	req, err := http.NewRequest("POST", c.Model.ApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	}

	// 对于流式响应，使用无超时的客户端（或非常长的超时时间）
	// 因为流式响应可能需要很长时间才能完成
	streamTimeout := c.Timeout
	if streamTimeout < 5*time.Minute {
		// 如果超时时间小于5分钟，设置为5分钟（流式响应通常需要更长时间）
		streamTimeout = 5 * time.Minute
	}
	client := &http.Client{Timeout: streamTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	ch := make(chan *StreamChunk, 10)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		// 增加缓冲区大小，避免读取大块数据时的问题
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024) // 1MB 缓冲区

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// OpenAI流式响应格式: data: {"choices":[...]}
			// 或者: data: [DONE]
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					ch <- &StreamChunk{Done: true}
					return
				}

				var streamResp struct {
					Choices []struct {
						Delta struct {
							Content   string `json:"content"`
							Reasoning string `json:"reasoning,omitempty"` // DeepSeek R1等模型的思考内容
						} `json:"delta"`
						FinishReason string `json:"finish_reason"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
					// 忽略解析错误，继续处理下一行
					continue
				}

				if len(streamResp.Choices) > 0 {
					choice := streamResp.Choices[0]
					if choice.FinishReason != "" {
						ch <- &StreamChunk{Done: true}
						return
					}
					// 优先发送思考内容
					if choice.Delta.Reasoning != "" {
						ch <- &StreamChunk{Think: choice.Delta.Reasoning}
					}
					// 发送回答内容
					if choice.Delta.Content != "" {
						ch <- &StreamChunk{Content: choice.Delta.Content}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			// 检查是否是超时错误
			errStr := err.Error()
			if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
				// 超时错误，发送完成信号而不是错误（因为可能已经接收了部分数据）
				// 流式响应可能已经接收了部分数据，所以发送完成信号
				ch <- &StreamChunk{Done: true}
			} else {
				ch <- &StreamChunk{Error: fmt.Errorf("读取流式响应失败: %v", err)}
			}
			return
		}

		// 如果没有收到Done信号，发送Done
		ch <- &StreamChunk{Done: true}
	}()

	return ch, nil
}

// OpenAIClient OpenAI官方API客户端
type OpenAIClient struct {
	*BaseClient
}

// Chat 实现非流式聊天
func (c *OpenAIClient) Chat(messages []Message, options *ChatOptions) (*ChatResponse, error) {
	// OpenAI使用OpenAI兼容接口，可以直接复用
	compatibleClient := &OpenAICompatibleClient{BaseClient: c.BaseClient}
	return compatibleClient.Chat(messages, options)
}

// ChatStream 实现流式聊天
func (c *OpenAIClient) ChatStream(messages []Message, options *ChatOptions) (<-chan *StreamChunk, error) {
	compatibleClient := &OpenAICompatibleClient{BaseClient: c.BaseClient}
	return compatibleClient.ChatStream(messages, options)
}

// TestConnection 测试连接
func (c *OpenAIClient) TestConnection() error {
	testMessages := []Message{
		{Role: "user", Content: "Hello"},
	}
	_, err := c.Chat(testMessages, nil)
	if err != nil {
		return fmt.Errorf("测试AI模型连接失败 (URL: %s, Model: %s): %v", c.Model.ApiUrl, c.Model.ModelName, err)
	}
	return nil
}

// DeepSeekClient DeepSeek API客户端
type DeepSeekClient struct {
	*BaseClient
}

// normalizeDeepSeekModelName 规范化DeepSeek模型名称
func normalizeDeepSeekModelName(modelName string) string {
	modelName = strings.ToLower(strings.TrimSpace(modelName))
	
	// 模型名称映射表
	modelMap := map[string]string{
		"deepseek-v3.2":     "deepseek-chat",
		"deepseek-v3":       "deepseek-chat",
		"deepseek-chat-v3":  "deepseek-chat",
		"deepseek-chat-v2":  "deepseek-chat",
		"deepseek-reasoner": "deepseek-reasoner",
		"deepseek-r1":       "deepseek-reasoner",
		"deepseek-coder":    "deepseek-coder",
	}
	
	// 如果找到映射，返回映射后的名称
	if mapped, ok := modelMap[modelName]; ok {
		return mapped
	}
	
	// 如果已经是正确的格式（以 deepseek- 开头），直接返回
	if strings.HasPrefix(modelName, "deepseek-") {
		return modelName
	}
	
	// 默认返回 deepseek-chat
	return "deepseek-chat"
}

// Chat 实现非流式聊天
func (c *DeepSeekClient) Chat(messages []Message, options *ChatOptions) (*ChatResponse, error) {
	// 规范化模型名称
	normalizedModelName := normalizeDeepSeekModelName(c.Model.ModelName)
	
	// 创建模型副本并更新模型名称
	modelCopy := *c.Model
	modelCopy.ModelName = normalizedModelName
	
	// DeepSeek也使用OpenAI兼容接口
	compatibleClient := &OpenAICompatibleClient{
		BaseClient: &BaseClient{
			Model:   &modelCopy,
			ApiKey:  c.ApiKey,
			Timeout: c.Timeout,
		},
	}
	return compatibleClient.Chat(messages, options)
}

// ChatStream 实现流式聊天
func (c *DeepSeekClient) ChatStream(messages []Message, options *ChatOptions) (<-chan *StreamChunk, error) {
	// 规范化模型名称
	normalizedModelName := normalizeDeepSeekModelName(c.Model.ModelName)
	
	// 创建模型副本并更新模型名称
	modelCopy := *c.Model
	modelCopy.ModelName = normalizedModelName
	
	// DeepSeek也使用OpenAI兼容接口
	compatibleClient := &OpenAICompatibleClient{
		BaseClient: &BaseClient{
			Model:   &modelCopy,
			ApiKey:  c.ApiKey,
			Timeout: c.Timeout,
		},
	}
	return compatibleClient.ChatStream(messages, options)
}

// TestConnection 测试连接
func (c *DeepSeekClient) TestConnection() error {
	testMessages := []Message{
		{Role: "user", Content: "Hello"},
	}
	_, err := c.Chat(testMessages, nil)
	if err != nil {
		normalizedModelName := normalizeDeepSeekModelName(c.Model.ModelName)
		return fmt.Errorf("测试AI模型连接失败 (URL: %s, Model: %s -> %s): %v", c.Model.ApiUrl, c.Model.ModelName, normalizedModelName, err)
	}
	return nil
}

// QwenClient Qwen API客户端
type QwenClient struct {
	*BaseClient
}

// Chat 实现非流式聊天
func (c *QwenClient) Chat(messages []Message, options *ChatOptions) (*ChatResponse, error) {
	// Qwen也使用OpenAI兼容接口
	compatibleClient := &OpenAICompatibleClient{BaseClient: c.BaseClient}
	return compatibleClient.Chat(messages, options)
}

// ChatStream 实现流式聊天
func (c *QwenClient) ChatStream(messages []Message, options *ChatOptions) (<-chan *StreamChunk, error) {
	compatibleClient := &OpenAICompatibleClient{BaseClient: c.BaseClient}
	return compatibleClient.ChatStream(messages, options)
}

// TestConnection 测试连接
func (c *QwenClient) TestConnection() error {
	testMessages := []Message{
		{Role: "user", Content: "Hello"},
	}
	_, err := c.Chat(testMessages, nil)
	if err != nil {
		return fmt.Errorf("测试AI模型连接失败 (URL: %s, Model: %s): %v", c.Model.ApiUrl, c.Model.ModelName, err)
	}
	return nil
}
