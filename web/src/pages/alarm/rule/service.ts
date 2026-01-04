import request from 'umi-request';
import { TableListParams } from './data.d';

export async function queryRule(params?: TableListParams) {
  return request('/api/v1/alarm/rule', {
    params,
  });
}

export async function removeRule(params: { key: number[] }) {
  return request('/api/v1/alarm/rule', {
    method: 'DELETE',
    data: {
      ...params,
    },
  });
}

export async function addRule(params: TableListParams) {
  return request('/api/v1/alarm/rule', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateRule(params: TableListParams) {
  return request('/api/v1/alarm/rule', {
    method: 'PUT',
    data: {
      ...params,
    },
  });
}
