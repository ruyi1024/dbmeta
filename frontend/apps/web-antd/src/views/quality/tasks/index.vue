<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';
import { checkPermission } from '#/utils/check-permission';

interface TaskListItem {
  createdAt?: string;
  createdBy?: string;
  databaseName: string;
  datasourceId?: number;
  duration?: number;
  endTime?: string;
  errorMessage?: string;
  id: number;
  resultSummary?: string;
  scheduleConfig?: string;
  startTime?: string;
  status: string;
  tableFilter?: string;
  taskName: string;
  taskType: string;
  updatedAt?: string;
}

interface DatabaseInfo {
  alias_name?: string;
  database_name: string;
  id: number;
}

const loading = ref(false);
const saving = ref(false);
const modalVisible = ref(false);

const dataSource = ref<TaskListItem[]>([]);
const databaseList = ref<DatabaseInfo[]>([]);

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  status: undefined as string | undefined,
  taskType: undefined as string | undefined,
});

const formModel = reactive({
  createdBy: 'admin',
  databaseName: '',
  datasourceId: undefined as number | undefined,
  scheduleConfig: '',
  tableFilter: '',
  taskName: '',
  taskType: '全量',
});

const columns: TableColumnsType<TaskListItem> = [
  { title: $t('page.qualityTasks.columns.taskName'), dataIndex: 'taskName', key: 'taskName', width: 200 },
  { title: $t('page.qualityTasks.columns.taskType'), dataIndex: 'taskType', key: 'taskType', width: 120 },
  { title: $t('page.qualityTasks.columns.datasource'), dataIndex: 'datasourceId', key: 'datasourceId', width: 100 },
  { title: $t('page.qualityTasks.columns.databaseName'), dataIndex: 'databaseName', key: 'databaseName', width: 150 },
  { title: $t('page.qualityTasks.columns.status'), dataIndex: 'status', key: 'status', width: 100 },
  { title: $t('page.qualityTasks.columns.startTime'), dataIndex: 'startTime', key: 'startTime', width: 180 },
  { title: $t('page.qualityTasks.columns.endTime'), dataIndex: 'endTime', key: 'endTime', width: 180 },
  { title: $t('page.qualityTasks.columns.duration'), dataIndex: 'duration', key: 'duration', width: 100 },
  { title: $t('page.qualityTasks.columns.createdBy'), dataIndex: 'createdBy', key: 'createdBy', width: 100 },
  { title: $t('page.qualityTasks.columns.createdAt'), dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: $t('page.qualityTasks.columns.operation'), key: 'option', width: 180, fixed: 'right' },
];

function statusTag(status: string) {
  if (status === 'running') return { color: 'processing', text: $t('page.qualityTasks.status.running') };
  if (status === 'success') return { color: 'success', text: $t('page.qualityTasks.status.success') };
  if (status === 'failed') return { color: 'error', text: $t('page.qualityTasks.status.failed') };
  return { color: 'default', text: $t('page.qualityTasks.status.pending') };
}

function taskTypeLabel(taskType: string) {
  switch (taskType) {
    case '全量':
      return $t('page.qualityTasks.taskTypeOption.full');
    case '增量':
      return $t('page.qualityTasks.taskTypeOption.incremental');
    case '定时':
      return $t('page.qualityTasks.taskTypeOption.scheduled');
    default:
      return taskType;
  }
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

async function fetchTasks() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/dataquality/tasks', {
      params: {
        ...queryForm,
        current: pagination.current,
        pageSize: pagination.pageSize,
      },
    });
    const payload = (response as any)?.data ?? response;
    const list = payload?.data?.list ?? payload?.list ?? [];
    const total = payload?.data?.total ?? payload?.total ?? 0;
    dataSource.value = Array.isArray(list) ? list : [];
    pagination.total = Number(total) || dataSource.value.length;
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityTasks.message.fetchFailed'));
  } finally {
    loading.value = false;
  }
}

async function fetchDatabaseList() {
  try {
    const response = await baseRequestClient.get('/v1/meta/database/list', {
      params: { is_deleted: 0 },
    });
    const payload = (response as any)?.data ?? response;
    const list = payload?.data ?? payload;
    databaseList.value = Array.isArray(list) ? list : [];
  } catch {
    databaseList.value = [];
  }
}

function openCreate() {
  if (!checkPermission()) return;
  formModel.taskName = '';
  formModel.taskType = '全量';
  formModel.databaseName = '';
  formModel.datasourceId = undefined;
  formModel.tableFilter = '';
  formModel.scheduleConfig = '';
  modalVisible.value = true;
  fetchDatabaseList();
}

function onDatabaseChange(value: string) {
  const db = databaseList.value.find((item) => item.database_name === value);
  if (db) formModel.datasourceId = db.id;
}

async function submitCreate() {
  if (!checkPermission()) return;
  if (!formModel.taskName || !formModel.databaseName) {
    message.warning($t('page.qualityTasks.message.nameDbRequired'));
    return;
  }
  saving.value = true;
  try {
    const response = await baseRequestClient.post('/v1/dataquality/tasks', {
      ...formModel,
      status: 'pending',
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.code && payload.code !== 200) {
      message.error(payload?.msg || $t('page.qualityTasks.message.createFailed'));
      return;
    }
    message.success($t('page.qualityTasks.message.createSuccess'));
    modalVisible.value = false;
    fetchTasks();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityTasks.message.createFailed'));
  } finally {
    saving.value = false;
  }
}

async function handleStart(record: TaskListItem) {
  try {
    const response = await baseRequestClient.put('/v1/dataquality/tasks/status', {
      id: record.id,
      status: 'running',
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.code && payload.code !== 200) {
      message.error(payload?.msg || $t('page.qualityTasks.message.startFailed'));
      return;
    }
    message.success($t('page.qualityTasks.message.started'));
    fetchTasks();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityTasks.message.startFailed'));
  }
}

async function handleDelete(id: number) {
  if (!checkPermission()) return;
  try {
    const response = await baseRequestClient.delete(`/v1/dataquality/tasks/${id}`);
    const payload = (response as any)?.data ?? response;
    if (payload?.code && payload.code !== 200) {
      message.error(payload?.msg || $t('page.qualityTasks.message.deleteFailed'));
      return;
    }
    message.success($t('page.qualityTasks.message.deleteSuccess'));
    fetchTasks();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityTasks.message.deleteFailed'));
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchTasks();
}

function handleReset() {
  queryForm.taskType = undefined;
  queryForm.status = undefined;
  pagination.current = 1;
  fetchTasks();
}

function handleTableChange(page: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  fetchTasks();
}

onMounted(fetchTasks);
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.qualityTasks.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.qualityTasks.form.taskType')" class="query-item">
            <Select v-model:value="queryForm.taskType" allow-clear class="query-control">
              <Select.Option value="全量">{{ $t('page.qualityTasks.taskTypeOption.full') }}</Select.Option>
              <Select.Option value="增量">{{ $t('page.qualityTasks.taskTypeOption.incremental') }}</Select.Option>
              <Select.Option value="定时">{{ $t('page.qualityTasks.taskTypeOption.scheduled') }}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item :label="$t('page.qualityTasks.form.status')" class="query-item">
            <Select v-model:value="queryForm.status" allow-clear class="query-control">
              <Select.Option value="pending">{{ $t('page.qualityTasks.status.pending') }}</Select.Option>
              <Select.Option value="running">{{ $t('page.qualityTasks.status.running') }}</Select.Option>
              <Select.Option value="success">{{ $t('page.qualityTasks.status.success') }}</Select.Option>
              <Select.Option value="failed">{{ $t('page.qualityTasks.status.failed') }}</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="openCreate">{{ $t('page.qualityTasks.action.newTask') }}</Button>
            <Button type="primary" ghost @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: TaskListItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'taskType'">
            {{ taskTypeLabel(record.taskType) }}
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">{{ statusTag(record.status).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'startTime'">
            {{ formatDate(record.startTime) }}
          </template>
          <template v-else-if="column.key === 'endTime'">
            {{ formatDate(record.endTime) }}
          </template>
          <template v-else-if="column.key === 'createdAt'">
            {{ formatDate(record.createdAt) }}
          </template>
          <template v-else-if="column.key === 'duration'">
            {{
              record.duration
                ? $t('page.qualityTasks.durationSeconds', { n: record.duration })
                : '-'
            }}
          </template>
          <template v-else-if="column.key === 'option'">
            <Space>
              <Button
                v-if="record.status === 'pending'"
                type="link"
                size="small"
                @click="handleStart(record)"
              >
                {{ $t('page.qualityTasks.action.start') }}
              </Button>
              <Popconfirm :title="$t('page.qualityTasks.confirmDelete')" @confirm="handleDelete(record.id)">
                <Button type="link" danger size="small">{{ $t('page.common.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalVisible"
      :title="$t('page.qualityTasks.modal.createTitle')"
      :confirm-loading="saving"
      width="640px"
      @ok="submitCreate"
    >
      <Form layout="vertical">
        <Form.Item :label="$t('page.qualityTasks.formModal.taskName')" required>
          <Input v-model:value="formModel.taskName" :placeholder="$t('page.qualityTasks.placeholder.taskName')" />
        </Form.Item>
        <Form.Item :label="$t('page.qualityTasks.formModal.taskType')" required>
          <Select v-model:value="formModel.taskType">
            <Select.Option value="全量">{{ $t('page.qualityTasks.taskTypeOption.full') }}</Select.Option>
            <Select.Option value="增量">{{ $t('page.qualityTasks.taskTypeOption.incremental') }}</Select.Option>
            <Select.Option value="定时">{{ $t('page.qualityTasks.taskTypeOption.scheduled') }}</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item :label="$t('page.qualityTasks.formModal.database')" required>
          <Select
            v-model:value="formModel.databaseName"
            show-search
            @change="onDatabaseChange"
          >
            <Select.Option
              v-for="item in databaseList"
              :key="item.id"
              :value="item.database_name"
            >
              {{ item.alias_name ? `${item.alias_name}(${item.database_name})` : item.database_name }}
            </Select.Option>
          </Select>
        </Form.Item>
        <Form.Item :label="$t('page.qualityTasks.formModal.tableFilter')">
          <Input.TextArea
            v-model:value="formModel.tableFilter"
            :rows="3"
            :placeholder="$t('page.qualityTasks.placeholder.tableFilter')"
          />
        </Form.Item>
        <Form.Item :label="$t('page.qualityTasks.formModal.scheduleConfig')">
          <Input.TextArea
            v-model:value="formModel.scheduleConfig"
            :rows="3"
            :placeholder="$t('page.qualityTasks.placeholder.scheduleConfig')"
          />
        </Form.Item>
      </Form>
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
  margin-top: 12px;
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
