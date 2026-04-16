<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

defineOptions({ name: 'MetaDatabaseBusinessPage' });

interface LinkRow {
  app_description?: string;
  app_name: string;
  app_owner?: string;
  database_name: string;
  gmt_created?: string;
  gmt_updated?: string;
  id: number;
  remark?: string;
}

interface BusinessOption {
  app_name: string;
}

const loading = ref(false);
const dataSource = ref<LinkRow[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  database_name: '',
  app_name: '',
});

const modalOpen = ref(false);
const modalMode = ref<'create' | 'edit'>('create');
const modalLoading = ref(false);
const formModel = reactive({
  id: 0,
  database_name: '',
  app_name: '',
  remark: '',
});

const appOptions = ref<{ label: string; value: string }[]>([]);

const columns: TableColumnsType<LinkRow> = [
  { title: $t('page.metaDatabaseBusiness.columns.id'), dataIndex: 'id', key: 'id', width: 72 },
  { title: $t('page.metaDatabaseBusiness.columns.databaseName'), dataIndex: 'database_name', key: 'database_name', sorter: true },
  { title: $t('page.metaDatabaseBusiness.columns.appName'), dataIndex: 'app_name', key: 'app_name', sorter: true },
  {
    title: $t('page.metaDatabaseBusiness.columns.appDescription'),
    dataIndex: 'app_description',
    key: 'app_description',
    ellipsis: true,
    width: 220,
  },
  { title: $t('page.metaDatabaseBusiness.columns.appOwner'), dataIndex: 'app_owner', key: 'app_owner', width: 120 },
  { title: $t('page.metaDatabaseBusiness.columns.remark'), dataIndex: 'remark', key: 'remark', ellipsis: true },
  { title: $t('page.metaDatabaseBusiness.columns.createdAt'), dataIndex: 'gmt_created', key: 'gmt_created', sorter: true, width: 170 },
  { title: $t('page.metaDatabaseBusiness.columns.updatedAt'), dataIndex: 'gmt_updated', key: 'gmt_updated', width: 170 },
  { title: $t('page.metaDatabaseBusiness.columns.operation'), dataIndex: 'option', key: 'option', fixed: 'right', width: 140 },
];

async function loadAppOptions() {
  try {
    const response = await baseRequestClient.get('/v1/meta/business-info/list', {});
    const payload = (response as any)?.data ?? response;
    const list: BusinessOption[] = Array.isArray(payload?.data) ? payload.data : [];
    const seen = new Set<string>();
    appOptions.value = list
      .map((x) => String(x.app_name || '').trim())
      .filter((name) => {
        if (!name || seen.has(name)) return false;
        seen.add(name);
        return true;
      })
      .map((name) => ({ label: name, value: name }));
  } catch {
    appOptions.value = [];
  }
}

async function fetchList(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/database-business/list', {
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
    message.error(error?.message || $t('page.metaDatabaseBusiness.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchList();
}

function handleReset() {
  queryForm.database_name = '';
  queryForm.app_name = '';
  pagination.current = 1;
  fetchList();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  if (sorter?.field && sorter?.order) {
    fetchList({ [sorter.field]: sorter.order });
    return;
  }
  fetchList();
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

function resetForm() {
  formModel.id = 0;
  formModel.database_name = '';
  formModel.app_name = '';
  formModel.remark = '';
}

async function openCreate() {
  modalMode.value = 'create';
  resetForm();
  await loadAppOptions();
  if (appOptions.value.length === 0) {
    message.warning($t('page.metaDatabaseBusiness.message.addAppFirst'));
  }
  modalOpen.value = true;
}

async function openEdit(record: LinkRow) {
  modalMode.value = 'edit';
  await loadAppOptions();
  formModel.id = record.id;
  formModel.database_name = record.database_name || '';
  formModel.app_name = record.app_name || '';
  formModel.remark = record.remark || '';
  const cur = formModel.app_name;
  if (cur && !appOptions.value.some((o) => o.value === cur)) {
    appOptions.value = [{ label: cur, value: cur }, ...appOptions.value];
  }
  modalOpen.value = true;
}

async function handleModalOk() {
  if (!formModel.database_name.trim()) {
    message.warning($t('page.metaDatabaseBusiness.message.databaseNameRequired'));
    return;
  }
  if (!String(formModel.app_name).trim()) {
    message.warning($t('page.metaDatabaseBusiness.message.appNameRequired'));
    return;
  }
  modalLoading.value = true;
  try {
    const base = {
      database_name: formModel.database_name.trim(),
      app_name: String(formModel.app_name).trim(),
      remark: formModel.remark,
    };
    const payload =
      modalMode.value === 'create' ? base : { ...base, id: formModel.id };
    const response =
      modalMode.value === 'create'
        ? await baseRequestClient.post('/v1/meta/database-business/list', payload)
        : await baseRequestClient.put('/v1/meta/database-business/list', payload);
    const resData = (response as any)?.data ?? response;
    if (resData?.success === false) {
      message.error(resData?.msg || $t('page.metaDatabaseBusiness.message.saveFailed'));
      return;
    }
    message.success(
      modalMode.value === 'create'
        ? $t('page.metaDatabaseBusiness.message.createSuccess')
        : $t('page.metaDatabaseBusiness.message.saveSuccess'),
    );
    modalOpen.value = false;
    fetchList();
  } catch (error: any) {
    message.error(error?.message || $t('page.metaDatabaseBusiness.message.saveFailed'));
  } finally {
    modalLoading.value = false;
  }
}

async function handleDelete(record: LinkRow) {
  try {
    const response = await baseRequestClient.delete(`/v1/meta/database-business/${record.id}`);
    const resData = (response as any)?.data ?? response;
    if (resData?.success === false) {
      message.error(resData?.msg || $t('page.metaDatabaseBusiness.message.deleteFailed'));
      return;
    }
    message.success($t('page.metaDatabaseBusiness.message.deleted'));
    fetchList();
  } catch (error: any) {
    message.error(error?.message || $t('page.metaDatabaseBusiness.message.deleteFailed'));
  }
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.metaDatabaseBusiness.title')">
      <p class="mb-4 text-sm text-gray-500">
        {{ $t('page.metaDatabaseBusiness.intro') }}
      </p>
      <Form class="query-form">
        <div class="query-bar">
          <Form.Item :label="$t('page.metaDatabaseBusiness.columns.databaseName')" class="query-item query-field">
            <Input
              v-model:value="queryForm.database_name"
              :placeholder="$t('page.metaDatabaseBusiness.placeholder.fuzzyQuery')"
              allow-clear
              class="query-input"
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaDatabaseBusiness.columns.appName')" class="query-item query-field">
            <Input
              v-model:value="queryForm.app_name"
              :placeholder="$t('page.metaDatabaseBusiness.placeholder.fuzzyQuery')"
              allow-clear
              class="query-input"
            />
          </Form.Item>
          <Space class="query-bar-actions" :size="8">
            <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
            <Button type="primary" @click="openCreate">{{ $t('page.metaDatabaseBusiness.action.createLink') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ x: 1180 }"
        :row-key="(record: LinkRow) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'gmt_created'">
            {{ formatDate(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatDate(record.gmt_updated) }}
          </template>
          <template v-else-if="column.key === 'option'">
            <Space>
              <a @click="openEdit(record as LinkRow)">{{ $t('page.common.edit') }}</a>
              <Popconfirm :title="$t('page.metaDatabaseBusiness.confirmDelete')" @confirm="handleDelete(record as LinkRow)">
                <a class="text-red-500">{{ $t('page.common.delete') }}</a>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <Modal
        v-model:open="modalOpen"
        :title="modalMode === 'create' ? $t('page.metaDatabaseBusiness.modal.createTitle') : $t('page.metaDatabaseBusiness.modal.editTitle')"
        :confirm-loading="modalLoading"
        width="520px"
        destroy-on-close
        @ok="handleModalOk"
      >
        <Form layout="vertical">
          <Form.Item :label="$t('page.metaDatabaseBusiness.columns.databaseName')" required>
            <Input
              v-model:value="formModel.database_name"
              :placeholder="$t('page.metaDatabaseBusiness.placeholder.databaseNameMatch')"
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaDatabaseBusiness.columns.appName')" required>
            <Select
              v-model:value="formModel.app_name"
              allow-clear
              show-search
              :options="appOptions"
              :placeholder="$t('page.metaDatabaseBusiness.placeholder.selectApp')"
              :filter-option="
                (input: string, option: any) =>
                  (option?.label ?? '')
                    .toString()
                    .toLowerCase()
                    .includes(input.toLowerCase())
              "
            />
          </Form.Item>
          <Form.Item :label="$t('page.metaDatabaseBusiness.columns.remark')">
            <Input v-model:value="formModel.remark" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  </div>
</template>

<style scoped>
.query-form {
  margin-bottom: 28px;
}

.query-form :deep(.ant-form-item) {
  margin-bottom: 0;
}

.query-bar {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
}

:deep(.query-field .ant-form-item-row) {
  align-items: center;
  display: flex;
}

:deep(.query-field .ant-form-item-label) {
  flex: 0 0 auto;
  padding-right: 8px;
  text-align: right;
}

:deep(.query-field .ant-form-item-control) {
  flex: 0 0 auto;
}

.query-input {
  max-width: 100%;
  width: 200px;
}

.query-bar-actions {
  flex-shrink: 0;
  margin-left: 4px;
}

@media (min-width: 900px) {
  .query-bar-actions {
    margin-left: auto;
  }
}

@media (max-width: 640px) {
  .query-input {
    width: 160px;
  }

  .query-bar-actions {
    margin-left: 0;
  }
}
</style>
