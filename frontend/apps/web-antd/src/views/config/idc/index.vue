<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import dayjs from 'dayjs';

import { $t } from '#/locales';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  message,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'ConfigIdcPage' });

interface IdcRow {
  city?: string;
  description?: string;
  gmt_created?: string;
  gmt_updated?: string;
  id?: number;
  idc_key?: string;
  idc_name?: string;
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
  return r;
}

function formatTime(v?: string) {
  if (!v) return '-';
  const d = dayjs(v);
  return d.isValid() ? d.format('YYYY-MM-DD HH:mm:ss') : v;
}

const loading = ref(false);
const allRows = ref<IdcRow[]>([]);

const searchForm = reactive({
  city: '',
  idc_key: '',
  idc_name: '',
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

const formModel = reactive<IdcRow>({
  city: '',
  description: '',
  id: undefined,
  idc_key: '',
  idc_name: '',
});

function resetFormModel() {
  formModel.city = '';
  formModel.description = '';
  formModel.id = undefined;
  formModel.idc_key = '';
  formModel.idc_name = '';
}

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, string> = {};
    if (searchForm.idc_key.trim()) params.idc_key = searchForm.idc_key.trim();
    if (searchForm.idc_name.trim()) params.idc_name = searchForm.idc_name.trim();
    if (searchForm.city.trim()) params.city = searchForm.city.trim();

    const response = await baseRequestClient.get('/v1/datasource_idc/list', {
      params,
    });
    const body = extractApiBody(response);
    const listRaw = body.data;
    const list = Array.isArray(listRaw) ? (listRaw as IdcRow[]) : [];
    allRows.value = list;
    pagination.total = list.length;
    pagination.current = 1;
  } catch (e: unknown) {
    allRows.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || $t('page.settingIdc.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchList();
}

function handleReset() {
  searchForm.idc_key = '';
  searchForm.idc_name = '';
  searchForm.city = '';
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

function openEdit(record: IdcRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.idc_key = record.idc_key ?? '';
  formModel.idc_name = record.idc_name ?? '';
  formModel.city = record.city ?? '';
  formModel.description = record.description ?? '';
  modalOpen.value = true;
}

async function submitModal() {
  if (!formModel.idc_key?.trim()) {
    message.warning($t('page.settingIdc.message.idcKeyRequired'));
    return Promise.reject();
  }
  if (!formModel.idc_name?.trim()) {
    message.warning($t('page.settingIdc.message.idcNameRequired'));
    return Promise.reject();
  }
  if (!formModel.city?.trim()) {
    message.warning($t('page.settingIdc.message.cityRequired'));
    return Promise.reject();
  }

  saving.value = true;
  try {
    const payload = {
      city: formModel.city.trim(),
      description: formModel.description?.trim() || '',
      idc_key: formModel.idc_key.trim(),
      idc_name: formModel.idc_name.trim(),
    };
    if (modalMode.value === 'create') {
      const response = await baseRequestClient.post('/v1/datasource_idc/list', payload);
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.settingIdc.message.addFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingIdc.message.createSuccess'));
    } else {
      const response = await baseRequestClient.put('/v1/datasource_idc/list', {
        ...payload,
        id: formModel.id,
      });
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? $t('page.settingIdc.message.updateFailed')));
        throw new Error('biz');
      }
      message.success($t('page.settingIdc.message.updateSuccess'));
    }
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || $t('page.settingIdc.message.saveFailed'));
    }
    throw e;
  } finally {
    saving.value = false;
  }
}

async function handleDelete(record: IdcRow) {
  if (record.id === undefined) return;
  try {
    const response = await baseRequestClient.delete('/v1/datasource_idc/list', {
      data: { id: record.id },
    } as any);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.settingIdc.message.deleteFailed')));
      return;
    }
    message.success($t('page.settingIdc.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.settingIdc.message.deleteFailed'));
  }
}

const columns = computed<TableColumnsType<IdcRow>>(() => [
  { title: $t('page.settingIdc.columns.idc_key'), dataIndex: 'idc_key', key: 'idc_key', width: 160 },
  { title: $t('page.settingIdc.columns.idc_name'), dataIndex: 'idc_name', key: 'idc_name', width: 180 },
  { title: $t('page.settingIdc.columns.city'), dataIndex: 'city', key: 'city', width: 150 },
  { title: $t('page.settingIdc.columns.description'), dataIndex: 'description', key: 'description' },
  { title: $t('page.settingIdc.columns.gmt_created'), dataIndex: 'gmt_created', key: 'gmt_created', width: 180 },
  { title: $t('page.settingIdc.columns.gmt_updated'), dataIndex: 'gmt_updated', key: 'gmt_updated', width: 180 },
  { title: $t('page.settingIdc.columns.action'), key: 'action', fixed: 'right', width: 140 },
]);

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.settingIdc.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.settingIdc.form.idc_key')" class="query-item">
            <Input
              v-model:value="searchForm.idc_key"
              allow-clear
              class="query-control"
              :placeholder="$t('page.settingIdc.placeholder.idc_key')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.settingIdc.form.idc_name')" class="query-item">
            <Input
              v-model:value="searchForm.idc_name"
              allow-clear
              class="query-control"
              :placeholder="$t('page.settingIdc.placeholder.idc_name')"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item :label="$t('page.settingIdc.form.city')" class="query-item">
            <Input
              v-model:value="searchForm.city"
              allow-clear
              class="query-control"
              :placeholder="$t('page.settingIdc.placeholder.city')"
              @press-enter="handleSearch"
            />
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
        :row-key="(record: IdcRow, index: number) => record.id ?? `idc-${pagination.current}-${index}`"
        :scroll="{ x: 1200 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'gmt_created'">
            {{ formatTime(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatTime(record.gmt_updated) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.common.edit') }}</Button>
              <Popconfirm
                :title="$t('page.settingIdc.confirmDelete')"
                placement="left"
                @confirm="handleDelete(record)"
              >
                <Button type="link" size="small" danger>{{ $t('page.common.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? $t('page.settingIdc.modal.createTitle') : $t('page.settingIdc.modal.editTitle')"
      :confirm-loading="saving"
      width="640px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item :label="$t('page.settingIdc.form.idc_key')" required>
          <Input
            v-model:value="formModel.idc_key"
            :placeholder="$t('page.settingIdc.placeholder.idc_key_example')"
            :disabled="modalMode === 'edit'"
          />
        </Form.Item>
        <Form.Item :label="$t('page.settingIdc.form.idc_name')" required>
          <Input v-model:value="formModel.idc_name" :placeholder="$t('page.settingIdc.placeholder.idc_name')" />
        </Form.Item>
        <Form.Item :label="$t('page.settingIdc.form.city')" required>
          <Input v-model:value="formModel.city" :placeholder="$t('page.settingIdc.placeholder.city')" />
        </Form.Item>
        <Form.Item :label="$t('page.settingIdc.form.description')">
          <Input.TextArea v-model:value="formModel.description" :rows="4" :placeholder="$t('page.settingIdc.placeholder.description')" />
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
