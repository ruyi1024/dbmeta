<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { EchartsUIType } from '@vben/plugins/echarts';
import type { TableColumnsType } from 'ant-design-vue';

import { computed, nextTick, onMounted, ref } from 'vue';

import { AnalysisChartCard, AnalysisOverview } from '@vben/common-ui';
import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { Card, Col, Row, Spin, Table, Typography } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

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
      title: '今日查询',
      totalTitle: '当日 SQL 执行次数',
      totalValue: v('todayQueryCount'),
      value: v('todayQueryCount'),
    },
    {
      icon: 'lucide:bar-chart-3',
      title: '查询总数',
      totalTitle: '累计查询次数',
      totalValue: v('totalQueryCount'),
      value: v('totalQueryCount'),
    },
    {
      icon: 'lucide:shield-alert',
      title: '高危查询拦截',
      totalTitle: '高危查询拦截累计',
      totalValue: v('totalInterceptCount'),
      value: v('totalInterceptCount'),
    },
    {
      icon: 'lucide:database',
      title: '敏感数据库',
      totalTitle: '涉及库数量',
      totalValue: v('sensitiveDatabaseCount'),
      value: v('sensitiveDatabaseCount'),
    },
    {
      icon: 'lucide:table-2',
      title: '敏感数据表',
      totalTitle: '涉及表数量',
      totalValue: v('sensitiveTableCount'),
      value: v('sensitiveTableCount'),
    },
    {
      icon: 'lucide:columns-3',
      title: '敏感数据字段',
      totalTitle: '涉及字段数量',
      totalValue: v('sensitiveColumnCount'),
      value: v('sensitiveColumnCount'),
    },
    {
      icon: 'lucide:shield-check',
      title: '敏感数据保护次数',
      totalTitle: '敏感库相关查询',
      totalValue: v('sensitiveQueryCount'),
      value: v('sensitiveQueryCount'),
    },
  ];
});

const columnsIntercept: TableColumnsType = [
  { title: '拦截原因', dataIndex: 'result', key: 'result', ellipsis: true },
  { title: '执行类型', dataIndex: 'sql_type', key: 'sql_type' },
  { title: '类型', dataIndex: 'datasource_type', key: 'datasource_type' },
  { title: '数据库', dataIndex: 'database', key: 'database' },
  { title: '用户', dataIndex: 'username', key: 'username' },
  { title: '日期', dataIndex: 'gmt_created', key: 'gmt_created', width: 120 },
];

const columnsSensitive: TableColumnsType = [
  { title: '敏感类型', dataIndex: 'rule_name', key: 'rule_name' },
  { title: '源类型', dataIndex: 'datasource_type', key: 'datasource_type' },
  { title: '数据库', dataIndex: 'database_name', key: 'database_name' },
  { title: '数据表', dataIndex: 'table_name', key: 'table_name' },
  { title: '数据字段', dataIndex: 'column_name', key: 'column_name' },
  { title: '日期', dataIndex: 'gmt_created', key: 'gmt_created', width: 120 },
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
    return [{ type: '暂无数据', value: 1 }];
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
        text: '暂无数据',
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
        name: rows[0]?.category ?? '数值',
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
        text: '暂无数据',
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
  renderPie(renderStatusPie, 'SQL执行状态', dashboardData.value.queryStatusPieDataList);
  renderPie(renderTypePie, 'SQL执行类型', dashboardData.value.queryTypePieDataList);
  renderPie(renderDsPie, '敏感数据源', dashboardData.value.sensitiveDsTypePieDataList);
  renderPie(
    renderSensitiveTypePie,
    '敏感类型',
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
      <Title :level="5" class="!mb-4 text-foreground/90"> 核心指标 </Title>
      <Spin :spinning="loading">
        <AnalysisOverview :items="overviewItems" />
      </Spin>
    </div>

    <div class="mt-2 grid grid-cols-1 gap-4 md:grid-cols-2">
      <AnalysisChartCard title="SQL执行状态分布">
        <EchartsUI ref="statusPieRef" class="h-[330px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="SQL执行类型分布">
        <EchartsUI ref="typePieRef" class="h-[330px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="近15日SQL查询趋势">
        <EchartsUI ref="queryLineRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="近15日SQL拦截趋势">
        <EchartsUI ref="interceptLineRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard class="md:col-span-2" title="年度SQL查询和风险拦截统计">
        <EchartsUI ref="monthBarRef" class="h-[360px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="敏感数据数据源分布">
        <EchartsUI ref="dsPieRef" class="h-[330px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="敏感数据类型分布">
        <EchartsUI ref="sensitiveTypePieRef" class="h-[330px]" />
      </AnalysisChartCard>
    </div>

    <Row :gutter="[16, 16]" class="mt-5">
      <Col :lg="12" :span="24">
        <Card title="SQL执行最新拦截记录" :loading="loading">
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
        <Card title="敏感信息最新探测记录" :loading="loading">
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
