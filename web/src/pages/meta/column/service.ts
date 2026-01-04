import { request } from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function queryColumn(params?: TableListParams) {
  return request('/api/v1/meta/column/list', {
    params,
  });
}

export async function batchUpdateAiFixed(params: { ids: number[]; ai_fixed: number }) {
  return request('/api/v1/meta/column/batch-update-ai-fixed', {
    method: 'PUT',
    data: params,
  });
}

