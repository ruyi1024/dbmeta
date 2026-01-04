import { request } from "@@/plugin-request/request";

// 获取数据源列表
export async function getDatasourceList() {
  return request('/api/v1/datasource/list', {
    method: 'GET',
  });
}

// 获取进程列表
export async function getProcessList(params: { datasource_id: number }) {
  return request('/api/v1/monitor/processlist', {
    method: 'POST',
    data: params,
  });
}
