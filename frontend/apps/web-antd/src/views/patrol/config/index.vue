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
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString('zh-CN', { hour12: false });
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
  showTotal: (total: number) => `共 ${total} 条`,
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
  showTotal: (total: number) => `共 ${total} 条`,
  total: 0,
});

const operatorOptions = [
  { label: '大于(>)', value: '>' },
  { label: '小于(<)', value: '<' },
  { label: '等于(=)', value: '=' },
  { label: '大于等于(>=)', value: '>=' },
  { label: '小于等于(<=)', value: '<=' },
  { label: '不等于(!=)', value: '!=' },
];

function operatorText(op?: string) {
  return (
    {
      '!=': '不等于',
      '<': '小于',
      '<=': '小于等于',
      '=': '等于',
      '>': '大于',
      '>=': '大于等于',
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
      message.error(String(payload?.msg ?? '查询失败'));
      listData.value = [];
      listPagination.total = 0;
      return;
    }
    listData.value = Array.isArray(payload?.data) ? payload.data : [];
    listPagination.total = Number(payload?.total ?? listData.value.length) || listData.value.length;
  } catch (e: unknown) {
    listData.value = [];
    listPagination.total = 0;
    message.error((e as Error)?.message || '查询失败');
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
      message.error(String(payload?.msg ?? '查询日志失败'));
      logData.value = [];
      logPagination.total = 0;
      return;
    }
    logData.value = Array.isArray(payload?.data) ? payload.data : [];
    logPagination.total = Number(payload?.total ?? logData.value.length) || logData.value.length;
  } catch (e: unknown) {
    logData.value = [];
    logPagination.total = 0;
    message.error((e as Error)?.message || '查询日志失败');
  } finally {
    logLoading.value = false;
  }
}

function openCreate() {
  createEditMode.value = 'create';
  resetFormModel();
  createEditOpen.value = true;
}

async function openEdit(record: AlarmRow) {
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
  if (!formModel.alarm_name?.trim()) return message.warning('请输入告警名称');
  if (!formModel.datasource_type?.trim()) return message.warning('请选择数据源类型');
  if (!formModel.datasource_id) return message.warning('请选择数据源');
  if (!formModel.sql_query?.trim()) return message.warning('请输入SQL查询');
  if (!formModel.rule_operator?.trim()) return message.warning('请选择规则操作符');
  if (formModel.rule_value === undefined || formModel.rule_value === null) return message.warning('请输入规则值');
  if (!formModel.cron_expression?.trim()) return message.warning('请输入Cron表达式');
  if (!formModel.email_to?.trim()) return message.warning('请输入接收邮箱');

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
      message.error(String(body?.msg ?? '保存失败'));
      return;
    }
    message.success(createEditMode.value === 'create' ? '创建成功' : '更新成功');
    createEditOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '保存失败');
  } finally {
    createEditSaving.value = false;
  }
}

async function handleDelete(record: AlarmRow) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.delete(`/v1/data/alarm/delete/${record.id}`);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '删除失败'));
      return;
    }
    message.success('删除成功');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '删除失败');
  }
}

async function handleExecute(record: AlarmRow) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.post('/v1/data/alarm/execute', { id: record.id });
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '执行失败'));
      return;
    }
    message.success('执行成功');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '执行失败');
  }
}

async function handleToggle(record: AlarmRow, checked: boolean) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.put('/v1/data/alarm/toggle-status', {
      id: record.id,
      status: checked ? 1 : 0,
    });
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '状态切换失败'));
      return;
    }
    message.success(checked ? '已启用' : '已禁用');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '状态切换失败');
  }
}

function openLogs(record: AlarmRow) {
  currentAlarm.value = record;
  logPagination.current = 1;
  logOpen.value = true;
  void fetchLogs();
}

const listColumns: TableColumnsType<AlarmRow> = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: '告警名称', dataIndex: 'alarm_name', key: 'alarm_name', width: 180 },
  { title: '告警描述', dataIndex: 'alarm_description', key: 'alarm_description', width: 220 },
  { title: '数据源类型', dataIndex: 'datasource_type', key: 'datasource_type', width: 120 },
  { title: '规则', key: 'rule', width: 160 },
  { title: '接收邮箱', dataIndex: 'email_to', key: 'email_to', width: 220 },
  { title: 'Cron表达式', dataIndex: 'cron_expression', key: 'cron_expression', width: 150 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 90 },
  { title: '上次运行', dataIndex: 'last_run_time', key: 'last_run_time', width: 170 },
  { title: '下次运行', dataIndex: 'next_run_time', key: 'next_run_time', width: 170 },
  { title: '操作', key: 'action', width: 220, fixed: 'right' },
];

const logColumns: TableColumnsType<AlarmLogRow> = [
  { title: '开始时间', dataIndex: 'start_time', key: 'start_time', width: 170 },
  { title: '完成时间', dataIndex: 'complete_time', key: 'complete_time', width: 170 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '数据量', dataIndex: 'data_count', key: 'data_count', width: 90 },
  { title: '规则命中', dataIndex: 'rule_matched', key: 'rule_matched', width: 100 },
  { title: '邮件发送', dataIndex: 'email_sent', key: 'email_sent', width: 100 },
  { title: '错误信息', dataIndex: 'error_message', key: 'error_message', width: 220 },
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
    <Card title="巡检配置">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="告警名称" class="query-item">
            <Input v-model:value="listQuery.alarm_name" allow-clear class="query-control" placeholder="请输入告警名称" @press-enter="fetchList" />
          </Form.Item>
          <Form.Item label="数据源类型" class="query-item">
            <Select v-model:value="listQuery.datasource_type" allow-clear class="query-control" placeholder="全部" :options="datasourceTypeOptions" />
          </Form.Item>
          <Form.Item label="状态" class="query-item">
            <Select
              v-model:value="listQuery.status"
              allow-clear
              class="query-control"
              placeholder="全部"
              :options="[
                { value: 0, label: '禁用' },
                { value: 1, label: '启用' },
              ]"
            />
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="fetchList">查询</Button>
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
              重置
            </Button>
            <Button type="primary" ghost @click="openCreate">创建告警</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="listColumns"
        :data-source="listData"
        :loading="listLoading"
        :pagination="listPagination"
        :row-key="(record: AlarmRow, index: number) => record.id ?? `alarm-${index}`"
        :scroll="{ x: 1900 }"
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
            数据量 {{ operatorText(record.rule_operator) }} {{ record.rule_value ?? 0 }}
          </template>
          <template v-else-if="column.key === 'status'">
            <Switch :checked="Number(record.status) === 1" @change="(checked: boolean) => handleToggle(record, checked)" />
          </template>
          <template v-else-if="column.key === 'last_run_time'">{{ formatTime(record.last_run_time) }}</template>
          <template v-else-if="column.key === 'next_run_time'">{{ formatTime(record.next_run_time) }}</template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">编辑</Button>
              <Button type="link" size="small" @click="handleExecute(record)">执行</Button>
              <Button type="link" size="small" @click="openLogs(record)">日志</Button>
              <Popconfirm title="确定要删除这个告警吗？" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="createEditOpen"
      :title="createEditMode === 'create' ? '创建巡检告警' : '编辑巡检告警'"
      :confirm-loading="createEditSaving"
      width="860px"
      destroy-on-close
      @ok="submitForm"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item label="告警名称" required>
            <Input v-model:value="formModel.alarm_name" placeholder="请输入告警名称" />
          </Form.Item>
          <Form.Item label="状态">
            <Select
              v-model:value="formModel.status"
              :options="[
                { value: 0, label: '禁用' },
                { value: 1, label: '启用' },
              ]"
            />
          </Form.Item>
          <Form.Item label="数据源类型" required>
            <Select
              v-model:value="formModel.datasource_type"
              placeholder="请选择数据源类型"
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
          <Form.Item label="数据源" required>
            <Select
              v-model:value="formModel.datasource_id"
              placeholder="请选择数据源"
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
          <Form.Item label="数据库（可选）">
            <Select v-model:value="formModel.database_name" allow-clear placeholder="可选择指定数据库" :options="databaseOptions" />
          </Form.Item>
          <Form.Item label="Cron表达式" required>
            <Input v-model:value="formModel.cron_expression" placeholder="如 */10 * * * *" />
          </Form.Item>
          <Form.Item label="规则操作符" required>
            <Select v-model:value="formModel.rule_operator" :options="operatorOptions" />
          </Form.Item>
          <Form.Item label="规则值" required>
            <InputNumber v-model:value="formModel.rule_value" :min="0" class="w-full" />
          </Form.Item>
          <Form.Item label="接收邮箱" required class="col-span-2">
            <Input v-model:value="formModel.email_to" placeholder="多个邮箱请用 ; 分隔" />
          </Form.Item>
          <Form.Item label="告警描述" class="col-span-2">
            <Input.TextArea v-model:value="formModel.alarm_description" :rows="2" placeholder="可选" />
          </Form.Item>
          <Form.Item label="告警邮件内容（可选）" class="col-span-2">
            <Input.TextArea v-model:value="formModel.email_content" :rows="3" placeholder="可选，发送邮件时展示在结果上方" />
          </Form.Item>
          <Form.Item label="SQL查询" required class="col-span-2">
            <Input.TextArea v-model:value="formModel.sql_query" :rows="5" placeholder="请输入用于巡检判断的SQL" />
          </Form.Item>
        </div>
      </Form>
    </Modal>

    <Modal
      v-model:open="logOpen"
      :title="`执行日志 - ${currentAlarm?.alarm_name || '-'}`"
      width="960px"
      :footer="null"
      destroy-on-close
    >
      <Table
        :columns="logColumns"
        :data-source="logData"
        :loading="logLoading"
        :pagination="logPagination"
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
              {{ record.status || '-' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'rule_matched'">
            <Badge :status="record.rule_matched ? 'error' : 'default'" :text="record.rule_matched ? '命中' : '未命中'" />
          </template>
          <template v-else-if="column.key === 'email_sent'">
            <Badge :status="record.email_sent ? 'success' : 'default'" :text="record.email_sent ? '已发送' : '未发送'" />
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
