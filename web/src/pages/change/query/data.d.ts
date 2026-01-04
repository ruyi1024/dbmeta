// 发布清单数据类型
export interface ReleaseItem {
  id: string;
  appName: string;
  appDescription: string;
  releaseRequirement: string;
  status: 'pending' | 'deploying' | 'success' | 'failed';
  startTime: string;
  endTime?: string;
  publisher: string;
  version: string;
  releaseType: 'normal' | 'urgent';
  hasDbChange: boolean;
}

// 运维变更数据类型
export interface OperationChangeItem {
  id: string;
  changeStartTime: string;
  changeEndTime?: string;
  changeType: 'operation' | 'network' | 'database' | 'bigdata' | 'security';
  changeLevel: 'level1' | 'level2' | 'level3';
  changeName: string;
  changePerson: string;
  currentStatus: 'not_started' | 'in_progress' | 'completed';
}

// 自动化变更数据类型
export interface AutoChangeItem {
  id: string;
  changeNo: string;
  title: string;
  type: 'deploy' | 'config' | 'rollback' | 'maintenance';
  status: 'scheduled' | 'running' | 'success' | 'failed' | 'cancelled';
  executor: string;
  startTime: string;
  endTime?: string;
  duration?: string;
  environment: string;
}

// 查询参数类型
export interface ChangeQueryParams {
  title?: string;
  status?: string;
  type?: string;
  applicant?: string;
  assignee?: string;
  executor?: string;
  priority?: string;
  environment?: string;
  startTime?: string;
  endTime?: string;
  pageSize?: number;
  currentPage?: number;
} 