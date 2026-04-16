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
import { $t } from '#/locales';

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
  { title: $t('page.metaInstance.columns.type'), dataIndex: 'type', key: 'type', sorter: true },
  { title: $t('page.metaInstance.columns.name'), dataIndex: 'name', key: 'name', sorter: true },
  { title: $t('page.metaInstance.columns.host'), dataIndex: 'host', key: 'host', sorter: true },
  { title: $t('page.metaInstance.columns.port'), dataIndex: 'port', key: 'port' },
  { title: $t('page.metaInstance.columns.role'), dataIndex: 'role', key: 'role' },
  { title: $t('page.metaInstance.columns.enable'), dataIndex: 'enable', key: 'enable' },
  { title: $t('page.metaInstance.columns.createdAt'), dataIndex: 'gmt_created', key: 'gmt_created', sorter: true },
  { title: $t('page.metaInstance.columns.updatedAt'), dataIndex: 'gmt_updated', key: 'gmt_updated', sorter: true },
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
    message.error(error?.message || $t('page.metaInstance.message.fetchFailed'));
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
  return String(role) === '1'
    ? $t('page.metaInstance.option.primary')
    : String(role) === '2'
      ? $t('page.metaInstance.option.secondary')
      : '-';
}

function enableText(enable: number | string) {
  return String(enable) === '1'
    ? $t('page.metaInstance.option.yes')
    : String(enable) === '0'
      ? $t('page.metaInstance.option.no')
      : '-';
}

onMounted(() => {
  fetchInstances();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.metaInstance.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.metaInstance.columns.type')" class="query-item">
            <Input
              v-model:value="queryForm.type"
              :placeholder="$t('page.metaInstance.placeholder.type')"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaInstance.columns.name')" class="query-item">
            <Input
              v-model:value="queryForm.name"
              :placeholder="$t('page.metaInstance.placeholder.name')"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaInstance.columns.host')" class="query-item">
            <Input
              v-model:value="queryForm.host"
              :placeholder="$t('page.metaInstance.placeholder.host')"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaInstance.columns.port')" class="query-item">
            <Input
              v-model:value="queryForm.port"
              :placeholder="$t('page.metaInstance.placeholder.port')"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaInstance.columns.role')" class="query-item">
            <Select v-model:value="queryForm.role" allow-clear class="query-control">
              <Select.Option value="1">{{ $t('page.metaInstance.option.primary') }}</Select.Option>
              <Select.Option value="2">{{ $t('page.metaInstance.option.secondary') }}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item :label="$t('page.metaInstance.columns.enable')" class="query-item">
            <Select v-model:value="queryForm.enable" allow-clear class="query-control">
              <Select.Option value="1">{{ $t('page.metaInstance.option.yes') }}</Select.Option>
              <Select.Option value="0">{{ $t('page.metaInstance.option.no') }}</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
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
