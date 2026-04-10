<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { nextTick, onMounted, ref } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { Card, Col, Progress, Row, Statistic } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

interface PieItem {
  type: string;
  value: number;
}

interface QualityData {
  columnAccuracyRate?: number | string;
  columnCommentAccuracyDataList?: PieItem[];
  columnCommentRate?: number | string;
  columnCount?: number | string;
  columnQualityDataList?: PieItem[];
  databaseBusinessRate?: number | string;
  databaseCount?: number | string;
  databaseQualityDataList?: PieItem[];
  tableAccuracyRate?: number | string;
  tableCommentAccuracyDataList?: PieItem[];
  tableCommentRate?: number | string;
  tableCount?: number | string;
  tableQualityDataList?: PieItem[];
}

const loading = ref(true);
const qualityData = ref<QualityData>({});

const dbQualityChartRef = ref<EchartsUIType>();
const tableQualityChartRef = ref<EchartsUIType>();
const columnQualityChartRef = ref<EchartsUIType>();
const tableAccChartRef = ref<EchartsUIType>();
const columnAccChartRef = ref<EchartsUIType>();

const { renderEcharts: renderDbQuality } = useEcharts(dbQualityChartRef);
const { renderEcharts: renderTableQuality } = useEcharts(tableQualityChartRef);
const { renderEcharts: renderColumnQuality } = useEcharts(columnQualityChartRef);
const { renderEcharts: renderTableAcc } = useEcharts(tableAccChartRef);
const { renderEcharts: renderColumnAcc } = useEcharts(columnAccChartRef);

function toNumber(value: number | string | undefined) {
  if (typeof value === 'number') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function normalizePieData(data?: PieItem[]) {
  if (!data || data.length === 0) return [{ type: '暂无数据', value: 1 }];
  return data.map((item) => ({
    type: item.type,
    value: toNumber(item.value),
  }));
}

function getQualityColor(rate: number) {
  if (rate >= 80) return '#52c41a';
  if (rate >= 60) return '#faad14';
  return '#ff4d4f';
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
        data: pieData.map((item) => ({ name: item.type, value: item.value })),
        itemStyle: {
          borderRadius: 8,
          borderWidth: 2,
        },
        label: {
          formatter: '{b}: {d}%',
        },
        name: title,
        radius: ['45%', '70%'],
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
    renderDbQuality,
    '数据库业务关联情况',
    qualityData.value.databaseQualityDataList,
  );
  renderPie(renderTableQuality, '数据表注释完备情况', qualityData.value.tableQualityDataList);
  renderPie(
    renderColumnQuality,
    '数据字段注释完备情况',
    qualityData.value.columnQualityDataList,
  );
  renderPie(
    renderTableAcc,
    '表注释准确度分布',
    qualityData.value.tableCommentAccuracyDataList,
  );
  renderPie(
    renderColumnAcc,
    '字段注释准确度分布',
    qualityData.value.columnCommentAccuracyDataList,
  );
}

function resolveQualityData(response: any): QualityData {
  if (response?.data?.databaseCount !== undefined) return response.data as QualityData;
  if (response?.databaseCount !== undefined) return response as QualityData;
  if (response?.data?.data) return response.data.data as QualityData;
  return {};
}

async function fetchQualityData() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/quality/info');
    qualityData.value = resolveQualityData(response);
  } finally {
    loading.value = false;
    await nextTick();
    renderAllCharts();
  }
}

onMounted(fetchQualityData);
</script>

<template>
  <div class="p-5">
    <Row :gutter="[16, 16]">
      <Col :md="8" :xs="24">
        <Card :loading="loading">
          <Statistic title="数据库总数" :value="toNumber(qualityData.databaseCount)" />
        </Card>
      </Col>
      <Col :md="8" :xs="24">
        <Card :loading="loading">
          <Statistic title="数据表总数" :value="toNumber(qualityData.tableCount)" />
        </Card>
      </Col>
      <Col :md="8" :xs="24">
        <Card :loading="loading">
          <Statistic title="数据字段总数" :value="toNumber(qualityData.columnCount)" />
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mt-4">
      <Col :md="8" :xs="24">
        <Card title="数据库业务关联率" :loading="loading">
          <div class="circle-wrap">
            <Progress
              type="circle"
              :percent="toNumber(qualityData.databaseBusinessRate)"
              :stroke-color="getQualityColor(toNumber(qualityData.databaseBusinessRate))"
            />
          </div>
        </Card>
      </Col>
      <Col :md="8" :xs="24">
        <Card title="数据表注释完备率" :loading="loading">
          <div class="circle-wrap">
            <Progress
              type="circle"
              :percent="toNumber(qualityData.tableCommentRate)"
              :stroke-color="getQualityColor(toNumber(qualityData.tableCommentRate))"
            />
          </div>
        </Card>
      </Col>
      <Col :md="8" :xs="24">
        <Card title="数据字段注释完备率" :loading="loading">
          <div class="circle-wrap">
            <Progress
              type="circle"
              :percent="toNumber(qualityData.columnCommentRate)"
              :stroke-color="getQualityColor(toNumber(qualityData.columnCommentRate))"
            />
          </div>
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mt-4">
      <Col :md="12" :xs="24">
        <Card title="数据表备注准确度" :loading="loading">
          <div class="circle-wrap">
            <Progress
              type="circle"
              :percent="toNumber(qualityData.tableAccuracyRate)"
              :stroke-color="getQualityColor(toNumber(qualityData.tableAccuracyRate))"
            />
          </div>
        </Card>
      </Col>
      <Col :md="12" :xs="24">
        <Card title="数据字段备注准确度" :loading="loading">
          <div class="circle-wrap">
            <Progress
              type="circle"
              :percent="toNumber(qualityData.columnAccuracyRate)"
              :stroke-color="getQualityColor(toNumber(qualityData.columnAccuracyRate))"
            />
          </div>
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mt-4">
      <Col :md="12" :xs="24">
        <Card title="数据库业务关联情况" :loading="loading">
          <EchartsUI ref="dbQualityChartRef" class="h-[330px]" />
        </Card>
      </Col>
      <Col :md="12" :xs="24">
        <Card title="数据表注释完备情况" :loading="loading">
          <EchartsUI ref="tableQualityChartRef" class="h-[330px]" />
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mt-4">
      <Col :md="12" :xs="24">
        <Card title="数据字段注释完备情况" :loading="loading">
          <EchartsUI ref="columnQualityChartRef" class="h-[330px]" />
        </Card>
      </Col>
      <Col :md="12" :xs="24">
        <Card title="表注释准确度分布" :loading="loading">
          <EchartsUI ref="tableAccChartRef" class="h-[330px]" />
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mt-4">
      <Col :md="12" :xs="24">
        <Card title="字段注释准确度分布" :loading="loading">
          <EchartsUI ref="columnAccChartRef" class="h-[330px]" />
        </Card>
      </Col>
    </Row>
  </div>
</template>

<style scoped>
.circle-wrap {
  display: flex;
  justify-content: center;
  padding: 16px 0;
}
</style>
