<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

defineOptions({ name: 'DataCapacityDatabaseQuery' });

interface DatabaseCapacityRow {
  id: number;
  databaseName: string;
  datasourceType: string;
  host?: string;
  port?: string;
  tableCount: number;
  dataSize: string;
  dataSizeBytes: number;
  rowCount: number;
  dataSizeIncr: string;
  dataSizeIncrBytes: number;
  rowCountIncr: number;
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

function parsePagedPumpkin(response: unknown): { rows: DatabaseCapacityRow[]; total: number } {
  const raw = unwrapAxiosData(response);
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) {
    return { rows: [], total: 0 };
  }
  const b = raw as Record<string, unknown>;
  const list = b.data;
  const total = Number(b.total ?? 0) || 0;
  if (!Array.isArray(list)) {
    return { rows: [], total: 0 };
  }
  const rows = list.map((item: any, index: number) => {
    const dataSizeBytes = typeof item.dataSizeBytes === 'number' ? item.dataSizeBytes : 0;
    const dataSizeIncrBytes =
      typeof item.dataSizeIncrBytes === 'number' ? item.dataSizeIncrBytes : 0;
    return {
      id: item.id ?? index,
      databaseName: String(item.databaseName ?? ''),
      datasourceType: String(item.datasourceType ?? ''),
      host: item.host ?? '',
      port: item.port ?? '',
      tableCount: Number(item.tableCount) || 0,
      dataSize: String(item.dataSize ?? ''),
      dataSizeBytes,
      rowCount: Number(item.rowCount) || 0,
      dataSizeIncr: String(item.dataSizeIncr ?? '0 B'),
      dataSizeIncrBytes,
      rowCountIncr: Number(item.rowCountIncr) || 0,
    } as DatabaseCapacityRow;
  });
  return { rows, total };
}

const loading = ref(false);
const dataSource = ref<DatabaseCapacityRow[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => `${$t('page.common.total')} ${total} ${$t('page.common.records')}`,
  total: 0,
});

const queryForm = reactive({
  databaseName: '',
  datasourceType: '',
  host: '',
  port: '',
});

const sortField = ref<string | undefined>();
const sortOrder = ref<'ascend' | 'descend' | undefined>();

const columns: TableColumnsType<DatabaseCapacityRow> = [
  {
    title: $t('page.capacity.databaseQuery.columns.databaseName'),
    dataIndex: 'databaseName',
    ellipsis: true,
    key: 'databaseName',
    sorter: true,
    width: 180,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.datasourceType'),
    dataIndex: 'datasourceType',
    key: 'datasourceType',
    sorter: true,
    width: 120,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.host'),
    dataIndex: 'host',
    ellipsis: true,
    key: 'host',
    width: 150,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.port'),
    dataIndex: 'port',
    key: 'port',
    width: 80,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.tableCount'),
    dataIndex: 'tableCount',
    key: 'tableCount',
    sorter: true,
    width: 100,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.dataSize'),
    dataIndex: 'dataSize',
    key: 'dataSize',
    sorter: true,
    width: 130,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.rowCount'),
    dataIndex: 'rowCount',
    key: 'rowCount',
    sorter: true,
    width: 130,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.dataSizeIncr'),
    dataIndex: 'dataSizeIncr',
    key: 'dataSizeIncr',
    sorter: true,
    width: 140,
  },
  {
    title: $t('page.capacity.databaseQuery.columns.rowCountIncr'),
    dataIndex: 'rowCountIncr',
    key: 'rowCountIncr',
    sorter: true,
    width: 140,
  },
];

function formatRowCount(count: number): string {
  if (!count) return '0';
  if (count >= 1_000_000) {
    return `${(count / 1_000_000).toFixed(2)}M`;
  }
  if (count >= 1000) {
    return `${(count / 1000).toFixed(2)}K`;
  }
  return count.toLocaleString();
}

function formatRowIncrDisplay(count: number): string {
  if (count === 0) return '0';
  const abs = Math.abs(count);
  let formatted: string;
  if (abs >= 1_000_000) {
    formatted = `${(count / 1_000_000).toFixed(2)}M`;
  } else if (abs >= 1000) {
    formatted = `${(count / 1000).toFixed(2)}K`;
  } else {
    formatted = String(count);
  }
  const sign = count > 0 ? '+' : '';
  return `${sign}${formatted}`;
}

async function fetchDatabaseCapacity() {
  loading.value = true;
  try {
    const params: Record<string, string | number> = {
      current: pagination.current,
      pageSize: pagination.pageSize,
    };
    if (queryForm.databaseName) {
      params.databaseName = queryForm.databaseName;
    }
    if (queryForm.datasourceType) {
      params.datasourceType = queryForm.datasourceType;
    }
    if (queryForm.host) {
      params.host = queryForm.host;
    }
    if (queryForm.port) {
      params.port = queryForm.port;
    }
    if (sortField.value && sortOrder.value) {
      params.sortField = sortField.value;
      params.sortOrder = sortOrder.value === 'ascend' ? 'asc' : 'desc';
    }

    const response = await baseRequestClient.get('/v1/pumpkin/capacity/database/top10', {
      params,
    });
    const { rows, total } = parsePagedPumpkin(response);
    dataSource.value = rows;
    pagination.total = total;
  } catch (error: any) {
    message.error(error?.message || $t('page.capacity.databaseQuery.message.fetchFailed'));
    dataSource.value = [];
    pagination.total = 0;
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchDatabaseCapacity();
}

function handleReset() {
  queryForm.databaseName = '';
  queryForm.datasourceType = '';
  queryForm.host = '';
  queryForm.port = '';
  sortField.value = undefined;
  sortOrder.value = undefined;
  pagination.current = 1;
  fetchDatabaseCapacity();
}

function handleTableChange(pag: any, _filters: unknown, sorter: any) {
  pagination.current = pag?.current ?? 1;
  pagination.pageSize = pag?.pageSize ?? 10;

  let s = sorter;
  if (Array.isArray(sorter)) {
    s = sorter[0];
  }
  const colKey = s?.field ?? s?.columnKey;
  if (colKey && s?.order) {
    sortField.value = String(colKey);
    sortOrder.value = s.order;
  } else {
    sortField.value = undefined;
    sortOrder.value = undefined;
  }
  fetchDatabaseCapacity();
}

onMounted(() => {
  fetchDatabaseCapacity();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.capacity.databaseQuery.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.capacity.databaseQuery.form.databaseName')" class="query-item">
            <Input
              v-model:value="queryForm.databaseName"
              allow-clear
              class="query-control"
              :placeholder="$t('page.capacity.databaseQuery.placeholder.databaseName')"
            />
          </Form.Item>
          <Form.Item :label="$t('page.capacity.databaseQuery.form.datasourceType')" class="query-item">
            <Input
              v-model:value="queryForm.datasourceType"
              allow-clear
              class="query-control"
              :placeholder="$t('page.capacity.databaseQuery.placeholder.datasourceType')"
            />
          </Form.Item>
          <Form.Item :label="$t('page.capacity.databaseQuery.form.host')" class="query-item">
            <Input
              v-model:value="queryForm.host"
              allow-clear
              class="query-control"
              :placeholder="$t('page.capacity.databaseQuery.placeholder.host')"
            />
          </Form.Item>
          <Form.Item :label="$t('page.capacity.databaseQuery.form.port')" class="query-item">
            <Input
              v-model:value="queryForm.port"
              allow-clear
              class="query-control"
              :placeholder="$t('page.capacity.databaseQuery.placeholder.port')"
            />
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
        :row-key="(record: DatabaseCapacityRow) => record.id"
        :scroll="{ x: 1400 }"
        size="middle"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'tableCount'">
            {{ record.tableCount || 0 }}
          </template>
          <template v-else-if="column.key === 'dataSize'">
            <span class="font-medium text-[#1890ff]">{{ record.dataSize }}</span>
          </template>
          <template v-else-if="column.key === 'rowCount'">
            {{ formatRowCount(record.rowCount) }}
          </template>
          <template v-else-if="column.key === 'dataSizeIncr'">
            <template v-if="(record.dataSizeIncrBytes || 0) === 0">
              <span>0 B</span>
            </template>
            <template v-else>
              <span class="font-medium text-foreground">
                <span
                  :style="{
                    color: record.dataSizeIncrBytes > 0 ? '#52c41a' : '#ff4d4f',
                    fontSize: '12px',
                    marginRight: '4px',
                  }"
                >
                  {{ record.dataSizeIncrBytes > 0 ? '↑' : '↓' }}
                </span>
                {{ record.dataSizeIncr }}
              </span>
            </template>
          </template>
          <template v-else-if="column.key === 'rowCountIncr'">
            <span v-if="record.rowCountIncr === 0" class="font-medium">0</span>
            <span v-else class="font-medium text-foreground">
              <span
                :style="{
                  color: record.rowCountIncr > 0 ? '#52c41a' : '#ff4d4f',
                  fontSize: '12px',
                  marginRight: '4px',
                }"
              >
                {{ record.rowCountIncr > 0 ? '↑' : '↓' }}
              </span>
              {{ formatRowIncrDisplay(record.rowCountIncr) }}
            </span>
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
