<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  message,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

defineOptions({ name: 'DataSecuritySensitiveRule' });

interface SensitiveRuleRow {
  id?: number;
  rule_type?: string;
  rule_key?: string;
  rule_name?: string;
  rule_express?: string;
  rule_pct?: number;
  level?: number;
  status?: number;
  enable?: number;
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
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString(undefined, { hour12: false });
}

function levelTag(level?: number) {
  if (level === 0) return { color: 'orange', text: $t('page.securitySensitiveRule.level.low') };
  if (level === 1) return { color: 'red', text: $t('page.securitySensitiveRule.level.high') };
  return { color: 'default', text: level !== undefined ? String(level) : '-' };
}

function statusTag(status?: number) {
  if (status === -1) return { color: 'orange', text: $t('page.securitySensitiveRule.status.suspected') };
  if (status === 0) return { color: 'default', text: $t('page.securitySensitiveRule.status.nonSensitive') };
  if (status === 1) return { color: 'green', text: $t('page.securitySensitiveRule.status.confirmed') };
  return { color: 'default', text: status !== undefined ? String(status) : '-' };
}

const loading = ref(false);
const allRows = ref<SensitiveRuleRow[]>([]);

const searchForm = reactive({
  rule_key: '',
  rule_name: '',
  enable: undefined as number | undefined,
});

const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number) => `${$t('page.common.total')} ${total} ${$t('page.common.records')}`,
  pageSizeOptions: ['10', '15', '30', '50'],
});

const pagedRows = computed(() => {
  const current = pagination.current ?? 1;
  const pageSize = pagination.pageSize ?? 10;
  const start = (current - 1) * pageSize;
  return allRows.value.slice(start, start + pageSize);
});

const modalOpen = ref(false);
const modalMode = ref<'create' | 'edit'>('create');
const saving = ref(false);

const formModel = reactive<SensitiveRuleRow>({
  id: undefined,
  rule_type: 'data',
  rule_key: '',
  rule_name: '',
  rule_express: '',
  rule_pct: 0,
  level: 0,
  status: 1,
  enable: 1,
});

function resetFormModel() {
  formModel.id = undefined;
  formModel.rule_type = 'data';
  formModel.rule_key = '';
  formModel.rule_name = '';
  formModel.rule_express = '';
  formModel.rule_pct = 0;
  formModel.level = 0;
  formModel.status = 1;
  formModel.enable = 1;
}

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, string> = {};
    if (searchForm.rule_key.trim()) {
      params.rule_key = searchForm.rule_key.trim();
    }
    if (searchForm.rule_name.trim()) {
      params.rule_name = searchForm.rule_name.trim();
    }
    if (searchForm.enable !== undefined && searchForm.enable !== null) {
      params.enable = String(searchForm.enable);
    }
    const response = await baseRequestClient.get('/v1/sensitive/rule', { params });
    const body = extractApiBody(response);
    const raw = body?.data;
    const list = Array.isArray(raw) ? (raw as SensitiveRuleRow[]) : [];
    allRows.value = list;
    pagination.total = list.length;
    pagination.current = 1;
  } catch (e: unknown) {
    allRows.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || $t('page.securitySensitiveRule.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchList();
}

function resetSearch() {
  searchForm.rule_key = '';
  searchForm.rule_name = '';
  searchForm.enable = undefined;
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

function openCreate() {
  modalMode.value = 'create';
  resetFormModel();
  modalOpen.value = true;
}

function openEdit(record: SensitiveRuleRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.rule_type = record.rule_type ?? 'data';
  formModel.rule_key = record.rule_key ?? '';
  formModel.rule_name = record.rule_name ?? '';
  formModel.rule_express = record.rule_express ?? '';
  formModel.rule_pct = record.rule_pct ?? 0;
  formModel.level = record.level ?? 0;
  formModel.status = record.status ?? 0;
  formModel.enable = record.enable ?? 1;
  modalOpen.value = true;
}

async function submitModal() {
  if (!formModel.rule_key?.trim()) {
    message.warning($t('page.securitySensitiveRule.message.fillRuleKey'));
    return Promise.reject();
  }
  if (!formModel.rule_name?.trim()) {
    message.warning($t('page.securitySensitiveRule.message.fillRuleName'));
    return Promise.reject();
  }
  if (!formModel.rule_express?.trim()) {
    message.warning($t('page.securitySensitiveRule.message.fillRuleExpress'));
    return Promise.reject();
  }
  saving.value = true;
  try {
    const payload = {
      rule_type: formModel.rule_type,
      rule_key: formModel.rule_key.trim(),
      rule_name: formModel.rule_name.trim(),
      rule_express: formModel.rule_express.trim(),
      rule_pct: Number(formModel.rule_pct) || 0,
      level: Number(formModel.level) ?? 0,
      status: Number(formModel.status) ?? 0,
      enable: Number(formModel.enable) ?? 1,
    };
    if (modalMode.value === 'create') {
      const response = await baseRequestClient.post('/v1/sensitive/rule', payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.securitySensitiveRule.message.createFailed')));
        throw new Error('biz');
      }
      message.success($t('page.securitySensitiveRule.message.createSuccess'));
    } else {
      const response = await baseRequestClient.put('/v1/sensitive/rule', {
        ...payload,
        id: formModel.id,
      });
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.securitySensitiveRule.message.updateFailed')));
        throw new Error('biz');
      }
      message.success($t('page.securitySensitiveRule.message.updateSuccess'));
    }
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || $t('page.securitySensitiveRule.message.saveFailed'));
    }
    throw e;
  } finally {
    saving.value = false;
  }
}

async function handleDelete(record: SensitiveRuleRow) {
  if (record.id === undefined) {
    return;
  }
  try {
    const response = await baseRequestClient.delete('/v1/sensitive/rule', {
      data: { id: record.id },
    } as any);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.securitySensitiveRule.message.deleteFailed')));
      return;
    }
    message.success($t('page.securitySensitiveRule.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.securitySensitiveRule.message.deleteFailed'));
  }
}

const columns: TableColumnsType<SensitiveRuleRow> = [
  { title: $t('page.securitySensitiveRule.columns.ruleKey'), dataIndex: 'rule_key', key: 'rule_key', width: 120, ellipsis: true },
  { title: $t('page.securitySensitiveRule.columns.ruleName'), dataIndex: 'rule_name', key: 'rule_name', width: 140, ellipsis: true },
  { title: $t('page.securitySensitiveRule.columns.ruleType'), dataIndex: 'rule_type', key: 'rule_type', width: 100 },
  { title: $t('page.securitySensitiveRule.columns.ruleExpress'), dataIndex: 'rule_express', key: 'rule_express', ellipsis: true },
  { title: $t('page.securitySensitiveRule.columns.rulePct'), dataIndex: 'rule_pct', key: 'rule_pct', width: 90 },
  { title: $t('page.securitySensitiveRule.columns.level'), dataIndex: 'level', key: 'level', width: 88 },
  { title: $t('page.securitySensitiveRule.columns.status'), dataIndex: 'status', key: 'status', width: 100 },
  { title: $t('page.securitySensitiveRule.columns.enable'), dataIndex: 'enable', key: 'enable', width: 80 },
  { title: $t('page.securitySensitiveRule.columns.createdAt'), dataIndex: 'gmt_created', key: 'gmt_created', width: 170 },
  { title: $t('page.securitySensitiveRule.columns.updatedAt'), dataIndex: 'gmt_updated', key: 'gmt_updated', width: 170 },
  { title: $t('page.securitySensitiveRule.columns.action'), key: 'action', width: 140, fixed: 'right' },
];

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.securitySensitiveRule.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.securitySensitiveRule.form.ruleKey')" class="query-item">
            <Input
              v-model:value="searchForm.rule_key"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveRule.placeholder.ruleKey')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.securitySensitiveRule.form.ruleName')" class="query-item">
            <Input
              v-model:value="searchForm.rule_name"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveRule.placeholder.ruleName')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.securitySensitiveRule.form.enable')" class="query-item">
            <Select
              v-model:value="searchForm.enable"
              allow-clear
              class="query-control"
              :placeholder="$t('page.securitySensitiveRule.placeholder.all')"
              :options="[
                { value: 0, label: $t('page.securitySensitiveRule.enable.disabled') },
                { value: 1, label: $t('page.securitySensitiveRule.enable.enabled') },
              ]"
            />
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="resetSearch">{{ $t('page.common.reset') }}</Button>
            <Button type="primary" ghost @click="openCreate">{{ $t('page.securitySensitiveRule.action.create') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="pagedRows"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: SensitiveRuleRow, index: number) => record.id ?? `r-${pagination.current}-${index}`"
        :scroll="{ x: 1600 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'rule_type'">
            {{ record.rule_type === 'column' ? $t('page.securitySensitiveRule.ruleType.column') : record.rule_type === 'data' ? $t('page.securitySensitiveRule.ruleType.data') : record.rule_type || '-' }}
          </template>
          <template v-else-if="column.key === 'rule_express'">
            <Tooltip :title="record.rule_express">
              <span class="inline-block max-w-[240px] truncate">{{ record.rule_express || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'level'">
            <Tag :color="levelTag(record.level).color">{{ levelTag(record.level).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">{{ statusTag(record.status).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'enable'">
            <Tag :color="record.enable === 1 ? 'green' : 'default'">{{ record.enable === 1 ? $t('page.securitySensitiveRule.enable.enabled') : $t('page.securitySensitiveRule.enable.disabled') }}</Tag>
          </template>
          <template v-else-if="column.key === 'gmt_created'">
            {{ formatTime(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatTime(record.gmt_updated) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.securitySensitiveRule.action.edit') }}</Button>
              <Popconfirm
                :title="$t('page.securitySensitiveRule.deleteConfirm')"
                placement="left"
                @confirm="handleDelete(record)"
              >
                <Button type="link" size="small" danger>{{ $t('page.securitySensitiveRule.action.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? $t('page.securitySensitiveRule.modal.createTitle') : $t('page.securitySensitiveRule.modal.editTitle')"
      :confirm-loading="saving"
      width="640px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item :label="$t('page.securitySensitiveRule.modal.ruleType')" required>
          <Select
            v-model:value="formModel.rule_type"
            :options="[
              { value: 'data', label: $t('page.securitySensitiveRule.ruleType.data') },
              { value: 'column', label: $t('page.securitySensitiveRule.ruleType.column') },
            ]"
          />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.ruleKey')" required>
          <Input v-model:value="formModel.rule_key" :placeholder="$t('page.securitySensitiveRule.modal.ruleKeyPlaceholder')" :disabled="modalMode === 'edit'" />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.ruleName')" required>
          <Input v-model:value="formModel.rule_name" :placeholder="$t('page.securitySensitiveRule.modal.ruleNamePlaceholder')" />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.ruleExpress')" required>
          <Input.TextArea v-model:value="formModel.rule_express" :rows="4" :placeholder="$t('page.securitySensitiveRule.modal.ruleExpressPlaceholder')" />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.rulePct')" required>
          <InputNumber v-model:value="formModel.rule_pct" :min="0" :max="100" class="w-full" :placeholder="$t('page.securitySensitiveRule.modal.rulePctPlaceholder')" />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.level')" required>
          <Select
            v-model:value="formModel.level"
            :options="[
              { value: 0, label: $t('page.securitySensitiveRule.level.low') },
              { value: 1, label: $t('page.securitySensitiveRule.level.high') },
            ]"
          />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.status')" required>
          <Select
            v-model:value="formModel.status"
            :options="[
              { value: -1, label: $t('page.securitySensitiveRule.status.suspected') },
              { value: 0, label: $t('page.securitySensitiveRule.status.nonSensitive') },
              { value: 1, label: $t('page.securitySensitiveRule.status.confirmed') },
            ]"
          />
        </Form.Item>
        <Form.Item :label="$t('page.securitySensitiveRule.modal.enable')" required>
          <Select
            v-model:value="formModel.enable"
            :options="[
              { value: 0, label: $t('page.securitySensitiveRule.enable.disabled') },
              { value: 1, label: $t('page.securitySensitiveRule.enable.enabled') },
            ]"
          />
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
