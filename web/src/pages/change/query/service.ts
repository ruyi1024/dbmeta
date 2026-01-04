import { request } from 'umi';
import type { 
  ReleaseItem, 
  OperationChangeItem, 
  AutoChangeItem, 
  ChangeQueryParams 
} from './data';

// 发布清单查询
export async function queryReleaseList(params?: ChangeQueryParams) {
  return request<{
    data: ReleaseItem[];
    total: number;
    success: boolean;
  }>('/api/v1/change/release/list', {
    method: 'GET',
    params,
  });
}

// 运维变更查询
export async function queryOperationChangeList(params?: ChangeQueryParams) {
  return request<{
    data: OperationChangeItem[];
    total: number;
    success: boolean;
  }>('/api/v1/change/operation/list', {
    method: 'GET',
    params,
  });
}

// 自动化变更查询
export async function queryAutoChangeList(params?: ChangeQueryParams) {
  return request<{
    data: AutoChangeItem[];
    total: number;
    success: boolean;
  }>('/api/v1/change/auto/list', {
    method: 'GET',
    params,
  });
}

// 获取发布清单详情
export async function getReleaseDetail(id: string) {
  return request<{
    data: ReleaseItem;
    success: boolean;
  }>(`/api/v1/change/release/${id}`, {
    method: 'GET',
  });
}

// 获取运维变更详情
export async function getOperationChangeDetail(id: string) {
  return request<{
    data: OperationChangeItem;
    success: boolean;
  }>(`/api/v1/change/operation/${id}`, {
    method: 'GET',
  });
}

// 获取自动化变更详情
export async function getAutoChangeDetail(id: string) {
  return request<{
    data: AutoChangeItem;
    success: boolean;
  }>(`/api/v1/change/auto/${id}`, {
    method: 'GET',
  });
}

// 导出发布清单
export async function exportReleaseList(params?: ChangeQueryParams) {
  return request('/api/v1/change/release/export', {
    method: 'GET',
    params,
    responseType: 'blob',
  });
}

// 导出运维变更
export async function exportOperationChangeList(params?: ChangeQueryParams) {
  return request('/api/v1/change/operation/export', {
    method: 'GET',
    params,
    responseType: 'blob',
  });
}

// 导出自动化变更
export async function exportAutoChangeList(params?: ChangeQueryParams) {
  return request('/api/v1/change/auto/export', {
    method: 'GET',
    params,
    responseType: 'blob',
  });
} 