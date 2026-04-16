<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { EchartsUIType } from '@vben/plugins/echarts';

import { computed, nextTick, onMounted, ref } from 'vue';

import { AnalysisChartCard, AnalysisOverview } from '@vben/common-ui';
import {
  SvgBellIcon,
  SvgCakeIcon,
  SvgCardIcon,
  SvgDownloadIcon,
} from '@vben/icons';
import { EchartsUI, useEcharts } from '@vben/plugins/echarts';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

interface PieItem {
  type: string;
  value: number;
}

interface DashboardData {
  columnCount?: number | string;
  columnPieDataList?: PieItem[];
  databaseCount?: number | string;
  databasePieDataList?: PieItem[];
  datasourceCount?: number | string;
  datasourceEnvCount?: number | string;
  datasourceIdcCount?: number | string;
  datasourcePieDataList?: PieItem[];
  datasourceTypeCount?: number | string;
  tableCount?: number | string;
  tablePieDataList?: PieItem[];
}

const loading = ref(true);
const dashboardData = ref<DashboardData>({});
const overviewItems = computed<AnalysisOverviewItem[]>(() => [
  {
    icon: SvgCardIcon,
    title: $t('page.metaDashboard.overview.datasourceCount'),
    totalTitle: $t('page.metaDashboard.overview.datasourceCount'),
    totalValue: toNumber(dashboardData.value.datasourceCount),
    value: toNumber(dashboardData.value.datasourceCount),
  },
  {
    icon: SvgCakeIcon,
    title: $t('page.metaDashboard.overview.databaseCount'),
    totalTitle: $t('page.metaDashboard.overview.databaseCount'),
    totalValue: toNumber(dashboardData.value.databaseCount),
    value: toNumber(dashboardData.value.databaseCount),
  },
  {
    icon: SvgDownloadIcon,
    title: $t('page.metaDashboard.overview.tableCount'),
    totalTitle: $t('page.metaDashboard.overview.tableCount'),
    totalValue: toNumber(dashboardData.value.tableCount),
    value: toNumber(dashboardData.value.tableCount),
  },
  {
    icon: SvgBellIcon,
    title: $t('page.metaDashboard.overview.columnCount'),
    totalTitle: $t('page.metaDashboard.overview.columnCount'),
    totalValue: toNumber(dashboardData.value.columnCount),
    value: toNumber(dashboardData.value.columnCount),
  },
]);

const datasourceChartRef = ref<EchartsUIType>();
const databaseChartRef = ref<EchartsUIType>();
const tableChartRef = ref<EchartsUIType>();
const columnChartRef = ref<EchartsUIType>();

const { renderEcharts: renderDatasourceChart } = useEcharts(datasourceChartRef);
const { renderEcharts: renderDatabaseChart } = useEcharts(databaseChartRef);
const { renderEcharts: renderTableChart } = useEcharts(tableChartRef);
const { renderEcharts: renderColumnChart } = useEcharts(columnChartRef);

function normalizePieData(data?: PieItem[]) {
  if (!data || data.length === 0) {
    return [{ type: $t('page.metaDashboard.noData'), value: 1 }];
  }
  return data;
}

function resolveDashboardData(response: any): DashboardData {
  if (response?.data?.datasourceTypeCount !== undefined) {
    return response.data as DashboardData;
  }
  if (response?.datasourceTypeCount !== undefined) {
    return response as DashboardData;
  }
  if (response?.data?.data) {
    return response.data.data as DashboardData;
  }
  return {};
}

function toNumber(value: number | string | undefined) {
  if (typeof value === 'number') {
    return value;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function normalizeDashboardData(data: DashboardData): DashboardData {
  return {
    ...data,
    columnCount: toNumber(data.columnCount),
    databaseCount: toNumber(data.databaseCount),
    datasourceCount: toNumber(data.datasourceCount),
    datasourceEnvCount: toNumber(data.datasourceEnvCount),
    datasourceIdcCount: toNumber(data.datasourceIdcCount),
    datasourceTypeCount: toNumber(data.datasourceTypeCount),
    tableCount: toNumber(data.tableCount),
  };
}

function renderPie(
  render: (option: Record<string, any>) => void,
  title: string,
  data?: PieItem[],
) {
  const pieData = normalizePieData(data);
  render({
    legend: {
      bottom: '2%',
      left: 'center',
    },
    series: [
      {
        avoidLabelOverlap: false,
        data: pieData.map((item) => ({ name: item.type, value: item.value })),
        emphasis: {
          label: {
            fontSize: 12,
            fontWeight: 'bold',
            show: true,
          },
        },
        itemStyle: {
          borderRadius: 10,
          borderWidth: 2,
        },
        label: {
          formatter: '{b}: {d}%',
          show: true,
        },
        name: title,
        radius: ['40%', '65%'],
        type: 'pie',
      },
    ],
    tooltip: {
      trigger: 'item',
    },
  });
}

function renderAllCharts() {
  renderPie(
    renderDatasourceChart,
    $t('page.metaDashboard.chart.datasourceDistribution'),
    dashboardData.value.datasourcePieDataList,
  );
  renderPie(renderDatabaseChart, $t('page.metaDashboard.chart.databaseDistribution'), dashboardData.value.databasePieDataList);
  renderPie(renderTableChart, $t('page.metaDashboard.chart.tableDistribution'), dashboardData.value.tablePieDataList);
  renderPie(renderColumnChart, $t('page.metaDashboard.chart.columnDistribution'), dashboardData.value.columnPieDataList);
}

async function fetchDashboard() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/dashboard/info');
    dashboardData.value = normalizeDashboardData(resolveDashboardData(response));
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
  <div class="p-5">
    <AnalysisOverview :items="overviewItems" />

    <div class="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2">
      <AnalysisChartCard :title="$t('page.metaDashboard.chart.datasourceDistribution')">
        <EchartsUI ref="datasourceChartRef" class="h-[340px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.metaDashboard.chart.databaseDistribution')">
        <EchartsUI ref="databaseChartRef" class="h-[340px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.metaDashboard.chart.tableDistribution')">
        <EchartsUI ref="tableChartRef" class="h-[340px]" />
      </AnalysisChartCard>
      <AnalysisChartCard :title="$t('page.metaDashboard.chart.columnDistribution')">
        <EchartsUI ref="columnChartRef" class="h-[340px]" />
      </AnalysisChartCard>
    </div>

    <a-row :gutter="[16, 16]" class="mt-5">
      <a-col :lg="8" :md="8" :sm="12" :xs="24">
        <a-card :loading="loading">
          <a-statistic :title="$t('page.metaDashboard.kpi.idcCount')" :value="dashboardData.datasourceIdcCount || 0" />
        </a-card>
      </a-col>
      <a-col :lg="8" :md="8" :sm="12" :xs="24">
        <a-card :loading="loading">
          <a-statistic :title="$t('page.metaDashboard.kpi.envCount')" :value="dashboardData.datasourceEnvCount || 0" />
        </a-card>
      </a-col>
      <a-col :lg="8" :md="8" :sm="12" :xs="24">
        <a-card :loading="loading">
          <a-statistic :title="$t('page.metaDashboard.kpi.tableCount')" :value="dashboardData.tableCount || 0" />
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>
