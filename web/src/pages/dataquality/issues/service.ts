import { request } from 'umi';

export interface IssueListItem {
  key: number;
  tableName: string;
  columnName: string;
  issueType: string;
  issueLevel: string;
  issueDesc: string;
  issueCount: number;
  status: number;
  handler?: string;
  handleRemark?: string;
  lastCheckTime: string;
}

export async function queryIssues(params: any) {
  return request('/api/v1/dataquality/issues', {
    method: 'GET',
    params,
  });
}

export async function updateIssueStatus(data: any) {
  return request('/api/v1/dataquality/issues/status', {
    method: 'PUT',
    data,
  });
}

