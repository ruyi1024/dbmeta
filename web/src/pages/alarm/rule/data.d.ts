export interface TableListItem {
  id: number;
  title: string;
  event_type: string;
  event_group: string;
  event_key: string;
  event_entity: string;
  alarm_rule: string;
  alarm_value: string;
  alarm_sleep: number;
  alarm_times: number;
  channel_id: number;
  level_id: number;
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
  title?: string;
  event_type?: string;
  event_group?: string;
  event_key: string;
  event_entity: string;
  level_id?: number;
  channel_id?: number;
  enable: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
