export interface TaskListItem {
  id: number;
  taskName: string;
  taskType: string;
  datasourceId?: number;
  databaseName: string;
  tableFilter?: string;
  scheduleConfig?: string;
  status: string;
  startTime?: string;
  endTime?: string;
  duration?: number;
  resultSummary?: string;
  errorMessage?: string;
  createdBy?: string;
  createdAt: string;
  updatedAt: string;
}

