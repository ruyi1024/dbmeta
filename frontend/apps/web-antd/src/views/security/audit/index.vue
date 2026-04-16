<script lang="ts" setup>
import type { TableColumnsType } from 'ant-design-vue';
import type { SorterResult, TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Col,
  Input,
  message,
  Row,
  Space,
  Table,
  Tag,
  Tooltip,
} from 'ant-design-vue';
import dayjs from 'dayjs';
import { saveAs } from 'file-saver';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

defineOptions({ name: 'DataSecurityQueryAudit' });

interface QueryLogRow {
  content?: string;
  database?: string;
  datasource_type?: string;
  gmt_created?: string;
  id?: number;
  query_type?: string;
  result?: string;
  sql_type?: string;
  status?: string;
  username?: string;
}

const loading = ref(false);
const dataSource = ref<QueryLogRow[]>([]);
const keyword = ref('');
const sortField = ref('id');
const sortOrder = ref<'ascend' | 'descend'>('descend');

const pagination = reactive({
  current: 1,
  pageSize: 10,
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => `${$t('page.common.total')} ${total} ${$t('page.common.records')}`,
  total: 0,
});

function resolveAuditList(response: unknown): { list: QueryLogRow[]; total: number } {
  if (!response || typeof response !== 'object') {
    return { list: [], total: 0 };
  }
  const r = response as Record<string, unknown>;
  const httpBody =
    'status' in r && typeof r.status === 'number' && r.data !== undefined ? r.data : r;
  const hb = httpBody as Record<string, unknown>;
  if (hb?.success === false) {
    return { list: [], total: 0 };
  }
  const list = Array.isArray(hb?.data) ? (hb.data as QueryLogRow[]) : [];
  const total = Number(hb?.total ?? 0);
  return { list, total };
}

async function fetchList() {
  loading.value = true;
  try {
    const offset = pagination.pageSize * (pagination.current >= 1 ? pagination.current - 1 : 0);
    const response = await baseRequestClient.get('/v1/audit/query_log', {
      params: {
        keyword: keyword.value || undefined,
        limit: pagination.pageSize,
        offset,
        sorterField: sortField.value || 'id',
        sorterOrder: sortOrder.value === 'ascend' ? 'ascend' : 'descend',
      },
    });
    const { list, total } = resolveAuditList(response);
    dataSource.value = list;
    pagination.total = total;
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.securityAudit.message.loadFailed'));
    dataSource.value = [];
    pagination.total = 0;
  } finally {
    loading.value = false;
  }
}

function onSearch(val: string) {
  keyword.value = val;
  pagination.current = 1;
  fetchList();
}

function handleTableChange(
  pag: TablePaginationConfig,
  _f: Record<string, unknown>,
  sorter: SorterResult<QueryLogRow> | SorterResult<QueryLogRow>[],
) {
  if (pag.current !== undefined) {
    pagination.current = pag.current;
  }
  if (pag.pageSize !== undefined) {
    pagination.pageSize = pag.pageSize;
  }
  const one = Array.isArray(sorter) ? sorter[0] : sorter;
  const col = one?.column;
  const rawField = col?.key ?? one?.field;
  const field = Array.isArray(rawField) ? rawField[0] : rawField;
  const order = one?.order;
  if (field !== undefined && field !== null && (order === 'ascend' || order === 'descend')) {
    sortField.value = String(field);
    sortOrder.value = order;
  } else {
    sortField.value = 'id';
    sortOrder.value = 'descend';
  }
  fetchList();
}

function formatTime(v?: string) {
  if (!v) return '-';
  return dayjs(v).isValid() ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : v;
}

function statusTag(status?: string) {
  const s = String(status || '');
  if (s === 'success') return { color: 'success' as const, text: $t('page.securityAudit.status.success') };
  if (s === 'intercept') return { color: 'error' as const, text: $t('page.securityAudit.status.intercept') };
  if (s === 'failed') return { color: 'warning' as const, text: $t('page.securityAudit.status.failed') };
  return { color: 'default' as const, text: s || '-' };
}

function exportCsv() {
  if (!dataSource.value.length) {
    message.warning($t('page.securityAudit.message.noDataToExport'));
    return;
  }
  const headers = [
    $t('page.securityAudit.columns.recordTime'),
    $t('page.securityAudit.columns.username'),
    $t('page.securityAudit.columns.datasourceType'),
    $t('page.securityAudit.columns.database'),
    $t('page.securityAudit.columns.queryType'),
    $t('page.securityAudit.columns.sqlType'),
    $t('page.securityAudit.columns.content'),
    $t('page.securityAudit.columns.status'),
    $t('page.securityAudit.columns.result'),
  ];
  const rows = dataSource.value.map((r) =>
    [
      formatTime(r.gmt_created),
      r.username ?? '',
      r.datasource_type ?? '',
      r.database ?? '',
      r.query_type ?? '',
      r.sql_type ?? '',
      (r.content ?? '').replaceAll('"', '""'),
      r.status ?? '',
      (r.result ?? '').replaceAll('"', '""'),
    ].map((cell) => `"${String(cell)}"`),
  );
  const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\r\n');
  const blob = new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8;' });
  saveAs(blob, `${$t('page.securityAudit.exportFilePrefix')}-${dayjs().format('YYYYMMDD-HHmmss')}.csv`);
  message.success($t('page.securityAudit.message.exported'));
}

const columns: TableColumnsType<QueryLogRow> = [
  {
    title: $t('page.securityAudit.columns.recordTime'),
    dataIndex: 'gmt_created',
    key: 'gmt_created',
    sorter: true,
    width: 170,
    customRender: ({ record }) => formatTime(record.gmt_created),
  },
  { title: $t('page.securityAudit.columns.username'), dataIndex: 'username', key: 'username', width: 100 },
  { title: $t('page.securityAudit.columns.datasourceType'), dataIndex: 'datasource_type', key: 'datasource_type', width: 100 },
  {
    title: $t('page.securityAudit.columns.database'),
    dataIndex: 'database',
    key: 'database',
    ellipsis: true,
    width: 120,
  },
  { title: $t('page.securityAudit.columns.queryType'), dataIndex: 'query_type', key: 'query_type', sorter: true, width: 140 },
  { title: $t('page.securityAudit.columns.sqlType'), dataIndex: 'sql_type', key: 'sql_type', sorter: true, width: 100 },
  {
    title: $t('page.securityAudit.columns.content'),
    dataIndex: 'content',
    key: 'content',
    ellipsis: true,
    width: 220,
  },
  {
    title: $t('page.securityAudit.columns.status'),
    dataIndex: 'status',
    key: 'status',
    sorter: true,
    width: 110,
  },
  {
    title: $t('page.securityAudit.columns.result'),
    dataIndex: 'result',
    key: 'result',
    ellipsis: true,
    width: 260,
  },
];

onMounted(fetchList);
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.securityAudit.title')" size="small">
      <Row :gutter="[12, 12]" align="middle">
        <Col :flex="'auto'">
          <Space wrap>
            <Input.Search
              v-model:value="keyword"
              allow-clear
              :placeholder="$t('page.securityAudit.searchPlaceholder')"
              style="width: 280px"
              @search="onSearch"
            />
            <Tooltip :title="$t('page.securityAudit.refreshTip')">
              <Button @click="fetchList">
                {{ $t('page.common.refresh') }}
              </Button>
            </Tooltip>
          </Space>
        </Col>
        <Col>
          <Button type="primary" ghost @click="exportCsv">{{ $t('page.securityAudit.exportCurrentPage') }}</Button>
        </Col>
      </Row>

      <Table
        class="mt-3"
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record, index) => record.id ?? `audit-${pagination.current}-${index}`"
        :scroll="{ x: 1400 }"
        size="small"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record, text }">
          <template v-if="column.key === 'content'">
            <Tooltip v-if="text" :title="String(text)" placement="topLeft">
              <span>{{ text }}</span>
            </Tooltip>
            <span v-else>-</span>
          </template>
          <template v-else-if="column.key === 'result'">
            <Tooltip v-if="text" :title="String(text)" placement="topLeft">
              <span>{{ text }}</span>
            </Tooltip>
            <span v-else>-</span>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">
              {{ statusTag(record.status).text }}
            </Tag>
          </template>
        </template>
      </Table>
    </Card>
  </div>
</template>
