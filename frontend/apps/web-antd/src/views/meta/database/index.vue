<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'MetaDatabasePage' });

interface DatabaseItem {
  alias_name?: string;
  app_desc?: string;
  app_name?: string;
  app_owner?: string;
  app_owner_email?: string;
  app_owner_phone?: string;
  characters?: string;
  database_name: string;
  datasource_type: string;
  gmt_created?: string;
  gmt_updated?: string;
  host: string;
  id: number;
  is_deleted: number | string;
  port: number | string;
}

const loading = ref(false);
const dataSource = ref<DatabaseItem[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  database_name: '',
  datasource_type: '',
  host: '',
  is_deleted: '',
  port: '',
});
const editVisible = ref(false);
const editLoading = ref(false);
const editingId = ref<number>(0);
const editForm = reactive({
  alias_name: '',
  app_desc: '',
  app_name: '',
  app_owner: '',
  app_owner_email: '',
  app_owner_phone: '',
  is_deleted: 0,
});

const columns: TableColumnsType<DatabaseItem> = [
  { title: '数据库名', dataIndex: 'database_name', key: 'database_name', sorter: true },
  { title: '数据库别名', dataIndex: 'alias_name', key: 'alias_name' },
  { title: '库字符集', dataIndex: 'characters', key: 'characters' },
  { title: '数据库类型', dataIndex: 'datasource_type', key: 'datasource_type', sorter: true },
  { title: '所属主机', dataIndex: 'host', key: 'host' },
  { title: '所属端口', dataIndex: 'port', key: 'port' },
  { title: '应用名称', dataIndex: 'app_name', key: 'app_name' },
  { title: '应用描述', dataIndex: 'app_desc', key: 'app_desc' },
  { title: '应用负责人', dataIndex: 'app_owner', key: 'app_owner' },
  { title: '负责人邮箱', dataIndex: 'app_owner_email', key: 'app_owner_email' },
  { title: '负责人电话', dataIndex: 'app_owner_phone', key: 'app_owner_phone' },
  { title: '是否删除', dataIndex: 'is_deleted', key: 'is_deleted' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', sorter: true },
  { title: '操作', dataIndex: 'option', key: 'option', fixed: 'right', width: 140 },
];

async function fetchDatabases(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/database/list', {
      params: {
        ...queryForm,
        sorter: sorter ? JSON.stringify(sorter) : undefined,
      },
    });
    const payload = (response as any)?.data ?? response;
    const list = Array.isArray(payload?.data)
      ? payload.data
      : Array.isArray(payload)
        ? payload
        : [];
    dataSource.value = list;
    pagination.total = Number(payload?.total ?? list.length) || list.length;
  } catch (error: any) {
    message.error(error?.message || '数据库查询失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchDatabases();
}

function handleReset() {
  queryForm.database_name = '';
  queryForm.datasource_type = '';
  queryForm.host = '';
  queryForm.port = '';
  queryForm.is_deleted = '';
  pagination.current = 1;
  fetchDatabases();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  if (sorter?.field && sorter?.order) {
    fetchDatabases({ [sorter.field]: sorter.order });
    return;
  }
  fetchDatabases();
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

function deletedText(value: number | string) {
  return String(value) === '1' ? '是' : '否';
}

function openEdit(record: Record<string, any>) {
  editingId.value = record.id;
  editForm.alias_name = record.alias_name || '';
  editForm.app_name = record.app_name || '';
  editForm.app_desc = record.app_desc || '';
  editForm.app_owner = record.app_owner || '';
  editForm.app_owner_email = record.app_owner_email || '';
  editForm.app_owner_phone = record.app_owner_phone || '';
  editForm.is_deleted = Number(record.is_deleted) || 0;
  editVisible.value = true;
}

async function handleUpdateSubmit() {
  if (!editingId.value) return;
  editLoading.value = true;
  try {
    const response = await baseRequestClient.put('/v1/meta/database/list', {
      ...editForm,
      id: editingId.value,
      is_deleted: Number(editForm.is_deleted) || 0,
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(payload?.msg || '修改失败');
      return;
    }
    message.success('修改成功');
    editVisible.value = false;
    fetchDatabases();
  } catch (error: any) {
    message.error(error?.message || '修改失败');
  } finally {
    editLoading.value = false;
  }
}

onMounted(fetchDatabases);
</script>

<template>
  <div class="p-5">
    <Card title="数据库列表">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="数据库名" class="query-item">
            <Input
              v-model:value="queryForm.database_name"
              placeholder="请输入数据库名"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="数据库类型" class="query-item">
            <Input
              v-model:value="queryForm.datasource_type"
              placeholder="请输入数据库类型"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属主机" class="query-item">
            <Input
              v-model:value="queryForm.host"
              placeholder="请输入所属主机"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属端口" class="query-item">
            <Input
              v-model:value="queryForm.port"
              placeholder="请输入所属端口"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="是否删除" class="query-item">
            <Select v-model:value="queryForm.is_deleted" allow-clear class="query-control">
              <Select.Option value="0">否</Select.Option>
              <Select.Option value="1">是</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: DatabaseItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'is_deleted'">
            {{ deletedText(record.is_deleted) }}
          </template>
          <template v-else-if="column.key === 'option'">
            <a @click="openEdit(record)">修改业务信息</a>
          </template>
          <template v-else-if="column.key === 'gmt_created'">
            {{ formatDate(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatDate(record.gmt_updated) }}
          </template>
        </template>
      </Table>

      <Modal
        v-model:open="editVisible"
        title="修改业务信息"
        :confirm-loading="editLoading"
        @ok="handleUpdateSubmit"
      >
        <Form layout="vertical">
          <Form.Item label="数据库别名">
            <Input v-model:value="editForm.alias_name" />
          </Form.Item>
          <Form.Item label="应用名称">
            <Input v-model:value="editForm.app_name" />
          </Form.Item>
          <Form.Item label="应用描述">
            <Input v-model:value="editForm.app_desc" />
          </Form.Item>
          <Form.Item label="应用负责人">
            <Input v-model:value="editForm.app_owner" />
          </Form.Item>
          <Form.Item label="负责人邮箱">
            <Input v-model:value="editForm.app_owner_email" />
          </Form.Item>
          <Form.Item label="负责人电话">
            <Input v-model:value="editForm.app_owner_phone" />
          </Form.Item>
          <Form.Item label="是否删除">
            <Select v-model:value="editForm.is_deleted">
              <Select.Option :value="0">否</Select.Option>
              <Select.Option :value="1">是</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  </div>
</template>

<style scoped>
.query-grid {
  column-gap: 16px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  row-gap: 12px;
}

:deep(.query-item) {
  margin-bottom: 0;
  min-width: 0;
}

:deep(.query-item .ant-form-item-row) {
  align-items: center;
  display: flex;
}

:deep(.query-item .ant-form-item-label) {
  flex: 0 0 5.5rem;
  max-width: 7rem;
  padding-right: 8px;
  text-align: right;
}

:deep(.query-item .ant-form-item-control) {
  flex: 1;
  min-width: 0;
}

:deep(.query-control) {
  max-width: 100%;
  min-width: 0;
  width: 100%;
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

  .query-actions {
    justify-content: flex-start;
  }
}
</style>
