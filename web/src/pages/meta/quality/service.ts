import { request } from '@@/plugin-request/request';
import type { QualityParams, QualityData } from './data.d.ts';

export async function queryQualityData(params?: QualityParams) {
  return request<{
    success: boolean;
    data: QualityData;
  }>('/api/v1/meta/quality/info', {
    method: 'GET',
    params,
  });
} 