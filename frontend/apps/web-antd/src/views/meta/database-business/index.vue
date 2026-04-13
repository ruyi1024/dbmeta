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
  { title: 'ID', dataIndex: 'id', key: 'id', width: 72 },
  { title: '数据库名', dataIndex: 'database_name', key: 'database_name', sorter: true },
  { title: '应用名称', dataIndex: 'app_name', key: 'app_name', sorter: true },
  {
    title: '应用说明',
    dataIndex: 'app_description',
    key: 'app_description',
    ellipsis: true,
    width: 220,
  },
  { title: '负责人', dataIndex: 'app_owner', key: 'app_owner', width: 120 },
  { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true, width: 170 },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', width: 170 },
  { title: '操作', dataIndex: 'option', key: 'option', fixed: 'right', width: 140 },
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
    message.error(error?.message || '加载失败');
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
    message.warning('请先在「业务信息」中新增至少一条应用');
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
    message.warning('请填写数据库名');
    return;
  }
  if (!String(formModel.app_name).trim()) {
    message.warning('请选择应用名称');
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
      message.error(resData?.msg || '保存失败');
      return;
    }
    message.success(modalMode.value === 'create' ? '新增成功' : '保存成功');
    modalOpen.value = false;
    fetchList();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    modalLoading.value = false;
  }
}

async function handleDelete(record: LinkRow) {
  try {
    const response = await baseRequestClient.delete(`/v1/meta/database-business/${record.id}`);
    const resData = (response as any)?.data ?? response;
    if (resData?.success === false) {
      message.error(resData?.msg || '删除失败');
      return;
    }
    message.success('已删除');
    fetchList();
  } catch (error: any) {
    message.error(error?.message || '删除失败');
  }
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card title="库表业务关联">
      <p class="mb-4 text-sm text-gray-500">
        以<strong>数据库名</strong>与<strong>应用名</strong>为关联键，将元数据中的库与「业务信息」中的应用绑定；应用名须已在业务信息中存在。
      </p>
      <Form class="query-form">
        <div class="query-bar">
          <Form.Item label="数据库名" class="query-item query-field">
            <Input
              v-model:value="queryForm.database_name"
              placeholder="模糊查询"
              allow-clear
              class="query-input"
            />
          </Form.Item>
          <Form.Item label="应用名称" class="query-item query-field">
            <Input
              v-model:value="queryForm.app_name"
              placeholder="模糊查询"
              allow-clear
              class="query-input"
            />
          </Form.Item>
          <Space class="query-bar-actions" :size="8">
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
            <Button type="primary" @click="openCreate">新增关联</Button>
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
              <a @click="openEdit(record as LinkRow)">编辑</a>
              <Popconfirm title="确定删除该关联？" @confirm="handleDelete(record as LinkRow)">
                <a class="text-red-500">删除</a>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <Modal
        v-model:open="modalOpen"
        :title="modalMode === 'create' ? '新增关联' : '编辑关联'"
        :confirm-loading="modalLoading"
        width="520px"
        destroy-on-close
        @ok="handleModalOk"
      >
        <Form layout="vertical">
          <Form.Item label="数据库名" required>
            <Input
              v-model:value="formModel.database_name"
              placeholder="与元数据中的 database_name 一致"
            />
          </Form.Item>
          <Form.Item label="应用名称" required>
            <Select
              v-model:value="formModel.app_name"
              allow-clear
              show-search
              :options="appOptions"
              placeholder="请选择已在「业务信息」中维护的应用"
              :filter-option="
                (input: string, option: any) =>
                  (option?.label ?? '')
                    .toString()
                    .toLowerCase()
                    .includes(input.toLowerCase())
              "
            />
          </Form.Item>
          <Form.Item label="备注">
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
