import {request} from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function query(params?: TableListParams) {
  return request('/api/v1/privilege/list', {
    params,
  });
}

export async function remove(params: { key: number[] }) {
  return request('/api/v1/privilege/list', {
    method: 'DELETE',
    data: {
      ...params,
    },
  });
}
