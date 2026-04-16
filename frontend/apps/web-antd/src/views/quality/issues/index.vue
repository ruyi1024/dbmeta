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
import { $t } from '#/locales';

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
  { title: $t('page.qualityIssues.columns.databaseName'), dataIndex: 'databaseName', key: 'databaseName', width: 150 },
  { title: $t('page.qualityIssues.columns.tableName'), dataIndex: 'tableName', key: 'tableName', width: 150 },
  { title: $t('page.qualityIssues.columns.columnName'), dataIndex: 'columnName', key: 'columnName', width: 150 },
  { title: $t('page.qualityIssues.columns.issueType'), dataIndex: 'issueType', key: 'issueType', width: 120 },
  { title: $t('page.qualityIssues.columns.issueLevel'), dataIndex: 'issueLevel', key: 'issueLevel', width: 100 },
  { title: $t('page.qualityIssues.columns.issueDesc'), dataIndex: 'issueDesc', key: 'issueDesc' },
  { title: $t('page.qualityIssues.columns.issueCount'), dataIndex: 'issueCount', key: 'issueCount', width: 100, sorter: true },
  { title: $t('page.qualityIssues.columns.status'), dataIndex: 'status', key: 'status', width: 100 },
  { title: $t('page.qualityIssues.columns.handler'), dataIndex: 'handler', key: 'handler', width: 100 },
  { title: $t('page.qualityIssues.columns.lastCheckTime'), dataIndex: 'lastCheckTime', key: 'lastCheckTime', width: 180 },
  { title: $t('page.qualityIssues.columns.option'), key: 'option', width: 100, fixed: 'right' },
];

function levelTag(level: string) {
  if (level === 'high') return { color: 'red', text: $t('page.qualityIssues.level.high') };
  if (level === 'medium') return { color: 'orange', text: $t('page.qualityIssues.level.medium') };
  if (level === 'low') return { color: 'blue', text: $t('page.qualityIssues.level.low') };
  return { color: 'default', text: level || $t('page.qualityIssues.level.unknown') };
}

function statusTag(status: number) {
  if (status === 1) return { color: 'red', text: $t('page.qualityIssues.status.pending') };
  if (status === 2) return { color: 'orange', text: $t('page.qualityIssues.status.processing') };
  if (status === 3) return { color: 'green', text: $t('page.qualityIssues.status.done') };
  return { color: 'default', text: $t('page.qualityIssues.status.ignored') };
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
    message.error(error?.message || $t('page.qualityIssues.message.fetchFailed'));
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
      message.error(payload?.msg || $t('page.qualityIssues.message.updateFailed'));
      return;
    }
    message.success($t('page.qualityIssues.message.updateSuccess'));
    modalVisible.value = false;
    fetchIssues();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityIssues.message.updateFailed'));
  } finally {
    saving.value = false;
  }
}

onMounted(fetchIssues);
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.qualityIssues.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.qualityIssues.form.tableName')" class="query-item">
            <Input v-model:value="queryForm.tableName" allow-clear class="query-control" :placeholder="$t('page.qualityIssues.placeholder.tableName')" />
          </Form.Item>
          <Form.Item :label="$t('page.qualityIssues.form.columnName')" class="query-item">
            <Input v-model:value="queryForm.columnName" allow-clear class="query-control" :placeholder="$t('page.qualityIssues.placeholder.columnName')" />
          </Form.Item>
          <Form.Item :label="$t('page.qualityIssues.form.issueType')" class="query-item">
            <Select v-model:value="queryForm.issueType" allow-clear class="query-control">
              <Select.Option value="完整性">{{ $t('page.qualityIssues.issueType.completeness') }}</Select.Option>
              <Select.Option value="准确性">{{ $t('page.qualityIssues.issueType.accuracy') }}</Select.Option>
              <Select.Option value="唯一性">{{ $t('page.qualityIssues.issueType.uniqueness') }}</Select.Option>
              <Select.Option value="一致性">{{ $t('page.qualityIssues.issueType.consistency') }}</Select.Option>
              <Select.Option value="及时性">{{ $t('page.qualityIssues.issueType.timeliness') }}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item :label="$t('page.qualityIssues.form.issueLevel')" class="query-item">
            <Select v-model:value="queryForm.issueLevel" allow-clear class="query-control">
              <Select.Option value="high">{{ $t('page.qualityIssues.level.high') }}</Select.Option>
              <Select.Option value="medium">{{ $t('page.qualityIssues.level.medium') }}</Select.Option>
              <Select.Option value="low">{{ $t('page.qualityIssues.level.low') }}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item :label="$t('page.qualityIssues.form.status')" class="query-item">
            <Select v-model:value="queryForm.status" allow-clear class="query-control">
              <Select.Option :value="1">{{ $t('page.qualityIssues.status.pending') }}</Select.Option>
              <Select.Option :value="2">{{ $t('page.qualityIssues.status.processing') }}</Select.Option>
              <Select.Option :value="3">{{ $t('page.qualityIssues.status.done') }}</Select.Option>
              <Select.Option :value="0">{{ $t('page.qualityIssues.status.ignored') }}</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Tag color="blue">{{ $t('page.qualityIssues.aiDiagnosis') }}</Tag>
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
            <Button type="link" size="small" @click="openHandleModal(record)">{{ $t('page.qualityIssues.action.handle') }}</Button>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalVisible"
      :title="$t('page.qualityIssues.modal.title')"
      :confirm-loading="saving"
      @ok="submitHandle"
    >
      <Form layout="vertical">
        <Form.Item :label="$t('page.qualityIssues.modal.handleStatus')" required>
          <Select v-model:value="editForm.status">
            <Select.Option :value="1">{{ $t('page.qualityIssues.status.pending') }}</Select.Option>
            <Select.Option :value="2">{{ $t('page.qualityIssues.status.processing') }}</Select.Option>
            <Select.Option :value="3">{{ $t('page.qualityIssues.status.done') }}</Select.Option>
            <Select.Option :value="0">{{ $t('page.qualityIssues.status.ignored') }}</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item :label="$t('page.qualityIssues.modal.handler')">
          <Input v-model:value="editForm.handler" :placeholder="$t('page.qualityIssues.placeholder.handler')" />
        </Form.Item>
        <Form.Item :label="$t('page.qualityIssues.modal.handleRemark')">
          <Input.TextArea v-model:value="editForm.handleRemark" :rows="4" :placeholder="$t('page.qualityIssues.placeholder.handleRemark')" />
        </Form.Item>
      </Form>
    </Modal>
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
