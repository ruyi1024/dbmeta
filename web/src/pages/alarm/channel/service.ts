import {request} from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function queryChannel(params?: TableListParams) {
  return request('/api/v1/alarm/channel', {
    params,
  });
}

export async function removeChannel(params: { key: number[] }) {
  return request('/api/v1/alarm/channel', {
    method: 'DELETE',
    data: {
      ...params,
    },
  });
}

export async function addChannel(params: TableListParams) {
  return request('/api/v1/alarm/channel', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateChannel(params: TableListParams) {
  return request('/api/v1/alarm/channel', {
    method: 'PUT',
    data: {
      ...params,
    },
  });
}
