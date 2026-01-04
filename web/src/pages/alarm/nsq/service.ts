import { request } from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function queryLevel(params?: TableListParams) {
  return request('/api/v1/alarm/level', {
    params,
  });
}

export async function removeLevel(params: { id: number }) {
  return request('/api/v1/alarm/level', {
    method: 'DELETE',
    data: {
      ...params,
    },
  });
}

export async function addLevel(params: TableListParams) {
  return request('/api/v1/alarm/level', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateLevel(params: TableListParams) {
  return request('/api/v1/alarm/level', {
    method: 'PUT',
    data: {
      ...params,
    },
  });
}
