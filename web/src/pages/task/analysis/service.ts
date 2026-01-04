import { request } from '@@/plugin-request/request';
import type { 
  AnalysisTaskParams, 
  AnalysisTaskFormData, 
  AnalysisTaskLogParams 
} from './data.d';

// 获取分析任务列表
export async function queryAnalysisTasks(params?: AnalysisTaskParams) {
  return request('/api/v1/task/analysis/list', {
    method: 'GET',
    params,
  });
}

// 创建分析任务
export async function createAnalysisTask(data: AnalysisTaskFormData) {
  return request('/api/v1/task/analysis/create', {
    method: 'POST',
    data,
  });
}

// 更新分析任务
export async function updateAnalysisTask(data: AnalysisTaskFormData & { id: number }) {
  return request('/api/v1/task/analysis/update', {
    method: 'PUT',
    data,
  });
}

// 删除分析任务
export async function deleteAnalysisTask(id: number) {
  return request(`/api/v1/task/analysis/delete/${id}`, {
    method: 'DELETE',
  });
}

// 启用/禁用分析任务
export async function toggleAnalysisTaskStatus(id: number, status: number) {
  return request('/api/v1/task/analysis/toggle-status', {
    method: 'PUT',
    data: { id, status },
  });
}

// 手动执行分析任务
export async function executeAnalysisTask(id: number) {
  return request('/api/v1/task/analysis/execute', {
    method: 'POST',
    data: { id },
  });
}

// 获取分析任务执行日志
export async function queryAnalysisTaskLogs(params?: AnalysisTaskLogParams) {
  return request('/api/v1/task/analysis/logs', {
    method: 'GET',
    params,
  });
}

// 获取分析任务详情
export async function getAnalysisTaskDetail(id: number) {
  return request(`/api/v1/task/analysis/detail/${id}`, {
    method: 'GET',
  });
}

// 测试SQL查询
export async function testSqlQuery(sql: string) {
  return request('/api/v1/task/analysis/test-sql', {
    method: 'POST',
    data: { sql },
  });
}

// 测试Dify连接
export async function testDifyConnection() {
  return request('/api/v1/task/analysis/test-dify', {
    method: 'POST',
  });
}

// 获取数据源类型列表
export async function getDatasourceTypeList() {
  return request('/api/v1/task/analysis/datasource-type', {
    method: 'GET',
  });
}

// 获取数据源列表
export async function getDatasourceList(params?: { type?: string; env?: string }) {
  return request('/api/v1/task/analysis/datasource', {
    method: 'GET',
    params,
  });
}

// 获取AI模型列表（启用的模型）
export async function getEnabledAIModels() {
  return request('/api/v1/ai/models/enabled', {
    method: 'GET',
  });
} 