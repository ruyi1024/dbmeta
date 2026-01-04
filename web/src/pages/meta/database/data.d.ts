export interface TableListItem {
  id: number;
  datasource_type: string;
  host: string;
  port: string;
  database_name: string;
  alias_name?: string;
  characters: string;
  app_name: string;
  app_desc: string;
  app_owner: string;
  app_owner_email: string;
  app_owner_phone: string;
  is_deleted: number;
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
  alias_name?: string;
  characters?: string;
  app_name?: string;
  app_desc?: string;
  app_owner?: string;
  app_owner_email?: string;
  app_owner_phone?: string;
  is_deleted?: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
