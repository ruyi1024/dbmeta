import { request } from '@@/plugin-request/request';
import type { 
  DataAlarmParams, 
  DataAlarmFormData, 
  DataAlarmLogParams 
} from './data.d';

// 获取数据告警列表
export async function queryDataAlarms(params?: DataAlarmParams) {
  return request('/api/v1/data/alarm/list', {
    method: 'GET',
    params,
  });
}

// 创建数据告警
export async function createDataAlarm(data: DataAlarmFormData) {
  return request('/api/v1/data/alarm/create', {
    method: 'POST',
    data,
  });
}

// 更新数据告警
export async function updateDataAlarm(data: DataAlarmFormData & { id: number }) {
  return request('/api/v1/data/alarm/update', {
    method: 'PUT',
    data,
  });
}

// 删除数据告警
export async function deleteDataAlarm(id: number) {
  return request(`/api/v1/data/alarm/delete/${id}`, {
    method: 'DELETE',
  });
}

// 启用/禁用数据告警
export async function toggleDataAlarmStatus(id: number, status: number) {
  return request('/api/v1/data/alarm/toggle-status', {
    method: 'PUT',
    data: { id, status },
  });
}

// 手动执行数据告警
export async function executeDataAlarm(id: number) {
  return request('/api/v1/data/alarm/execute', {
    method: 'POST',
    data: { id },
  });
}

// 获取数据告警执行日志
export async function queryDataAlarmLogs(params?: DataAlarmLogParams) {
  return request('/api/v1/data/alarm/logs', {
    method: 'GET',
    params,
  });
}

// 获取数据告警详情
export async function getDataAlarmDetail(id: number) {
  return request(`/api/v1/data/alarm/detail/${id}`, {
    method: 'GET',
  });
}

// 测试SQL查询
export async function testSqlQuery(sql: string, datasource_type: string, datasource_id: number, database_name?: string) {
  return request('/api/v1/data/alarm/test-sql', {
    method: 'POST',
    data: { sql, datasource_type, datasource_id, database_name },
  });
}

// 获取数据源类型列表
export async function getDatasourceTypeList() {
  return request('/api/v1/data/alarm/datasource-type', {
    method: 'GET',
  });
}

// 获取数据源列表
export async function getDatasourceList(params?: { type?: string; env?: string }) {
  return request('/api/v1/data/alarm/datasource', {
    method: 'GET',
    params,
  });
}

// 获取数据库列表
export async function getDatabaseList(datasourceId: number) {
  return request('/api/v1/data/alarm/database', {
    method: 'GET',
    params: { datasource_id: datasourceId },
  });
}

