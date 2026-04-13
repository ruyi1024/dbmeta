<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

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

const providerOptions = [
  { label: 'Ollama', value: 'ollama' },
  { label: 'LM Studio', value: 'lm_studio' },
  { label: 'vLLM', value: 'vllm' },
  { label: 'Dify本地', value: 'dify_local' },
  { label: 'OpenAI', value: 'openai' },
  { label: 'DeepSeek', value: 'deepseek' },
  { label: 'Qwen', value: 'qwen' },
];

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
  showTotal: (total: number) => `共 ${total} 条`,
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
  if (!formModel.name?.trim()) return '请填写模型名称';
  if (!formModel.provider?.trim()) return '请选择提供商';
  if (!formModel.api_url?.trim()) return '请填写API地址';
  if (!formModel.model_name?.trim()) return '请填写模型标识';
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
    message.error((e as Error)?.message || '加载模型列表失败');
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
      message.error(String(body.error ?? body.message ?? '连接测试失败'));
      return;
    }
    message.success('连接测试成功');
  } catch (e: unknown) {
    message.error(extractErrorMessage(e, '连接测试失败'));
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
      message.error(String(body.error ?? body.message ?? '连接测试失败'));
      return;
    }
    message.success('连接测试成功');
  } catch (e: unknown) {
    message.error(extractErrorMessage(e, '连接测试失败'));
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
        message.error(String(body.message ?? body.error ?? '新增失败'));
        throw new Error('biz');
      }
      message.success('新增成功');
    } else {
      const response = await baseRequestClient.put(`/v1/ai/models/${formModel.id}`, payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.message ?? body.error ?? '修改失败'));
        throw new Error('biz');
      }
      message.success('修改成功');
    }
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || '保存失败');
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
      message.error(String(body.message ?? body.error ?? '删除失败'));
      return;
    }
    message.success('删除成功');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '删除失败');
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
      message.error(String(body.message ?? body.error ?? '操作失败'));
      return;
    }
    message.success(checked ? '已启用' : '已禁用');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '操作失败');
  }
}

const columns: TableColumnsType<AIModelRow> = [
  { title: '模型名称', dataIndex: 'name', key: 'name', width: 160 },
  { title: '提供商', dataIndex: 'provider', key: 'provider', width: 120 },
  { title: '模型标识', dataIndex: 'model_name', key: 'model_name', width: 160 },
  { title: 'API地址', dataIndex: 'api_url', key: 'api_url', width: 220 },
  { title: '优先级', dataIndex: 'priority', key: 'priority', width: 80 },
  { title: '启用', dataIndex: 'enabled', key: 'enabled', width: 80 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '描述', dataIndex: 'description', key: 'description', width: 180 },
  { title: '操作', key: 'action', width: 180, fixed: 'right' },
];

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
        table_column_accuracy_model_id?: number;
        table_column_comment_model_id?: number;
      } | undefined) ??
      defBody;
    const p = payload as {
      grading_model_id?: number;
      table_column_accuracy_model_id?: number;
      table_column_comment_model_id?: number;
    };
    const gid = p.grading_model_id;
    gradingDefaultModelId.value = gid !== undefined && gid !== null ? Number(gid) : undefined;
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
    message.error((e as Error)?.message || '加载默认模型失败');
  } finally {
    defaultLoading.value = false;
  }
}

async function saveDefaults() {
  defaultSaving.value = true;
  try {
    const response = await baseRequestClient.put('/v1/ai/model-defaults', {
      grading_model_id: gradingDefaultModelId.value ?? null,
      table_column_accuracy_model_id: tableColumnAccuracyDefaultModelId.value ?? null,
      table_column_comment_model_id: tableColumnCommentDefaultModelId.value ?? null,
    });
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.message ?? '保存失败'));
      return;
    }
    message.success('默认模型已保存');
    await fetchDefaults();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '保存失败');
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
      <Tabs.TabPane key="settings" tab="模型设置">
        <Card :bordered="false">
          <Form class="mb-4">
            <div class="query-grid">
              <Form.Item label="模型名称" class="query-item">
                <Input v-model:value="searchForm.name" allow-clear class="query-control" placeholder="请输入模型名称" @press-enter="handleSearch" />
              </Form.Item>
              <Form.Item label="提供商" class="query-item">
                <Select v-model:value="searchForm.provider" allow-clear class="query-control" placeholder="请选择提供商" :options="providerOptions" />
              </Form.Item>
              <Form.Item label="模型标识" class="query-item">
                <Input v-model:value="searchForm.model_name" allow-clear class="query-control" placeholder="请输入模型标识" @press-enter="handleSearch" />
              </Form.Item>
            </div>
            <div class="query-actions">
              <Space>
                <Button type="primary" @click="handleSearch">查询</Button>
                <Button @click="handleReset">重置</Button>
                <Button type="primary" ghost @click="openCreate">新建</Button>
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
                  <Button type="link" size="small" @click="handleTestConnectionById(record.id)">测试</Button>
                  <Button type="link" size="small" @click="openEdit(record)">修改</Button>
                  <Popconfirm title="确定删除此模型吗？" placement="left" @confirm="handleDelete(record)">
                    <Button type="link" size="small" danger>删除</Button>
                  </Popconfirm>
                </Space>
              </template>
            </template>
          </Table>
        </Card>
      </Tabs.TabPane>

      <Tabs.TabPane key="defaults" tab="默认模型">
        <Card :bordered="false" :loading="defaultLoading">
          <p class="mb-3 text-muted-foreground text-sm">
            为业务场景指定默认 AI 模型（从已启用的模型中选择）。未指定时将按各任务原有回退逻辑（例如 Dify
            智能体）。
          </p>
          <Form layout="vertical" class="max-w-xl">
            <Form.Item label="数据分级默认模型">
              <Select
                v-model:value="gradingDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                placeholder="不指定则使用系统内置策略"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item label="表字段备注生成默认模型">
              <Select
                v-model:value="tableColumnCommentDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                placeholder="用于 AI 生成表注释、字段注释任务；不指定则回退 Dify"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item label="表字段准确度评估默认模型">
              <Select
                v-model:value="tableColumnAccuracyDefaultModelId"
                allow-clear
                show-search
                option-filter-prop="label"
                placeholder="用于表字段与注释准确度评估任务；不指定则回退系统默认逻辑"
                :options="enabledModelOptions"
                class="w-full"
              />
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" :loading="defaultSaving" @click="saveDefaults">保存</Button>
                <Button @click="fetchDefaults">刷新</Button>
              </Space>
            </Form.Item>
          </Form>
        </Card>
      </Tabs.TabPane>
    </Tabs>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? '新建模型' : '修改模型'"
      :confirm-loading="saving"
      width="760px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item label="模型名称" required>
            <Input v-model:value="formModel.name" placeholder="如：GPT-4、Qwen-7B" />
          </Form.Item>
          <Form.Item label="提供商" required>
            <Select v-model:value="formModel.provider" placeholder="请选择提供商" :options="providerOptions" />
          </Form.Item>
          <Form.Item label="API地址" required>
            <Input v-model:value="formModel.api_url" placeholder="如：https://api.openai.com/v1/chat/completions" />
          </Form.Item>
          <Form.Item :label="modalMode === 'create' ? 'API密钥' : 'API密钥（留空不更新）'">
            <Input.Password v-model:value="formModel.api_key" placeholder="请输入API密钥" />
          </Form.Item>
          <Form.Item label="模型标识" required>
            <Input v-model:value="formModel.model_name" placeholder="如：gpt-4、qwen-max" />
          </Form.Item>
          <Form.Item label="优先级">
            <InputNumber v-model:value="formModel.priority" :min="0" :max="100" class="w-full" />
          </Form.Item>
          <Form.Item label="启用">
            <Select v-model:value="formModel.enabled" :options="[{ value: 0, label: '禁用' }, { value: 1, label: '启用' }]" />
          </Form.Item>
          <Form.Item label="超时(秒)">
            <InputNumber v-model:value="formModel.timeout" :min="1" :max="300" class="w-full" />
          </Form.Item>
          <Form.Item label="最大Token">
            <InputNumber v-model:value="formModel.max_tokens" :min="1" :max="100000" class="w-full" />
          </Form.Item>
          <Form.Item label="温度参数">
            <InputNumber v-model:value="formModel.temperature" :min="0" :max="2" :step="0.1" class="w-full" />
          </Form.Item>
          <Form.Item label="流式响应">
            <Select v-model:value="formModel.stream_enabled" :options="[{ value: 0, label: '否' }, { value: 1, label: '是' }]" />
          </Form.Item>
          <Form.Item label="描述" class="col-span-2">
            <Input.TextArea v-model:value="formModel.description" :rows="4" placeholder="请输入模型描述" />
          </Form.Item>
        </div>
      </Form>
      <template #footer>
        <Space>
          <Button @click="modalOpen = false">取消</Button>
          <Button :loading="testing" @click="handleTestConnectionBeforeSave">测试连接</Button>
          <Button type="primary" :loading="saving" @click="submitModal">保存</Button>
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
  width: 300px;
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
