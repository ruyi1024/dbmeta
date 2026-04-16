<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';
import type { TableColumnsType } from 'ant-design-vue';

import { computed, nextTick, onMounted, ref } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import {
  Card,
  Col,
  Progress,
  Row,
  Statistic,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

interface PieItem {
  type: string;
  value: number;
}

interface IssueItem {
  columnName?: string;
  issueCount?: number;
  issueDesc?: string;
  issueLevel?: 'high' | 'low' | 'medium' | string;
  issueType?: string;
  key?: number | string;
  lastCheckTime?: string;
  tableName?: string;
}

interface RecommendationItem {
  desc: string;
  priority: string;
  title: string;
  type: 'high' | 'low' | 'medium';
}

interface AiAnalysis {
  analysisTime?: string;
  insights?: string[];
  overallLevel?: string;
  overallScore?: number;
  recommendations?: RecommendationItem[];
  trendAnalysis?: string;
}

interface DashboardData {
  accuracyData?: PieItem[];
  aiAnalysis?: AiAnalysis;
  completenessData?: PieItem[];
  consistencyData?: PieItem[];
  dataConsistency?: number;
  dataTimeliness?: number;
  dataUniqueness?: number;
  fieldAccuracy?: number;
  fieldCompleteness?: number;
  issueList?: IssueItem[];
  tableCompleteness?: number;
  totalColumns?: number;
  totalIssues?: number;
  totalTables?: number;
  uniquenessData?: PieItem[];
}

const mockData: DashboardData = {
  accuracyData: [
    { type: $t('page.qualityDashboard.pie.accurate'), value: 14_470 },
    { type: $t('page.qualityDashboard.pie.formatError'), value: 850 },
    { type: $t('page.qualityDashboard.pie.rangeError'), value: 360 },
  ],
  aiAnalysis: {
    analysisTime: '2024-01-15 10:30:00',
    insights: [
      $t('page.qualityDashboard.ai.insight1'),
      $t('page.qualityDashboard.ai.insight2'),
      $t('page.qualityDashboard.ai.insight3'),
      $t('page.qualityDashboard.ai.insight4'),
    ],
    overallLevel: $t('page.qualityDashboard.ai.overallLevel'),
    overallScore: 88.2,
    recommendations: [
      {
        desc: $t('page.qualityDashboard.ai.recommendation1Desc'),
        priority: $t('page.qualityDashboard.level.high'),
        title: $t('page.qualityDashboard.ai.recommendation1Title'),
        type: 'high',
      },
      {
        desc: $t('page.qualityDashboard.ai.recommendation2Desc'),
        priority: $t('page.qualityDashboard.level.medium'),
        title: $t('page.qualityDashboard.ai.recommendation2Title'),
        type: 'medium',
      },
      {
        desc: $t('page.qualityDashboard.ai.recommendation3Desc'),
        priority: $t('page.qualityDashboard.level.low'),
        title: $t('page.qualityDashboard.ai.recommendation3Title'),
        type: 'low',
      },
    ],
    trendAnalysis: $t('page.qualityDashboard.ai.trendAnalysis'),
  },
  completenessData: [
    { type: $t('page.qualityDashboard.pie.complete'), value: 13_700 },
    { type: $t('page.qualityDashboard.pie.missing'), value: 1980 },
  ],
  consistencyData: [
    { type: $t('page.qualityDashboard.pie.consistent'), value: 13_420 },
    { type: $t('page.qualityDashboard.pie.inconsistent'), value: 2260 },
  ],
  dataConsistency: 85.6,
  dataTimeliness: 88.7,
  dataUniqueness: 94.1,
  fieldAccuracy: 92.3,
  fieldCompleteness: 87.5,
  issueList: [],
  tableCompleteness: 89.2,
  totalColumns: 15_680,
  totalIssues: 342,
  totalTables: 1250,
  uniquenessData: [
    { type: $t('page.qualityDashboard.pie.unique'), value: 14_750 },
    { type: $t('page.qualityDashboard.pie.duplicate'), value: 930 },
  ],
};

const loading = ref(true);
const dashboardData = ref<DashboardData>(mockData);

const completenessChartRef = ref<EchartsUIType>();
const accuracyChartRef = ref<EchartsUIType>();
const consistencyChartRef = ref<EchartsUIType>();
const uniquenessChartRef = ref<EchartsUIType>();

const { renderEcharts: renderCompleteness } = useEcharts(completenessChartRef);
const { renderEcharts: renderAccuracy } = useEcharts(accuracyChartRef);
const { renderEcharts: renderConsistency } = useEcharts(consistencyChartRef);
const { renderEcharts: renderUniqueness } = useEcharts(uniquenessChartRef);

function getQualityColor(rate: number) {
  if (rate >= 90) return '#52c41a';
  if (rate >= 80) return '#1890ff';
  if (rate >= 70) return '#faad14';
  return '#ff4d4f';
}

function getIssueLevelColor(level: string) {
  if (level === 'high') return 'red';
  if (level === 'medium') return 'orange';
  if (level === 'low') return 'blue';
  return 'default';
}

function getIssueLevelText(level: string) {
  if (level === 'high') return $t('page.qualityDashboard.level.high');
  if (level === 'medium') return $t('page.qualityDashboard.level.medium');
  if (level === 'low') return $t('page.qualityDashboard.level.low');
  return $t('page.qualityDashboard.level.unknown');
}

function normalizePieData(data?: PieItem[]) {
  if (!data || data.length === 0) return [{ type: $t('page.qualityDashboard.noData'), value: 1 }];
  return data;
}

function renderPie(
  render: (option: Record<string, any>) => void,
  title: string,
  data?: PieItem[],
  colors?: string[],
) {
  const pieData = normalizePieData(data);
  render({
    color: colors,
    legend: { bottom: '2%', left: 'center' },
    series: [
      {
        data: pieData.map((item) => ({ name: item.type, value: item.value })),
        label: { formatter: '{b}: {d}%' },
        name: title,
        radius: ['45%', '70%'],
        type: 'pie',
      },
    ],
    tooltip: { trigger: 'item' },
  });
}

function renderAllCharts() {
  renderPie(
    renderCompleteness,
    $t('page.qualityDashboard.chart.fieldCompleteness'),
    dashboardData.value.completenessData,
    ['#52c41a', '#ff4d4f'],
  );
  renderPie(
    renderAccuracy,
    $t('page.qualityDashboard.chart.fieldAccuracy'),
    dashboardData.value.accuracyData,
    ['#1890ff', '#faad14', '#ff4d4f'],
  );
  renderPie(
    renderConsistency,
    $t('page.qualityDashboard.chart.dataConsistency'),
    dashboardData.value.consistencyData,
    ['#52c41a', '#ff7875'],
  );
  renderPie(
    renderUniqueness,
    $t('page.qualityDashboard.chart.dataUniqueness'),
    dashboardData.value.uniquenessData,
    ['#52c41a', '#ff7875'],
  );
}

function resolveDataQualityResponse(response: any): DashboardData | null {
  const payload = (response as any)?.data ?? response;
  if (payload?.code === 200 && payload?.data) return payload.data as DashboardData;
  if (payload?.data?.totalTables !== undefined) return payload.data as DashboardData;
  if (payload?.totalTables !== undefined) return payload as DashboardData;
  return null;
}

const overallScore = computed(() => {
  const d = dashboardData.value;
  const score =
    ((d.fieldCompleteness || 0) +
      (d.fieldAccuracy || 0) +
      (d.tableCompleteness || 0) +
      (d.dataConsistency || 0) +
      (d.dataUniqueness || 0) +
      (d.dataTimeliness || 0)) /
    6;
  return Math.round(score);
});

const issueColumns: TableColumnsType<IssueItem> = [
  { dataIndex: 'tableName', key: 'tableName', title: $t('page.qualityDashboard.columns.tableName'), width: 150 },
  { dataIndex: 'columnName', key: 'columnName', title: $t('page.qualityDashboard.columns.columnName'), width: 150 },
  { dataIndex: 'issueType', key: 'issueType', title: $t('page.qualityDashboard.columns.issueType'), width: 120 },
  { dataIndex: 'issueLevel', key: 'issueLevel', title: $t('page.qualityDashboard.columns.issueLevel'), width: 100 },
  { dataIndex: 'issueDesc', key: 'issueDesc', title: $t('page.qualityDashboard.columns.issueDesc') },
  { dataIndex: 'issueCount', key: 'issueCount', title: $t('page.qualityDashboard.columns.issueCount'), width: 100 },
  { dataIndex: 'lastCheckTime', key: 'lastCheckTime', title: $t('page.qualityDashboard.columns.lastCheckTime'), width: 180 },
  { key: 'aiAnalysis', title: $t('page.qualityDashboard.columns.aiAnalysis'), width: 120 },
];

async function fetchDashboardData() {
  try {
    const response = await baseRequestClient.get('/v1/dataquality/dashboard/info');
    const parsed = resolveDataQualityResponse(response);
    dashboardData.value = parsed || mockData;
  } catch {
    dashboardData.value = mockData;
  } finally {
    loading.value = false;
    await nextTick();
    renderAllCharts();
  }
}

onMounted(fetchDashboardData);
</script>

<template>
  <div class="p-5">
    <Row :gutter="[16, 16]" class="mb-4">
      <Col :lg="6" :md="12" :xs="24"><Card><Statistic :title="$t('page.qualityDashboard.kpi.totalTables')" :value="dashboardData.totalTables || 0" /></Card></Col>
      <Col :lg="6" :md="12" :xs="24"><Card><Statistic :title="$t('page.qualityDashboard.kpi.totalColumns')" :value="dashboardData.totalColumns || 0" /></Card></Col>
      <Col :lg="6" :md="12" :xs="24"><Card><Statistic :title="$t('page.qualityDashboard.kpi.totalIssues')" :value="dashboardData.totalIssues || 0" /></Card></Col>
      <Col :lg="6" :md="12" :xs="24">
        <Card>
          <Statistic :title="$t('page.qualityDashboard.kpi.overallScore')" :value="overallScore" suffix="/ 100" :value-style="{ color: getQualityColor(overallScore) }" />
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mb-4">
      <Col v-for="item in [
        { key: 'fieldCompleteness', label: $t('page.qualityDashboard.metric.fieldCompleteness'), desc: $t('page.qualityDashboard.metricDesc.fieldCompleteness') },
        { key: 'fieldAccuracy', label: $t('page.qualityDashboard.metric.fieldAccuracy'), desc: $t('page.qualityDashboard.metricDesc.fieldAccuracy') },
        { key: 'tableCompleteness', label: $t('page.qualityDashboard.metric.tableCompleteness'), desc: $t('page.qualityDashboard.metricDesc.tableCompleteness') },
        { key: 'dataConsistency', label: $t('page.qualityDashboard.metric.dataConsistency'), desc: $t('page.qualityDashboard.metricDesc.dataConsistency') },
        { key: 'dataUniqueness', label: $t('page.qualityDashboard.metric.dataUniqueness'), desc: $t('page.qualityDashboard.metricDesc.dataUniqueness') },
        { key: 'dataTimeliness', label: $t('page.qualityDashboard.metric.dataTimeliness'), desc: $t('page.qualityDashboard.metricDesc.dataTimeliness') },
      ]" :key="item.key" :lg="8" :md="12" :xs="24">
        <Card :title="item.label">
          <div class="text-2xl font-semibold">{{ (dashboardData as any)[item.key] || 0 }}%</div>
          <Progress :percent="(dashboardData as any)[item.key] || 0" :show-info="false" :stroke-color="getQualityColor((dashboardData as any)[item.key] || 0)" />
          <div class="mt-2 text-xs text-gray-500">{{ item.desc }}</div>
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mb-4">
      <Col :md="12" :xs="24"><Card :title="$t('page.qualityDashboard.chart.fieldCompleteness')"><EchartsUI ref="completenessChartRef" class="h-[300px]" /></Card></Col>
      <Col :md="12" :xs="24"><Card :title="$t('page.qualityDashboard.chart.fieldAccuracy')"><EchartsUI ref="accuracyChartRef" class="h-[300px]" /></Card></Col>
    </Row>
    <Row :gutter="[16, 16]" class="mb-4">
      <Col :md="12" :xs="24"><Card :title="$t('page.qualityDashboard.chart.dataConsistency')"><EchartsUI ref="consistencyChartRef" class="h-[300px]" /></Card></Col>
      <Col :md="12" :xs="24"><Card :title="$t('page.qualityDashboard.chart.dataUniqueness')"><EchartsUI ref="uniquenessChartRef" class="h-[300px]" /></Card></Col>
    </Row>

    <Card :title="$t('page.qualityDashboard.issueListTitle')" :bordered="false">
      <Table
        :columns="issueColumns"
        :data-source="dashboardData.issueList || []"
        :loading="loading"
        :pagination="{ pageSize: 10, showSizeChanger: true, showTotal: (total:number) => `${$t('page.common.total')} ${total} ${$t('page.qualityDashboard.issuesUnit')}` }"
        :row-key="(record: IssueItem, index?: number) => record.key || `${record.tableName || 't'}-${record.columnName || 'c'}-${index ?? 0}`"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'issueLevel'">
            <Tag :color="getIssueLevelColor(record.issueLevel || '')">{{ getIssueLevelText(record.issueLevel || '') }}</Tag>
          </template>
          <template v-else-if="column.key === 'issueDesc'">
            <Tooltip :title="record.issueDesc || ''">
              <span class="inline-block max-w-[260px] truncate">{{ record.issueDesc || '' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'aiAnalysis'">
            <Tooltip
              :title="$t('page.qualityDashboard.aiTooltip', { level: getIssueLevelText(record.issueLevel || '') })"
            >
              <Tag color="blue">{{ $t('page.qualityDashboard.columns.aiAnalysis') }}</Tag>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'issueCount'">
            {{ record.issueCount || 0 }}
          </template>
        </template>
      </Table>
    </Card>
  </div>
</template>

<style scoped></style>
