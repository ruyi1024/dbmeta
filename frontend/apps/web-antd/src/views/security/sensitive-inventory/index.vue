<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Button, Card, Form, Input, Space, Table, Tag, Tooltip } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

defineOptions({ name: 'DataSecuritySensitiveInventory' });

interface SensitiveMetaRow {
  id?: number;
  datasource_type?: string;
  host?: string;
  port?: string;
  database_name?: string;
  table_name?: string;
  table_comment?: string;
  column_name?: string;
  column_comment?: string;
  rule_type?: string;
  rule_key?: string;
  rule_name?: string;
  level?: number;
  status?: number;
  gmt_created?: string;
  gmt_updated?: string;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') {
    return {};
  }
  const r = response as Record<string, unknown>;
  if (
    'data' in r &&
    r.data !== undefined &&
    typeof r.data === 'object' &&
    'status' in r &&
    typeof (r as { status?: unknown }).status === 'number'
  ) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r as Record<string, unknown>;
}

function formatTime(v?: string) {
  if (!v) return '-';
  return dayjs(v).isValid() ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : v;
}

/** 与库表注释一致：1 低敏、2 高敏；兼容旧数据 0/1 展示 */
function levelTag(level?: number) {
  if (level === 0) {
    return { color: 'orange', text: $t('page.securitySensitiveInventory.level.low') };
  }
  if (level === 1) {
    return { color: 'orange', text: $t('page.securitySensitiveInventory.level.low') };
  }
  if (level === 2) {
    return { color: 'red', text: $t('page.securitySensitiveInventory.level.high') };
  }
  return { color: 'default', text: level !== undefined ? String(level) : '-' };
}

function statusTag(status?: number) {
  if (status === -1) {
    return { color: 'orange', text: $t('page.securitySensitiveInventory.status.suspected') };
  }
  if (status === 0) {
    return { color: 'default', text: $t('page.securitySensitiveInventory.status.nonSensitive') };
  }
  if (status === 1) {
    return { color: 'green', text: $t('page.securitySensitiveInventory.status.confirmed') };
  }
  return { color: 'default', text: status !== undefined ? String(status) : '-' };
}

const loading = ref(false);
const allRows = ref<SensitiveMetaRow[]>([]);

const searchForm = reactive({
  datasource_type: '',
  host: '',
  port: '',
  database_name: '',
  table_name: '',
});

const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `${$t('page.common.total')} ${total} ${$t('page.common.records')}`,
  pageSizeOptions: ['10', '15', '30', '50', '100'],
});

const pagedRows = computed(() => {
  const current = pagination.current ?? 1;
  const pageSize = pagination.pageSize ?? 10;
  const start = (current - 1) * pageSize;
  return allRows.value.slice(start, start + pageSize);
});

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, string> = {};
    if (searchForm.datasource_type.trim()) {
      params.datasource_type = searchForm.datasource_type.trim();
    }
    if (searchForm.host.trim()) {
      params.host = searchForm.host.trim();
    }
    if (searchForm.port.trim()) {
      params.port = searchForm.port.trim();
    }
    if (searchForm.database_name.trim()) {
      params.database_name = searchForm.database_name.trim();
    }
    if (searchForm.table_name.trim()) {
      params.table_name = searchForm.table_name.trim();
    }
    const response = await baseRequestClient.get('/v1/sensitive/meta', { params });
    const body = extractApiBody(response);
    const raw = body?.data;
    const list = Array.isArray(raw) ? (raw as SensitiveMetaRow[]) : [];
    allRows.value = list;
    pagination.total = list.length;
    pagination.current = 1;
  } catch {
    allRows.value = [];
    pagination.total = 0;
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchList();
}

function handleTableChange(pag: TablePaginationConfig) {
  if (pag.current !== undefined) {
    pagination.current = pag.current;
  }
  if (pag.pageSize !== undefined) {
    pagination.pageSize = pag.pageSize;
  }
}

function resetSearch() {
  searchForm.datasource_type = '';
  searchForm.host = '';
  searchForm.port = '';
  searchForm.database_name = '';
  searchForm.table_name = '';
  pagination.current = 1;
  void fetchList();
}

const columns: TableColumnsType<SensitiveMetaRow> = [
  { title: $t('page.securitySensitiveInventory.columns.datasourceType'), dataIndex: 'datasource_type', key: 'datasource_type', width: 100, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.host'), dataIndex: 'host', key: 'host', width: 140, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.port'), dataIndex: 'port', key: 'port', width: 72 },
  { title: $t('page.securitySensitiveInventory.columns.databaseName'), dataIndex: 'database_name', key: 'database_name', width: 150, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.tableName'), dataIndex: 'table_name', key: 'table_name', width: 150, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.tableComment'), dataIndex: 'table_comment', key: 'table_comment', width: 140, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.columnName'), dataIndex: 'column_name', key: 'column_name', width: 150, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.columnComment'), dataIndex: 'column_comment', key: 'column_comment', width: 140, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.ruleType'), dataIndex: 'rule_type', key: 'rule_type', width: 100, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.ruleKey'), dataIndex: 'rule_key', key: 'rule_key', width: 110, ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.ruleName'), dataIndex: 'rule_name', key: 'rule_name', ellipsis: true },
  { title: $t('page.securitySensitiveInventory.columns.level'), dataIndex: 'level', key: 'level', width: 88 },
  { title: $t('page.securitySensitiveInventory.columns.status'), dataIndex: 'status', key: 'status', width: 100 },
  { title: $t('page.securitySensitiveInventory.columns.createdAt'), dataIndex: 'gmt_created', key: 'gmt_created', width: 180 },
  { title: $t('page.securitySensitiveInventory.columns.updatedAt'), dataIndex: 'gmt_updated', key: 'gmt_updated', width: 180 },
];

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.securitySensitiveInventory.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.securitySensitiveInventory.form.datasourceType')" class="query-item">
            <Input
              v-model:value="searchForm.datasource_type"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveInventory.placeholder.datasourceType')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.securitySensitiveInventory.form.host')" class="query-item">
            <Input
              v-model:value="searchForm.host"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveInventory.placeholder.host')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.securitySensitiveInventory.form.port')" class="query-item">
            <Input
              v-model:value="searchForm.port"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveInventory.placeholder.port')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.securitySensitiveInventory.form.databaseName')" class="query-item">
            <Input
              v-model:value="searchForm.database_name"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveInventory.placeholder.fuzzy')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.securitySensitiveInventory.form.tableName')" class="query-item">
            <Input
              v-model:value="searchForm.table_name"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveInventory.placeholder.fuzzy')"
              @press-enter="handleSearch"
            />
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="resetSearch">{{ $t('page.common.reset') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="pagedRows"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: SensitiveMetaRow, index: number) => record.id ?? `s-${pagination.current}-${index}`"
        :scroll="{ x: 'max-content' }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'level'">
            <Tag :color="levelTag(record.level).color">{{ levelTag(record.level).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">{{ statusTag(record.status).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'rule_name'">
            <Tooltip :title="record.rule_name">
              <span class="inline-block max-w-[280px] truncate">{{ record.rule_name || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'table_comment'">
            <Tooltip :title="record.table_comment">
              <span class="inline-block max-w-[200px] truncate">{{ record.table_comment || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'column_comment'">
            <Tooltip :title="record.column_comment">
              <span class="inline-block max-w-[200px] truncate">{{ record.column_comment || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'gmt_created'">
            {{ formatTime(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatTime(record.gmt_updated) }}
          </template>
        </template>
      </Table>
    </Card>
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
