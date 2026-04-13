<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'MetaBusinessInfoPage' });

interface BusinessInfoRow {
  app_description?: string;
  app_name: string;
  app_owner?: string;
  app_owner_email?: string;
  app_owner_phone?: string;
  gmt_created?: string;
  gmt_updated?: string;
  id: number;
  remark?: string;
}

const loading = ref(false);
const dataSource = ref<BusinessInfoRow[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  app_name: '',
  app_owner: '',
});

const modalOpen = ref(false);
const modalMode = ref<'create' | 'edit'>('create');
const modalLoading = ref(false);
const formModel = reactive({
  id: 0,
  app_name: '',
  app_description: '',
  app_owner: '',
  app_owner_email: '',
  app_owner_phone: '',
  remark: '',
});

const columns: TableColumnsType<BusinessInfoRow> = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 72 },
  { title: '应用名称', dataIndex: 'app_name', key: 'app_name', sorter: true },
  { title: '应用描述', dataIndex: 'app_description', key: 'app_description', ellipsis: true },
  { title: '应用负责人', dataIndex: 'app_owner', key: 'app_owner' },
  { title: '负责人邮箱', dataIndex: 'app_owner_email', key: 'app_owner_email' },
  { title: '负责人电话', dataIndex: 'app_owner_phone', key: 'app_owner_phone' },
  { title: '备注', dataIndex: 'remark', key: 'remark', ellipsis: true },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true, width: 170 },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', width: 170 },
  { title: '操作', dataIndex: 'option', key: 'option', fixed: 'right', width: 140 },
];

async function fetchList(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/business-info/list', {
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
  queryForm.app_name = '';
  queryForm.app_owner = '';
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
  formModel.app_name = '';
  formModel.app_description = '';
  formModel.app_owner = '';
  formModel.app_owner_email = '';
  formModel.app_owner_phone = '';
  formModel.remark = '';
}

function openCreate() {
  modalMode.value = 'create';
  resetForm();
  modalOpen.value = true;
}

function openEdit(record: BusinessInfoRow) {
  modalMode.value = 'edit';
  formModel.id = record.id;
  formModel.app_name = record.app_name || '';
  formModel.app_description = record.app_description || '';
  formModel.app_owner = record.app_owner || '';
  formModel.app_owner_email = record.app_owner_email || '';
  formModel.app_owner_phone = record.app_owner_phone || '';
  formModel.remark = record.remark || '';
  modalOpen.value = true;
}

async function handleModalOk() {
  if (!formModel.app_name.trim()) {
    message.warning('请填写应用名称');
    return;
  }
  modalLoading.value = true;
  try {
    const base = {
      app_name: formModel.app_name.trim(),
      app_description: formModel.app_description,
      app_owner: formModel.app_owner,
      app_owner_email: formModel.app_owner_email,
      app_owner_phone: formModel.app_owner_phone,
      remark: formModel.remark,
    };
    const payload =
      modalMode.value === 'create' ? base : { ...base, id: formModel.id };
    const response =
      modalMode.value === 'create'
        ? await baseRequestClient.post('/v1/meta/business-info/list', payload)
        : await baseRequestClient.put('/v1/meta/business-info/list', payload);
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

async function handleDelete(record: BusinessInfoRow) {
  try {
    const response = await baseRequestClient.delete(`/v1/meta/business-info/${record.id}`);
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

onMounted(fetchList);
</script>

<template>
  <div class="p-5">
    <Card title="业务信息">
      <p class="mb-4 text-sm text-gray-500">
        维护数据库相关的业务说明信息（应用名称、描述、负责人等），与技术元数据独立存储。
      </p>
      <Form class="query-form">
        <div class="query-bar">
          <Form.Item label="应用名称" class="query-item query-field">
            <Input
              v-model:value="queryForm.app_name"
              placeholder="支持模糊查询"
              allow-clear
              class="query-input"
            />
          </Form.Item>
          <Form.Item label="负责人" class="query-item query-field">
            <Input
              v-model:value="queryForm.app_owner"
              placeholder="支持模糊查询"
              allow-clear
              class="query-input"
            />
          </Form.Item>
          <Space class="query-bar-actions" :size="8">
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
            <Button type="primary" @click="openCreate">新增</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ x: 1200 }"
        :row-key="(record: BusinessInfoRow) => record.id"
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
              <a @click="openEdit(record)">编辑</a>
              <Popconfirm title="确定删除该条业务信息？" @confirm="handleDelete(record)">
                <a class="text-red-500">删除</a>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>

      <Modal
        v-model:open="modalOpen"
        :title="modalMode === 'create' ? '新增业务信息' : '编辑业务信息'"
        :confirm-loading="modalLoading"
        width="560px"
        destroy-on-close
        @ok="handleModalOk"
      >
        <Form layout="vertical">
          <Form.Item label="应用名称" required>
            <Input v-model:value="formModel.app_name" placeholder="唯一，如订单服务" />
          </Form.Item>
          <Form.Item label="应用描述">
            <Input.TextArea
              v-model:value="formModel.app_description"
              :rows="4"
              placeholder="业务背景、范围说明等"
            />
          </Form.Item>
          <Form.Item label="应用负责人">
            <Input v-model:value="formModel.app_owner" />
          </Form.Item>
          <Form.Item label="负责人邮箱">
            <Input v-model:value="formModel.app_owner_email" />
          </Form.Item>
          <Form.Item label="负责人电话">
            <Input v-model:value="formModel.app_owner_phone" />
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
