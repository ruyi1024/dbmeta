<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { EchartsUIType } from '@vben/plugins/echarts';
import type { TableColumnsType } from 'ant-design-vue';

import { computed, nextTick, onMounted, ref } from 'vue';

import { AnalysisChartCard, AnalysisOverview } from '@vben/common-ui';
import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { Card, Col, Row, Spin, Table, Typography } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

const { Title } = Typography;

defineOptions({ name: 'DataSecurityDashboard' });

interface PieItem {
  type: string;
  value: number;
}

interface TrendPoint {
  category?: string;
  time?: string;
  value?: number | string;
}

interface SafeDashboardData {
  query15DayInterceptLineDataList?: TrendPoint[];
  query15DayLineDataList?: TrendPoint[];
  queryMonthBarDataList?: TrendPoint[];
  queryNewInterceptDataList?: Record<string, unknown>[];
  queryNewSensitiveDataList?: Record<string, unknown>[];
  queryStatusPieDataList?: PieItem[];
  queryTypePieDataList?: PieItem[];
  sensitiveColumnCount?: number | string;
  sensitiveDatabaseCount?: number | string;
  sensitiveDsTypePieDataList?: PieItem[];
  sensitiveQueryCount?: number | string;
  sensitiveTableCount?: number | string;
  sensitiveTypePieDataList?: PieItem[];
  todayQueryCount?: number | string;
  totalInterceptCount?: number | string;
  totalQueryCount?: number | string;
}

const loading = ref(true);
const dashboardData = ref<SafeDashboardData>({});

const statusPieRef = ref<EchartsUIType>();
const typePieRef = ref<EchartsUIType>();
const dsPieRef = ref<EchartsUIType>();
const sensitiveTypePieRef = ref<EchartsUIType>();
const queryLineRef = ref<EchartsUIType>();
const interceptLineRef = ref<EchartsUIType>();
const monthBarRef = ref<EchartsUIType>();

const { renderEcharts: renderStatusPie } = useEcharts(statusPieRef);
const { renderEcharts: renderTypePie } = useEcharts(typePieRef);
const { renderEcharts: renderDsPie } = useEcharts(dsPieRef);
const { renderEcharts: renderSensitiveTypePie } = useEcharts(sensitiveTypePieRef);
const { renderEcharts: renderQueryLine } = useEcharts(queryLineRef);
const { renderEcharts: renderInterceptLine } = useEcharts(interceptLineRef);
const { renderEcharts: renderMonthBar } = useEcharts(monthBarRef);

const overviewItems = computed<AnalysisOverviewItem[]>(() => {
  const d = dashboardData.value;
  const v = (key: keyof SafeDashboardData) =>
    toNumber(d[key] as number | string | undefined);
  return [
    {
      icon: 'lucide:calendar-clock',
      title: $t('page.securityDashboard.overview.todayQuery'),
      totalTitle: $t('page.securityDashboard.overview.todayQueryTotal'),
      totalValue: v('todayQueryCount'),
      value: v('todayQueryCount'),
    },
    {
      icon: 'lucide:bar-chart-3',
      title: $t('page.securityDashboard.overview.totalQuery'),
      totalTitle: $t('page.securityDashboard.overview.totalQueryTotal'),
      totalValue: v('totalQueryCount'),
      value: v('totalQueryCount'),
    },
    {
      icon: 'lucide:shield-alert',
      title: $t('page.securityDashboard.overview.interceptQuery'),
      totalTitle: $t('page.securityDashboard.overview.interceptQueryTotal'),
      totalValue: v('totalInterceptCount'),
      value: v('totalInterceptCount'),
    },
    {
      icon: 'lucide:database',
      title: $t('page.securityDashboard.overview.sensitiveDatabase'),
      totalTitle: $t('page.securityDashboard.overview.sensitiveDatabaseTotal'),
      totalValue: v('sensitiveDatabaseCount'),
      value: v('sensitiveDatabaseCount'),
    },
    {
      icon: 'lucide:table-2',
      title: $t('page.securityDashboard.overview.sensitiveTable'),
      totalTitle: $t('page.securityDashboard.overview.sensitiveTableTotal'),
      totalValue: v('sensitiveTableCount'),
      value: v('sensitiveTableCount'),
    },
    {
      icon: 'lucide:columns-3',
      title: $t('page.securityDashboard.overview.sensitiveColumn'),
      totalTitle: $t('page.securityDashboard.overview.sensitiveColumnTotal'),
      totalValue: v('sensitiveColumnCount'),
      value: v('sensitiveColumnCount'),
    },
    {
      icon: 'lucide:shield-check',
      title: $t('page.securityDashboard.overview.sensitiveProtect'),
      totalTitle: $t('page.securityDashboard.overview.sensitiveProtectTotal'),
      totalValue: v('sensitiveQueryCount'),
      value: v('sensitiveQueryCount'),
    },
  ];
});

const columnsIntercept: TableColumnsType = [
  { title: $t('page.securityDashboard.columns.interceptReason'), dataIndex: 'result', key: 'result', ellipsis: true },
  { title: $t('page.securityDashboard.columns.sqlType'), dataIndex: 'sql_type', key: 'sql_type' },
  { title: $t('page.securityDashboard.columns.datasourceType'), dataIndex: 'datasource_type', key: 'datasource_type' },
  { title: $t('page.securityDashboard.columns.database'), dataIndex: 'database', key: 'database' },
  { title: $t('page.securityDashboard.columns.username'), dataIndex: 'username', key: 'username' },
  { title: $t('page.securityDashboard.columns.date'), dataIndex: 'gmt_created', key: 'gmt_created', width: 120 },
];

const columnsSensitive: TableColumnsType = [
  { title: $t('page.securityDashboard.columns.sensitiveType'), dataIndex: 'rule_name', key: 'rule_name' },
  { title: $t('page.securityDashboard.columns.datasourceType'), dataIndex: 'datasource_type', key: 'datasource_type' },
  { title: $t('page.securityDashboard.columns.database'), dataIndex: 'database_name', key: 'database_name' },
  { title: $t('page.securityDashboard.columns.tableName'), dataIndex: 'table_name', key: 'table_name' },
  { title: $t('page.securityDashboard.columns.columnName'), dataIndex: 'column_name', key: 'column_name' },
  { title: $t('page.securityDashboard.columns.date'), dataIndex: 'gmt_created', key: 'gmt_created', width: 120 },
];

function toNumber(value: number | string | undefined) {
  if (value === undefined || value === null) return 0;
  if (typeof value === 'number') return Number.isFinite(value) ? value : 0;
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function resolveDashboardPayload(response: unknown): SafeDashboardData {
  if (!response || typeof response !== 'object') {
    return {};
  }
  const r = response as Record<string, unknown>;
  const httpBody =
    'status' in r && typeof r.status === 'number' && r.data !== undefined
      ? r.data
      : r;
  const hb = httpBody as Record<string, unknown>;
  const inner = hb?.data;
  if (inner && typeof inner === 'object') {
    return inner as SafeDashboardData;
  }
  return (httpBody as SafeDashboardData) || {};
}

function normalizePieData(data?: PieItem[]) {
  if (!data?.length) {
    return [{ type: $t('page.securityDashboard.noData'), value: 1 }];
  }
  return data.map((d) => ({
    type: d.type,
    value: toNumber(d.value as number | string),
  }));
}

function renderPie(
  render: (opt: Record<string, unknown>) => void,
  title: string,
  data?: PieItem[],
) {
  const pieData = normalizePieData(data);
  render({
    legend: { bottom: '2%', left: 'center' },
    series: [
      {
        avoidLabelOverlap: false,
        data: pieData.map((item) => ({ name: item.type, value: item.value })),
        label: { formatter: '{b}: {d}%', show: true },
        name: title,
        radius: ['40%', '65%'],
        type: 'pie',
      },
    ],
    tooltip: { trigger: 'item' },
  });
}

function renderSimpleLine(
  render: (opt: Record<string, unknown>) => void,
  list?: TrendPoint[],
) {
  const rows = Array.isArray(list) ? list : [];
  if (!rows.length) {
    render({
      series: [],
      title: {
        left: 'center',
        text: $t('page.securityDashboard.noData'),
        textStyle: { color: '#999', fontSize: 14 },
        top: 'center',
      },
      xAxis: { show: false },
      yAxis: { show: false },
    });
    return;
  }
  const times = rows.map((r) => String(r.time ?? ''));
  const values = rows.map((r) => toNumber(r.value));
  render({
    grid: { bottom: 48, containLabel: true, left: 48, right: 24, top: 24 },
    series: [
      {
        areaStyle: { opacity: 0.08 },
        data: values,
        name: rows[0]?.category ?? $t('page.securityDashboard.value'),
        smooth: true,
        type: 'line',
      },
    ],
    tooltip: { trigger: 'axis' },
    xAxis: { boundaryGap: false, data: times, type: 'category' },
    yAxis: { type: 'value' },
  });
}

function renderGroupedBar(
  render: (opt: Record<string, unknown>) => void,
  list?: TrendPoint[],
) {
  const rows = Array.isArray(list) ? list : [];
  if (!rows.length) {
    render({
      series: [],
      title: {
        left: 'center',
        text: $t('page.securityDashboard.noData'),
        textStyle: { color: '#999', fontSize: 14 },
        top: 'center',
      },
      xAxis: { show: false },
      yAxis: { show: false },
    });
    return;
  }
  const times = [...new Set(rows.map((r) => String(r.time ?? '')))].sort();
  const cats = [...new Set(rows.map((r) => String(r.category ?? '')))];
  const series = cats.map((cat) => ({
    data: times.map((t) => {
      const row = rows.find((r) => String(r.time) === t && String(r.category) === cat);
      return row ? toNumber(row.value) : 0;
    }),
    name: cat,
    type: 'bar',
  }));
  render({
    grid: { bottom: 48, containLabel: true, left: 48, right: 24, top: 40 },
    legend: { top: 0 },
    series,
    tooltip: { trigger: 'axis' },
    xAxis: { data: times, type: 'category' },
    yAxis: { type: 'value' },
  });
}

function renderAllCharts() {
  renderPie(renderStatusPie, $t('page.securityDashboard.chart.sqlStatus'), dashboardData.value.queryStatusPieDataList);
  renderPie(renderTypePie, $t('page.securityDashboard.chart.sqlType'), dashboardData.value.queryTypePieDataList);
  renderPie(renderDsPie, $t('page.securityDashboard.chart.sensitiveDatasource'), dashboardData.value.sensitiveDsTypePieDataList);
  renderPie(
    renderSensitiveTypePie,
    $t('page.securityDashboard.chart.sensitiveType'),
    dashboardData.value.sensitiveTypePieDataList,
  );
  renderSimpleLine(renderQueryLine, dashboardData.value.query15DayLineDataList);
  renderSimpleLine(renderInterceptLine, dashboardData.value.query15DayInterceptLineDataList);
  renderGroupedBar(renderMonthBar, dashboardData.value.queryMonthBarDataList);
}

async function fetchDashboard() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get<unknown>('/v1/safe/dashboard/info');
    dashboardData.value = resolveDashboardPayload(response);
  } catch {
    dashboardData.value = {};
  } finally {
    loading.value = false;
    await nextTick();
    renderAllCharts();
  }
}

onMounted(fetchDashboard);
</script>

<template>
  <div class="safe-dashboard p-5">
    <div class="safe-dashboard__stats mb-6">
      <Title :level="5" class="!mb-4 text-foreground/90"> {{ $t('page.securityDashboard.kpiTitle') }} </Title>
      <Spin :spinning="loading">
        <AnalysisOverview :items="overviewItems" />
      </Spin>
    </div>

    <div class="mt-2 grid grid-cols-1 gap-4 md:grid-cols-2">
      <AnalysisChartCard :title="$t('page.securityDashboard.chart.sqlStatusDistribution')">
        <EchartsUI ref="statusPieRef" class="h-[330px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.securityDashboard.chart.sqlTypeDistribution')">
        <EchartsUI ref="typePieRef" class="h-[330px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.securityDashboard.chart.queryTrend15d')">
        <EchartsUI ref="queryLineRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.securityDashboard.chart.interceptTrend15d')">
        <EchartsUI ref="interceptLineRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard class="md:col-span-2" :title="$t('page.securityDashboard.chart.yearlyQueryIntercept')">
        <EchartsUI ref="monthBarRef" class="h-[360px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.securityDashboard.chart.sensitiveDatasourceDistribution')">
        <EchartsUI ref="dsPieRef" class="h-[330px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.securityDashboard.chart.sensitiveTypeDistribution')">
        <EchartsUI ref="sensitiveTypePieRef" class="h-[330px]" />
      </AnalysisChartCard>
    </div>

    <Row :gutter="[16, 16]" class="mt-5">
      <Col :lg="12" :span="24">
        <Card :title="$t('page.securityDashboard.latestIntercept')" :loading="loading">
          <Table
            :columns="columnsIntercept"
            :data-source="dashboardData.queryNewInterceptDataList ?? []"
            :pagination="false"
            :row-key="(_record, index) => `intercept-${index}`"
            size="small"
          />
        </Card>
      </Col>
      <Col :lg="12" :span="24">
        <Card :title="$t('page.securityDashboard.latestSensitive')" :loading="loading">
          <Table
            :columns="columnsSensitive"
            :data-source="dashboardData.queryNewSensitiveDataList ?? []"
            :pagination="false"
            :row-key="(_record, index) => `sensitive-${index}`"
            size="small"
          />
        </Card>
      </Col>
    </Row>
  </div>
</template>

<style scoped>
.safe-dashboard__stats {
  border-radius: 8px;
  border: 1px solid var(--ant-color-border-secondary, #f0f0f0);
  background: var(--ant-color-fill-quaternary, rgba(0, 0, 0, 0.02));
  padding: 16px 16px 8px;
}
</style>
