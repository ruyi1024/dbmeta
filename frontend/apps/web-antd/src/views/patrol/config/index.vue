<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Tooltip,
  message,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';
import { checkPermission } from '#/utils/check-permission';

defineOptions({ name: 'PatrolConfigPage' });

interface AlarmRow {
  alarm_description?: string;
  alarm_name?: string;
  cron_expression?: string;
  database_name?: string;
  datasource_id?: number;
  datasource_type?: string;
  email_content?: string;
  email_to?: string;
  id?: number;
  last_run_time?: string;
  next_run_time?: string;
  rule_operator?: string;
  rule_value?: number;
  sql_query?: string;
  status?: number;
}

interface AlarmLogRow {
  alarm_id?: number;
  alarm_name?: string;
  complete_time?: string;
  created_at?: string;
  data_count?: number;
  email_sent?: boolean;
  error_message?: string;
  id?: number;
  rule_matched?: boolean;
  start_time?: string;
  status?: string;
}

interface OptionItem {
  id?: number;
  name?: string;
}

function formatTime(v?: string) {
  if (!v) return '-';
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString(undefined, { hour12: false });
}

function normalizeAlarmRows(list: unknown): AlarmRow[] {
  if (!Array.isArray(list)) return [];
  return list.map((item: any) => ({
    alarm_description: String(item?.alarm_description ?? ''),
    alarm_name: String(item?.alarm_name ?? ''),
    cron_expression: String(item?.cron_expression ?? ''),
    database_name: String(item?.database_name ?? ''),
    datasource_id: item?.datasource_id === undefined ? undefined : Number(item.datasource_id),
    datasource_type: String(item?.datasource_type ?? ''),
    email_content: String(item?.email_content ?? ''),
    email_to: String(item?.email_to ?? ''),
    id: item?.id === undefined ? undefined : Number(item.id),
    last_run_time: item?.last_run_time ? String(item.last_run_time) : '',
    next_run_time: item?.next_run_time ? String(item.next_run_time) : '',
    rule_operator: String(item?.rule_operator ?? ''),
    rule_value: item?.rule_value === undefined ? undefined : Number(item.rule_value),
    sql_query: String(item?.sql_query ?? ''),
    status: item?.status === undefined ? undefined : Number(item.status),
  }));
}

function normalizeAlarmLogRows(list: unknown): AlarmLogRow[] {
  if (!Array.isArray(list)) return [];
  return list.map((item: any) => ({
    alarm_id: item?.alarm_id === undefined ? undefined : Number(item.alarm_id),
    alarm_name: String(item?.alarm_name ?? ''),
    complete_time: item?.complete_time ? String(item.complete_time) : '',
    created_at: item?.created_at ? String(item.created_at) : '',
    data_count: item?.data_count === undefined ? undefined : Number(item.data_count),
    email_sent: Boolean(item?.email_sent),
    error_message: String(item?.error_message ?? ''),
    id: item?.id === undefined ? undefined : Number(item.id),
    rule_matched: Boolean(item?.rule_matched),
    start_time: item?.start_time ? String(item.start_time) : '',
    status: String(item?.status ?? ''),
  }));
}

const listLoading = ref(false);
const listData = ref<AlarmRow[]>([]);
const listQuery = reactive({
  alarm_name: '',
  datasource_type: '',
  status: undefined as number | undefined,
});
const listPagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  pageSizeOptions: ['10', '15', '30', '50'],
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => `${$t('page.common.total')} ${total} ${$t('page.common.records')}`,
  total: 0,
});

const createEditOpen = ref(false);
const createEditMode = ref<'create' | 'edit'>('create');
const createEditSaving = ref(false);
const formDatasourceList = ref<OptionItem[]>([]);
const formDatasourceTypeList = ref<OptionItem[]>([]);
const formDatabaseList = ref<string[]>([]);
const formModel = reactive<AlarmRow>({
  alarm_description: '',
  alarm_name: '',
  cron_expression: '',
  database_name: '',
  datasource_id: undefined,
  datasource_type: '',
  email_content: '',
  email_to: '',
  id: undefined,
  rule_operator: '>',
  rule_value: 0,
  sql_query: '',
  status: 1,
});

const logOpen = ref(false);
const logLoading = ref(false);
const currentAlarm = ref<AlarmRow | null>(null);
const logData = ref<AlarmLogRow[]>([]);
const logPagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  pageSizeOptions: ['10', '15', '30'],
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => `${$t('page.common.total')} ${total} ${$t('page.common.records')}`,
  total: 0,
});

const operatorOptions = [
  { label: $t('page.patrolConfig.operator.gt'), value: '>' },
  { label: $t('page.patrolConfig.operator.lt'), value: '<' },
  { label: $t('page.patrolConfig.operator.eq'), value: '=' },
  { label: $t('page.patrolConfig.operator.gte'), value: '>=' },
  { label: $t('page.patrolConfig.operator.lte'), value: '<=' },
  { label: $t('page.patrolConfig.operator.neq'), value: '!=' },
];

function operatorText(op?: string) {
  return (
    {
      '!=': $t('page.patrolConfig.operatorText.neq'),
      '<': $t('page.patrolConfig.operatorText.lt'),
      '<=': $t('page.patrolConfig.operatorText.lte'),
      '=': $t('page.patrolConfig.operatorText.eq'),
      '>': $t('page.patrolConfig.operatorText.gt'),
      '>=': $t('page.patrolConfig.operatorText.gte'),
    }[op || ''] || op || '-'
  );
}

function resetFormModel() {
  formModel.id = undefined;
  formModel.alarm_name = '';
  formModel.alarm_description = '';
  formModel.datasource_type = '';
  formModel.datasource_id = undefined;
  formModel.database_name = '';
  formModel.sql_query = '';
  formModel.rule_operator = '>';
  formModel.rule_value = 0;
  formModel.email_content = '';
  formModel.email_to = '';
  formModel.cron_expression = '';
  formModel.status = 1;
  formDatasourceList.value = [];
  formDatabaseList.value = [];
}

async function loadDatasourceTypes() {
  const response = await baseRequestClient.get('/v1/data/alarm/datasource-type');
  const payload = (response as any)?.data ?? response;
  formDatasourceTypeList.value = Array.isArray(payload?.data) ? payload.data : [];
}

async function loadDatasourcesByType(type: string) {
  if (!type) {
    formDatasourceList.value = [];
    formModel.datasource_id = undefined;
    return;
  }
  const response = await baseRequestClient.get('/v1/data/alarm/datasource', {
    params: { type },
  });
  const payload = (response as any)?.data ?? response;
  formDatasourceList.value = Array.isArray(payload?.data) ? payload.data : [];
}

async function loadDatabasesByDatasource(datasourceId?: number) {
  if (!datasourceId) {
    formDatabaseList.value = [];
    formModel.database_name = '';
    return;
  }
  const response = await baseRequestClient.get('/v1/data/alarm/database', {
    params: { datasource_id: datasourceId },
  });
  const payload = (response as any)?.data ?? response;
  formDatabaseList.value = Array.isArray(payload?.data) ? payload.data : [];
}

async function fetchList() {
  listLoading.value = true;
  try {
    const params: Record<string, any> = {
      currentPage: listPagination.current ?? 1,
      pageSize: listPagination.pageSize ?? 10,
    };
    if (listQuery.alarm_name.trim()) params.alarm_name = listQuery.alarm_name.trim();
    if (listQuery.datasource_type.trim()) params.datasource_type = listQuery.datasource_type.trim();
    if (listQuery.status !== undefined) params.status = listQuery.status;

    const response = await baseRequestClient.get('/v1/data/alarm/list', { params });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(String(payload?.msg ?? $t('page.patrolConfig.message.queryFailed')));
      listData.value = [];
      listPagination.total = 0;
      return;
    }
    listData.value = normalizeAlarmRows(payload?.data);
    listPagination.total = Number(payload?.total ?? listData.value.length) || listData.value.length;
  } catch (e: unknown) {
    listData.value = [];
    listPagination.total = 0;
    message.error((e as Error)?.message || $t('page.patrolConfig.message.queryFailed'));
  } finally {
    listLoading.value = false;
  }
}

async function fetchLogs() {
  if (!currentAlarm.value?.id) return;
  logLoading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/data/alarm/logs', {
      params: {
        alarm_id: currentAlarm.value.id,
        currentPage: logPagination.current ?? 1,
        pageSize: logPagination.pageSize ?? 10,
      },
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(String(payload?.msg ?? $t('page.patrolConfig.message.queryLogFailed')));
      logData.value = [];
      logPagination.total = 0;
      return;
    }
    logData.value = normalizeAlarmLogRows(payload?.data);
    logPagination.total = Number(payload?.total ?? logData.value.length) || logData.value.length;
  } catch (e: unknown) {
    logData.value = [];
    logPagination.total = 0;
    message.error((e as Error)?.message || $t('page.patrolConfig.message.queryLogFailed'));
  } finally {
    logLoading.value = false;
  }
}

function openCreate() {
  if (!checkPermission()) return;
  createEditMode.value = 'create';
  resetFormModel();
  createEditOpen.value = true;
}

async function openEdit(record: AlarmRow) {
  if (!checkPermission()) return;
  createEditMode.value = 'edit';
  resetFormModel();
  formModel.id = record.id;
  formModel.alarm_name = record.alarm_name ?? '';
  formModel.alarm_description = record.alarm_description ?? '';
  formModel.datasource_type = record.datasource_type ?? '';
  await loadDatasourcesByType(formModel.datasource_type || '');
  formModel.datasource_id = record.datasource_id;
  await loadDatabasesByDatasource(formModel.datasource_id);
  formModel.database_name = record.database_name ?? '';
  formModel.sql_query = record.sql_query ?? '';
  formModel.rule_operator = record.rule_operator ?? '>';
  formModel.rule_value = Number(record.rule_value ?? 0);
  formModel.email_content = record.email_content ?? '';
  formModel.email_to = record.email_to ?? '';
  formModel.cron_expression = record.cron_expression ?? '';
  formModel.status = Number(record.status ?? 1);
  createEditOpen.value = true;
}

async function submitForm() {
  if (!checkPermission()) return;
  if (!formModel.alarm_name?.trim()) return message.warning($t('page.patrolConfig.message.enterAlarmName'));
  if (!formModel.datasource_type?.trim()) return message.warning($t('page.patrolConfig.message.selectDatasourceType'));
  if (!formModel.datasource_id) return message.warning($t('page.patrolConfig.message.selectDatasource'));
  if (!formModel.sql_query?.trim()) return message.warning($t('page.patrolConfig.message.enterSql'));
  if (!formModel.rule_operator?.trim()) return message.warning($t('page.patrolConfig.message.selectOperator'));
  if (formModel.rule_value === undefined || formModel.rule_value === null) return message.warning($t('page.patrolConfig.message.enterRuleValue'));
  if (!formModel.cron_expression?.trim()) return message.warning($t('page.patrolConfig.message.enterCron'));
  if (!formModel.email_to?.trim()) return message.warning($t('page.patrolConfig.message.enterEmail'));

  createEditSaving.value = true;
  try {
    const payload = {
      alarm_description: formModel.alarm_description?.trim() || '',
      alarm_name: formModel.alarm_name.trim(),
      cron_expression: formModel.cron_expression.trim(),
      database_name: formModel.database_name?.trim() || '',
      datasource_id: formModel.datasource_id,
      datasource_type: formModel.datasource_type,
      email_content: formModel.email_content?.trim() || '',
      email_to: formModel.email_to.trim(),
      id: formModel.id,
      rule_operator: formModel.rule_operator,
      rule_value: Number(formModel.rule_value ?? 0),
      sql_query: formModel.sql_query.trim(),
      status: Number(formModel.status ?? 0),
    };
    const response =
      createEditMode.value === 'create'
        ? await baseRequestClient.post('/v1/data/alarm/create', payload)
        : await baseRequestClient.put('/v1/data/alarm/update', payload);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? $t('page.patrolConfig.message.saveFailed')));
      return;
    }
    message.success(
      createEditMode.value === 'create'
        ? $t('page.patrolConfig.message.createSuccess')
        : $t('page.patrolConfig.message.updateSuccess'),
    );
    createEditOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.patrolConfig.message.saveFailed'));
  } finally {
    createEditSaving.value = false;
  }
}

async function handleDelete(record: AlarmRow) {
  if (!checkPermission()) return;
  if (!record.id) return;
  try {
    const response = await baseRequestClient.delete(`/v1/data/alarm/delete/${record.id}`);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? $t('page.patrolConfig.message.deleteFailed')));
      return;
    }
    message.success($t('page.patrolConfig.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.patrolConfig.message.deleteFailed'));
  }
}

async function handleExecute(record: AlarmRow) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.post('/v1/data/alarm/execute', { id: record.id });
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? $t('page.patrolConfig.message.executeFailed')));
      return;
    }
    message.success($t('page.patrolConfig.message.executeSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.patrolConfig.message.executeFailed'));
  }
}

async function handleToggle(record: AlarmRow, checked: boolean) {
  if (!checkPermission()) return;
  if (!record.id) return;
  try {
    const response = await baseRequestClient.put('/v1/data/alarm/toggle-status', {
      id: record.id,
      status: checked ? 1 : 0,
    });
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? $t('page.patrolConfig.message.toggleFailed')));
      return;
    }
    message.success(checked ? $t('page.patrolConfig.message.enabled') : $t('page.patrolConfig.message.disabled'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.patrolConfig.message.toggleFailed'));
  }
}

function openLogs(record: AlarmRow) {
  currentAlarm.value = record;
  logPagination.current = 1;
  logOpen.value = true;
  void fetchLogs();
}

const listColumns: TableColumnsType<AlarmRow> = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: $t('page.patrolConfig.columns.alarmName'), dataIndex: 'alarm_name', key: 'alarm_name', width: 180, ellipsis: true },
  { title: $t('page.patrolConfig.columns.alarmDescription'), dataIndex: 'alarm_description', key: 'alarm_description', width: 220, ellipsis: true },
  { title: $t('page.patrolConfig.columns.datasourceType'), dataIndex: 'datasource_type', key: 'datasource_type', width: 120 },
  { title: $t('page.patrolConfig.columns.databaseName'), dataIndex: 'database_name', key: 'database_name', width: 140, ellipsis: true },
  { title: $t('page.patrolConfig.columns.rule'), key: 'rule', width: 180 },
  { title: $t('page.patrolConfig.columns.emailTo'), dataIndex: 'email_to', key: 'email_to', width: 240, ellipsis: true },
  { title: $t('page.patrolConfig.columns.cronExpression'), dataIndex: 'cron_expression', key: 'cron_expression', width: 150, ellipsis: true },
  { title: $t('page.patrolConfig.columns.status'), dataIndex: 'status', key: 'status', width: 130 },
  { title: $t('page.patrolConfig.columns.lastRunTime'), dataIndex: 'last_run_time', key: 'last_run_time', width: 170 },
  { title: $t('page.patrolConfig.columns.nextRunTime'), dataIndex: 'next_run_time', key: 'next_run_time', width: 170 },
  { title: $t('page.patrolConfig.columns.action'), key: 'action', width: 220, fixed: 'right' },
];

const logColumns: TableColumnsType<AlarmLogRow> = [
  { title: $t('page.patrolConfig.logColumns.startTime'), dataIndex: 'start_time', key: 'start_time', width: 170 },
  { title: $t('page.patrolConfig.logColumns.completeTime'), dataIndex: 'complete_time', key: 'complete_time', width: 170 },
  { title: $t('page.patrolConfig.logColumns.status'), dataIndex: 'status', key: 'status', width: 100 },
  { title: $t('page.patrolConfig.logColumns.dataCount'), dataIndex: 'data_count', key: 'data_count', width: 90 },
  { title: $t('page.patrolConfig.logColumns.ruleMatched'), dataIndex: 'rule_matched', key: 'rule_matched', width: 100 },
  { title: $t('page.patrolConfig.logColumns.emailSent'), dataIndex: 'email_sent', key: 'email_sent', width: 100 },
  { title: $t('page.patrolConfig.logColumns.errorMessage'), dataIndex: 'error_message', key: 'error_message', width: 260, ellipsis: true },
];

const datasourceTypeOptions = computed(() =>
  formDatasourceTypeList.value.map((item) => ({ label: item.name, value: item.name })),
);
const datasourceOptions = computed(() =>
  formDatasourceList.value.map((item) => ({ label: item.name, value: item.id })),
);
const databaseOptions = computed(() => formDatabaseList.value.map((name) => ({ label: name, value: name })));

onMounted(async () => {
  await loadDatasourceTypes();
  await fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.patrolConfig.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.patrolConfig.query.alarmName')" class="query-item">
            <Input v-model:value="listQuery.alarm_name" allow-clear class="query-control" :placeholder="$t('page.patrolConfig.placeholder.alarmName')" @press-enter="fetchList" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.query.datasourceType')" class="query-item">
            <Select v-model:value="listQuery.datasource_type" allow-clear class="query-control" :placeholder="$t('page.patrolConfig.placeholder.all')" :options="datasourceTypeOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.query.status')" class="query-item">
            <Select
              v-model:value="listQuery.status"
              allow-clear
              class="query-control"
              :placeholder="$t('page.patrolConfig.placeholder.all')"
              :options="[
                { value: 0, label: $t('page.patrolConfig.status.disabled') },
                { value: 1, label: $t('page.patrolConfig.status.enabled') },
              ]"
            />
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="fetchList">{{ $t('page.common.search') }}</Button>
            <Button
              @click="
                () => {
                  listQuery.alarm_name = '';
                  listQuery.datasource_type = '';
                  listQuery.status = undefined;
                  listPagination.current = 1;
                  fetchList();
                }
              "
            >
              {{ $t('page.common.reset') }}
            </Button>
            <Button type="primary" ghost @click="openCreate">{{ $t('page.patrolConfig.action.createAlarm') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="listColumns"
        :data-source="listData"
        :loading="listLoading"
        :pagination="listPagination"
        bordered
        size="middle"
        table-layout="fixed"
        :row-class-name="(_record: AlarmRow, index: number) => (index % 2 === 1 ? 'table-row-striped' : '')"
        :row-key="(record: AlarmRow, index: number) => record.id ?? `alarm-${index}`"
        :scroll="{ x: 2050 }"
        @change="
          (pag: TablePaginationConfig) => {
            if (pag.current !== undefined) listPagination.current = pag.current;
            if (pag.pageSize !== undefined) listPagination.pageSize = pag.pageSize;
            fetchList();
          }
        "
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'alarm_description'">
            <Tooltip :title="record.alarm_description || '-'">
              <span class="inline-block max-w-[220px] truncate">{{ record.alarm_description || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'rule'">
            {{ $t('page.patrolConfig.ruleTextPrefix') }} {{ operatorText(record.rule_operator) }} {{ record.rule_value ?? 0 }}
          </template>
          <template v-else-if="column.key === 'database_name'">
            <Tooltip :title="record.database_name || '-'">
              <span class="inline-block max-w-[120px] truncate">{{ record.database_name || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'email_to'">
            <Tooltip :title="record.email_to || '-'">
              <span class="inline-block max-w-[220px] truncate">{{ record.email_to || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'cron_expression'">
            <code class="rounded bg-muted px-1 py-[1px] text-[12px]">{{ record.cron_expression || '-' }}</code>
          </template>
          <template v-else-if="column.key === 'status'">
            <div class="flex items-center gap-2">
              <Tag :color="Number(record.status) === 1 ? 'green' : 'default'">
                {{ Number(record.status) === 1 ? $t('page.patrolConfig.status.enabled') : $t('page.patrolConfig.status.disabled') }}
              </Tag>
              <Switch :checked="Number(record.status) === 1" size="small" @change="(checked: boolean) => handleToggle(record, checked)" />
            </div>
          </template>
          <template v-else-if="column.key === 'last_run_time'">{{ formatTime(record.last_run_time) }}</template>
          <template v-else-if="column.key === 'next_run_time'">{{ formatTime(record.next_run_time) }}</template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.patrolConfig.action.edit') }}</Button>
              <Button type="link" size="small" @click="handleExecute(record)">{{ $t('page.patrolConfig.action.execute') }}</Button>
              <Button type="link" size="small" @click="openLogs(record)">{{ $t('page.patrolConfig.action.logs') }}</Button>
              <Popconfirm :title="$t('page.patrolConfig.deleteConfirm')" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>{{ $t('page.patrolConfig.action.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="createEditOpen"
      :title="createEditMode === 'create' ? $t('page.patrolConfig.modal.createTitle') : $t('page.patrolConfig.modal.editTitle')"
      :confirm-loading="createEditSaving"
      width="860px"
      destroy-on-close
      @ok="submitForm"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item :label="$t('page.patrolConfig.form.alarmName')" required>
            <Input v-model:value="formModel.alarm_name" :placeholder="$t('page.patrolConfig.placeholder.alarmName')" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.status')">
            <Select
              v-model:value="formModel.status"
              :options="[
                { value: 0, label: $t('page.patrolConfig.status.disabled') },
                { value: 1, label: $t('page.patrolConfig.status.enabled') },
              ]"
            />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.datasourceType')" required>
            <Select
              v-model:value="formModel.datasource_type"
              :placeholder="$t('page.patrolConfig.placeholder.selectDatasourceType')"
              :options="datasourceTypeOptions"
              @change="
                (v: string) => {
                  formModel.datasource_type = v;
                  formModel.datasource_id = undefined;
                  formModel.database_name = '';
                  loadDatasourcesByType(v);
                }
              "
            />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.datasource')" required>
            <Select
              v-model:value="formModel.datasource_id"
              :placeholder="$t('page.patrolConfig.placeholder.selectDatasource')"
              :options="datasourceOptions"
              @change="
                (v: number) => {
                  formModel.datasource_id = v;
                  formModel.database_name = '';
                  loadDatabasesByDatasource(v);
                }
              "
            />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.databaseNameOptional')">
            <Select v-model:value="formModel.database_name" allow-clear :placeholder="$t('page.patrolConfig.placeholder.databaseOptional')" :options="databaseOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.cronExpression')" required>
            <Input v-model:value="formModel.cron_expression" :placeholder="$t('page.patrolConfig.placeholder.cronExample')" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.ruleOperator')" required>
            <Select v-model:value="formModel.rule_operator" :options="operatorOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.ruleValue')" required>
            <InputNumber v-model:value="formModel.rule_value" :min="0" class="w-full" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.emailTo')" required class="col-span-2">
            <Input v-model:value="formModel.email_to" :placeholder="$t('page.patrolConfig.placeholder.emailTo')" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.alarmDescription')" class="col-span-2">
            <Input.TextArea v-model:value="formModel.alarm_description" :rows="2" :placeholder="$t('page.patrolConfig.placeholder.optional')" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.emailContentOptional')" class="col-span-2">
            <Input.TextArea v-model:value="formModel.email_content" :rows="3" :placeholder="$t('page.patrolConfig.placeholder.emailContentOptional')" />
          </Form.Item>
          <Form.Item :label="$t('page.patrolConfig.form.sqlQuery')" required class="col-span-2">
            <Input.TextArea v-model:value="formModel.sql_query" :rows="5" :placeholder="$t('page.patrolConfig.placeholder.sqlQuery')" />
          </Form.Item>
        </div>
      </Form>
    </Modal>

    <Modal
      v-model:open="logOpen"
      :title="`${$t('page.patrolConfig.logModalTitle')} - ${currentAlarm?.alarm_name || '-'}`"
      width="960px"
      :footer="null"
      destroy-on-close
    >
      <Table
        :columns="logColumns"
        :data-source="logData"
        :loading="logLoading"
        :pagination="logPagination"
        bordered
        size="middle"
        table-layout="fixed"
        :row-class-name="(_record: AlarmLogRow, index: number) => (index % 2 === 1 ? 'table-row-striped' : '')"
        :row-key="(record: AlarmLogRow, index: number) => record.id ?? `alarm-log-${index}`"
        :scroll="{ x: 1200 }"
        @change="
          (pag: TablePaginationConfig) => {
            if (pag.current !== undefined) logPagination.current = pag.current;
            if (pag.pageSize !== undefined) logPagination.pageSize = pag.pageSize;
            fetchLogs();
          }
        "
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="record.status === 'success' ? 'green' : record.status === 'failed' ? 'red' : record.status === 'triggered' ? 'orange' : 'blue'">
              {{ record.status === 'success' ? $t('page.patrolConfig.logStatus.success') : record.status === 'failed' ? $t('page.patrolConfig.logStatus.failed') : record.status === 'triggered' ? $t('page.patrolConfig.logStatus.triggered') : record.status || '-' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'start_time'">{{ formatTime(record.start_time) }}</template>
          <template v-else-if="column.key === 'complete_time'">{{ formatTime(record.complete_time) }}</template>
          <template v-else-if="column.key === 'rule_matched'">
            <Badge :status="record.rule_matched ? 'error' : 'default'" :text="record.rule_matched ? $t('page.patrolConfig.match.hit') : $t('page.patrolConfig.match.miss')" />
          </template>
          <template v-else-if="column.key === 'email_sent'">
            <Badge :status="record.email_sent ? 'success' : 'default'" :text="record.email_sent ? $t('page.patrolConfig.email.sent') : $t('page.patrolConfig.email.unsent')" />
          </template>
          <template v-else-if="column.key === 'error_message'">
            <Tooltip :title="record.error_message || '-'">
              <span class="inline-block max-w-[220px] truncate">{{ record.error_message || '-' }}</span>
            </Tooltip>
          </template>
        </template>
      </Table>
    </Modal>
  </div>
</template>

<style scoped>
:deep(.table-row-striped > td) {
  background: rgba(0, 0, 0, 0.015);
}

.query-grid {
  column-gap: 12px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  row-gap: 8px;
}

.form-grid {
  column-gap: 12px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

:deep(.query-item) {
  margin-bottom: 0;
}

:deep(.query-item .ant-form-item-row) {
  align-items: center;
  display: flex;
}

:deep(.query-item .ant-form-item-label) {
  flex: 0 0 72px;
  max-width: 72px;
  padding-right: 8px;
  text-align: right;
}

:deep(.query-item .ant-form-item-control) {
  flex: 1;
  min-width: 0;
}

:deep(.query-control) {
  width: 300px;
}

.query-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
  margin-top: 12px;
}

:deep(.col-span-2) {
  grid-column: span 2 / span 2;
}

@media (max-width: 1400px) {
  .query-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 1100px) {
  .query-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .form-grid {
    grid-template-columns: 1fr;
  }

  :deep(.col-span-2) {
    grid-column: span 1 / span 1;
  }
}

@media (max-width: 768px) {
  .query-grid {
    grid-template-columns: 1fr;
  }

  :deep(.query-control) {
    width: 100%;
  }

  .query-actions {
    justify-content: flex-start;
  }
}
</style>
