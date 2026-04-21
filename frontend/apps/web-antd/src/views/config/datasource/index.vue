<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import { $t } from '#/locales';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Badge, Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tooltip, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'SettingDatasourcePage' });

interface DatasourceRow {
  alarm_enable?: number;
  dbid?: string;
  dbmeta_enable?: number;
  enable?: number;
  env?: string;
  execute_enable?: number;
  gmt_created?: string;
  gmt_updated?: string;
  host?: string;
  id?: number;
  idc?: string;
  monitor_enable?: number;
  name?: string;
  pass?: string;
  port?: string;
  sensitive_enable?: number;
  status?: number;
  status_text?: string;
  type?: string;
  user?: string;
}

interface OptionItem {
  id?: number;
  name?: string;
  env_key?: string;
  env_name?: string;
  idc_key?: string;
  idc_name?: string;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') return {};
  const r = response as Record<string, unknown>;
  if ('data' in r && r.data !== undefined && typeof r.data === 'object' && 'status' in r) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r;
}

const loading = ref(false);
const allRows = ref<DatasourceRow[]>([]);
const typeOptions = ref<OptionItem[]>([]);
const idcOptions = ref<OptionItem[]>([]);
const envOptions = ref<OptionItem[]>([]);

const searchForm = reactive({
  host: '',
  name: '',
  type: '',
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

const formModel = reactive<DatasourceRow>({
  alarm_enable: 0,
  dbid: '',
  dbmeta_enable: 0,
  enable: 1,
  env: '',
  execute_enable: 0,
  host: '',
  id: undefined,
  idc: '',
  monitor_enable: 0,
  name: '',
  pass: '',
  port: '',
  sensitive_enable: 0,
  type: '',
  user: '',
});

const boolOptions = computed(() => [
  { value: 0, label: $t('page.settingCommon.boolNo') },
  { value: 1, label: $t('page.settingCommon.boolYes') },
]);

function resetFormModel() {
  formModel.id = undefined;
  formModel.name = '';
  formModel.type = '';
  formModel.host = '';
  formModel.port = '';
  formModel.user = '';
  formModel.pass = '';
  formModel.dbid = '';
  formModel.idc = '';
  formModel.env = '';
  formModel.enable = 1;
  formModel.execute_enable = 0;
  formModel.dbmeta_enable = 0;
  formModel.sensitive_enable = 0;
  formModel.monitor_enable = 0;
  formModel.alarm_enable = 0;
}

async function loadOptions() {
  const [idcRes, envRes, typeRes] = await Promise.all([
    baseRequestClient.get('/v1/datasource_idc/list'),
    baseRequestClient.get('/v1/datasource_env/list'),
    baseRequestClient.get('/v1/datasource_type/list'),
  ]);
  idcOptions.value = (extractApiBody(idcRes).data as OptionItem[]) || [];
  envOptions.value = (extractApiBody(envRes).data as OptionItem[]) || [];
  typeOptions.value = (extractApiBody(typeRes).data as OptionItem[]) || [];
}

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, string> = {};
    if (searchForm.name.trim()) params.name = searchForm.name.trim();
    if (searchForm.type.trim()) params.type = searchForm.type.trim();
    if (searchForm.host.trim()) params.host = searchForm.host.trim();
    const response = await baseRequestClient.get('/v1/datasource/list', { params });
    const body = extractApiBody(response);
    const listRaw = body.data;
    const list = Array.isArray(listRaw) ? (listRaw as DatasourceRow[]) : [];
    allRows.value = list;
    pagination.total = list.length;
    pagination.current = 1;
  } catch (e: unknown) {
    allRows.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || $t('page.settingDatasource.message.loadFailed'));
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
  searchForm.type = '';
  searchForm.host = '';
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

function openEdit(record: DatasourceRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.name = record.name ?? '';
  formModel.type = record.type ?? '';
  formModel.host = record.host ?? '';
  formModel.port = record.port ?? '';
  formModel.user = record.user ?? '';
  formModel.pass = '';
  formModel.dbid = record.dbid ?? '';
  formModel.idc = record.idc ?? '';
  formModel.env = record.env ?? '';
  formModel.enable = Number(record.enable ?? 0);
  formModel.execute_enable = Number(record.execute_enable ?? 0);
  formModel.dbmeta_enable = Number(record.dbmeta_enable ?? 0);
  formModel.sensitive_enable = Number(record.sensitive_enable ?? 0);
  formModel.monitor_enable = Number(record.monitor_enable ?? 0);
  formModel.alarm_enable = Number(record.alarm_enable ?? 0);
  modalOpen.value = true;
}

function validateForm(): string | null {
  if (!formModel.name?.trim()) return $t('page.settingDatasource.validation.nameRequired');
  if (!formModel.type?.trim()) return $t('page.settingDatasource.validation.typeRequired');
  if (!formModel.host?.trim()) return $t('page.settingDatasource.validation.hostRequired');
  if (!formModel.port?.trim()) return $t('page.settingDatasource.validation.portRequired');
  if (!formModel.idc?.trim()) return $t('page.settingDatasource.validation.idcRequired');
  if (!formModel.env?.trim()) return $t('page.settingDatasource.validation.envRequired');
  if (modalMode.value === 'create' && !formModel.pass?.trim())
    return $t('page.settingDatasource.validation.passwordRequired');
  return null;
}

function buildPayload() {
  return {
    alarm_enable: Number(formModel.alarm_enable ?? 0),
    dbid: formModel.dbid?.trim() || '',
    dbmeta_enable: Number(formModel.dbmeta_enable ?? 0),
    enable: Number(formModel.enable ?? 0),
    env: formModel.env?.trim() || '',
    execute_enable: Number(formModel.execute_enable ?? 0),
    host: formModel.host?.trim() || '',
    idc: formModel.idc?.trim() || '',
    monitor_enable: Number(formModel.monitor_enable ?? 0),
    name: formModel.name?.trim() || '',
    pass: formModel.pass ?? '',
    port: formModel.port?.trim() || '',
    sensitive_enable: Number(formModel.sensitive_enable ?? 0),
    type: formModel.type?.trim() || '',
    user: formModel.user?.trim() || '',
  };
}

async function handleTestConnection() {
  const err = validateForm();
  if (err) {
    message.warning(err);
    return;
  }
  testing.value = true;
  try {
    const response = await baseRequestClient.post('/v1/datasource/check', buildPayload());
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.settingDatasource.message.checkFailed')));
      return;
    }
    message.success($t('page.settingDatasource.message.checkSuccess'));
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingDatasource.message.checkFailed'));
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
      const response = await baseRequestClient.post('/v1/datasource/list', payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.settingDatasource.message.addFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingDatasource.message.createSuccess'));
    } else {
      const response = await baseRequestClient.put('/v1/datasource/list', { ...payload, id: formModel.id });
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.settingDatasource.message.updateFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingDatasource.message.updateSuccess'));
    }
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || $t('page.settingDatasource.message.saveFailed'));
    }
    throw e;
  } finally {
    saving.value = false;
  }
}

async function handleDelete(record: DatasourceRow) {
  if (record.id === undefined) return;
  try {
    const response = await baseRequestClient.delete('/v1/datasource/list', {
      data: { id: record.id },
    } as any);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.settingDatasource.message.deleteFailed')));
      return;
    }
    message.success($t('page.settingDatasource.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingDatasource.message.deleteFailed'));
  }
}

const columns = computed<TableColumnsType<DatasourceRow>>(() => [
  { title: $t('page.settingDatasource.columns.name'), dataIndex: 'name', key: 'name', width: 180 },
  { title: $t('page.settingDatasource.columns.type'), dataIndex: 'type', key: 'type', width: 130 },
  { title: $t('page.settingDatasource.columns.host'), dataIndex: 'host', key: 'host', width: 180 },
  { title: $t('page.settingDatasource.columns.port'), dataIndex: 'port', key: 'port', width: 90 },
  { title: $t('page.settingDatasource.columns.idc'), dataIndex: 'idc', key: 'idc', width: 100 },
  { title: $t('page.settingDatasource.columns.env'), dataIndex: 'env', key: 'env', width: 100 },
  { title: $t('page.settingDatasource.columns.enable'), dataIndex: 'enable', key: 'enable', width: 70 },
  { title: $t('page.settingDatasource.columns.status'), dataIndex: 'status', key: 'status', width: 80 },
  { title: $t('page.settingDatasource.columns.status_text'), dataIndex: 'status_text', key: 'status_text', width: 180 },
  { title: $t('page.settingDatasource.columns.action'), key: 'action', width: 140, fixed: 'right' },
]);

onMounted(async () => {
  try {
    await loadOptions();
    await fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingCommon.initFailed'));
  }
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.settingDatasource.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.settingDatasource.columns.name')" class="query-item">
            <Input v-model:value="searchForm.name" allow-clear class="query-control" :placeholder="$t('page.settingDatasource.placeholder.name')" @press-enter="handleSearch" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.columns.type')" class="query-item">
            <Select
              v-model:value="searchForm.type"
              allow-clear
              class="query-control"
              :placeholder="$t('page.settingDatasource.placeholder.type')"
              :options="typeOptions.map((item) => ({ label: item.name, value: item.name }))"
            />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.columns.host')" class="query-item">
            <Input v-model:value="searchForm.host" allow-clear class="query-control" :placeholder="$t('page.settingDatasource.placeholder.host')" @press-enter="handleSearch" />
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
        :row-key="(record: DatasourceRow, index?: number) => record.id ?? `ds-${pagination.current}-${index ?? 0}`"
        :scroll="{ x: 1600 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'enable'">
            <Badge :status="Number(record.enable) === 1 ? 'success' : 'default'" />
          </template>
          <template v-else-if="column.key === 'status'">
            <Badge :status="Number(record.status) === 1 ? 'success' : 'error'" />
          </template>
          <template v-else-if="column.key === 'status_text'">
            <Tooltip :title="record.status_text || '-'">
              <span class="inline-block max-w-[150px] truncate">{{ record.status_text || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.common.edit') }}</Button>
              <Popconfirm :title="$t('page.settingDatasource.confirmDelete')" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>{{ $t('page.common.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? $t('page.settingDatasource.modal.createTitle') : $t('page.settingDatasource.modal.editTitle')"
      :confirm-loading="saving"
      width="760px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item :label="$t('page.settingDatasource.form.datasourceName')" required>
            <Input v-model:value="formModel.name" :placeholder="$t('page.settingDatasource.placeholder.name')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.type')" required>
            <Select
              v-model:value="formModel.type"
              :placeholder="$t('page.settingDatasource.placeholder.type')"
              :options="typeOptions.map((item) => ({ label: item.name, value: item.name }))"
            />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.host')" required>
            <Input v-model:value="formModel.host" :placeholder="$t('page.settingDatasource.placeholder.host')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.port')" required>
            <Input v-model:value="formModel.port" :placeholder="$t('page.settingDatasource.placeholder.port')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.user')">
            <Input v-model:value="formModel.user" :placeholder="$t('page.settingDatasource.placeholder.user')" />
          </Form.Item>
          <Form.Item :label="modalMode === 'create' ? $t('page.settingDatasource.form.password') : $t('page.settingDatasource.form.passwordEditHint')">
            <Input.Password v-model:value="formModel.pass" :placeholder="$t('page.settingDatasource.placeholder.password')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.dbid')">
            <Input v-model:value="formModel.dbid" :placeholder="$t('page.settingDatasource.placeholder.dbid')" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.idc')" required>
            <Select
              v-model:value="formModel.idc"
              :placeholder="$t('page.settingDatasource.placeholder.idc')"
              :options="idcOptions.map((item) => ({ label: item.idc_name || item.idc_key, value: item.idc_key }))"
            />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.env')" required>
            <Select
              v-model:value="formModel.env"
              :placeholder="$t('page.settingDatasource.placeholder.env')"
              :options="envOptions.map((item) => ({ label: item.env_name || item.env_key, value: item.env_key }))"
            />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.enable')">
            <Select v-model:value="formModel.enable" :options="boolOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.execute_enable')">
            <Select v-model:value="formModel.execute_enable" :options="boolOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.dbmeta_enable')">
            <Select v-model:value="formModel.dbmeta_enable" :options="boolOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.sensitive_enable')">
            <Select v-model:value="formModel.sensitive_enable" :options="boolOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.monitor_enable')">
            <Select v-model:value="formModel.monitor_enable" :options="boolOptions" />
          </Form.Item>
          <Form.Item :label="$t('page.settingDatasource.form.alarm_enable')">
            <Select v-model:value="formModel.alarm_enable" :options="boolOptions" />
          </Form.Item>
        </div>
      </Form>
      <template #footer>
        <Space>
          <Button @click="modalOpen = false">{{ $t('page.settingCommon.cancel') }}</Button>
          <Button :loading="testing" @click="handleTestConnection">{{ $t('page.settingCommon.testConnection') }}</Button>
          <Button type="primary" :loading="saving" @click="submitModal">{{ $t('page.settingCommon.save') }}</Button>
        </Space>
      </template>
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
