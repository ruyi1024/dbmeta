import { request } from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function query(params?: TableListParams) {
  return request('/api/v1/setting/api/list', {
    params,
  });
}

export async function remove(params: { key: number[] }) {
  return request('/api/v1/setting/api/list', {
    method: 'DELETE',
    data: {
      ...params,
    },
  });
}

export async function add(params: TableListParams) {
  return request('/api/v1/setting/api/list', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function update(params: TableListParams) {
  return request('/api/v1/setting/api/list', {
    method: 'PUT',
    data: {
      ...params,
    },
  });
}
