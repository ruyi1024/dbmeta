import { request } from 'umi';

/** 智能查数请求 */
export interface DbQueryRequest {
  question: string;
  datasource_id?: number;
  database_name?: string;
  datasource_type?: string;
  host?: string;
  port?: string;
  table_name?: string;
  page?: number;
  page_size?: number;
}

/** 智能查数响应 */
export interface DbQueryResponse {
  success: boolean;
  message?: string;
  data?: {
    sql_query: string;
    query_result: Array<Record<string, any>>;
    total: number;
    page: number;
    page_size: number;
  };
}

/** 数据库信息 */
export interface DatabaseInfo {
  id: number;
  database_name: string;
  alias_name?: string;
  datasource_type?: string;
  host?: string;
  port?: string;
}

/** 数据库列表响应 */
export interface DatabaseListResponse {
  success: boolean;
  msg?: string;
  data?: DatabaseInfo[];
  total?: number;
}

/** 执行智能查数 */
export async function queryDatabase(params: DbQueryRequest) {
  return request<DbQueryResponse>('/api/v1/ai/dbquery', {
    method: 'POST',
    data: params,
  });
}

/** 获取数据库列表 */
export async function getDatabaseList() {
  return request<DatabaseListResponse>('/api/v1/meta/database/list', {
    method: 'GET',
    params: {
      is_deleted: 0,
    },
  });
}

