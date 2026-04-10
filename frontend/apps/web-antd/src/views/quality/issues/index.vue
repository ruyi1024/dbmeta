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
  Tag,
  Tooltip,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

interface IssueListItem {
  columnName: string;
  databaseName?: string;
  handleRemark?: string;
  handler?: string;
  issueCount: number;
  issueDesc: string;
  issueLevel: string;
  issueType: string;
  key: number;
  lastCheckTime: string;
  status: number;
  tableName: string;
}

const loading = ref(false);
const dataSource = ref<IssueListItem[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  columnName: '',
  issueLevel: undefined as string | undefined,
  issueType: undefined as string | undefined,
  status: undefined as number | undefined,
  tableName: '',
});

const modalVisible = ref(false);
const saving = ref(false);
const currentRecord = ref<IssueListItem | null>(null);
const editForm = reactive({
  handleRemark: '',
  handler: '',
  status: 1,
});

const columns: TableColumnsType<IssueListItem> = [
  { title: '数据库名', dataIndex: 'databaseName', key: 'databaseName', width: 150 },
  { title: '表名', dataIndex: 'tableName', key: 'tableName', width: 150 },
  { title: '字段名', dataIndex: 'columnName', key: 'columnName', width: 150 },
  { title: '问题类型', dataIndex: 'issueType', key: 'issueType', width: 120 },
  { title: '严重程度', dataIndex: 'issueLevel', key: 'issueLevel', width: 100 },
  { title: '问题描述', dataIndex: 'issueDesc', key: 'issueDesc' },
  { title: '问题数量', dataIndex: 'issueCount', key: 'issueCount', width: 100, sorter: true },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '处理人', dataIndex: 'handler', key: 'handler', width: 100 },
  { title: '最后检查时间', dataIndex: 'lastCheckTime', key: 'lastCheckTime', width: 180 },
  { title: '操作', key: 'option', width: 100, fixed: 'right' },
];

function levelTag(level: string) {
  if (level === 'high') return { color: 'red', text: '高' };
  if (level === 'medium') return { color: 'orange', text: '中' };
  if (level === 'low') return { color: 'blue', text: '低' };
  return { color: 'default', text: level || '未知' };
}

function statusTag(status: number) {
  if (status === 1) return { color: 'red', text: '待处理' };
  if (status === 2) return { color: 'orange', text: '处理中' };
  if (status === 3) return { color: 'green', text: '已处理' };
  return { color: 'default', text: '已忽略' };
}

async function fetchIssues(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/dataquality/issues', {
      params: {
        ...queryForm,
        current: pagination.current,
        pageSize: pagination.pageSize,
        sorter: sorter ? JSON.stringify(sorter) : undefined,
      },
    });
    const payload = (response as any)?.data ?? response;
    const list = payload?.data?.list ?? payload?.list ?? [];
    const total = payload?.data?.total ?? payload?.total ?? 0;
    dataSource.value = Array.isArray(list) ? list : [];
    pagination.total = Number(total) || dataSource.value.length;
  } catch (error: any) {
    message.error(error?.message || '获取质量问题失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchIssues();
}

function handleReset() {
  queryForm.issueType = undefined;
  queryForm.issueLevel = undefined;
  queryForm.status = undefined;
  queryForm.tableName = '';
  queryForm.columnName = '';
  pagination.current = 1;
  fetchIssues();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  if (sorter?.field && sorter?.order) {
    fetchIssues({ [sorter.field]: sorter.order });
    return;
  }
  fetchIssues();
}

function openHandleModal(record: IssueListItem) {
  currentRecord.value = record;
  editForm.status = record.status;
  editForm.handler = record.handler || '';
  editForm.handleRemark = record.handleRemark || '';
  modalVisible.value = true;
}

async function submitHandle() {
  if (!currentRecord.value) return;
  saving.value = true;
  try {
    const response = await baseRequestClient.put('/v1/dataquality/issues/status', {
      handleRemark: editForm.handleRemark,
      handler: editForm.handler,
      id: currentRecord.value.key,
      status: editForm.status,
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.code && payload.code !== 200) {
      message.error(payload?.msg || '更新失败');
      return;
    }
    message.success('更新成功');
    modalVisible.value = false;
    fetchIssues();
  } catch (error: any) {
    message.error(error?.message || '更新失败');
  } finally {
    saving.value = false;
  }
}

onMounted(fetchIssues);
</script>

<template>
  <div class="p-5">
    <Card title="质量问题列表">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="表名" class="query-item">
            <Input v-model:value="queryForm.tableName" allow-clear class="query-control" placeholder="请输入表名" />
          </Form.Item>
          <Form.Item label="字段名" class="query-item">
            <Input v-model:value="queryForm.columnName" allow-clear class="query-control" placeholder="请输入字段名" />
          </Form.Item>
          <Form.Item label="问题类型" class="query-item">
            <Select v-model:value="queryForm.issueType" allow-clear class="query-control">
              <Select.Option value="完整性">完整性</Select.Option>
              <Select.Option value="准确性">准确性</Select.Option>
              <Select.Option value="唯一性">唯一性</Select.Option>
              <Select.Option value="一致性">一致性</Select.Option>
              <Select.Option value="及时性">及时性</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="严重程度" class="query-item">
            <Select v-model:value="queryForm.issueLevel" allow-clear class="query-control">
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="low">低</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="状态" class="query-item">
            <Select v-model:value="queryForm.status" allow-clear class="query-control">
              <Select.Option :value="1">待处理</Select.Option>
              <Select.Option :value="2">处理中</Select.Option>
              <Select.Option :value="3">已处理</Select.Option>
              <Select.Option :value="0">已忽略</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Tag color="blue">AI智能诊断</Tag>
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
        :row-key="(record: IssueListItem) => record.key"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'issueLevel'">
            <Tag :color="levelTag(record.issueLevel).color">{{ levelTag(record.issueLevel).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">{{ statusTag(record.status).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'issueDesc'">
            <Tooltip :title="record.issueDesc">
              <span class="inline-block max-w-[280px] truncate">{{ record.issueDesc }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'option'">
            <Button type="link" size="small" @click="openHandleModal(record)">处理</Button>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalVisible"
      title="处理质量问题"
      :confirm-loading="saving"
      @ok="submitHandle"
    >
      <Form layout="vertical">
        <Form.Item label="处理状态" required>
          <Select v-model:value="editForm.status">
            <Select.Option :value="1">待处理</Select.Option>
            <Select.Option :value="2">处理中</Select.Option>
            <Select.Option :value="3">已处理</Select.Option>
            <Select.Option :value="0">已忽略</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="处理人">
          <Input v-model:value="editForm.handler" placeholder="请输入处理人" />
        </Form.Item>
        <Form.Item label="处理备注">
          <Input.TextArea v-model:value="editForm.handleRemark" :rows="4" placeholder="请输入处理备注" />
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
