<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import { Button, Card, Form, Input, Space, Table, Tag, message } from 'ant-design-vue';
import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Page } from '@vben/common-ui';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'GradingLog' });

interface LogRow {
  id: number;
  assetId: number;
  gradeIdOld?: number;
  gradeNameOld?: string;
  gradeIdNew: number;
  gradeNameNew?: string;
  action: string;
  reason?: string;
  operator?: string;
  gmtCreated?: string;
}

const loading = ref(false);
const dataSource = ref<LogRow[]>([]);

const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 条`,
});

const queryForm = reactive({
  assetId: '',
});

function parsePage(response: unknown) {
  const body = (response as any)?.data ?? response;
  const inner = body?.data ?? body;
  const list = inner?.list ?? [];
  const total = inner?.total ?? 0;
  return {
    list: Array.isArray(list) ? list : [],
    total: Number(total) || 0,
  };
}

async function fetchLogs() {
  loading.value = true;
  try {
    const res = await baseRequestClient.get('/v1/grading/logs', {
      params: {
        current: pagination.current,
        pageSize: pagination.pageSize,
        assetId: queryForm.assetId.trim() || undefined,
      },
    });
    const { list, total } = parsePage(res);
    dataSource.value = list as LogRow[];
    pagination.total = total;
  } catch (e: unknown) {
    message.error((e as Error)?.message || '加载失败');
    dataSource.value = [];
  } finally {
    loading.value = false;
  }
}

function actionLabel(a: string) {
  if (a === 'create') return '新建';
  if (a === 'update') return '变更';
  if (a === 'delete') return '删除';
  return a;
}

function actionColor(a: string) {
  if (a === 'delete') return 'red';
  if (a === 'create') return 'green';
  return 'blue';
}

function handleTableChange(p: TablePaginationConfig) {
  pagination.current = p.current ?? 1;
  pagination.pageSize = p.pageSize ?? 10;
  void fetchLogs();
}

const columns: TableColumnsType<LogRow> = [
  { title: '时间', dataIndex: 'gmtCreated', key: 'gmtCreated', width: 175 },
  { title: '操作', key: 'action', width: 88 },
  { title: '资产ID', dataIndex: 'assetId', key: 'assetId', width: 100 },
  { title: '原分级', key: 'oldG', width: 120 },
  { title: '新分级', key: 'newG', width: 120 },
  { title: '操作人', dataIndex: 'operator', key: 'operator', width: 120 },
  { title: '原因', dataIndex: 'reason', key: 'reason', ellipsis: true },
];

onMounted(() => {
  void fetchLogs();
});
</script>

<template>
  <Page auto-content-height description="资产分级的新建、调整与删除记录。">
    <Card title="变更记录">
      <Form layout="inline" class="mb-3">
        <Form.Item label="资产ID">
          <Input v-model:value="queryForm.assetId" allow-clear placeholder="可选" style="width: 160px" />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" @click="() => { pagination.current = 1; fetchLogs(); }">查询</Button>
            <Button @click="fetchLogs">刷新</Button>
          </Space>
        </Form.Item>
      </Form>

      <Table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="pagination"
        bordered
        size="small"
        :scroll="{ x: 1000 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <Tag :color="actionColor(record.action)">{{ actionLabel(record.action) }}</Tag>
          </template>
          <template v-else-if="column.key === 'oldG'">
            {{ record.gradeNameOld || (record.gradeIdOld ? record.gradeIdOld : '-') }}
          </template>
          <template v-else-if="column.key === 'newG'">
            <span v-if="record.action === 'delete'">—</span>
            <span v-else>{{ record.gradeNameNew || record.gradeIdNew }}</span>
          </template>
        </template>
      </Table>
    </Card>
  </Page>
</template>
