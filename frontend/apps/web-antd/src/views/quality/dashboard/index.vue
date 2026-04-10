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
    { type: '准确', value: 14_470 },
    { type: '格式错误', value: 850 },
    { type: '范围错误', value: 360 },
  ],
  aiAnalysis: {
    analysisTime: '2024-01-15 10:30:00',
    insights: [
      '整体数据质量评分为88.2分，处于良好水平',
      '数据唯一性表现最佳，达到94.1%',
      '字段准确性需要重点关注，存在格式和范围错误',
      '建议优先处理高优先级问题，预计可提升整体质量5-8分',
    ],
    overallLevel: '良好',
    overallScore: 88.2,
    recommendations: [
      {
        desc: '检测到1980个字段存在数据缺失，建议优先处理user_info表的email字段，空值率超过20%',
        priority: '高',
        title: '字段完整性待提升',
        type: 'high',
      },
      {
        desc: '发现850个字段存在格式错误，主要集中在phone、email等联系信息字段，建议统一格式规范',
        priority: '中',
        title: '数据格式规范性问题',
        type: 'medium',
      },
      {
        desc: '部分关联表数据存在不一致情况，建议检查外键约束和数据同步机制',
        priority: '低',
        title: '数据一致性优化',
        type: 'low',
      },
    ],
    trendAnalysis: '近30天数据质量呈上升趋势，较上月提升2.3%',
  },
  completenessData: [
    { type: '完整', value: 13_700 },
    { type: '缺失', value: 1980 },
  ],
  consistencyData: [
    { type: '一致', value: 13_420 },
    { type: '不一致', value: 2260 },
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
    { type: '唯一', value: 14_750 },
    { type: '重复', value: 930 },
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
  if (level === 'high') return '高';
  if (level === 'medium') return '中';
  if (level === 'low') return '低';
  return '未知';
}

function normalizePieData(data?: PieItem[]) {
  if (!data || data.length === 0) return [{ type: '暂无数据', value: 1 }];
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
    '字段完整性分布',
    dashboardData.value.completenessData,
    ['#52c41a', '#ff4d4f'],
  );
  renderPie(
    renderAccuracy,
    '字段准确性分布',
    dashboardData.value.accuracyData,
    ['#1890ff', '#faad14', '#ff4d4f'],
  );
  renderPie(
    renderConsistency,
    '数据一致性分布',
    dashboardData.value.consistencyData,
    ['#52c41a', '#ff7875'],
  );
  renderPie(
    renderUniqueness,
    '数据唯一性分布',
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
  { dataIndex: 'tableName', key: 'tableName', title: '表名', width: 150 },
  { dataIndex: 'columnName', key: 'columnName', title: '字段名', width: 150 },
  { dataIndex: 'issueType', key: 'issueType', title: '问题类型', width: 120 },
  { dataIndex: 'issueLevel', key: 'issueLevel', title: '严重程度', width: 100 },
  { dataIndex: 'issueDesc', key: 'issueDesc', title: '问题描述' },
  { dataIndex: 'issueCount', key: 'issueCount', title: '问题数量', width: 100 },
  { dataIndex: 'lastCheckTime', key: 'lastCheckTime', title: '最后检查时间', width: 180 },
  { key: 'aiAnalysis', title: 'AI分析', width: 120 },
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
    <Card class="mb-4" :bordered="false">
      <Row :gutter="24">
        <Col :md="8" :xs="24">
          <div class="ai-score-card">
            <div class="text-sm text-gray-500">AI综合评分</div>
            <div class="mt-2 text-3xl font-semibold">
              {{ dashboardData.aiAnalysis?.overallScore || 0 }}<span class="text-base">分</span>
            </div>
            <div class="mt-2">
              <Tag :color="getQualityColor(dashboardData.aiAnalysis?.overallScore || 0)">
                {{ dashboardData.aiAnalysis?.overallLevel || '良好' }}
              </Tag>
            </div>
            <div class="mt-2 text-xs text-gray-500">分析时间：{{ dashboardData.aiAnalysis?.analysisTime || '--' }}</div>
          </div>
        </Col>
        <Col :md="16" :xs="24">
          <div class="text-sm font-medium">AI智能洞察</div>
          <div class="mt-2 space-y-2 text-sm">
            <div v-for="(insight, index) in dashboardData.aiAnalysis?.insights || []" :key="index">
              {{ insight }}
            </div>
          </div>
          <div class="mt-3 text-sm text-green-600">{{ dashboardData.aiAnalysis?.trendAnalysis || '' }}</div>
        </Col>
      </Row>
    </Card>

    <Card title="AI优化建议" class="mb-4" :bordered="false">
      <Row :gutter="[16, 16]">
        <Col
          v-for="(rec, index) in dashboardData.aiAnalysis?.recommendations || []"
          :key="index"
          :lg="8"
          :md="12"
          :xs="24"
        >
          <Card :bordered="true" size="small">
            <Tag :color="rec.type === 'high' ? 'red' : rec.type === 'medium' ? 'orange' : 'blue'">
              {{ rec.priority }}优先级
            </Tag>
            <div class="mt-2 font-medium">{{ rec.title }}</div>
            <div class="mt-2 text-sm text-gray-500">{{ rec.desc }}</div>
          </Card>
        </Col>
      </Row>
    </Card>

    <Row :gutter="[16, 16]" class="mb-4">
      <Col :lg="6" :md="12" :xs="24"><Card><Statistic title="数据表总数" :value="dashboardData.totalTables || 0" /></Card></Col>
      <Col :lg="6" :md="12" :xs="24"><Card><Statistic title="数据字段总数" :value="dashboardData.totalColumns || 0" /></Card></Col>
      <Col :lg="6" :md="12" :xs="24"><Card><Statistic title="质量问题总数" :value="dashboardData.totalIssues || 0" /></Card></Col>
      <Col :lg="6" :md="12" :xs="24">
        <Card>
          <Statistic title="整体质量评分" :value="overallScore" suffix="/ 100" :value-style="{ color: getQualityColor(overallScore) }" />
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mb-4">
      <Col v-for="item in [
        { key: 'fieldCompleteness', label: '字段完整性', desc: '字段数据完整率，反映数据缺失情况' },
        { key: 'fieldAccuracy', label: '字段准确性', desc: '字段数据准确率，反映格式和范围正确性' },
        { key: 'tableCompleteness', label: '表完整性', desc: '表结构完整性，反映表结构规范性' },
        { key: 'dataConsistency', label: '数据一致性', desc: '跨表数据一致性，反映关联数据正确性' },
        { key: 'dataUniqueness', label: '数据唯一性', desc: '主键和业务唯一性，反映重复数据情况' },
        { key: 'dataTimeliness', label: '数据及时性', desc: '数据更新及时性，反映数据新鲜度' },
      ]" :key="item.key" :lg="8" :md="12" :xs="24">
        <Card :title="item.label">
          <div class="text-2xl font-semibold">{{ (dashboardData as any)[item.key] || 0 }}%</div>
          <Progress :percent="(dashboardData as any)[item.key] || 0" :show-info="false" :stroke-color="getQualityColor((dashboardData as any)[item.key] || 0)" />
          <div class="mt-2 text-xs text-gray-500">{{ item.desc }}</div>
        </Card>
      </Col>
    </Row>

    <Row :gutter="[16, 16]" class="mb-4">
      <Col :md="12" :xs="24"><Card title="字段完整性分布"><EchartsUI ref="completenessChartRef" class="h-[300px]" /></Card></Col>
      <Col :md="12" :xs="24"><Card title="字段准确性分布"><EchartsUI ref="accuracyChartRef" class="h-[300px]" /></Card></Col>
    </Row>
    <Row :gutter="[16, 16]" class="mb-4">
      <Col :md="12" :xs="24"><Card title="数据一致性分布"><EchartsUI ref="consistencyChartRef" class="h-[300px]" /></Card></Col>
      <Col :md="12" :xs="24"><Card title="数据唯一性分布"><EchartsUI ref="uniquenessChartRef" class="h-[300px]" /></Card></Col>
    </Row>

    <Card title="质量问题列表" :bordered="false">
      <Table
        :columns="issueColumns"
        :data-source="dashboardData.issueList || []"
        :loading="loading"
        :pagination="{ pageSize: 10, showSizeChanger: true, showTotal: (total:number) => `共 ${total} 条问题` }"
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
              :title="`AI评估：该问题属于${getIssueLevelText(record.issueLevel || '')}优先级`"
            >
              <Tag color="blue">AI分析</Tag>
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

<style scoped>
.ai-score-card {
  border-right: 1px solid #f0f0f0;
  height: 100%;
  padding-right: 12px;
}

@media (max-width: 768px) {
  .ai-score-card {
    border-right: none;
    margin-bottom: 12px;
    padding-right: 0;
  }
}
</style>
