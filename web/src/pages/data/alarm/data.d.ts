export interface DataAlarmItem {
  id: number;
  alarm_name: string;
  alarm_description: string;
  datasource_type: string;
  datasource_id: number;
  database_name?: string;
  sql_query: string;
  rule_operator: string; // >, <, =, >=, <=, !=
  rule_value: number;
  email_content?: string;
  email_to: string;
  cron_expression: string;
  status: number; // 0: 禁用, 1: 启用
  last_run_time?: string;
  next_run_time?: string;
  created_at: string;
  updated_at: string;
}

export interface DataAlarmParams {
  id?: number;
  alarm_name?: string;
  datasource_type?: string;
  status?: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}

export interface DataAlarmFormData {
  alarm_name: string;
  alarm_description: string;
  datasource_type: string;
  datasource_id: number;
  database_name?: string;
  sql_query: string;
  rule_operator: string;
  rule_value: number;
  email_content?: string;
  email_to: string;
  cron_expression: string;
  status: number;
}

export interface DataAlarmLogItem {
  id: number;
  alarm_id: number;
  alarm_name: string;
  start_time: string;
  complete_time?: string;
  status: string; // running, success, failed, triggered
  data_count: number;
  rule_matched: boolean;
  email_sent: boolean;
  error_message?: string;
  created_at: string;
}

export interface DataAlarmLogParams {
  alarm_id?: number;
  status?: string;
  start_date?: string;
  end_date?: string;
  pageSize?: number;
  currentPage?: number;
}

