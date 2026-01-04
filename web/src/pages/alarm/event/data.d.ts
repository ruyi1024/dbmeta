export interface TableListItem {
  id: number;
  alarm_title: string;
  event_uuid: string;
  event_type: string;
  event_group: string;
  event_key: string;
  event_value: string;
  event_entity: string;
  event_unit: string;
  alarm_rule: string;
  alarm_value: string;
  alarm_level: string;
  alarm_sleep: number;
  alarm_times: number;
  channel_id: number;
  gmt_created: date;
  gmt_updated: date;
  status: number;
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
  event_type?: string;
  event_group?: string;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}


import { DataItem } from '@antv/g2plot/esm/interface/config';

export { DataItem };

export interface VisitDataType {
  x: string;
  y: number;
}

export type SearchDataType = {
  index: number;
  keyword: string;
  count: number;
  range: number;
  status: number;
};

export type OfflineDataType = {
  name: string;
  cvr: number;
};

export interface OfflineChartData {
  date: number;
  type: number;
  value: number;
}

export type RadarData = {
  name: string;
  label: string;
  value: number;
};

export interface AnalysisData {
  visitData: DataItem[];
  visitData2: DataItem[];
  salesData: DataItem[];
  searchData: DataItem[];
  offlineData: OfflineDataType[];
  offlineChartData: DataItem[];
  salesTypeData: DataItem[];
  salesTypeDataOnline: DataItem[];
  salesTypeDataOffline: DataItem[];
  radarData: RadarData[];
}
