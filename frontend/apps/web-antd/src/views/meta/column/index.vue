<script lang="ts" setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';

import {
  Badge,
  Button,
  Card,
  Checkbox,
  Form,
  Popover,
  Input,
  Progress,
  Select,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'MetaColumnPage' });

interface ColumnItem {
  ai_comment?: string;
  ai_fixed?: number;
  column_comment?: string;
  column_comment_accuracy?: number | string;
  column_name: string;
  data_type?: string;
  database_name: string;
  datasource_type: string;
  default_value?: string;
  gmt_created?: string;
  gmt_updated?: string;
  host: string;
  id: number;
  is_nullable?: string;
  port: number | string;
  table_name: string;
}

const loading = ref(false);
const dataSource = ref<ColumnItem[]>([]);
const selectedRowKeys = ref<number[]>([]);
/** 双击编辑「AI注释生成」 */
const editingAiCommentId = ref<number | null>(null);
const editingAiCommentDraft = ref('');
const aiCommentInputRef = ref<{ focus?: () => void } | null>(null);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  column_name: '',
  database_name: '',
  datasource_type: '',
  host: '',
  port: '',
  table_name: '',
  /** 是否有字段注释：'' 全部 | '1' 有 | '0' 无 */
  has_column_comment: '' as '' | '0' | '1',
  /** 是否有 AI 注释：'' 全部 | '1' 有 | '0' 无 */
  has_ai_comment: '' as '' | '0' | '1',
  /** AI 注释应用状态：'' 全部 | '0' | '1' | '2' | '3' */
  ai_fixed: '' as '' | '0' | '1' | '2' | '3',
});

const yesNoFilterOptions = [
  { label: '全部', value: '' },
  { label: '有', value: '1' },
  { label: '无', value: '0' },
];

const aiFixedFilterOptions = [
  { label: '全部', value: '' },
  { label: '待审核', value: '0' },
  { label: '不应用', value: '1' },
  { label: '待应用', value: '2' },
  { label: '已应用', value: '3' },
];

const VISIBILITY_STORAGE_KEY = 'meta_column_visible_columns';
const TIME_COLUMN_KEYS = new Set(['gmt_created', 'gmt_updated']);

const allColumns: TableColumnsType<ColumnItem> = [
  { title: '字段名', dataIndex: 'column_name', key: 'column_name', sorter: true },
  { title: '数据类型', dataIndex: 'data_type', key: 'data_type' },
  { title: '允许为空', dataIndex: 'is_nullable', key: 'is_nullable' },
  { title: '默认值', dataIndex: 'default_value', key: 'default_value' },
  { title: '字段备注', dataIndex: 'column_comment', key: 'column_comment' },
  { title: '注释准确度', dataIndex: 'column_comment_accuracy', key: 'column_comment_accuracy' },
  { title: 'AI注释生成', dataIndex: 'ai_comment', key: 'ai_comment' },
  { title: 'AI注释应用', dataIndex: 'ai_fixed', key: 'ai_fixed' },
  { title: '所属表', dataIndex: 'table_name', key: 'table_name', sorter: true },
  { title: '所属库', dataIndex: 'database_name', key: 'database_name', sorter: true },
  { title: '数据库类型', dataIndex: 'datasource_type', key: 'datasource_type', sorter: true },
  { title: '所属主机', dataIndex: 'host', key: 'host' },
  { title: '所属端口', dataIndex: 'port', key: 'port' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', sorter: true },
];

const columnPickerOptions = computed(() =>
  allColumns.map((c) => ({
    label: String(c.title ?? c.key),
    value: String(c.key),
  })),
);

function defaultVisibleColumnKeys(): string[] {
  return allColumns.map((c) => String(c.key)).filter((k) => !TIME_COLUMN_KEYS.has(k));
}

function loadVisibleColumnKeys(): string[] {
  try {
    const raw = localStorage.getItem(VISIBILITY_STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as string[];
      if (Array.isArray(parsed) && parsed.length > 0) {
        const valid = parsed.filter((k) => allColumns.some((c) => String(c.key) === k));
        if (valid.length > 0) {
          if (!valid.includes('column_comment_accuracy')) {
            valid.push('column_comment_accuracy');
          }
          return valid;
        }
      }
    }
  } catch {
    /* ignore */
  }
  return defaultVisibleColumnKeys();
}

const visibleColumnKeys = ref<string[]>(loadVisibleColumnKeys());

const displayColumns = computed(() =>
  allColumns.filter((c) => visibleColumnKeys.value.includes(String(c.key))),
);

watch(
  visibleColumnKeys,
  (v) => {
    if (v.length === 0) {
      visibleColumnKeys.value = defaultVisibleColumnKeys();
      message.warning('至少保留一列');
      return;
    }
    try {
      localStorage.setItem(VISIBILITY_STORAGE_KEY, JSON.stringify(v));
    } catch {
      /* ignore */
    }
  },
  { deep: true },
);

async function fetchColumns(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/column/list', {
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
    message.error(error?.message || '数据字段查询失败');
  } finally {
    loading.value = false;
  }
}

async function handleBatchUpdate(aiFixed: number) {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请先选择要操作的字段');
    return;
  }
  try {
    const response = await baseRequestClient.put('/v1/meta/column/batch-update-ai-fixed', {
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
    fetchColumns();
  } catch (error: any) {
    message.error(error?.message || '批量操作失败');
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchColumns();
}

function handleReset() {
  queryForm.column_name = '';
  queryForm.datasource_type = '';
  queryForm.host = '';
  queryForm.port = '';
  queryForm.database_name = '';
  queryForm.table_name = '';
  queryForm.has_column_comment = '';
  queryForm.has_ai_comment = '';
  queryForm.ai_fixed = '';
  pagination.current = 1;
  fetchColumns();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  if (sorter?.field && sorter?.order) {
    fetchColumns({ [sorter.field]: sorter.order });
    return;
  }
  fetchColumns();
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

function normalizeAccuracyValue(value?: number | string) {
  if (value === null || value === undefined || value === '') return null;
  const n = typeof value === 'number' ? value : Number(value);
  if (!Number.isFinite(n)) return null;
  if (n < 0) return 0;
  if (n > 1) return 1;
  return n;
}

function getAccuracyMeta(value?: number | string) {
  const score = normalizeAccuracyValue(value);
  if (score === null) {
    return { color: '#d9d9d9', label: '-', percent: 0, scoreText: '-' };
  }
  if (score > 0.95) {
    return { color: '#52c41a', label: '优', percent: Math.round(score * 100), scoreText: score.toFixed(1) };
  }
  if (score >= 0.8) {
    return { color: '#1677ff', label: '良', percent: Math.round(score * 100), scoreText: score.toFixed(1) };
  }
  if (score >= 0.6) {
    return { color: '#faad14', label: '中', percent: Math.round(score * 100), scoreText: score.toFixed(1) };
  }
  return { color: '#ff4d4f', label: '差', percent: Math.round(score * 100), scoreText: score.toFixed(1) };
}

function startEditAiComment(record: ColumnItem) {
  editingAiCommentId.value = record.id;
  editingAiCommentDraft.value = record.ai_comment ?? '';
  nextTick(() => {
    aiCommentInputRef.value?.focus?.();
  });
}

function cancelEditAiComment() {
  editingAiCommentId.value = null;
  editingAiCommentDraft.value = '';
}

async function commitAiCommentEdit(record: ColumnItem) {
  if (editingAiCommentId.value !== record.id) return;
  const next = editingAiCommentDraft.value.trim();
  const prev = (record.ai_comment ?? '').trim();
  if (next === prev) {
    cancelEditAiComment();
    return;
  }
  try {
    const response = await baseRequestClient.put('/v1/meta/column/update-ai-comment', {
      id: record.id,
      ai_comment: next,
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(payload?.msg || '保存失败');
      return;
    }
    record.ai_comment = next;
    message.success(payload?.msg || '已保存');
    cancelEditAiComment();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  }
}

onMounted(fetchColumns);
</script>

<template>
  <div class="p-5">
    <Card title="数据字段列表">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="字段名" class="query-item">
            <Input
              v-model:value="queryForm.column_name"
              placeholder="请输入字段名"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属表" class="query-item">
            <Input
              v-model:value="queryForm.table_name"
              placeholder="请输入所属表"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属库" class="query-item">
            <Input
              v-model:value="queryForm.database_name"
              placeholder="请输入所属库"
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
          <Form.Item label="有字段注释" class="query-item">
            <Select
              v-model:value="queryForm.has_column_comment"
              :options="yesNoFilterOptions"
              allow-clear
              placeholder="全部"
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="有AI注释" class="query-item">
            <Select
              v-model:value="queryForm.has_ai_comment"
              :options="yesNoFilterOptions"
              allow-clear
              placeholder="全部"
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="AI注释状态" class="query-item">
            <Select
              v-model:value="queryForm.ai_fixed"
              :options="aiFixedFilterOptions"
              allow-clear
              placeholder="全部"
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

      <div class="query-toolbar-divider"></div>

      <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
        <Space>
          <Button type="primary" :disabled="selectedRowKeys.length === 0" @click="handleBatchUpdate(2)">
            应用AI注释 ({{ selectedRowKeys.length }})
          </Button>
          <Button danger :disabled="selectedRowKeys.length === 0" @click="handleBatchUpdate(1)">
            不应用AI注释 ({{ selectedRowKeys.length }})
          </Button>
        </Space>
        <Popover trigger="click" placement="bottomRight">
          <template #content>
            <div class="column-picker">
              <div class="mb-2 text-xs text-gray-500">勾选要显示的列（创建/修改时间默认隐藏）</div>
              <Checkbox.Group v-model:value="visibleColumnKeys" class="column-picker-group">
                <div v-for="opt in columnPickerOptions" :key="opt.value" class="column-picker-item">
                  <Checkbox :value="opt.value">{{ opt.label }}</Checkbox>
                </div>
              </Checkbox.Group>
            </div>
          </template>
          <Button>列设置</Button>
        </Popover>
      </div>

      <Table
        :columns="displayColumns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-selection="{
          selectedRowKeys,
          onChange: handleRowSelectionChange,
          getCheckboxProps: (record: ColumnItem) => ({
            disabled: !record.ai_comment,
          }),
        }"
        :row-key="(record: ColumnItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'ai_comment'">
            <Input
              v-if="editingAiCommentId === record.id"
              ref="aiCommentInputRef"
              v-model:value="editingAiCommentDraft"
              size="small"
              :maxlength="100"
              show-count
              placeholder="请输入 AI 注释"
              class="w-full min-w-[200px]"
              @blur="commitAiCommentEdit(record as ColumnItem)"
              @keydown.esc.prevent="cancelEditAiComment"
            />
            <span
              v-else
              class="ai-comment-cell"
              title="双击修改"
              @dblclick="startEditAiComment(record as ColumnItem)"
            >
              <span v-if="record.ai_comment">{{ record.ai_comment }}</span>
              <span v-else style="color: #999">暂无AI注释</span>
            </span>
          </template>
          <template v-else-if="column.key === 'ai_fixed'">
            <Badge :status="aiFixedStatus(record.ai_fixed).status" :text="aiFixedStatus(record.ai_fixed).text" />
          </template>
          <template v-else-if="column.key === 'column_comment_accuracy'">
            <div class="accuracy-cell">
              <Progress
                :percent="getAccuracyMeta(record.column_comment_accuracy).percent"
                :stroke-color="getAccuracyMeta(record.column_comment_accuracy).color"
                :show-info="false"
                size="small"
                class="accuracy-progress"
              />
              <span
                class="accuracy-text"
                :style="{ color: getAccuracyMeta(record.column_comment_accuracy).color }"
              >
                {{ getAccuracyMeta(record.column_comment_accuracy).label }}
                <span class="accuracy-score">{{ getAccuracyMeta(record.column_comment_accuracy).scoreText }}</span>
              </span>
            </div>
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

.query-toolbar-divider {
  border-top: 1px solid #f0f0f0;
  margin-bottom: 12px;
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

.column-picker {
  background: hsl(var(--background, 0 0% 100%));
  border-radius: 8px;
  max-height: 360px;
  max-width: 280px;
  overflow: auto;
  padding: 12px;
}

.column-picker-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.column-picker-item {
  line-height: 1.5;
}

.ai-comment-cell {
  cursor: text;
  display: inline-block;
  max-width: 100%;
  word-break: break-word;
}

.accuracy-cell {
  align-items: center;
  display: flex;
  gap: 4px;
  min-width: 140px;
}

.accuracy-progress {
  flex: 0 0 72px;
  margin-bottom: 0;
}

.accuracy-text {
  font-size: 12px;
  font-weight: 600;
  min-width: 3rem;
  text-align: right;
}

.accuracy-score {
  color: #666;
  font-size: 11px;
  font-weight: 400;
  margin-left: 4px;
}
</style>
