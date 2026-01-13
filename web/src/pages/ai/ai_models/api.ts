import { request } from 'umi';

/** AI模型配置 */
export interface AIModel {
  id: number;
  name: string;
  provider: string;
  api_url: string;
  api_key?: string;
  model_name: string;
  priority: number;
  enabled: number;
  timeout: number;
  max_tokens: number;
  temperature: number;
  stream_enabled: number;
  description?: string;
  gmt_created: string;
  gmt_updated: string;
}

/** 获取模型列表 */
export async function getModels() {
  return request<{
    success: boolean;
    data: AIModel[];
  }>('/api/v1/ai/models', {
    method: 'GET',
  });
}

/** 获取启用的模型列表 */
export async function getEnabledModels() {
  return request<{
    success: boolean;
    data: AIModel[];
  }>('/api/v1/ai/models/enabled', {
    method: 'GET',
  });
}

/** 创建模型 */
export async function createModel(params: Partial<AIModel>) {
  return request<{
    success: boolean;
    data: AIModel;
    message: string;
  }>('/api/v1/ai/models', {
    method: 'POST',
    data: params,
  });
}

/** 更新模型 */
export async function updateModel(id: number, params: Partial<AIModel>) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/models/${id}`, {
    method: 'PUT',
    data: params,
  });
}

/** 删除模型 */
export async function deleteModel(id: number) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/models/${id}`, {
    method: 'DELETE',
  });
}

/** 测试模型连接 */
export async function testModel(id: number) {
  return request<{
    success: boolean;
    message: string;
    error?: string;
  }>(`/api/v1/ai/models/${id}/test`, {
    method: 'POST',
  });
}

/** 测试模型配置（不需要id，用于创建前测试） */
export async function testModelConfig(params: Partial<AIModel>) {
  return request<{
    success: boolean;
    message: string;
    error?: string;
  }>('/api/v1/ai/model/test-config', {
    method: 'POST',
    data: params,
  });
}

/** 启用/禁用模型 */
export async function toggleModel(id: number, enabled: number) {
  return request<{
    success: boolean;
    message: string;
  }>(`/api/v1/ai/models/${id}/toggle`, {
    method: 'PUT',
    data: { enabled },
  });
}

