export interface TableListItem {
  task_key: string;
  task_name: string;
  crontab: string;
  task_description: string;
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
  task_key: string;
  task_name: string;
  enable: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}

// 任务日志相关类型定义
export interface TaskLogItem {
  id: number;
  task_key: string;
  start_time: string;
  complete_time?: string;
  status: string;
  result: string;
  gmt_created: string;
}

export interface TaskLogParams {
  task_key?: string;
  status?: string;
  start_date?: string;
  end_date?: string;
  pageSize?: number;
  currentPage?: number;
  sorter?: { [key: string]: any };
}

export interface TaskLogData {
  list: TaskLogItem[];
  pagination: Partial<TableListPagination>;
}
