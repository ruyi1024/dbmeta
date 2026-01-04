import {request} from "@@/plugin-request/request";
import { TableListParams } from './data.d';

export async function queryEvent(params?: TableListParams) {
  return request('/api/v1/alarm/event', {
    params,
  });
}

export async function queryAlarmAnalysis() {
  return request('/api/v1/alarm/event/analysis', {});
}

import type { AnalysisData } from './data';

export async function fakeChartData(): Promise<{ data: AnalysisData }> {
  return request('/api/fake_analysis_chart_data');
}
