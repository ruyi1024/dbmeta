<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import dayjs from 'dayjs';

import { $t } from '#/locales';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Button, Card, Form, Input, Modal, Popconfirm, Space, Table, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'SettingEnvPage' });

interface EnvRow {
  description?: string;
  env_key?: string;
  env_name?: string;
  gmt_created?: string;
  gmt_updated?: string;
  id?: number;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') return {};
  const r = response as Record<string, unknown>;
  if ('data' in r && r.data !== undefined && typeof r.data === 'object' && 'status' in r) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r;
}

function formatTime(v?: string) {
  if (!v) return '-';
  const d = dayjs(v);
  return d.isValid() ? d.format('YYYY-MM-DD HH:mm:ss') : v;
}

const loading = ref(false);
const allRows = ref<EnvRow[]>([]);

const searchForm = reactive({
  env_key: '',
  env_name: '',
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
const formModel = reactive<EnvRow>({
  description: '',
  env_key: '',
  env_name: '',
  id: undefined,
});

function resetFormModel() {
  formModel.id = undefined;
  formModel.env_key = '';
  formModel.env_name = '';
  formModel.description = '';
}

async function fetchList() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/datasource_env/list');
    const body = extractApiBody(response);
    const listRaw = body.data;
    const list = Array.isArray(listRaw) ? (listRaw as EnvRow[]) : [];
    const filtered = list.filter((item) => {
      const keyMatch = !searchForm.env_key.trim() || String(item.env_key ?? '').includes(searchForm.env_key.trim());
      const nameMatch = !searchForm.env_name.trim() || String(item.env_name ?? '').includes(searchForm.env_name.trim());
      return keyMatch && nameMatch;
    });
    allRows.value = filtered;
    pagination.total = filtered.length;
    pagination.current = 1;
  } catch (e: unknown) {
    allRows.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || $t('page.settingEnv.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchList();
}

function handleReset() {
  searchForm.env_key = '';
  searchForm.env_name = '';
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

function openEdit(record: EnvRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.env_key = record.env_key ?? '';
  formModel.env_name = record.env_name ?? '';
  formModel.description = record.description ?? '';
  modalOpen.value = true;
}

async function submitModal() {
  if (!formModel.env_key?.trim()) {
    message.warning($t('page.settingEnv.message.envKeyRequired'));
    return Promise.reject();
  }
  if (!formModel.env_name?.trim()) {
    message.warning($t('page.settingEnv.message.envNameRequired'));
    return Promise.reject();
  }
  saving.value = true;
  try {
    const payload = {
      description: formModel.description?.trim() || '',
      env_key: formModel.env_key.trim(),
      env_name: formModel.env_name.trim(),
    };
    if (modalMode.value === 'create') {
      const response = await baseRequestClient.post('/v1/datasource_env/list', payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.settingEnv.message.addFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingEnv.message.createSuccess'));
    } else {
      const response = await baseRequestClient.put('/v1/datasource_env/list', { ...payload, id: formModel.id });
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.settingEnv.message.updateFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingEnv.message.updateSuccess'));
    }
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || $t('page.settingEnv.message.saveFailed'));
    }
    throw e;
  } finally {
    saving.value = false;
  }
}

async function handleDelete(record: EnvRow) {
  if (record.id === undefined) return;
  try {
    const response = await baseRequestClient.delete('/v1/datasource_env/list', {
      data: { id: record.id },
    } as any);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.settingEnv.message.deleteFailed')));
      return;
    }
    message.success($t('page.settingEnv.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingEnv.message.deleteFailed'));
  }
}

const columns = computed<TableColumnsType<EnvRow>>(() => [
  { title: $t('page.settingEnv.columns.env_key'), dataIndex: 'env_key', key: 'env_key', width: 180 },
  { title: $t('page.settingEnv.columns.env_name'), dataIndex: 'env_name', key: 'env_name', width: 180 },
  { title: $t('page.settingEnv.columns.description'), dataIndex: 'description', key: 'description' },
  { title: $t('page.settingEnv.columns.gmt_created'), dataIndex: 'gmt_created', key: 'gmt_created', width: 180 },
  { title: $t('page.settingEnv.columns.gmt_updated'), dataIndex: 'gmt_updated', key: 'gmt_updated', width: 180 },
  { title: $t('page.settingEnv.columns.action'), key: 'action', width: 140, fixed: 'right' },
]);

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.settingEnv.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.settingEnv.form.env_key')" class="query-item">
            <Input v-model:value="searchForm.env_key" allow-clear class="query-control" :placeholder="$t('page.settingEnv.placeholder.env_key')" @press-enter="handleSearch" />
          </Form.Item>
          <Form.Item :label="$t('page.settingEnv.form.env_name')" class="query-item">
            <Input v-model:value="searchForm.env_name" allow-clear class="query-control" :placeholder="$t('page.settingEnv.placeholder.env_name')" @press-enter="handleSearch" />
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
        :row-key="(record: EnvRow, index: number) => record.id ?? `env-${pagination.current}-${index}`"
        :scroll="{ x: 1000 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'gmt_created'">{{ formatTime(record.gmt_created) }}</template>
          <template v-else-if="column.key === 'gmt_updated'">{{ formatTime(record.gmt_updated) }}</template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.common.edit') }}</Button>
              <Popconfirm :title="$t('page.settingEnv.confirmDelete')" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>{{ $t('page.common.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal v-model:open="modalOpen" :title="modalMode === 'create' ? $t('page.settingEnv.modal.createTitle') : $t('page.settingEnv.modal.editTitle')" :confirm-loading="saving" width="640px" destroy-on-close @ok="submitModal">
      <Form layout="vertical" class="mt-2">
        <Form.Item :label="$t('page.settingEnv.form.env_key')" required>
          <Input v-model:value="formModel.env_key" :placeholder="$t('page.settingEnv.placeholder.env_key_example')" :disabled="modalMode === 'edit'" />
        </Form.Item>
        <Form.Item :label="$t('page.settingEnv.form.env_name')" required>
          <Input v-model:value="formModel.env_name" :placeholder="$t('page.settingEnv.placeholder.env_name')" />
        </Form.Item>
        <Form.Item :label="$t('page.settingEnv.form.description')">
          <Input.TextArea v-model:value="formModel.description" :rows="4" :placeholder="$t('page.settingEnv.placeholder.description')" />
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
