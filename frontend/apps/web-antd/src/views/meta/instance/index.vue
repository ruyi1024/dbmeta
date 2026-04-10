<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Select,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'MetaInstancePage' });

interface InstanceItem {
  enable: number | string;
  gmt_created?: string;
  gmt_updated?: string;
  host: string;
  id: number;
  name: string;
  port: number | string;
  role: number | string;
  type: string;
}

const loading = ref(false);
const dataSource = ref<InstanceItem[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  enable: '',
  host: '',
  name: '',
  port: '',
  role: '',
  type: '',
});
const columns: TableColumnsType<InstanceItem> = [
  { title: '实例类型', dataIndex: 'type', key: 'type', sorter: true },
  { title: '实例名', dataIndex: 'name', key: 'name', sorter: true },
  { title: '实例主机', dataIndex: 'host', key: 'host', sorter: true },
  { title: '实例端口', dataIndex: 'port', key: 'port' },
  { title: '角色', dataIndex: 'role', key: 'role' },
  { title: '是否启用', dataIndex: 'enable', key: 'enable' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', sorter: true },
];

async function fetchInstances(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get<{
      data?: InstanceItem[];
      msg?: string;
      success?: boolean;
      total?: number;
    }>('/v1/meta/instance/list', {
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
    message.error(error?.message || '实例查询失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchInstances();
}

function handleReset() {
  queryForm.type = '';
  queryForm.name = '';
  queryForm.host = '';
  queryForm.port = '';
  queryForm.role = '';
  queryForm.enable = '';
  pagination.current = 1;
  fetchInstances();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;

  if (sorter?.field && sorter?.order) {
    fetchInstances({
      [sorter.field]: sorter.order,
    });
    return;
  }
  fetchInstances();
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

function roleText(role: number | string) {
  return String(role) === '1' ? '主' : String(role) === '2' ? '备' : '-';
}

function enableText(enable: number | string) {
  return String(enable) === '1' ? '是' : String(enable) === '0' ? '否' : '-';
}

onMounted(() => {
  fetchInstances();
});
</script>

<template>
  <div class="p-5">
    <Card title="数据库实例列表">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="实例类型" class="query-item">
            <Input
              v-model:value="queryForm.type"
              placeholder="请输入实例类型"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="实例名" class="query-item">
            <Input
              v-model:value="queryForm.name"
              placeholder="请输入实例名"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="实例主机" class="query-item">
            <Input
              v-model:value="queryForm.host"
              placeholder="请输入实例主机"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="实例端口" class="query-item">
            <Input
              v-model:value="queryForm.port"
              placeholder="请输入实例端口"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="角色" class="query-item">
            <Select v-model:value="queryForm.role" allow-clear class="query-control">
              <Select.Option value="1">主</Select.Option>
              <Select.Option value="2">备</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="是否启用" class="query-item">
            <Select v-model:value="queryForm.enable" allow-clear class="query-control">
              <Select.Option value="1">是</Select.Option>
              <Select.Option value="0">否</Select.Option>
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
        :row-key="(record: InstanceItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'role'">
            {{ roleText(record.role) }}
          </template>
          <template v-else-if="column.key === 'enable'">
            {{ enableText(record.enable) }}
          </template>
          <template v-else-if="column.key === 'gmt_created'">
            {{ formatDate(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatDate(record.gmt_updated) }}
          </template>
        </template>
      </Table>
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
}

.query-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}

@media (max-width: 768px) {
  .query-actions {
    justify-content: flex-start;
  }
}
</style>
