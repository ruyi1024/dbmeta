<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'MetaTablePage' });

interface TableItem {
  ai_comment?: string;
  ai_fixed?: number;
  characters?: string;
  database_name: string;
  datasource_type: string;
  gmt_created?: string;
  gmt_updated?: string;
  host: string;
  id: number;
  port: number | string;
  table_comment?: string;
  table_name: string;
  table_type?: string;
}

const loading = ref(false);
const dataSource = ref<TableItem[]>([]);
const selectedRowKeys = ref<number[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  database_name: '',
  datasource_type: '',
  host: '',
  port: '',
  table_name: '',
});

const columns: TableColumnsType<TableItem> = [
  { title: '数据表名', dataIndex: 'table_name', key: 'table_name', sorter: true },
  { title: '表类型', dataIndex: 'table_type', key: 'table_type' },
  { title: '表字符集', dataIndex: 'characters', key: 'characters' },
  { title: '表备注', dataIndex: 'table_comment', key: 'table_comment' },
  { title: 'AI注释生成', dataIndex: 'ai_comment', key: 'ai_comment' },
  { title: 'AI注释应用', dataIndex: 'ai_fixed', key: 'ai_fixed' },
  { title: '所属数据库', dataIndex: 'database_name', key: 'database_name', sorter: true },
  { title: '数据库类型', dataIndex: 'datasource_type', key: 'datasource_type', sorter: true },
  { title: '所属主机', dataIndex: 'host', key: 'host' },
  { title: '所属端口', dataIndex: 'port', key: 'port' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', sorter: true },
];

async function fetchTables(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/table/list', {
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
    message.error(error?.message || '数据表查询失败');
  } finally {
    loading.value = false;
  }
}

async function handleBatchUpdate(aiFixed: number) {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择要操作的表');
    return;
  }
  try {
    const response = await baseRequestClient.put('/v1/meta/table/batch-update-ai-fixed', {
      ai_fixed: aiFixed,
      ids: selectedRowKeys.value,
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(payload?.msg || '批量操作失败');
      return;
    }
    message.success(payload?.msg || '批量操作成功');
    selectedRowKeys.value = [];
    fetchTables();
  } catch (error: any) {
    message.error(error?.message || '批量操作失败');
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchTables();
}

function handleReset() {
  queryForm.datasource_type = '';
  queryForm.host = '';
  queryForm.port = '';
  queryForm.database_name = '';
  queryForm.table_name = '';
  pagination.current = 1;
  fetchTables();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  if (sorter?.field && sorter?.order) {
    fetchTables({ [sorter.field]: sorter.order });
    return;
  }
  fetchTables();
}

function handleRowSelectionChange(keys: (number | string)[]) {
  selectedRowKeys.value = keys.map((key) => Number(key));
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

function aiFixedStatus(value?: number) {
  if (value === 1) return { status: 'error' as const, text: '不应用' };
  if (value === 2) return { status: 'warning' as const, text: '待应用' };
  if (value === 3) return { status: 'success' as const, text: '已应用' };
  return { status: 'default' as const, text: '待审核' };
}

onMounted(fetchTables);
</script>

<template>
  <div class="p-5">
    <Card title="数据表列表">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="数据表名" class="query-item">
            <Input
              v-model:value="queryForm.table_name"
              placeholder="请输入数据表名"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属库名" class="query-item">
            <Input
              v-model:value="queryForm.database_name"
              placeholder="请输入所属库名"
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
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
          </Space>
        </div>
      </Form>

      <div class="mb-3 flex justify-end">
        <Space>
          <Button danger :disabled="selectedRowKeys.length === 0" @click="handleBatchUpdate(1)">
            不应用AI注释 ({{ selectedRowKeys.length }})
          </Button>
          <Button type="primary" :disabled="selectedRowKeys.length === 0" @click="handleBatchUpdate(2)">
            应用AI注释 ({{ selectedRowKeys.length }})
          </Button>
        </Space>
      </div>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-selection="{
          selectedRowKeys,
          onChange: handleRowSelectionChange,
          getCheckboxProps: (record: TableItem) => ({
            disabled: !record.ai_comment,
          }),
        }"
        :row-key="(record: TableItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'ai_comment'">
            <span v-if="record.ai_comment">{{ record.ai_comment }}</span>
            <span v-else style="color: #999">暂无AI注释</span>
          </template>
          <template v-else-if="column.key === 'ai_fixed'">
            <Badge :status="aiFixedStatus(record.ai_fixed).status" :text="aiFixedStatus(record.ai_fixed).text" />
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
