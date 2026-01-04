export interface TableListItem {
  id: number;
  rule_type: string;
  rule_key: string;
  rule_name: string;
  rule_express: string;
  rule_pct: number;
  level: number;
  status: number;
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
  rule_type: string;
  rule_key: string;
  rule_name: string;
  rule_express: string;
  rule_pct: number;
  level: number;
  status: number;
  enable: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
