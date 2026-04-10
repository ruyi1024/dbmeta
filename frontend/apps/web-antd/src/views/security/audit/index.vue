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
  showTotal: (total: number) => `共 ${total} 条`,
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
    message.error((e as Error)?.message || '加载审计日志失败');
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
  if (s === 'success') return { color: 'success' as const, text: '执行成功' };
  if (s === 'intercept') return { color: 'error' as const, text: '风险拦截' };
  if (s === 'failed') return { color: 'warning' as const, text: '执行失败' };
  return { color: 'default' as const, text: s || '-' };
}

function exportCsv() {
  if (!dataSource.value.length) {
    message.warning('当前无数据可导出');
    return;
  }
  const headers = [
    '记录时间',
    '用户',
    '数据源',
    '数据库',
    '操作',
    'SQL类型',
    '执行内容',
    '执行状态',
    '执行结果',
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
  saveAs(blob, `数据查询审计-${dayjs().format('YYYYMMDD-HHmmss')}.csv`);
  message.success('已导出当前页数据');
}

const columns: TableColumnsType<QueryLogRow> = [
  {
    title: '记录时间',
    dataIndex: 'gmt_created',
    key: 'gmt_created',
    sorter: true,
    width: 170,
    customRender: ({ record }) => formatTime(record.gmt_created),
  },
  { title: '用户', dataIndex: 'username', key: 'username', width: 100 },
  { title: '数据源', dataIndex: 'datasource_type', key: 'datasource_type', width: 100 },
  {
    title: '数据库',
    dataIndex: 'database',
    key: 'database',
    ellipsis: true,
    width: 120,
  },
  { title: '操作', dataIndex: 'query_type', key: 'query_type', sorter: true, width: 140 },
  { title: 'SQL类型', dataIndex: 'sql_type', key: 'sql_type', sorter: true, width: 100 },
  {
    title: '执行内容',
    dataIndex: 'content',
    key: 'content',
    ellipsis: true,
    width: 220,
  },
  {
    title: '执行状态',
    dataIndex: 'status',
    key: 'status',
    sorter: true,
    width: 110,
  },
  {
    title: '执行结果',
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
    <Card title="数据查询审计" size="small">
      <Row :gutter="[12, 12]" align="middle">
        <Col :flex="'auto'">
          <Space wrap>
            <Input.Search
              v-model:value="keyword"
              allow-clear
              placeholder="支持按用户名搜索"
              style="width: 280px"
              @search="onSearch"
            />
            <Tooltip title="刷新列表">
              <Button @click="fetchList">
                刷新
              </Button>
            </Tooltip>
          </Space>
        </Col>
        <Col>
          <Button type="primary" ghost @click="exportCsv">导出当前页</Button>
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
