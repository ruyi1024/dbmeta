export interface AnalysisTaskItem {
  id: number;
  task_name: string;
  task_description: string;
  sql_queries: string[];
  prompt: string;
  cron_expression: string;
  report_email: string;
  status: number; // 0: 禁用, 1: 启用
  last_run_time?: string;
  next_run_time?: string;
  created_at: string;
  updated_at: string;
}

export interface AnalysisTaskParams {
  id?: number;
  task_name?: string;
  task_description?: string;
  sql_queries?: string[];
  prompt?: string;
  cron_expression?: string;
  report_email?: string;
  status?: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}

export interface AnalysisTaskFormData {
  task_name: string;
  task_description: string;
  sql_queries: string[];
  prompt: string;
  cron_expression: string;
  report_email: string;
  status: number;
}

export interface AnalysisTaskLogItem {
  id: number;
  task_id: number;
  task_name: string;
  start_time: string;
  complete_time?: string;
  status: string; // running, success, failed
  result: string;
  data_count: number;
  report_content?: string;
  error_message?: string;
  created_at: string;
}

export interface AnalysisTaskLogParams {
  task_id?: number;
  status?: string;
  start_date?: string;
  end_date?: string;
  pageSize?: number;
  currentPage?: number;
} 