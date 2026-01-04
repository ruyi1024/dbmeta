export interface TableListItem {
  id: number;
  datasource_type: string;
  host: string;
  port: string;
  database_name: string;
  table_name: string;
  column_name: string;
  column_comment: string;
  ai_comment: string;
  ai_fixed: number;
  data_type: string;
  is_nullable: string;
  default_value: string;
  ordinal_position: number;
  characters: string;
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
  column_name?: string;
  column_comment?: string;
  ai_comment?: string;
  ai_fixed?: number;
  data_type?: string;
  is_nullable?: string;
  default_value?: string;
  ordinal_position?: number;
  characters?: string;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
