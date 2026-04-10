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
  { title: '任务名称', dataIndex: 'taskName', key: 'taskName', width: 200 },
  { title: '任务类型', dataIndex: 'taskType', key: 'taskType', width: 120 },
  { title: '数据源', dataIndex: 'datasourceId', key: 'datasourceId', width: 100 },
  { title: '数据库', dataIndex: 'databaseName', key: 'databaseName', width: 150 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '开始时间', dataIndex: 'startTime', key: 'startTime', width: 180 },
  { title: '结束时间', dataIndex: 'endTime', key: 'endTime', width: 180 },
  { title: '执行时长', dataIndex: 'duration', key: 'duration', width: 100 },
  { title: '创建人', dataIndex: 'createdBy', key: 'createdBy', width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '操作', key: 'option', width: 180, fixed: 'right' },
];

function statusTag(status: string) {
  if (status === 'running') return { color: 'processing', text: '执行中' };
  if (status === 'success') return { color: 'success', text: '成功' };
  if (status === 'failed') return { color: 'error', text: '失败' };
  return { color: 'default', text: '待执行' };
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
    message.error(error?.message || '获取任务失败');
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
  if (!formModel.taskName || !formModel.databaseName) {
    message.warning('请完善任务名称和数据库');
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
      message.error(payload?.msg || '创建失败');
      return;
    }
    message.success('创建成功');
    modalVisible.value = false;
    fetchTasks();
  } catch (error: any) {
    message.error(error?.message || '创建失败');
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
      message.error(payload?.msg || '启动失败');
      return;
    }
    message.success('任务已启动');
    fetchTasks();
  } catch (error: any) {
    message.error(error?.message || '启动失败');
  }
}

async function handleDelete(id: number) {
  try {
    const response = await baseRequestClient.delete(`/v1/dataquality/tasks/${id}`);
    const payload = (response as any)?.data ?? response;
    if (payload?.code && payload.code !== 200) {
      message.error(payload?.msg || '删除失败');
      return;
    }
    message.success('删除成功');
    fetchTasks();
  } catch (error: any) {
    message.error(error?.message || '删除失败');
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
    <Card title="评估任务管理">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="任务类型" class="query-item">
            <Select v-model:value="queryForm.taskType" allow-clear class="query-control">
              <Select.Option value="全量">全量评估</Select.Option>
              <Select.Option value="增量">增量评估</Select.Option>
              <Select.Option value="定时">定时评估</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="状态" class="query-item">
            <Select v-model:value="queryForm.status" allow-clear class="query-control">
              <Select.Option value="pending">待执行</Select.Option>
              <Select.Option value="running">执行中</Select.Option>
              <Select.Option value="success">成功</Select.Option>
              <Select.Option value="failed">失败</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="openCreate">新建任务</Button>
            <Button type="primary" ghost @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
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
          <template v-if="column.key === 'status'">
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
            {{ record.duration ? `${record.duration}秒` : '-' }}
          </template>
          <template v-else-if="column.key === 'option'">
            <Space>
              <Button
                v-if="record.status === 'pending'"
                type="link"
                size="small"
                @click="handleStart(record)"
              >
                启动
              </Button>
              <Popconfirm title="确定要删除这个任务吗？" @confirm="handleDelete(record.id)">
                <Button type="link" danger size="small">删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalVisible"
      title="新建评估任务"
      :confirm-loading="saving"
      width="640px"
      @ok="submitCreate"
    >
      <Form layout="vertical">
        <Form.Item label="任务名称" required>
          <Input v-model:value="formModel.taskName" placeholder="请输入任务名称" />
        </Form.Item>
        <Form.Item label="任务类型" required>
          <Select v-model:value="formModel.taskType">
            <Select.Option value="全量">全量评估</Select.Option>
            <Select.Option value="增量">增量评估</Select.Option>
            <Select.Option value="定时">定时评估</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="数据库" required>
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
        <Form.Item label="表过滤条件(JSON)">
          <Input.TextArea
            v-model:value="formModel.tableFilter"
            :rows="3"
            placeholder='例如: {"include": ["table1", "table2"]}'
          />
        </Form.Item>
        <Form.Item label="调度配置(JSON)">
          <Input.TextArea
            v-model:value="formModel.scheduleConfig"
            :rows="3"
            placeholder='例如: {"cron": "0 0 2 * * ?"}'
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
