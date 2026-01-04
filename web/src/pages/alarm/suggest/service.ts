import {request} from "@@/plugin-request/request";
import { stringify } from 'qs';

export async function querySuggest(params?: string) {
  return request(`/api/v1/alarm/suggest?${stringify(params)}`);
}

export async function updateSuggest(params: { modify?: boolean; admin?:boolean; createdAt?: Date; password?: string; chineseName?: string; id?: number; username?: string; updatedAt?: Date }) {
  return request(`/api/v1/alarm/suggest`, {
    method: params.modify ? 'PUT' : 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeSuggest(params: { username?: string }) {
  return request('/api/v1/alarm/suggest', {
    method: 'DELETE',
    data: {
      ...params,
    },
  });
}
