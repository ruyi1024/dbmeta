export interface TableListItem {
  id: number;
  api_name: string;
  api_url: string;
  api_description?: string;
  protocol: string;
  method: string;
  headers?: string;
  params?: string;
  body?: string;
  token?: string;
  auth_type: string;
  expected_codes: string;
  timeout: number;
  retry_count: number;
  enable: number;
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
  api_name?: string;
  api_url?: string;
  api_description?: string;
  protocol?: string;
  method?: string;
  headers?: string;
  params?: string;
  body?: string;
  token?: string;
  auth_type?: string;
  expected_codes?: string;
  timeout?: number;
  retry_count?: number;
  enable?: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
