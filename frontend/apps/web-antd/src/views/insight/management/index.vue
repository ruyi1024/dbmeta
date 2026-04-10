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
  Table,
  Tag,
  Tooltip,
  message,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'InsightManagementPage' });

interface AnalysisTaskRow {
  ai_model_id?: number;
  created_at?: string;
  cron_expression?: string;
  datasource_id?: number;
  datasource_type?: string;
  id?: number;
  last_run_time?: string;
  next_run_time?: string;
  prompt?: string;
  report_email?: string;
  sql_queries?: string[];
  status?: number;
  task_description?: string;
  task_name?: string;
  updated_at?: string;
}

interface AnalysisTaskLogRow {
  complete_time?: string;
  created_at?: string;
  data_count?: number;
  error_message?: string;
  id?: number;
  report_content?: string;
  result?: string;
  start_time?: string;
  status?: string;
  task_id?: number;
  task_name?: string;
}

interface OptionItem {
  id?: number;
  name?: string;
  provider?: string;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') return {};
  const r = response as Record<string, unknown>;
  if ('data' in r && r.data !== undefined && typeof r.data === 'object' && 'status' in r) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r;
}

function formatTime(v?: string) {
  if (!v) return '-';
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString('zh-CN', { hour12: false });
}

const loading = ref(false);
const tasks = ref<AnalysisTaskRow[]>([]);

const listQuery = reactive({
  status: undefined as number | undefined,
  task_name: '',
});
const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  pageSizeOptions: ['10', '15', '30', '50'],
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
  total: 0,
});

const taskModalOpen = ref(false);
const taskModalMode = ref<'create' | 'edit'>('create');
const savingTask = ref(false);

const datasourceTypes = ref<OptionItem[]>([]);
const datasources = ref<OptionItem[]>([]);
const aiModels = ref<OptionItem[]>([]);

const taskForm = reactive({
  ai_model_id: undefined as number | undefined,
  cron_expression: '',
  datasource_id: undefined as number | undefined,
  datasource_type: '',
  id: undefined as number | undefined,
  prompt: '',
  report_email: '',
  sql_queries_text: '',
  status: 1,
  task_description: '',
  task_name: '',
});

const logModalOpen = ref(false);
const currentLogTaskId = ref<number | undefined>(undefined);
const currentLogTaskName = ref('');
const logLoading = ref(false);
const taskLogs = ref<AnalysisTaskLogRow[]>([]);
const logPagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  pageSizeOptions: ['10', '15', '30'],
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`,
  total: 0,
});

function sqlArrayFromText(text: string) {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
}

async function loadFormOptions() {
  const [typeRes, modelRes] = await Promise.all([
    baseRequestClient.get('/v1/task/analysis/datasource-type'),
    baseRequestClient.get('/v1/ai/models/enabled'),
  ]);
  datasourceTypes.value = ((extractApiBody(typeRes).data as OptionItem[]) || []).map((item) => ({
    id: item.id,
    name: item.name,
  }));
  aiModels.value = ((extractApiBody(modelRes).data as OptionItem[]) || []).map((item) => ({
    id: item.id,
    name: `${item.name} (${item.provider || 'unknown'})`,
  }));
}

async function loadDatasourceByType(type: string) {
  if (!type) {
    datasources.value = [];
    taskForm.datasource_id = undefined;
    return;
  }
  const response = await baseRequestClient.get('/v1/task/analysis/datasource', {
    params: { type },
  });
  datasources.value = ((extractApiBody(response).data as OptionItem[]) || []).map((item) => ({
    id: item.id,
    name: item.name,
  }));
}

async function fetchTasks() {
  loading.value = true;
  try {
    const params: Record<string, any> = {
      currentPage: pagination.current ?? 1,
      pageSize: pagination.pageSize ?? 10,
    };
    if (listQuery.task_name.trim()) params.task_name = listQuery.task_name.trim();
    if (listQuery.status !== undefined) params.status = listQuery.status;

    const response = await baseRequestClient.get('/v1/task/analysis/list', { params });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(String(payload?.msg ?? '加载失败'));
      tasks.value = [];
      pagination.total = 0;
      return;
    }
    const list = Array.isArray(payload?.data) ? (payload.data as AnalysisTaskRow[]) : [];
    tasks.value = list;
    pagination.total = Number(payload?.total ?? list.length) || list.length;
  } catch (e: unknown) {
    tasks.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function resetTaskForm() {
  taskForm.id = undefined;
  taskForm.task_name = '';
  taskForm.task_description = '';
  taskForm.datasource_type = '';
  taskForm.datasource_id = undefined;
  taskForm.ai_model_id = undefined;
  taskForm.sql_queries_text = '';
  taskForm.prompt = '';
  taskForm.cron_expression = '';
  taskForm.report_email = '';
  taskForm.status = 1;
  datasources.value = [];
}

function openCreateTask() {
  taskModalMode.value = 'create';
  resetTaskForm();
  taskModalOpen.value = true;
}

async function openEditTask(record: AnalysisTaskRow) {
  taskModalMode.value = 'edit';
  taskForm.id = record.id;
  taskForm.task_name = record.task_name || '';
  taskForm.task_description = record.task_description || '';
  taskForm.datasource_type = record.datasource_type || '';
  taskForm.datasource_id = record.datasource_id;
  taskForm.ai_model_id = record.ai_model_id;
  taskForm.sql_queries_text = (record.sql_queries || []).join('\n');
  taskForm.prompt = record.prompt || '';
  taskForm.cron_expression = record.cron_expression || '';
  taskForm.report_email = record.report_email || '';
  taskForm.status = Number(record.status ?? 0);
  await loadDatasourceByType(taskForm.datasource_type);
  taskModalOpen.value = true;
}

async function submitTask() {
  if (!taskForm.task_name.trim()) return message.warning('请输入任务名称');
  if (!taskForm.datasource_type) return message.warning('请选择数据源类型');
  if (!taskForm.datasource_id) return message.warning('请选择数据源');
  if (!taskForm.ai_model_id) return message.warning('请选择AI模型');
  if (!taskForm.cron_expression.trim()) return message.warning('请输入Cron表达式');
  if (!taskForm.report_email.trim()) return message.warning('请输入报告邮箱');

  const sqlQueries = sqlArrayFromText(taskForm.sql_queries_text);
  if (sqlQueries.length === 0) return message.warning('请至少输入一条SQL');

  savingTask.value = true;
  try {
    const payload = {
      ai_model_id: taskForm.ai_model_id,
      cron_expression: taskForm.cron_expression.trim(),
      datasource_id: taskForm.datasource_id,
      datasource_type: taskForm.datasource_type,
      id: taskForm.id,
      prompt: taskForm.prompt.trim(),
      report_email: taskForm.report_email.trim(),
      sql_queries: sqlQueries,
      status: Number(taskForm.status ?? 0),
      task_description: taskForm.task_description.trim(),
      task_name: taskForm.task_name.trim(),
    };
    const response =
      taskModalMode.value === 'create'
        ? await baseRequestClient.post('/v1/task/analysis/create', payload)
        : await baseRequestClient.put('/v1/task/analysis/update', payload);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '保存失败'));
      return;
    }
    message.success(taskModalMode.value === 'create' ? '创建成功' : '修改成功');
    taskModalOpen.value = false;
    void fetchTasks();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '保存失败');
  } finally {
    savingTask.value = false;
  }
}

async function handleDeleteTask(record: AnalysisTaskRow) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.delete(`/v1/task/analysis/delete/${record.id}`);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '删除失败'));
      return;
    }
    message.success('删除成功');
    void fetchTasks();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '删除失败');
  }
}

async function handleExecuteTask(record: AnalysisTaskRow) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.post('/v1/task/analysis/execute', { id: record.id });
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '执行失败'));
      return;
    }
    message.success('任务执行已启动');
    void fetchTasks();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '执行失败');
  }
}

async function handleToggleStatus(record: AnalysisTaskRow, checked: boolean) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.put('/v1/task/analysis/toggle-status', {
      id: record.id,
      status: checked ? 1 : 0,
    });
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '切换状态失败'));
      return;
    }
    message.success(checked ? '已启用' : '已禁用');
    void fetchTasks();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '切换状态失败');
  }
}

async function fetchLogs() {
  if (!currentLogTaskId.value) return;
  logLoading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/task/analysis/logs', {
      params: {
        currentPage: logPagination.current ?? 1,
        pageSize: logPagination.pageSize ?? 10,
        task_id: currentLogTaskId.value,
      },
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(String(payload?.msg ?? '加载日志失败'));
      taskLogs.value = [];
      logPagination.total = 0;
      return;
    }
    const list = Array.isArray(payload?.data) ? (payload.data as AnalysisTaskLogRow[]) : [];
    taskLogs.value = list;
    logPagination.total = Number(payload?.total ?? list.length) || list.length;
  } catch (e: unknown) {
    taskLogs.value = [];
    logPagination.total = 0;
    message.error((e as Error)?.message || '加载日志失败');
  } finally {
    logLoading.value = false;
  }
}

function openLogs(record: AnalysisTaskRow) {
  currentLogTaskId.value = record.id;
  currentLogTaskName.value = record.task_name || '';
  logPagination.current = 1;
  logModalOpen.value = true;
  void fetchLogs();
}

function handleTaskTableChange(pag: TablePaginationConfig) {
  if (pag.current !== undefined) pagination.current = pag.current;
  if (pag.pageSize !== undefined) pagination.pageSize = pag.pageSize;
  void fetchTasks();
}

function handleLogTableChange(pag: TablePaginationConfig) {
  if (pag.current !== undefined) logPagination.current = pag.current;
  if (pag.pageSize !== undefined) logPagination.pageSize = pag.pageSize;
  void fetchLogs();
}

const taskColumns: TableColumnsType<AnalysisTaskRow> = [
  { title: '任务名称', dataIndex: 'task_name', key: 'task_name', width: 180 },
  { title: '任务描述', dataIndex: 'task_description', key: 'task_description', width: 240 },
  { title: 'Cron表达式', dataIndex: 'cron_expression', key: 'cron_expression', width: 150 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 90 },
  { title: '最后执行', dataIndex: 'last_run_time', key: 'last_run_time', width: 170 },
  { title: '下次执行', dataIndex: 'next_run_time', key: 'next_run_time', width: 170 },
  { title: '创建时间', dataIndex: 'created_at', key: 'created_at', width: 170 },
  { title: '操作', key: 'action', width: 260, fixed: 'right' },
];

const logColumns: TableColumnsType<AnalysisTaskLogRow> = [
  { title: '开始时间', dataIndex: 'start_time', key: 'start_time', width: 170 },
  { title: '完成时间', dataIndex: 'complete_time', key: 'complete_time', width: 170 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '数据量', dataIndex: 'data_count', key: 'data_count', width: 90 },
  { title: '结果', dataIndex: 'result', key: 'result', width: 220 },
  { title: '错误信息', dataIndex: 'error_message', key: 'error_message', width: 220 },
];

const datasourceOptions = computed(() =>
  datasources.value.map((item) => ({ label: item.name, value: item.id })),
);
const datasourceTypeOptions = computed(() =>
  datasourceTypes.value.map((item) => ({ label: item.name, value: item.name })),
);
const aiModelOptions = computed(() =>
  aiModels.value.map((item) => ({ label: item.name, value: item.id })),
);

onMounted(async () => {
  await loadFormOptions();
  await fetchTasks();
});
</script>

<template>
  <div class="p-5">
    <Card title="数据洞察管理">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="任务名称" class="query-item">
            <Input
              v-model:value="listQuery.task_name"
              allow-clear
              class="query-control"
              placeholder="请输入任务名称"
              @press-enter="fetchTasks"
            />
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
            <Button type="primary" @click="fetchTasks">查询</Button>
            <Button
              @click="
                () => {
                  listQuery.task_name = '';
                  listQuery.status = undefined;
                  pagination.current = 1;
                  fetchTasks();
                }
              "
            >
              重置
            </Button>
            <Button type="primary" ghost @click="openCreateTask">创建任务</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="taskColumns"
        :data-source="tasks"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: AnalysisTaskRow, index: number) => record.id ?? `task-${index}`"
        :scroll="{ x: 1700 }"
        @change="handleTaskTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'task_description'">
            <Tooltip :title="record.task_description || '-'">
              <span class="inline-block max-w-[220px] truncate">{{ record.task_description || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'status'">
            <Switch
              :checked="Number(record.status) === 1"
              @change="(checked: boolean) => handleToggleStatus(record, checked)"
            />
          </template>
          <template v-else-if="column.key === 'last_run_time'">{{ formatTime(record.last_run_time) }}</template>
          <template v-else-if="column.key === 'next_run_time'">{{ formatTime(record.next_run_time) }}</template>
          <template v-else-if="column.key === 'created_at'">{{ formatTime(record.created_at) }}</template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEditTask(record)">编辑</Button>
              <Button type="link" size="small" @click="handleExecuteTask(record)">执行</Button>
              <Button type="link" size="small" @click="openLogs(record)">日志</Button>
              <Popconfirm title="确认删除该任务？删除后不可恢复。" placement="left" @confirm="handleDeleteTask(record)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="taskModalOpen"
      :title="taskModalMode === 'create' ? '创建洞察任务' : '编辑洞察任务'"
      :confirm-loading="savingTask"
      width="860px"
      destroy-on-close
      @ok="submitTask"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item label="任务名称" required>
            <Input v-model:value="taskForm.task_name" placeholder="请输入任务名称" />
          </Form.Item>
          <Form.Item label="状态">
            <Select
              v-model:value="taskForm.status"
              :options="[
                { value: 0, label: '禁用' },
                { value: 1, label: '启用' },
              ]"
            />
          </Form.Item>
          <Form.Item label="数据源类型" required>
            <Select
              v-model:value="taskForm.datasource_type"
              placeholder="请选择类型"
              :options="datasourceTypeOptions"
              @change="
                (v: string) => {
                  taskForm.datasource_type = v;
                  taskForm.datasource_id = undefined;
                  loadDatasourceByType(v);
                }
              "
            />
          </Form.Item>
          <Form.Item label="数据源" required>
            <Select v-model:value="taskForm.datasource_id" placeholder="请选择数据源" :options="datasourceOptions" />
          </Form.Item>
          <Form.Item label="AI模型" required>
            <Select v-model:value="taskForm.ai_model_id" placeholder="请选择AI模型" :options="aiModelOptions" />
          </Form.Item>
          <Form.Item label="Cron表达式" required>
            <Input v-model:value="taskForm.cron_expression" placeholder="如 */10 * * * *" />
          </Form.Item>
          <Form.Item label="报告邮箱" required class="col-span-2">
            <Input
              v-model:value="taskForm.report_email"
              placeholder="支持多个邮箱，使用 ; 分隔，如 a@x.com;b@y.com"
            />
          </Form.Item>
          <Form.Item label="任务描述" class="col-span-2">
            <Input.TextArea v-model:value="taskForm.task_description" :rows="2" placeholder="请输入任务描述" />
          </Form.Item>
          <Form.Item label="SQL列表（每行一条）" required class="col-span-2">
            <Input.TextArea
              v-model:value="taskForm.sql_queries_text"
              :rows="5"
              placeholder="SELECT * FROM table_a LIMIT 100
SELECT count(*) AS total FROM table_b"
            />
          </Form.Item>
          <Form.Item label="分析提示词" class="col-span-2">
            <Input.TextArea
              v-model:value="taskForm.prompt"
              :rows="4"
              placeholder="请输入分析提示词，指导模型如何分析并输出报告。"
            />
          </Form.Item>
        </div>
      </Form>
    </Modal>

    <Modal
      v-model:open="logModalOpen"
      :title="`执行日志 - ${currentLogTaskName || '-'}`"
      width="1000px"
      :footer="null"
      destroy-on-close
    >
      <Table
        :columns="logColumns"
        :data-source="taskLogs"
        :loading="logLoading"
        :pagination="logPagination"
        :row-key="(record: AnalysisTaskLogRow, index: number) => record.id ?? `log-${index}`"
        :scroll="{ x: 1300 }"
        @change="handleLogTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="record.status === 'success' ? 'green' : record.status === 'failed' ? 'red' : 'blue'">
              {{ record.status || '-' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'result'">
            <Tooltip :title="record.result || '-'">
              <span class="inline-block max-w-[200px] truncate">{{ record.result || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'error_message'">
            <Tooltip :title="record.error_message || '-'">
              <span class="inline-block max-w-[200px] truncate">{{ record.error_message || '-' }}</span>
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
