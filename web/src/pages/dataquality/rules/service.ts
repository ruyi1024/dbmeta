import { request } from 'umi';

export interface RuleListItem {
  id: number;
  ruleName: string;
  ruleType: string;
  ruleDesc: string;
  ruleConfig: string;
  threshold: number;
  severity: string;
  enabled: number;
  createdBy?: string;
  createdAt: string;
  updatedAt: string;
}

export async function queryRules(params: any) {
  return request('/api/v1/dataquality/rules', {
    method: 'GET',
    params,
  });
}

export async function createRule(data: RuleListItem) {
  return request('/api/v1/dataquality/rules', {
    method: 'POST',
    data,
  });
}

export async function updateRule(data: RuleListItem) {
  return request('/api/v1/dataquality/rules', {
    method: 'PUT',
    data,
  });
}

export async function deleteRule(id: number) {
  return request(`/api/v1/dataquality/rules/${id}`, {
    method: 'DELETE',
  });
}

