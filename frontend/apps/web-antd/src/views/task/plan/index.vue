<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Table,
  Tooltip,
  Typography,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';
import { checkPermission } from '#/utils/check-permission';

defineOptions({ name: 'TaskPlanPage' });

const { Text } = Typography;

interface TaskOptionRow {
  task_key: string;
  task_name: string;
  task_description: string;
  crontab: string;
  enable: number;
  last_run_time?: string;
  next_run_time?: string;
  gmt_created?: string;
  gmt_updated?: string;
}

interface TrendPoint {
  x?: string;
  y?: number;
}

interface TaskStatsState {
  todayExecuteCount: number;
  todayFailedCount: number;
  hour24ExecuteCount: number;
  successRateStr: string;
  todayTrend: TrendPoint[];
  todayFailedTrend: TrendPoint[];
  hour24Trend: TrendPoint[];
  successRateTrend: TrendPoint[];
  lastExecuteTime: string;
}

interface TaskLogRow {
  id: number;
  task_key: string;
  start_time: string;
  complete_time?: string;
  status: string;
  result: string;
  gmt_created: string;
}

function unwrapAxiosData(response: unknown): unknown {
  if (!response || typeof response !== 'object') {
    return response;
  }
  const r = response as Record<string, unknown>;
  if ('data' in r && 'status' in r && typeof r.status === 'number') {
    return r.data;
  }
  return response;
}

function parseTaskStats(response: unknown): TaskStatsState | null {
  const raw = unwrapAxiosData(response);
  if (!raw || typeof raw !== 'object') {
    return null;
  }
  const b = raw as Record<string, unknown>;
  if (b.success !== true) {
    return null;
  }
  return {
    todayExecuteCount: Number(b.todayTotal) || 0,
    todayFailedCount: Number(b.todayFailedTotal) || 0,
    hour24ExecuteCount: Number(b.hour24Total) || 0,
    successRateStr: String(b.successRateStr ?? '0%'),
    todayTrend: Array.isArray(b.todayTrend) ? (b.todayTrend as TrendPoint[]) : [],
    todayFailedTrend: Array.isArray(b.todayFailedTrend) ? (b.todayFailedTrend as TrendPoint[]) : [],
    hour24Trend: Array.isArray(b.hour24Trend) ? (b.hour24Trend as TrendPoint[]) : [],
    successRateTrend: Array.isArray(b.successRateTrend) ? (b.successRateTrend as TrendPoint[]) : [],
    lastExecuteTime: String(b.lastExecuteTime ?? ''),
  };
}

function parseTaskList(response: unknown): TaskOptionRow[] {
  const raw = unwrapAxiosData(response);
  if (!raw || typeof raw !== 'object') {
    return [];
  }
  const b = raw as Record<string, unknown>;
  const data = b.data;
  return Array.isArray(data) ? (data as TaskOptionRow[]) : [];
}

const loading = ref(false);
const dataSource = ref<TaskOptionRow[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => $t('page.taskPlan.paginationTotal', { total }),
  total: 0,
});

const queryForm = reactive({
  task_key: '',
  task_name: '',
  enable: undefined as number | undefined,
});

const sortField = ref<string | undefined>();
const sortOrder = ref<'ascend' | 'descend' | undefined>();
const lastSortSignature = ref('');

const stats = reactive<TaskStatsState>({
  todayExecuteCount: 0,
  todayFailedCount: 0,
  hour24ExecuteCount: 0,
  successRateStr: '0%',
  todayTrend: [],
  todayFailedTrend: [],
  hour24Trend: [],
  successRateTrend: [],
  lastExecuteTime: '',
});

const chartTodayRef = ref<EchartsUIType>();
const chartFailedRef = ref<EchartsUIType>();
const chartH24Ref = ref<EchartsUIType>();
const chartRateRef = ref<EchartsUIType>();

const { renderEcharts: renderToday } = useEcharts(chartTodayRef);
const { renderEcharts: renderFailed } = useEcharts(chartFailedRef);
const { renderEcharts: renderH24 } = useEcharts(chartH24Ref);
const { renderEcharts: renderRate } = useEcharts(chartRateRef);

function formatTrendAxisLabel(value: string) {
  if (!value) {
    return '';
  }
  // 近 24 小时后端为 "MM-DD HH:00"，缩成 "HH:00" 避免挤在一起
  if (value.includes(' ')) {
    const parts = value.split(/\s+/);
    return parts[parts.length - 1] ?? value;
  }
  return value;
}

function renderMiniArea(
  render: (o: Record<string, unknown>) => unknown,
  trend: TrendPoint[],
  color: string,
) {
  const xs = trend.map((t) => String(t.x ?? ''));
  const ys = trend.map((t) => Number(t.y) || 0);
  render({
    grid: { bottom: 22, containLabel: false, left: 2, right: 2, top: 4 },
    series: [
      {
        areaStyle: { color, opacity: 0.12 },
        data: ys,
        lineStyle: { color, width: 1.5 },
        smooth: true,
        symbol: 'none',
        type: 'line',
      },
    ],
    xAxis: {
      axisLabel: {
        color: '#8c8c8c',
        fontSize: 10,
        formatter: (v: string) => formatTrendAxisLabel(v),
        hideOverlap: true,
        interval: 3,
      },
      axisLine: { lineStyle: { color: '#f0f0f0' } },
      axisTick: { show: false },
      boundaryGap: false,
      data: xs,
      type: 'category',
    },
    yAxis: { show: false, splitLine: { show: false }, type: 'value' },
  });
}

function renderAllStatCharts() {
  renderMiniArea(renderToday, stats.todayTrend, '#1979C9');
  renderMiniArea(renderFailed, stats.todayFailedTrend, '#F5222D');
  renderMiniArea(renderH24, stats.hour24Trend, '#52C41A');
  renderMiniArea(renderRate, stats.successRateTrend, '#FAAD14');
}

let statsTimer: ReturnType<typeof setInterval> | undefined;

async function fetchStats() {
  try {
    const res = await baseRequestClient.get('/v1/task/today/stats');
    const parsed = parseTaskStats(res);
    if (parsed) {
      Object.assign(stats, parsed);
      await nextTick();
      renderAllStatCharts();
    }
  } catch {
    /* 忽略统计失败 */
  }
}

async function fetchTaskList() {
  loading.value = true;
  try {
    const params: Record<string, string> = {};
    if (queryForm.task_key) {
      params.task_key = queryForm.task_key;
    }
    if (queryForm.task_name) {
      params.task_name = queryForm.task_name;
    }
    if (queryForm.enable !== undefined && queryForm.enable !== null) {
      params.enable = String(queryForm.enable);
    }
    if (sortField.value && sortOrder.value) {
      params.sorter = JSON.stringify({
        [sortField.value]: sortOrder.value,
      });
    }
    const response = await baseRequestClient.get('/v1/task/option', { params });
    const list = parseTaskList(response);
    dataSource.value = list;
    pagination.total = list.length;
  } catch (error: any) {
    message.error(error?.message || $t('page.taskPlan.message.listLoadFailed'));
    dataSource.value = [];
    pagination.total = 0;
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchTaskList();
}

function handleReset() {
  queryForm.task_key = '';
  queryForm.task_name = '';
  queryForm.enable = undefined;
  sortField.value = undefined;
  sortOrder.value = undefined;
  lastSortSignature.value = '';
  pagination.current = 1;
  fetchTaskList();
}

function handleTableChange(pag: any, _filters: unknown, sorter: any) {
  if (pag) {
    pagination.current = pag.current ?? 1;
    pagination.pageSize = pag.pageSize ?? 10;
  }
  let s = sorter;
  if (Array.isArray(sorter)) {
    s = sorter[0];
  }
  const colKey = s?.field ?? s?.columnKey;
  const order = s?.order;
  const sig =
    colKey && order ? `${String(colKey)}:${String(order)}` : '';
  if (sig !== lastSortSignature.value) {
    lastSortSignature.value = sig;
    if (colKey && order) {
      sortField.value = String(colKey);
      sortOrder.value = order;
    } else {
      sortField.value = undefined;
      sortOrder.value = undefined;
    }
    fetchTaskList();
  }
}

const createOpen = ref(false);
const editOpen = ref(false);
const editRecord = ref<TaskOptionRow | null>(null);

const formCreate = reactive({
  task_key: '',
  task_name: '',
  task_description: '',
  crontab: '',
  enable: 1,
});

const formEdit = reactive({
  task_name: '',
  task_description: '',
  crontab: '',
  enable: 1,
});

function openCreate() {
  if (!checkPermission($t('page.taskPlan.message.permissionDenied'))) return;
  formCreate.task_key = '';
  formCreate.task_name = '';
  formCreate.task_description = '';
  formCreate.crontab = '';
  formCreate.enable = 1;
  createOpen.value = true;
}

function openEdit(record: TaskOptionRow) {
  if (!checkPermission($t('page.taskPlan.message.permissionDenied'))) return;
  editRecord.value = record;
  formEdit.task_name = record.task_name;
  formEdit.task_description = record.task_description;
  formEdit.crontab = record.crontab;
  formEdit.enable = Number(record.enable) === 1 ? 1 : 0;
  editOpen.value = true;
}

async function submitCreate() {
  if (!checkPermission($t('page.taskPlan.message.permissionDenied'))) return;
  if (!formCreate.task_key?.trim() || !formCreate.task_name?.trim()) {
    message.warning($t('page.taskPlan.message.keyNameRequired'));
    return;
  }
  try {
    const res = await baseRequestClient.post('/v1/task/option', {
      task_key: formCreate.task_key.trim(),
      task_name: formCreate.task_name.trim(),
      task_description: formCreate.task_description,
      crontab: formCreate.crontab,
      enable: formCreate.enable,
    });
    const raw = unwrapAxiosData(res) as Record<string, unknown>;
    if (raw?.success === false) {
      message.error(String(raw?.msg || $t('page.taskPlan.message.addFailed')));
      return;
    }
    message.success($t('page.taskPlan.message.addSuccess'));
    createOpen.value = false;
    await fetchTaskList();
  } catch (error: any) {
    message.error(error?.message || $t('page.taskPlan.message.addFailed'));
  }
}

async function submitEdit() {
  if (!checkPermission($t('page.taskPlan.message.permissionDenied'))) return;
  const key = editRecord.value?.task_key;
  if (!key) {
    return;
  }
  try {
    const res = await baseRequestClient.put('/v1/task/option', {
      task_key: key,
      task_name: formEdit.task_name,
      task_description: formEdit.task_description,
      crontab: formEdit.crontab,
      enable: formEdit.enable,
    });
    const raw = unwrapAxiosData(res) as Record<string, unknown>;
    if (raw?.success === false) {
      message.error(String(raw?.msg || $t('page.taskPlan.message.updateFailed')));
      return;
    }
    message.success($t('page.taskPlan.message.updateSuccess'));
    editOpen.value = false;
    await fetchTaskList();
  } catch (error: any) {
    message.error(error?.message || $t('page.taskPlan.message.updateFailed'));
  }
}

async function removeTask(taskKey: string) {
  if (!checkPermission($t('page.taskPlan.message.permissionDenied'))) return;
  try {
    const res = await baseRequestClient.delete('/v1/task/option', {
      data: { task_key: taskKey },
    });
    const raw = unwrapAxiosData(res) as Record<string, unknown>;
    if (raw?.success === false) {
      message.error(String(raw?.msg || $t('page.taskPlan.message.deleteFailed')));
      return;
    }
    message.success($t('page.taskPlan.message.deleteSuccess'));
    await fetchTaskList();
  } catch (error: any) {
    message.error(error?.message || $t('page.taskPlan.message.deleteFailed'));
  }
}

function confirmExecute(record: TaskOptionRow) {
  const title =
    record.enable === 1
      ? $t('page.taskPlan.execute.titleEnabled')
      : $t('page.taskPlan.execute.titleDisabled');
  const content =
    record.enable === 1
      ? $t('page.taskPlan.execute.contentEnabled', { name: record.task_name })
      : $t('page.taskPlan.execute.contentDisabled', { name: record.task_name });
  Modal.confirm({
    content,
    onOk: () => executeTask(record.task_key, record.task_name),
    title,
  });
}

async function executeTask(taskKey: string, taskName: string) {
  const hide = message.loading($t('page.taskPlan.message.executing'), 0);
  try {
    const res = await baseRequestClient.post('/v1/task/option/execute', {
      task_key: taskKey,
    });
    const raw = unwrapAxiosData(res) as Record<string, unknown>;
    hide();
    if (raw?.success === true) {
      message.success($t('page.taskPlan.message.executeStarted', { name: taskName }));
    } else {
      message.error(String(raw?.msg || $t('page.taskPlan.message.executeFailed')));
    }
  } catch (error: any) {
    hide();
    message.error(error?.message || $t('page.taskPlan.message.executeFailed'));
  }
}

const logOpen = ref(false);
const logTitle = ref('');
const logTaskKey = ref('');
const logLoading = ref(false);
const logDataSource = ref<TaskLogRow[]>([]);
const logPagination = reactive({
  current: 1,
  pageSize: 10,
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => $t('page.taskPlan.paginationTotal', { total }),
  total: 0,
});

function openLogs(taskKey: string, taskName: string) {
  logTaskKey.value = taskKey;
  logTitle.value = `${taskName} - ${$t('page.taskPlan.log.titleSuffix')}`;
  logPagination.current = 1;
  logOpen.value = true;
  fetchLogs();
}

async function fetchLogs() {
  if (!logTaskKey.value) {
    return;
  }
  logLoading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/task/log', {
      params: {
        currentPage: logPagination.current,
        pageSize: logPagination.pageSize,
        task_key: logTaskKey.value,
      },
    });
    const raw = unwrapAxiosData(response) as Record<string, unknown>;
    const list = raw?.data;
    logDataSource.value = Array.isArray(list) ? (list as TaskLogRow[]) : [];
    logPagination.total = Number(raw?.total ?? 0) || 0;
  } catch (error: any) {
    message.error(error?.message || $t('page.taskPlan.message.logLoadFailed'));
    logDataSource.value = [];
    logPagination.total = 0;
  } finally {
    logLoading.value = false;
  }
}

function handleLogTableChange(pag: any) {
  logPagination.current = pag?.current ?? 1;
  logPagination.pageSize = pag?.pageSize ?? 10;
  fetchLogs();
}

watch(logOpen, (open) => {
  if (!open) {
    logTaskKey.value = '';
  }
});

const columns = [
  { dataIndex: 'task_key', key: 'task_key', sorter: true, title: $t('page.taskPlan.columns.taskKey'), width: 160 },
  { dataIndex: 'task_name', key: 'task_name', sorter: true, title: $t('page.taskPlan.columns.taskName'), width: 140 },
  {
    dataIndex: 'task_description',
    ellipsis: true,
    key: 'task_description',
    title: $t('page.taskPlan.columns.taskDescription'),
    width: 220,
  },
  { dataIndex: 'crontab', key: 'crontab', title: $t('page.taskPlan.columns.crontab'), width: 120 },
  {
    key: 'enable',
    title: $t('page.taskPlan.columns.enable'),
    width: 90,
  },
  { dataIndex: 'last_run_time', key: 'last_run_time', sorter: true, title: $t('page.taskPlan.columns.lastRunTime'), width: 170 },
  { dataIndex: 'next_run_time', key: 'next_run_time', sorter: true, title: $t('page.taskPlan.columns.nextRunTime'), width: 170 },
  { key: 'actions', title: $t('page.taskPlan.columns.actions'), width: 320, fixed: 'right' as const },
];

const logColumns = [
  { dataIndex: 'id', key: 'id', title: $t('page.taskPlan.log.columns.id'), width: 80 },
  { dataIndex: 'start_time', key: 'start_time', title: $t('page.taskPlan.log.columns.startTime'), width: 180 },
  { dataIndex: 'complete_time', key: 'complete_time', title: $t('page.taskPlan.log.columns.completeTime'), width: 180 },
  { dataIndex: 'status', key: 'status', title: $t('page.taskPlan.log.columns.status'), width: 100 },
  { dataIndex: 'result', ellipsis: true, key: 'result', title: $t('page.taskPlan.log.columns.result'), width: 300 },
  { dataIndex: 'gmt_created', key: 'gmt_created', title: $t('page.taskPlan.log.columns.createdAt'), width: 180 },
];

function formatTime(v?: string) {
  return v || '-';
}

function statusBadge(status: string) {
  if (status === 'running') {
    return { color: 'processing', text: $t('page.taskPlan.logStatus.running') };
  }
  if (status === 'success') {
    return { color: 'success', text: $t('page.taskPlan.logStatus.success') };
  }
  if (status === 'failed') {
    return { color: 'error', text: $t('page.taskPlan.logStatus.failed') };
  }
  return { color: 'default', text: status || '-' };
}

onMounted(async () => {
  await fetchStats();
  await fetchTaskList();
  await nextTick();
  renderAllStatCharts();
  statsTimer = setInterval(fetchStats, 5 * 60 * 1000);
});

onUnmounted(() => {
  if (statsTimer) {
    clearInterval(statsTimer);
  }
});
</script>

<template>
  <div class="task-plan p-5">
    <div class="mb-4 text-base font-medium text-foreground/90">{{ $t('page.taskPlan.pageTitle') }}</div>

    <Card class="mb-4" :bordered="false" :title="$t('page.taskPlan.statsCardTitle')">
      <Row :gutter="[16, 16]">
        <Col :lg="6" :md="12" :span="24" :xl="6" :xs="24">
          <Card size="small" class="stat-card">
            <div class="flex items-start justify-between gap-2">
              <div>
                <Text type="secondary" class="text-xs">{{ $t('page.taskPlan.stat.todayExecute') }}</Text>
                <Tooltip :title="$t('page.taskPlan.stat.todayExecuteTip')">
                  <span class="ml-1 cursor-help text-[#8c8c8c]">ⓘ</span>
                </Tooltip>
                <div class="mt-1 text-2xl font-semibold">{{ stats.todayExecuteCount }}</div>
                <div v-if="stats.lastExecuteTime" class="mt-1 text-xs text-[#8c8c8c]">
                  {{ $t('page.taskPlan.stat.lastExecute') }}{{ stats.lastExecuteTime }}
                </div>
              </div>
            </div>
            <EchartsUI ref="chartTodayRef" class="mt-2 w-full" height="78px" width="100%" />
          </Card>
        </Col>
        <Col :lg="6" :md="12" :span="24" :xl="6" :xs="24">
          <Card size="small" class="stat-card">
            <div>
              <Text type="secondary" class="text-xs">{{ $t('page.taskPlan.stat.todayFailed') }}</Text>
              <Tooltip :title="$t('page.taskPlan.stat.todayFailedTip')">
                <span class="ml-1 cursor-help text-[#8c8c8c]">ⓘ</span>
              </Tooltip>
              <div class="mt-1 text-2xl font-semibold">{{ stats.todayFailedCount }}</div>
              <div v-if="stats.lastExecuteTime" class="mt-1 text-xs text-[#8c8c8c]">
                {{ $t('page.taskPlan.stat.lastExecute') }}{{ stats.lastExecuteTime }}
              </div>
            </div>
            <EchartsUI ref="chartFailedRef" class="mt-2 w-full" height="78px" width="100%" />
          </Card>
        </Col>
        <Col :lg="6" :md="12" :span="24" :xl="6" :xs="24">
          <Card size="small" class="stat-card">
            <div>
              <Text type="secondary" class="text-xs">{{ $t('page.taskPlan.stat.hour24') }}</Text>
              <Tooltip :title="$t('page.taskPlan.stat.hour24Tip')">
                <span class="ml-1 cursor-help text-[#8c8c8c]">ⓘ</span>
              </Tooltip>
              <div class="mt-1 text-2xl font-semibold">{{ stats.hour24ExecuteCount }}</div>
              <div v-if="stats.lastExecuteTime" class="mt-1 text-xs text-[#8c8c8c]">
                {{ $t('page.taskPlan.stat.lastExecute') }}{{ stats.lastExecuteTime }}
              </div>
            </div>
            <EchartsUI ref="chartH24Ref" class="mt-2 w-full" height="78px" width="100%" />
          </Card>
        </Col>
        <Col :lg="6" :md="12" :span="24" :xl="6" :xs="24">
          <Card size="small" class="stat-card">
            <div>
              <Text type="secondary" class="text-xs">{{ $t('page.taskPlan.stat.successRate') }}</Text>
              <Tooltip :title="$t('page.taskPlan.stat.successRateTip')">
                <span class="ml-1 cursor-help text-[#8c8c8c]">ⓘ</span>
              </Tooltip>
              <div class="mt-1 text-2xl font-semibold">{{ stats.successRateStr }}</div>
              <div v-if="stats.lastExecuteTime" class="mt-1 text-xs text-[#8c8c8c]">
                {{ $t('page.taskPlan.stat.lastExecute') }}{{ stats.lastExecuteTime }}
              </div>
            </div>
            <EchartsUI ref="chartRateRef" class="mt-2 w-full" height="78px" width="100%" />
          </Card>
        </Col>
      </Row>
    </Card>

    <Card :title="$t('page.taskPlan.dataList')">
      <Form class="task-query-form" layout="inline">
        <Form.Item :label="$t('page.taskPlan.form.taskKey')">
          <Input
            v-model:value="queryForm.task_key"
            allow-clear
            class="task-query-input"
            :placeholder="$t('page.taskPlan.placeholder.taskKey')"
          />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.taskName')">
          <Input
            v-model:value="queryForm.task_name"
            allow-clear
            class="task-query-input"
            :placeholder="$t('page.taskPlan.placeholder.taskName')"
          />
        </Form.Item>
        <Form.Item class="task-query-enable-item" :label="$t('page.taskPlan.form.enable')">
          <Select
            v-model:value="queryForm.enable"
            allow-clear
            class="task-query-enable-select"
            :placeholder="$t('page.taskPlan.placeholder.all')"
          >
            <Select.Option :value="1">{{ $t('page.taskPlan.enabled.on') }}</Select.Option>
            <Select.Option :value="0">{{ $t('page.taskPlan.enabled.off') }}</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item class="task-query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
            <Button type="primary" @click="openCreate">{{ $t('page.taskPlan.action.new') }}</Button>
          </Space>
        </Form.Item>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(r: TaskOptionRow) => r.task_key"
        :scroll="{ x: 1400 }"
        size="middle"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'enable'">
            <span :style="{ color: record.enable === 1 ? '#52c41a' : '#8c8c8c' }">
              {{ record.enable === 1 ? $t('page.taskPlan.enabled.on') : $t('page.taskPlan.enabled.off') }}
            </span>
          </template>
          <template v-else-if="column.key === 'last_run_time'">
            {{ formatTime(record.last_run_time) }}
          </template>
          <template v-else-if="column.key === 'next_run_time'">
            {{ formatTime(record.next_run_time) }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <Space>
              <Button size="small" type="link" @click="openEdit(record as TaskOptionRow)">{{ $t('page.taskPlan.action.edit') }}</Button>
              <Popconfirm
                placement="left"
                :title="$t('page.taskPlan.confirmDelete')"
                @confirm="removeTask(record.task_key)"
              >
                <Button danger size="small" type="link">{{ $t('page.taskPlan.action.delete') }}</Button>
              </Popconfirm>
              <Button size="small" type="link" @click="confirmExecute(record as TaskOptionRow)">{{ $t('page.taskPlan.action.runManual') }}</Button>
              <Button size="small" type="link" @click="openLogs((record as TaskOptionRow).task_key, (record as TaskOptionRow).task_name)">
                {{ $t('page.taskPlan.action.viewLogs') }}
              </Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="createOpen"
      destroy-on-close
      :title="$t('page.taskPlan.modal.createTitle')"
      :footer="null"
      @cancel="createOpen = false"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item :label="$t('page.taskPlan.form.taskKey')" required>
          <Input v-model:value="formCreate.task_key" :placeholder="$t('page.taskPlan.placeholder.taskKeyUnique')" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.taskName')" required>
          <Input v-model:value="formCreate.task_name" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.taskDescription')" required>
          <Input.TextArea v-model:value="formCreate.task_description" :rows="3" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.scheduleCrontab')">
          <Input v-model:value="formCreate.crontab" :placeholder="$t('page.taskPlan.placeholder.cron')" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.enable')">
          <Select v-model:value="formCreate.enable" style="width: 100%">
            <Select.Option :value="1">{{ $t('page.taskPlan.enabled.yes') }}</Select.Option>
            <Select.Option :value="0">{{ $t('page.taskPlan.enabled.no') }}</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item>
          <Button type="primary" block @click="submitCreate">{{ $t('page.taskPlan.action.submit') }}</Button>
        </Form.Item>
      </Form>
    </Modal>

    <Modal
      v-model:open="editOpen"
      destroy-on-close
      :title="$t('page.taskPlan.modal.editTitle')"
      :footer="null"
      @cancel="editOpen = false"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item :label="$t('page.taskPlan.form.taskKey')">
          <Input :value="editRecord?.task_key" disabled />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.taskName')" required>
          <Input v-model:value="formEdit.task_name" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.taskDescription')" required>
          <Input.TextArea v-model:value="formEdit.task_description" :rows="3" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.schedule')">
          <Input v-model:value="formEdit.crontab" />
        </Form.Item>
        <Form.Item :label="$t('page.taskPlan.form.enable')">
          <Select v-model:value="formEdit.enable" style="width: 100%">
            <Select.Option :value="1">{{ $t('page.taskPlan.enabled.yes') }}</Select.Option>
            <Select.Option :value="0">{{ $t('page.taskPlan.enabled.no') }}</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item>
          <Button type="primary" block @click="submitEdit">{{ $t('page.taskPlan.action.save') }}</Button>
        </Form.Item>
      </Form>
    </Modal>

    <Modal
      v-model:open="logOpen"
      destroy-on-close
      :footer="null"
      :title="logTitle"
      width="1200px"
      @cancel="logOpen = false"
    >
      <Table
        :columns="logColumns"
        :data-source="logDataSource"
        :loading="logLoading"
        :pagination="logPagination"
        :row-key="(r: TaskLogRow) => r.id"
        :scroll="{ y: 400 }"
        size="small"
        @change="handleLogTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'complete_time'">
            {{ record.complete_time || $t('page.taskPlan.log.incomplete') }}
          </template>
          <template v-else-if="column.key === 'status'">
            <span>
              {{ statusBadge(record.status).text }}
            </span>
          </template>
        </template>
      </Table>
    </Modal>
  </div>
</template>

<style scoped>
.stat-card :deep(.ant-card-body) {
  padding: 12px 16px;
}

/* 数据列表搜索栏与表格间距 */
.task-query-form {
  margin-bottom: 28px;
}

/* 数据列表搜索栏：避免 wrapper-col flex 把「启用」与按钮挤错位 */
.task-query-form.ant-form-inline {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  row-gap: 10px;
}

.task-query-form :deep(.ant-form-item) {
  margin-bottom: 0;
  margin-inline-end: 0;
}

.task-query-input {
  max-width: 100%;
  width: 200px;
}

.task-query-enable-item {
  flex-shrink: 0;
}

.task-query-enable-select {
  min-width: 120px;
  width: 120px;
}

.task-query-actions {
  flex-shrink: 0;
}
</style>
