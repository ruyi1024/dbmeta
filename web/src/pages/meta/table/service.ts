import { request } from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function queryTable(params?: TableListParams) {
  return request('/api/v1/meta/table/list', {
    params,
  });
}

export async function batchUpdateAiFixed(params: { ids: number[]; ai_fixed: number }) {
  return request('/api/v1/meta/table/batch-update-ai-fixed', {
    method: 'PUT',
    data: params,
  });
}

