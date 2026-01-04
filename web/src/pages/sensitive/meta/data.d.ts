export interface TableListItem {
  id: number;
  datasource_type: string;
  host: string;
  port: string;
  database_name: string;
  table_name: string;
  table_comment: string;
  column_name: string;
  column_comment: string;
  rule_type: string;
  rule_key: string;
  rule_name: string;
  sensitive_count: number;
  simple_count: number;
  level: number;
  status: number;
  gmt_created: date;
  gmt_updated: date;
}

export interface TableListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface TableListData {
  list: TableListItem[];
  pagination: Partial<TableListPagination>;
}

export interface TableListParams {
  id?: number;
  datasource_type?: string;
  host?: string;
  port?: string;
  database_name?: string;
  table_name?: string;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
