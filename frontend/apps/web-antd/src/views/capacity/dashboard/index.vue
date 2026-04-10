<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { AnalysisChartCard } from '@vben/common-ui';
import { createIconifyIcon } from '@vben/icons';
import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { Card, Col, Row, Spin, Statistic, Typography } from 'ant-design-vue';
import { computed, nextTick, onMounted, reactive, ref } from 'vue';

import { baseRequestClient } from '#/api/request';

const { Title } = Typography;

defineOptions({ name: 'DataCapacityDashboard' });

const IconDatabase = createIconifyIcon('lucide:database');
const IconTable = createIconifyIcon('lucide:table-2');
const IconHdd = createIconifyIcon('lucide:hard-drive');
const IconTrending = createIconifyIcon('lucide:trending-up');

/** 与数据安全大盘 `renderGroupedBar` / `renderSimpleLine` 的 grid 一致 */
const CHART_GRID = {
  bottom: 48,
  containLabel: true,
  left: 48,
  right: 24,
  top: 40,
} as const;

/** AntV 系配色，与安全大盘图表观感一致 */
const BAR_PALETTE = {
  database: { top: '#5B8FF9', bottom: '#8EB8FF', emphasis: 'rgba(91, 143, 249, 0.45)' },
  table: { top: '#61DDAA', bottom: '#95E8CF', emphasis: 'rgba(97, 221, 170, 0.45)' },
  fragmentation: { top: '#F6BD16', bottom: '#FCDA6A', emphasis: 'rgba(246, 189, 22, 0.45)' },
  rows: { top: '#7262fd', bottom: '#A69FFB', emphasis: 'rgba(114, 98, 253, 0.45)' },
} as const;

function barLinearGradient(top: string, bottom: string) {
  return {
    type: 'linear' as const,
    x: 0,
    y: 0,
    x2: 0,
    y2: 1,
    colorStops: [
      { offset: 0, color: top },
      { offset: 1, color: bottom },
    ],
  };
}

const CHART_TOOLTIP_BASE = {
  axisPointer: {
    type: 'shadow' as const,
    shadowStyle: { color: 'rgba(0, 0, 0, 0.04)' },
  },
  borderWidth: 0,
  padding: [10, 14] as const,
  extraCssText:
    'border-radius:8px;box-shadow:0 3px 14px rgba(0,0,0,0.1);backdrop-filter:blur(8px);',
};

interface CapacityStats {
  totalDatabases: number;
  totalTables: number;
  totalDataSize: string;
  totalRows: number;
  dailyGrowth: string;
  dailyGrowthRows: number;
}

/** Axios 完整响应 → HTTP body（与 meta/instance 等页一致） */
function unwrapAxiosData(response: unknown): unknown {
  if (!response || typeof response !== 'object') {
    return response;
  }
  const r = response as Record<string, unknown>;
  if (
    'data' in r &&
    'status' in r &&
    typeof r.status === 'number'
  ) {
    return r.data;
  }
  return response;
}

/** 解析 pumpkin 接口：{ success, data } 或扁平 stats */
function parsePumpkinStatsPayload(response: unknown): Partial<CapacityStats> | null {
  const raw = unwrapAxiosData(response);
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return null;
  }
  const body = raw as Record<string, unknown>;
  const inner = body.data;
  if (
    inner &&
    typeof inner === 'object' &&
    !Array.isArray(inner) &&
    ('totalDatabases' in inner || 'totalTables' in inner)
  ) {
    return inner as Partial<CapacityStats>;
  }
  if ('totalDatabases' in body || 'totalTables' in body) {
    return body as Partial<CapacityStats>;
  }
  return null;
}

function parsePumpkinListPayload(response: unknown): any[] {
  const raw = unwrapAxiosData(response);
  if (Array.isArray(raw)) {
    return raw;
  }
  if (!raw || typeof raw !== 'object') {
    return [];
  }
  const body = raw as Record<string, unknown>;
  if (body.success === true && Array.isArray(body.data)) {
    return body.data;
  }
  if (Array.isArray(body.data)) {
    return body.data;
  }
  return [];
}

function bytesToGB(bytes: number): number {
  return bytes / (1024 * 1024 * 1024);
}

function shortName(name: string, max = 10): string {
  if (!name) return '';
  return name.length > max ? `${name.slice(0, max)}...` : name;
}

const loading = ref(true);
const stats = reactive<CapacityStats>({
  totalDatabases: 0,
  totalTables: 0,
  totalDataSize: '0 B',
  totalRows: 0,
  dailyGrowth: '0 B',
  dailyGrowthRows: 0,
});

const databaseList = ref<any[]>([]);
const tableList = ref<any[]>([]);
const fragmentationList = ref<any[]>([]);
const tableRowsList = ref<any[]>([]);

const databaseChartData = computed(() => {
  const sorted = [...databaseList.value].sort(
    (a, b) => (b.dataSizeBytes || 0) - (a.dataSizeBytes || 0),
  );
  return sorted.slice(0, 10).map((item) => {
    const bytes = item.dataSizeBytes || 0;
    return {
      name: shortName(String(item.databaseName ?? '')),
      value: bytes > 0 ? bytesToGB(bytes) : 0,
      raw: item,
    };
  });
});

const tableChartData = computed(() => {
  const sorted = [...tableList.value].sort(
    (a, b) => (b.dataSizeBytes || 0) - (a.dataSizeBytes || 0),
  );
  return sorted.slice(0, 10).map((item) => {
    const bytes = item.dataSizeBytes || 0;
    return {
      name: shortName(String(item.tableName ?? '')),
      value: bytes > 0 ? bytesToGB(bytes) : 0,
      raw: item,
    };
  });
});

const fragmentationChartData = computed(() => {
  const sorted = [...fragmentationList.value].sort(
    (a, b) => (b.fragmentationRateValue || 0) - (a.fragmentationRateValue || 0),
  );
  return sorted.slice(0, 10).map((item: any) => ({
    name: shortName(String(item.tableName ?? '')),
    value: item.fragmentationRateValue || 0,
    raw: item,
  }));
});

const tableRowsChartData = computed(() => {
  const sorted = [...tableRowsList.value].sort(
    (a, b) => (b.rowCountValue || 0) - (a.rowCountValue || 0),
  );
  return sorted.slice(0, 10).map((item: any) => ({
    name: shortName(String(item.tableName ?? '')),
    value: item.rowCountValue || 0,
    raw: item,
  }));
});

const dbBarRef = ref<EchartsUIType>();
const tableBarRef = ref<EchartsUIType>();
const fragBarRef = ref<EchartsUIType>();
const rowsBarRef = ref<EchartsUIType>();

const { renderEcharts: renderDbBar } = useEcharts(dbBarRef);
const { renderEcharts: renderTableBar } = useEcharts(tableBarRef);
const { renderEcharts: renderFragBar } = useEcharts(fragBarRef);
const { renderEcharts: renderRowsBar } = useEcharts(rowsBarRef);

function emptyChart(
  render: (o: Record<string, unknown>) => Promise<unknown> | unknown,
) {
  return render({
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
}

function renderGbBar(
  render: (o: Record<string, unknown>) => Promise<unknown> | unknown,
  list: { name: string; value: number; raw: any }[],
  paletteKey: 'database' | 'table',
) {
  if (!list.length) {
    return emptyChart(render);
  }
  const names = list.map((d) => d.name);
  const values = list.map((d) => d.value);
  const pal = BAR_PALETTE[paletteKey];
  return render({
    grid: { ...CHART_GRID },
    series: [
      {
        barMaxWidth: 40,
        data: values,
        emphasis: {
          focus: 'series',
          itemStyle: {
            shadowBlur: 12,
            shadowColor: pal.emphasis,
          },
        },
        itemStyle: {
          borderRadius: [6, 6, 0, 0],
          color: barLinearGradient(pal.top, pal.bottom),
        },
        label: {
          fontSize: 11,
          fontWeight: 500,
          formatter: (p: { value?: number }) => {
            const v = p.value ?? 0;
            if (v === 0 || Number.isNaN(v)) return '0.00 GB';
            return `${v.toFixed(2)} GB`;
          },
          position: 'top',
          show: true,
        },
        type: 'bar',
      },
    ],
    tooltip: {
      ...CHART_TOOLTIP_BASE,
      formatter: (params: any) => {
        const p = Array.isArray(params) ? params[0] : params;
        const idx = p?.dataIndex ?? 0;
        const row = list[idx]?.raw;
        if (!row) return '';
        if (row.databaseName !== undefined) {
          return `${row.databaseName}<br/>类型: ${row.datasourceType || '-'}<br/>主机: ${row.host || '-'}<br/>端口: ${row.port || '-'}<br/>容量: ${row.dataSize || '-'}`;
        }
        const bytes = row.dataSizeBytes ?? 0;
        const fmt = formatBytes(bytes);
        return `${row.tableName || ''}<br/>数据大小: ${fmt}`;
      },
      trigger: 'axis',
    },
    xAxis: {
      axisLabel: { fontSize: 11, interval: 0, rotate: 28 },
      axisLine: { lineStyle: { type: 'solid' as const, width: 1 } },
      axisTick: { show: false },
      data: names,
      type: 'category',
    },
    yAxis: {
      name: '容量 (GB)',
      nameGap: 12,
      nameTextStyle: { fontSize: 12, fontWeight: 500 },
      axisLabel: {
        fontSize: 11,
        formatter: (v: string) => {
          const n = Number.parseFloat(v);
          return Number.isNaN(n) ? '0.00 GB' : `${n.toFixed(2)} GB`;
        },
      },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        lineStyle: { opacity: 0.65, type: 'dashed' as const, width: 1 },
      },
      type: 'value',
    },
  });
}

type EchartsRender = (o: Record<string, unknown>) => Promise<unknown> | unknown;

function formatBytes(b: number): string {
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(2)} KB`;
  if (b < 1024 * 1024 * 1024) return `${(b / (1024 * 1024)).toFixed(2)} MB`;
  if (b < 1024 * 1024 * 1024 * 1024) return `${(b / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  return `${(b / (1024 * 1024 * 1024 * 1024)).toFixed(2)} TB`;
}

function renderFragBarChart(render: EchartsRender, list: { name: string; value: number; raw: any }[]) {
  if (!list.length) {
    return emptyChart(render);
  }
  const names = list.map((d) => d.name);
  const values = list.map((d) => d.value);
  const pal = BAR_PALETTE.fragmentation;
  return render({
    grid: { ...CHART_GRID },
    series: [
      {
        barMaxWidth: 40,
        data: values,
        emphasis: {
          focus: 'series',
          itemStyle: {
            shadowBlur: 12,
            shadowColor: pal.emphasis,
          },
        },
        itemStyle: {
          borderRadius: [6, 6, 0, 0],
          color: barLinearGradient(pal.top, pal.bottom),
        },
        label: {
          fontSize: 11,
          fontWeight: 500,
          formatter: (p: { value?: number }) => {
            const v = p.value ?? 0;
            if (v === 0 || Number.isNaN(v)) return '0.00%';
            return `${v.toFixed(2)}%`;
          },
          position: 'top',
          show: true,
        },
        type: 'bar',
      },
    ],
    tooltip: {
      ...CHART_TOOLTIP_BASE,
      formatter: (params: any) => {
        const p = Array.isArray(params) ? params[0] : params;
        const idx = p?.dataIndex ?? 0;
        const row = list[idx]?.raw;
        if (!row) return '';
        return `${row.tableName || ''}<br/>库: ${row.databaseName || '-'}<br/>类型: ${row.datasourceType || '-'}<br/>主机: ${row.host || '-'} 端口: ${row.port || '-'}<br/>碎片率: ${row.fragmentationRate || '-'}`;
      },
      trigger: 'axis',
    },
    xAxis: {
      axisLabel: { fontSize: 11, interval: 0, rotate: 28 },
      axisLine: { lineStyle: { type: 'solid' as const, width: 1 } },
      axisTick: { show: false },
      data: names,
      type: 'category',
    },
    yAxis: {
      name: '碎片率 (%)',
      nameGap: 12,
      nameTextStyle: { fontSize: 12, fontWeight: 500 },
      axisLabel: {
        fontSize: 11,
        formatter: (v: string) => {
          const n = Number.parseFloat(v);
          return Number.isNaN(n) ? '0.00%' : `${n.toFixed(2)}%`;
        },
      },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        lineStyle: { opacity: 0.65, type: 'dashed' as const, width: 1 },
      },
      type: 'value',
    },
  });
}

function formatRowAxis(v: number): string {
  if (v >= 1_000_000_000) return `${(v / 1_000_000_000).toFixed(2)}B`;
  if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(2)}M`;
  if (v >= 1000) return `${(v / 1000).toFixed(2)}K`;
  return `${v}`;
}

function renderRowsBarChart(render: EchartsRender, list: { name: string; value: number; raw: any }[]) {
  if (!list.length) {
    return emptyChart(render);
  }
  const names = list.map((d) => d.name);
  const values = list.map((d) => d.value);
  const pal = BAR_PALETTE.rows;
  return render({
    grid: { ...CHART_GRID },
    series: [
      {
        barMaxWidth: 40,
        data: values,
        emphasis: {
          focus: 'series',
          itemStyle: {
            shadowBlur: 12,
            shadowColor: pal.emphasis,
          },
        },
        itemStyle: {
          borderRadius: [6, 6, 0, 0],
          color: barLinearGradient(pal.top, pal.bottom),
        },
        label: {
          fontSize: 11,
          fontWeight: 500,
          formatter: (p: { value?: number }) => {
            const v = p.value ?? 0;
            return formatRowAxis(v);
          },
          position: 'top',
          show: true,
        },
        type: 'bar',
      },
    ],
    tooltip: {
      ...CHART_TOOLTIP_BASE,
      formatter: (params: any) => {
        const p = Array.isArray(params) ? params[0] : params;
        const idx = p?.dataIndex ?? 0;
        const row = list[idx]?.raw;
        if (!row) return '';
        return `${row.tableName || ''}<br/>库: ${row.databaseName || '-'}<br/>类型: ${row.datasourceType || '-'}<br/>主机: ${row.host || '-'} 端口: ${row.port || '-'}<br/>记录数: ${row.rowCount || '-'}`;
      },
      trigger: 'axis',
    },
    xAxis: {
      axisLabel: { fontSize: 11, interval: 0, rotate: 28 },
      axisLine: { lineStyle: { type: 'solid' as const, width: 1 } },
      axisTick: { show: false },
      data: names,
      type: 'category',
    },
    yAxis: {
      name: '记录数',
      nameGap: 12,
      nameTextStyle: { fontSize: 12, fontWeight: 500 },
      axisLabel: {
        fontSize: 11,
        formatter: (v: string) => formatRowAxis(Number.parseFloat(v) || 0),
      },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        lineStyle: { opacity: 0.65, type: 'dashed' as const, width: 1 },
      },
      type: 'value',
    },
  });
}

async function renderAllCharts() {
  await Promise.all([
    Promise.resolve(renderGbBar(renderDbBar, databaseChartData.value, 'database')),
    Promise.resolve(renderGbBar(renderTableBar, tableChartData.value, 'table')),
    Promise.resolve(renderFragBarChart(renderFragBar, fragmentationChartData.value)),
    Promise.resolve(renderRowsBarChart(renderRowsBar, tableRowsChartData.value)),
  ]);
}

function applyStats(sd: Partial<CapacityStats> | null) {
  if (!sd) return;
  stats.totalDatabases = Number(sd.totalDatabases) || 0;
  stats.totalTables = Number(sd.totalTables) || 0;
  stats.totalDataSize = String(sd.totalDataSize ?? '0 B');
  stats.totalRows = Number(sd.totalRows) || 0;
  stats.dailyGrowth = String(sd.dailyGrowth ?? '0 B');
  stats.dailyGrowthRows = Number(sd.dailyGrowthRows) || 0;
}

async function fetchDashboard() {
  loading.value = true;
  try {
    const results = await Promise.allSettled([
      baseRequestClient.get('/v1/pumpkin/capacity/stats'),
      baseRequestClient.get('/v1/pumpkin/capacity/database/top10/chart'),
      baseRequestClient.get('/v1/pumpkin/capacity/table/top10'),
      baseRequestClient.get('/v1/pumpkin/capacity/table/fragmentation/top10'),
      baseRequestClient.get('/v1/pumpkin/capacity/table/rows/top10'),
    ]);

    const [statsR, dbR, tableR, fragR, rowsR] = results;

    if (statsR.status === 'fulfilled') {
      applyStats(parsePumpkinStatsPayload(statsR.value));
    }
    databaseList.value =
      dbR.status === 'fulfilled' ? parsePumpkinListPayload(dbR.value) : [];
    tableList.value =
      tableR.status === 'fulfilled' ? parsePumpkinListPayload(tableR.value) : [];
    fragmentationList.value =
      fragR.status === 'fulfilled' ? parsePumpkinListPayload(fragR.value) : [];
    tableRowsList.value =
      rowsR.status === 'fulfilled' ? parsePumpkinListPayload(rowsR.value) : [];
  } catch {
    stats.totalDatabases = 0;
    stats.totalTables = 0;
    stats.totalDataSize = '0 B';
    stats.totalRows = 0;
    stats.dailyGrowth = '0 B';
    stats.dailyGrowthRows = 0;
    databaseList.value = [];
    tableList.value = [];
    fragmentationList.value = [];
    tableRowsList.value = [];
  } finally {
    loading.value = false;
    await nextTick();
    await renderAllCharts();
    requestAnimationFrame(() => {
      void renderAllCharts();
    });
  }
}

onMounted(fetchDashboard);
</script>

<template>
  <div class="capacity-dashboard p-5">
    <div class="capacity-dashboard__stats mb-6">
      <Title :level="5" class="!mb-4 text-foreground/90">核心指标</Title>
      <Spin :spinning="loading">
        <Row :gutter="[16, 16]">
          <Col :xs="24" :sm="12" :md="4">
            <Card size="small">
              <Statistic title="数据库数量" :value="stats.totalDatabases">
                <template #prefix>
                  <IconDatabase class="text-[#1890ff]" />
                </template>
              </Statistic>
            </Card>
          </Col>
          <Col :xs="24" :sm="12" :md="4">
            <Card size="small">
              <Statistic title="数据表数量" :value="stats.totalTables">
                <template #prefix>
                  <IconTable class="text-[#52c41a]" />
                </template>
              </Statistic>
            </Card>
          </Col>
          <Col :xs="24" :sm="12" :md="4">
            <Card size="small">
              <Statistic title="总数据容量" :value="stats.totalDataSize">
                <template #prefix>
                  <IconHdd class="text-[#faad14]" />
                </template>
              </Statistic>
            </Card>
          </Col>
          <Col :xs="24" :sm="12" :md="4">
            <Card size="small">
              <Statistic title="总数据记录" :value="stats.totalRows">
                <template #prefix>
                  <IconTable class="text-[#722ed1]" />
                </template>
              </Statistic>
            </Card>
          </Col>
          <Col :xs="24" :sm="12" :md="4">
            <Card size="small">
              <Statistic title="天增长数据量" :value="stats.dailyGrowth">
                <template #prefix>
                  <IconTrending class="text-[#f5222d]" />
                </template>
              </Statistic>
            </Card>
          </Col>
          <Col :xs="24" :sm="12" :md="4">
            <Card size="small">
              <Statistic title="天增长记录数" :value="stats.dailyGrowthRows">
                <template #prefix>
                  <IconTrending class="text-[#52c41a]" />
                </template>
              </Statistic>
            </Card>
          </Col>
        </Row>
      </Spin>
    </div>

    <div class="mt-2 grid grid-cols-1 gap-4 md:grid-cols-2">
      <AnalysisChartCard title="数据库容量 TOP 排行">
        <EchartsUI ref="dbBarRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="数据表容量 TOP 排行">
        <EchartsUI ref="tableBarRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="表碎片率 TOP 排行">
        <EchartsUI ref="fragBarRef" class="h-[320px]" />
      </AnalysisChartCard>
      <AnalysisChartCard title="表记录数 TOP 排行">
        <EchartsUI ref="rowsBarRef" class="h-[320px]" />
      </AnalysisChartCard>
    </div>
  </div>
</template>

<style scoped>
.capacity-dashboard__stats {
  background: var(--ant-color-fill-quaternary, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--ant-color-border-secondary, #f0f0f0);
  border-radius: 8px;
  padding: 16px 16px 8px;
}
</style>
