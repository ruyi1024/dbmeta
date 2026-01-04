import { request } from "@@/plugin-request/request";
import { stringify } from 'qs';

export async function queryStatus(params?: string) {
  return request(`/api/v1/monitor/mysql/status?${stringify(params)}`);
}

