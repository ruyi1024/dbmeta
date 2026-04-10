<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Tooltip, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

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
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString('zh-CN', { hour12: false });
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
  showTotal: (total: number) => `共 ${total} 条`,
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
      message.error(String(payload?.msg ?? '加载用户列表失败'));
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
    message.error((e as Error)?.message || '加载用户列表失败');
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
    message.warning('请填写用户名');
    return Promise.reject();
  }
  if (!formModel.chineseName.trim()) {
    message.warning('请填写姓名');
    return Promise.reject();
  }
  if (modalMode.value === 'create' && !formModel.password.trim()) {
    message.warning('新建用户时必须填写密码');
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
      message.error(String(body?.msg ?? '保存失败'));
      throw new Error('biz');
    }
    message.success(modalMode.value === 'create' ? '新增成功' : '修改成功');
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

async function handleDelete(record: UserRow) {
  if (!record.username) return;
  try {
    const response = await baseRequestClient.delete('/v1/users/manager/lists', {
      data: { username: record.username },
    } as any);
    const body = (response as any)?.data ?? response;
    if (body?.success === false) {
      message.error(String(body?.msg ?? '删除失败'));
      return;
    }
    message.success('删除成功');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '删除失败');
  }
}

const columns: TableColumnsType<UserRow> = [
  { title: '用户名', dataIndex: 'username', key: 'username', sorter: true, width: 160 },
  { title: '姓名', dataIndex: 'chineseName', key: 'chineseName', sorter: true, width: 140 },
  { title: '管理员', dataIndex: 'admin', key: 'admin', sorter: true, width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', sorter: true, width: 180 },
  { title: '修改时间', dataIndex: 'updatedAt', key: 'updatedAt', sorter: true, width: 180 },
  { title: '备注', dataIndex: 'remark', key: 'remark' },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card title="用户管理">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="关键词" class="query-item">
            <Input
              v-model:value="searchForm.keyword"
              allow-clear
              class="query-control"
              placeholder="支持搜索账号、姓名"
              @press-enter="handleSearch"
            />
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
            <Button type="primary" ghost @click="openCreate">新增用户</Button>
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
            <Tag :color="record.admin ? 'green' : 'default'">{{ record.admin ? '是' : '否' }}</Tag>
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
              <Button type="link" size="small" @click="openEdit(record)">修改</Button>
              <Popconfirm title="确认删除该用户？删除后不可恢复。" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? '新增用户' : '修改用户'"
      :confirm-loading="saving"
      width="640px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item label="用户名" required>
          <Input v-model:value="formModel.username" placeholder="请输入用户名" :disabled="modalMode === 'edit'" />
        </Form.Item>
        <Form.Item label="姓名" required>
          <Input v-model:value="formModel.chineseName" placeholder="请输入姓名" />
        </Form.Item>
        <Form.Item :label="modalMode === 'create' ? '密码' : '密码（留空不修改）'">
          <Input.Password v-model:value="formModel.password" placeholder="请输入密码" />
        </Form.Item>
        <Form.Item label="管理员">
          <Select
            v-model:value="formModel.admin"
            :options="[
              { value: 0, label: '否' },
              { value: 1, label: '是' },
            ]"
          />
        </Form.Item>
        <Form.Item label="备注">
          <Input.TextArea v-model:value="formModel.remark" :rows="3" placeholder="请输入备注" />
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
