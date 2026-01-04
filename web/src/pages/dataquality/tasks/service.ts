import { request } from 'umi';

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

export async function queryTasks(params: any) {
  return request('/api/v1/dataquality/tasks', {
    method: 'GET',
    params,
  });
}

export async function createTask(data: TaskListItem) {
  return request('/api/v1/dataquality/tasks', {
    method: 'POST',
    data,
  });
}

export async function updateTaskStatus(data: { id: number; status: string }) {
  return request('/api/v1/dataquality/tasks/status', {
    method: 'PUT',
    data,
  });
}

export async function deleteTask(id: number) {
  return request(`/api/v1/dataquality/tasks/${id}`, {
    method: 'DELETE',
  });
}

