<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import { $t } from '#/locales';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tooltip,
  message,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'SettingAiModelsPage' });

interface AIModelRow {
  api_key?: string;
  api_url?: string;
  description?: string;
  enabled?: number;
  id?: number;
  max_tokens?: number;
  model_name?: string;
  name?: string;
  priority?: number;
  provider?: string;
  stream_enabled?: number;
  temperature?: number;
  timeout?: number;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') return {};
  const r = response as Record<string, unknown>;
  if ('data' in r && r.data !== undefined && typeof r.data === 'object' && 'status' in r) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r;
}

function extractErrorMessage(err: unknown, fallback: string): string {
  const e = err as
    | {
        message?: string;
        response?: { data?: { error?: string; message?: string; msg?: string } };
      }
    | undefined;
  return (
    e?.response?.data?.error ||
    e?.response?.data?.message ||
    e?.response?.data?.msg ||
    e?.message ||
    fallback
  );
}

const providerOptions = computed(() => [
  { label: $t('page.settingAiModels.provider.ollama'), value: 'ollama' },
  { label: $t('page.settingAiModels.provider.lm_studio'), value: 'lm_studio' },
  { label: $t('page.settingAiModels.provider.vllm'), value: 'vllm' },
  { label: $t('page.settingAiModels.provider.openai'), value: 'openai' },
  { label: $t('page.settingAiModels.provider.deepseek'), value: 'deepseek' },
  { label: $t('page.settingAiModels.provider.qwen'), value: 'qwen' },
]);

const enabledModalOptions = computed(() => [
  { value: 0, label: $t('page.settingAiModels.enabledOption.off') },
  { value: 1, label: $t('page.settingAiModels.enabledOption.on') },
]);

const streamBoolOptions = computed(() => [
  { value: 0, label: $t('page.settingCommon.boolNo') },
  { value: 1, label: $t('page.settingCommon.boolYes') },
]);

const activeTab = ref<'settings' | 'defaults'>('settings');

const loading = ref(false);
const allRows = ref<AIModelRow[]>([]);

/** 默认模型（数据分级等） */
const defaultLoading = ref(false);
const defaultSaving = ref(false);
const gradingDefaultModelId = ref<number | undefined>(undefined);
/** AI 生成表/字段备注（元数据注释任务） */
const tableColumnCommentDefaultModelId = ref<number | undefined>(undefined);
/** 表字段准确度评估 */
const tableColumnAccuracyDefaultModelId = ref<number | undefined>(undefined);
/** 智能生成 SQL */
const sqlGenerationDefaultModelId = ref<number | undefined>(undefined);
const enabledModelOptions = ref<{ label: string; value: number }[]>([]);

const searchForm = reactive({
  model_name: '',
  name: '',
  provider: '',
});

const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  pageSizeOptions: ['10', '15', '30', '50'],
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => $t('page.settingCommon.paginationTotal', { total }),
  total: 0,
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
const testing = ref(false);

const formModel = reactive<AIModelRow>({
  api_key: '',
  api_url: '',
  description: '',
  enabled: 1,
  id: undefined,
  max_tokens: 2000,
  model_name: '',
  name: '',
  priority: 0,
  provider: '',
  stream_enabled: 0,
  temperature: 0.7,
  timeout: 30,
});

function resetFormModel() {
  formModel.id = undefined;
  formModel.name = '';
  formModel.provider = '';
  formModel.api_url = '';
  formModel.api_key = '';
  formModel.model_name = '';
  formModel.priority = 0;
  formModel.enabled = 1;
  formModel.timeout = 30;
  formModel.max_tokens = 2000;
  formModel.temperature = 0.7;
  formModel.stream_enabled = 0;
  formModel.description = '';
}

function buildPayload() {
  return {
    api_key: formModel.api_key ?? '',
    api_url: formModel.api_url?.trim() || '',
    description: formModel.description?.trim() || '',
    enabled: Number(formModel.enabled ?? 0),
    max_tokens: Number(formModel.max_tokens ?? 0),
    model_name: formModel.model_name?.trim() || '',
    name: formModel.name?.trim() || '',
    priority: Number(formModel.priority ?? 0),
    provider: formModel.provider?.trim() || '',
    stream_enabled: Number(formModel.stream_enabled ?? 0),
    temperature: Number(formModel.temperature ?? 0),
    timeout: Number(formModel.timeout ?? 0),
  };
}

function validateForm() {
  if (!formModel.name?.trim()) return $t('page.settingAiModels.validation.nameRequired');
  if (!formModel.provider?.trim()) return $t('page.settingAiModels.validation.providerRequired');
  if (!formModel.api_url?.trim()) return $t('page.settingAiModels.validation.apiUrlRequired');
  if (!formModel.model_name?.trim()) return $t('page.settingAiModels.validation.modelIdRequired');
  return '';
}

async function fetchList() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/ai/models');
    const body = extractApiBody(response);
    const raw = body.data;
    const list = Array.isArray(raw) ? (raw as AIModelRow[]) : [];
    const filtered = list.filter((item) => {
      const nameMatch = !searchForm.name.trim() || String(item.name ?? '').includes(searchForm.name.trim());
      const providerMatch = !searchForm.provider.trim() || String(item.provider ?? '') === searchForm.provider.trim();
      const modelMatch = !searchForm.model_name.trim() || String(item.model_name ?? '').includes(searchForm.model_name.trim());
      return nameMatch && providerMatch && modelMatch;
    });
    allRows.value = filtered;
    pagination.total = filtered.length;
    pagination.current = 1;
  } catch (e: unknown) {
    allRows.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || $t('page.settingAiModels.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchList();
}

function handleReset() {
  searchForm.name = '';
  searchForm.provider = '';
  searchForm.model_name = '';
  pagination.current = 1;
  void fetchList();
}

function handleTableChange(pag: TablePaginationConfig) {
  if (pag.current !== undefined) pagination.current = pag.current;
  if (pag.pageSize !== undefined) pagination.pageSize = pag.pageSize;
}

function openCreate() {
  modalMode.value = 'create';
  resetFormModel();
  modalOpen.value = true;
}

function openEdit(record: AIModelRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.name = record.name ?? '';
  formModel.provider = record.provider ?? '';
  formModel.api_url = record.api_url ?? '';
  formModel.api_key = '';
  formModel.model_name = record.model_name ?? '';
  formModel.priority = Number(record.priority ?? 0);
  formModel.enabled = Number(record.enabled ?? 0);
  formModel.timeout = Number(record.timeout ?? 30);
  formModel.max_tokens = Number(record.max_tokens ?? 2000);
  formModel.temperature = Number(record.temperature ?? 0.7);
  formModel.stream_enabled = Number(record.stream_enabled ?? 0);
  formModel.description = record.description ?? '';
  modalOpen.value = true;
}

async function handleTestConnectionById(id?: number) {
  if (!id) return;
  try {
    const response = await baseRequestClient.post(`/v1/ai/models/${id}/test`);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.error ?? body.message ?? $t('page.settingAiModels.message.testFailed')));
      return;
    }
    message.success($t('page.settingAiModels.message.testSuccess'));
  } catch (e: unknown) {
    message.error(extractErrorMessage(e, $t('page.settingAiModels.message.testFailed')));
  }
}

async function handleTestConnectionBeforeSave() {
  const err = validateForm();
  if (err) {
    message.warning(err);
    return;
  }
  testing.value = true;
  try {
    const response = await baseRequestClient.post('/v1/ai/model/test-config', buildPayload());
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.error ?? body.message ?? $t('page.settingAiModels.message.testFailed')));
      return;
    }
    message.success($t('page.settingAiModels.message.testSuccess'));
  } catch (e: unknown) {
    message.error(extractErrorMessage(e, $t('page.settingAiModels.message.testFailed')));
  } finally {
    testing.value = false;
  }
}

async function submitModal() {
  const err = validateForm();
  if (err) {
    message.warning(err);
    return Promise.reject();
  }
  saving.value = true;
  try {
    const payload = buildPayload();
    if (modalMode.value === 'create') {
      const response = await baseRequestClient.post('/v1/ai/models', payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.message ?? body.error ?? $t('page.settingAiModels.message.addFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingAiModels.message.createSuccess'));
    } else {
      const response = await baseRequestClient.put(`/v1/ai/models/${formModel.id}`, payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.message ?? body.error ?? $t('page.settingAiModels.message.updateFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingAiModels.message.updateSuccess'));
    }
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || $t('page.settingAiModels.message.saveFailed'));
    }
    throw e;
  } finally {
    saving.value = false;
  }
}

async function handleDelete(record: AIModelRow) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.delete(`/v1/ai/models/${record.id}`);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.message ?? body.error ?? $t('page.settingAiModels.message.deleteFailed')));
      return;
    }
    message.success($t('page.settingAiModels.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingAiModels.message.deleteFailed'));
  }
}

async function handleToggle(record: AIModelRow, checked: boolean) {
  if (!record.id) return;
  try {
    const response = await baseRequestClient.put(`/v1/ai/models/${record.id}/toggle`, {
      enabled: checked ? 1 : 0,
    });
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.message ?? body.error ?? $t('page.settingAiModels.message.toggleFailed')));
      return;
    }
    message.success(checked ? $t('page.settingAiModels.message.enabled') : $t('page.settingAiModels.message.disabled'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingAiModels.message.toggleFailed'));
  }
}

const columns = computed<TableColumnsType<AIModelRow>>(() => [
  { title: $t('page.settingAiModels.columns.name'), dataIndex: 'name', key: 'name', width: 160 },
  { title: $t('page.settingAiModels.columns.provider'), dataIndex: 'provider', key: 'provider', width: 120 },
  { title: $t('page.settingAiModels.columns.model_name'), dataIndex: 'model_name', key: 'model_name', width: 160 },
  { title: $t('page.settingAiModels.columns.api_url'), dataIndex: 'api_url', key: 'api_url', width: 220 },
  { title: $t('page.settingAiModels.columns.priority'), dataIndex: 'priority', key: 'priority', width: 80 },
  { title: $t('page.settingAiModels.columns.enabled'), dataIndex: 'enabled', key: 'enabled', width: 80 },
  { title: $t('page.settingAiModels.columns.status'), dataIndex: 'status', key: 'status', width: 80 },
  { title: $t('page.settingAiModels.columns.description'), dataIndex: 'description', key: 'description', width: 180 },
  { title: $t('page.settingAiModels.columns.action'), key: 'action', width: 180, fixed: 'right' },
]);

async function fetchDefaults() {
  defaultLoading.value = true;
  try {
    const [defRes, enRes] = await Promise.all([
      baseRequestClient.get('/v1/ai/model-defaults'),
      baseRequestClient.get('/v1/ai/models/enabled'),
    ]);
    const defBody = extractApiBody(defRes) as Record<string, unknown>;
    const payload =
      (defBody.data as {
        grading_model_id?: number;
        sql_generation_model_id?: number;
        table_column_accuracy_model_id?: number;
        table_column_comment_model_id?: number;
      } | undefined) ??
      defBody;
    const p = payload as {
      grading_model_id?: number;
      sql_generation_model_id?: number;
      table_column_accuracy_model_id?: number;
      table_column_comment_model_id?: number;
    };
    const gid = p.grading_model_id;
    gradingDefaultModelId.value = gid !== undefined && gid !== null ? Number(gid) : undefined;
    const sgid = p.sql_generation_model_id;
    sqlGenerationDefaultModelId.value =
      sgid !== undefined && sgid !== null ? Number(sgid) : undefined;
    const taid = p.table_column_accuracy_model_id;
    tableColumnAccuracyDefaultModelId.value =
      taid !== undefined && taid !== null ? Number(taid) : undefined;
    const tcid = p.table_column_comment_model_id;
    tableColumnCommentDefaultModelId.value =
      tcid !== undefined && tcid !== null ? Number(tcid) : undefined;

    const enBody = extractApiBody(enRes);
    const raw = enBody.data;
    const list = Array.isArray(raw) ? (raw as AIModelRow[]) : [];
    enabledModelOptions.value = list
      .filter((m) => m.id != null)
      .map((m) => ({
        label: `${m.name ?? m.model_name} (${m.provider ?? ''}/${m.model_name ?? ''})`,
        value: Number(m.id),
      }));
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingAiModels.message.defaultsLoadFailed'));
  } finally {
    defaultLoading.value = false;
  }
}

async function saveDefaults() {
  defaultSaving.value = true;
  try {
    const response = await baseRequestClient.put('/v1/ai/model-defaults', {
      grading_model_id: gradingDefaultModelId.value ?? null,
      sql_generation_model_id: sqlGenerationDefaultModelId.value ?? null,
      table_column_accuracy_model_id: tableColumnAccuracyDefaultModelId.value ?? null,
      table_column_comment_model_id: tableColumnCommentDefaultModelId.value ?? null,
    });
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.message ?? $t('page.settingAiModels.message.saveFailed')));
      return;
    }
    message.success($t('page.settingAiModels.message.defaultsSaved'));
    await fetchDefaults();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingAiModels.message.saveFailed'));
  } finally {
    defaultSaving.value = false;
  }
}

onMounted(() => {
  void fetchList();
  void fetchDefaults();
});
</script>

<template>
  <div class="p-5">
    <Tabs v-model:activeKey="activeTab" class="ai-models-tabs" type="card">
      <Tabs.TabPane key="settings" :tab="$t('page.settingAiModels.tab.settings')">
        <Card :bordered="false">
          <Form class="mb-4">
            <div class="query-grid">
              <Form.Item :label="$t('page.settingAiModels.columns.name')" class="query-item">
                <Input v-model:value="searchForm.name" allow-clear class="query-control" :placeholder="$t('page.settingAiModels.placeholder.name')" @press-enter="handleSearch" />
              </Form.Item>
              <Form.Item :label="$t('page.settingAiModels.columns.provider')" class="query-item">
                <Select v-model:value="searchForm.provider" allow-clear class="query-control" :placeholder="$t('page.settingAiModels.placeholder.provider')" :options="providerOptions" />
              </Form.Item>
              <Form.Item :label="$t('page.settingAiModels.columns.model_name')" class="query-item">
                <Input v-model:value="searchForm.model_name" allow-clear class="query-control" :placeholder="$t('page.settingAiModels.placeholder.model_name')" @press-enter="handleSearch" />
              </Form.Item>
            </div>
            <div class="query-actions">
              <Space>
                <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
                <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
                <Button type="primary" ghost @click="openCreate">{{ $t('page.common.create') }}</Button>
              </Space>
            </div>
          </Form>

          <Table
            :columns="columns"
            :data-source="pagedRows"
            :loading="loading"
            :pagination="pagination"
            :row-key="(record: AIModelRow, index?: number) => record.id ?? `ai-model-${pagination.current}-${index ?? 0}`"
            :scroll="{ x: 1700 }"
            @change="handleTableChange"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'api_url'">
                <Tooltip :title="record.api_url || '-'">
                  <span class="inline-block max-w-[200px] truncate">{{ record.api_url || '-' }}</span>
                </Tooltip>
              </template>
              <template v-else-if="column.key === 'enabled'">
                <Switch :checked="Number(record.enabled) === 1" @change="(checked) => handleToggle(record, Boolean(checked))" />
              </template>
              <template v-else-if="column.key === 'status'">
                <Badge :status="Number(record.enabled) === 1 ? 'success' : 'default'" />
              </template>
              <template v-else-if="column.key === 'description'">
                <Tooltip :title="record.description || '-'">
                  <span class="inline-block max-w-[160px] truncate">{{ record.description || '-' }}</span>
                </Tooltip>
              </template>
              <template v-else-if="column.key === 'action'">
                <Space>
                  <Button type="link" size="small" @click="handleTestConnectionById(record.id)">{{ $t('page.settingCommon.test') }}</Button>
                  <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.common.edit') }}</Button>
                  <Popconfirm :title="$t('page.settingAiModels.confirmDelete')" placement="left" @confirm="handleDelete(record)">
                    <Button type="link" size="small" danger>{{ $t('page.common.delete') }}</Button>
                  </Popconfirm>
                </Space>
              </template>
            </template>
          </Table>
        </Card>
      </Tabs.TabPane>

      <Tabs.TabPane key="defaults" :tab="$t('page.settingAiModels.tab.defaults')">
        <Card :bordered="false" :loading="defaultLoading">
          <p class="mb-3 text-muted-foreground text-sm">
            {{ $t('page.settingAiModels.defaultsIntro') }}
          </p>
          <Form layout="vertical" class="max-w-xl">
            <Form.Item :label="$t('page.settingAiModels.defaults.grading')">
              <Select
                v-model:value="gradingDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                :placeholder="$t('page.settingAiModels.defaults.gradingPlaceholder')"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item :label="$t('page.settingAiModels.defaults.sqlGeneration')">
              <Select
                v-model:value="sqlGenerationDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                :placeholder="$t('page.settingAiModels.defaults.sqlGenerationPlaceholder')"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item :label="$t('page.settingAiModels.defaults.tableComment')">
              <Select
                v-model:value="tableColumnCommentDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                :placeholder="$t('page.settingAiModels.defaults.tableCommentPlaceholder')"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item :label="$t('page.settingAiModels.defaults.accuracy')">
              <Select
                v-model:value="tableColumnAccuracyDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                :placeholder="$t('page.settingAiModels.defaults.accuracyPlaceholder')"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" :loading="defaultSaving" @click="saveDefaults">{{ $t('page.settingCommon.save') }}</Button>
                <Button @click="fetchDefaults">{{ $t('page.settingCommon.refresh') }}</Button>
              </Space>
            </Form.Item>
          </Form>
        </Card>
      </Tabs.TabPane>
    </Tabs>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? $t('page.settingAiModels.modal.createTitle') : $t('page.settingAiModels.modal.editTitle')"
      :confirm-loading="saving"
      width="760px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item :label="$t('page.settingAiModels.form.name')" required>
            <Input v-model:value="formModel.name" :placeholder="$t('page.settingAiModels.placeholder.name_example')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.provider')" required>
            <Select v-model:value="formModel.provider" :placeholder="$t('page.settingAiModels.placeholder.provider')" :options="providerOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.api_url')" required>
            <Input v-model:value="formModel.api_url" :placeholder="$t('page.settingAiModels.placeholder.api_url_example')" />
          </Form.Item>
          <Form.Item :label="modalMode === 'create' ? $t('page.settingAiModels.form.api_key') : $t('page.settingAiModels.form.api_key_edit_hint')">
            <Input.Password v-model:value="formModel.api_key" :placeholder="$t('page.settingAiModels.placeholder.api_key')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.model_name')" required>
            <Input v-model:value="formModel.model_name" :placeholder="$t('page.settingAiModels.placeholder.model_id_example')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.priority')">
            <InputNumber v-model:value="formModel.priority" :min="0" :max="100" class="w-full" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.enabled')">
            <Select v-model:value="formModel.enabled" :options="enabledModalOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.timeout')">
            <InputNumber v-model:value="formModel.timeout" :min="1" :max="300" class="w-full" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.max_tokens')">
            <InputNumber v-model:value="formModel.max_tokens" :min="1" :max="100000" class="w-full" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.temperature')">
            <InputNumber v-model:value="formModel.temperature" :min="0" :max="2" :step="0.1" class="w-full" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.stream_enabled')">
            <Select v-model:value="formModel.stream_enabled" :options="streamBoolOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingAiModels.form.description')" class="col-span-2">
            <Input.TextArea v-model:value="formModel.description" :rows="4" :placeholder="$t('page.settingAiModels.placeholder.description')" />
          </Form.Item>
        </div>
      </Form>
      <template #footer>
        <Space>
          <Button @click="modalOpen = false">{{ $t('page.settingCommon.cancel') }}</Button>
          <Button :loading="testing" @click="handleTestConnectionBeforeSave">{{ $t('page.settingCommon.testConnection') }}</Button>
          <Button type="primary" :loading="saving" @click="submitModal">{{ $t('page.settingCommon.save') }}</Button>
        </Space>
      </template>
    </Modal>
  </div>
</template>

<style scoped>
.ai-models-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 0;
}

.ai-models-tabs :deep(.ant-tabs-content) {
  padding-top: 16px;
}

.query-grid {
  column-gap: 12px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  row-gap: 8px;
}

.form-grid {
  column-gap: 12px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
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
  max-width: 100%;
  width: 100%;
}

.query-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
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

@media (max-width: 900px) {
  .form-grid {
    grid-template-columns: 1fr;
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
