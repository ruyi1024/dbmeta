<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Tooltip, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

import dayjs from 'dayjs';

defineOptions({ name: 'UsersManagerPage' });

interface UserRow {
  admin?: boolean;
  chineseName?: string;
  createdAt?: string;
  id?: number;
  remark?: string;
  updatedAt?: string;
  username?: string;
}

function formatTime(v?: string) {
  if (!v) return '-';
  return dayjs(v).isValid() ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : v;
}

const loading = ref(false);
const dataSource = ref<UserRow[]>([]);
const sorterState = reactive({
  sorterField: '',
  sorterOrder: '',
});
const searchForm = reactive({
  keyword: '',
});
const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  pageSizeOptions: ['10', '20', '50', '100', '200'],
  showQuickJumper: true,
  showSizeChanger: true,
  showTotal: (total: number) => $t('page.usersManager.paginationTotal', { total }),
  total: 0,
});

const modalOpen = ref(false);
const modalMode = ref<'create' | 'edit'>('create');
const saving = ref(false);
const formModel = reactive({
  admin: 0,
  chineseName: '',
  id: undefined as number | undefined,
  password: '',
  remark: '',
  username: '',
});

function resetFormModel() {
  formModel.id = undefined;
  formModel.username = '';
  formModel.chineseName = '';
  formModel.password = '';
  formModel.admin = 0;
  formModel.remark = '';
}

async function fetchList() {
  loading.value = true;
  try {
    const current = pagination.current ?? 1;
    const pageSize = pagination.pageSize ?? 10;
    const params = {
      keyword: searchForm.keyword.trim(),
      limit: pageSize,
      offset: pageSize * (current >= 2 ? current - 1 : 0),
      sorterField: sorterState.sorterField,
      sorterOrder: sorterState.sorterOrder,
    };
    const response = await baseRequestClient.get('/v1/users/manager/lists', { params });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(String(payload?.msg ?? $t('page.usersManager.message.loadFailed')));
      dataSource.value = [];
      pagination.total = 0;
      return;
    }
    const list = Array.isArray(payload?.data) ? (payload.data as UserRow[]) : [];
    dataSource.value = list;
    pagination.total = Number(payload?.total ?? list.length) || list.length;
  } catch (e: unknown) {
    dataSource.value = [];
    pagination.total = 0;
    message.error((e as Error)?.message || $t('page.usersManager.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchList();
}

function handleReset() {
  searchForm.keyword = '';
  sorterState.sorterField = '';
  sorterState.sorterOrder = '';
  pagination.current = 1;
  void fetchList();
}

function handleTableChange(pag: TablePaginationConfig, _filters: any, sorter: any) {
  if (pag.current !== undefined) pagination.current = pag.current;
  if (pag.pageSize !== undefined) pagination.pageSize = pag.pageSize;
  sorterState.sorterField = sorter?.field ? String(sorter.field) : '';
  sorterState.sorterOrder = sorter?.order ? String(sorter.order) : '';
  void fetchList();
}

function openCreate() {
  modalMode.value = 'create';
  resetFormModel();
  modalOpen.value = true;
}

function openEdit(record: UserRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.username = record.username ?? '';
  formModel.chineseName = record.chineseName ?? '';
  formModel.password = '';
  formModel.admin = record.admin ? 1 : 0;
  formModel.remark = record.remark ?? '';
  modalOpen.value = true;
}

async function submitModal() {
  if (!formModel.username.trim()) {
    message.warning($t('page.usersManager.message.usernameRequired'));
    return Promise.reject();
  }
  if (!formModel.chineseName.trim()) {
    message.warning($t('page.usersManager.message.nameRequired'));
    return Promise.reject();
  }
  if (modalMode.value === 'create' && !formModel.password.trim()) {
    message.warning($t('page.usersManager.message.passwordRequiredCreate'));
    return Promise.reject();
  }
  saving.value = true;
  try {
    const payload = {
      admin: formModel.admin === 1,
      chineseName: formModel.chineseName.trim(),
      id: formModel.id,
      password: formModel.password,
      remark: formModel.remark.trim(),
      username: formModel.username.trim(),
    };
    const response =
      modalMode.value === 'create'
        ? await baseRequestClient.post('/v1/users/manager/lists', payload)
        : await baseRequestClient.put('/v1/users/manager/lists', payload);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? $t('page.usersManager.message.saveFailed')));
      throw new Error('biz');
    }
    message.success(
      modalMode.value === 'create'
        ? $t('page.usersManager.message.createSuccess')
        : $t('page.usersManager.message.updateSuccess'),
    );
    modalOpen.value = false;
    void fetchList();
  } catch (e: unknown) {
    if ((e as Error)?.message !== 'biz') {
      message.error((e as Error)?.message || $t('page.usersManager.message.saveFailed'));
    }
    throw e;
  } finally {
    saving.value = false;
  }
}

async function handleDelete(record: UserRow) {
  if (!record.username) return;
  try {
    const response = await baseRequestClient.delete('/v1/users/manager/lists', {
      data: { username: record.username },
    } as any);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? $t('page.usersManager.message.deleteFailed')));
      return;
    }
    message.success($t('page.usersManager.message.deleteSuccess'));
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.usersManager.message.deleteFailed'));
  }
}

const columns: TableColumnsType<UserRow> = [
  { title: $t('page.usersManager.columns.username'), dataIndex: 'username', key: 'username', sorter: true, width: 160 },
  { title: $t('page.usersManager.columns.chineseName'), dataIndex: 'chineseName', key: 'chineseName', sorter: true, width: 140 },
  { title: $t('page.usersManager.columns.admin'), dataIndex: 'admin', key: 'admin', sorter: true, width: 100 },
  { title: $t('page.usersManager.columns.createdAt'), dataIndex: 'createdAt', key: 'createdAt', sorter: true, width: 180 },
  { title: $t('page.usersManager.columns.updatedAt'), dataIndex: 'updatedAt', key: 'updatedAt', sorter: true, width: 180 },
  { title: $t('page.usersManager.columns.remark'), dataIndex: 'remark', key: 'remark' },
  { title: $t('page.usersManager.columns.action'), key: 'action', width: 140, fixed: 'right' },
];

const adminOptions = computed(() => [
  { value: 0, label: $t('page.usersManager.adminTag.no') },
  { value: 1, label: $t('page.usersManager.adminTag.yes') },
]);

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.usersManager.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.usersManager.form.keyword')" class="query-item">
            <Input
              v-model:value="searchForm.keyword"
              allow-clear
              class="query-control"
              :placeholder="$t('page.usersManager.placeholder.keyword')"
              @press-enter="handleSearch"
            />
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
            <Button type="primary" ghost @click="openCreate">{{ $t('page.usersManager.action.addUser') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: UserRow, index: number) => record.id ?? `user-${index}`"
        :scroll="{ x: 1200 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'admin'">
            <Tag :color="record.admin ? 'green' : 'default'">{{
              record.admin ? $t('page.usersManager.adminTag.yes') : $t('page.usersManager.adminTag.no')
            }}</Tag>
          </template>
          <template v-else-if="column.key === 'createdAt'">{{ formatTime(record.createdAt) }}</template>
          <template v-else-if="column.key === 'updatedAt'">{{ formatTime(record.updatedAt) }}</template>
          <template v-else-if="column.key === 'remark'">
            <Tooltip :title="record.remark || '-'">
              <span class="inline-block max-w-[260px] truncate">{{ record.remark || '-' }}</span>
            </Tooltip>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record as UserRow)">{{ $t('page.common.edit') }}</Button>
              <Popconfirm :title="$t('page.usersManager.confirmDelete')" placement="left" @confirm="handleDelete(record as UserRow)">
                <Button type="link" size="small" danger>{{ $t('page.common.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? $t('page.usersManager.modal.createTitle') : $t('page.usersManager.modal.editTitle')"
      :confirm-loading="saving"
      width="640px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item :label="$t('page.usersManager.formModal.username')" required>
          <Input
            v-model:value="formModel.username"
            :placeholder="$t('page.usersManager.placeholder.username')"
            :disabled="modalMode === 'edit'"
          />
        </Form.Item>
        <Form.Item :label="$t('page.usersManager.formModal.chineseName')" required>
          <Input v-model:value="formModel.chineseName" :placeholder="$t('page.usersManager.placeholder.chineseName')" />
        </Form.Item>
        <Form.Item
          :label="
            modalMode === 'create'
              ? $t('page.usersManager.formModal.password')
              : $t('page.usersManager.formModal.passwordEditHint')
          "
        >
          <Input.Password v-model:value="formModel.password" :placeholder="$t('page.usersManager.placeholder.password')" />
        </Form.Item>
        <Form.Item :label="$t('page.usersManager.formModal.admin')">
          <Select v-model:value="formModel.admin" :options="adminOptions" />
        </Form.Item>
        <Form.Item :label="$t('page.usersManager.formModal.remark')">
          <Input.TextArea v-model:value="formModel.remark" :rows="3" :placeholder="$t('page.usersManager.placeholder.remark')" />
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
