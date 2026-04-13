<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'MetaDatabasePage' });

interface DatabaseItem {
  alias_name?: string;
  characters?: string;
  database_name: string;
  datasource_type: string;
  gmt_created?: string;
  gmt_updated?: string;
  host: string;
  id: number;
  is_deleted: number | string;
  ops_owner?: string;
  ops_owner_phone?: string;
  port: number | string;
}

const loading = ref(false);
const dataSource = ref<DatabaseItem[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  database_name: '',
  datasource_type: '',
  host: '',
  is_deleted: '',
  port: '',
});
const editVisible = ref(false);
const editLoading = ref(false);
const editingId = ref<number>(0);

const linkModalOpen = ref(false);
const linkLoading = ref(false);
const linkSaving = ref(false);
const linkDatabaseName = ref('');
const businessOptions = ref<{ label: string; value: string }[]>([]);
const selectedAppNames = ref<string[]>([]);
const editForm = reactive({
  alias_name: '',
  ops_owner: '',
  ops_owner_phone: '',
  is_deleted: 0,
});

const columns: TableColumnsType<DatabaseItem> = [
  { title: '数据库名', dataIndex: 'database_name', key: 'database_name', sorter: true },
  { title: '数据库别名', dataIndex: 'alias_name', key: 'alias_name' },
  { title: '库字符集', dataIndex: 'characters', key: 'characters' },
  { title: '数据库类型', dataIndex: 'datasource_type', key: 'datasource_type', sorter: true },
  { title: '所属主机', dataIndex: 'host', key: 'host' },
  { title: '所属端口', dataIndex: 'port', key: 'port' },
  { title: '运维负责人', dataIndex: 'ops_owner', key: 'ops_owner' },
  { title: '运维负责人电话', dataIndex: 'ops_owner_phone', key: 'ops_owner_phone' },
  { title: '是否删除', dataIndex: 'is_deleted', key: 'is_deleted' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', sorter: true },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', sorter: true },
  { title: '操作', dataIndex: 'option', key: 'option', fixed: 'right', width: 200 },
];

async function fetchDatabases(sorter?: Record<string, string>) {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/meta/database/list', {
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
    message.error(error?.message || '数据库查询失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchDatabases();
}

function handleReset() {
  queryForm.database_name = '';
  queryForm.datasource_type = '';
  queryForm.host = '';
  queryForm.port = '';
  queryForm.is_deleted = '';
  pagination.current = 1;
  fetchDatabases();
}

function handleTableChange(page: any, _filters: any, sorter: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  if (sorter?.field && sorter?.order) {
    fetchDatabases({ [sorter.field]: sorter.order });
    return;
  }
  fetchDatabases();
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

function deletedText(value: number | string) {
  return String(value) === '1' ? '是' : '否';
}

function openEdit(record: Record<string, any>) {
  editingId.value = record.id;
  editForm.alias_name = record.alias_name || '';
  editForm.ops_owner = record.ops_owner || '';
  editForm.ops_owner_phone = record.ops_owner_phone || '';
  editForm.is_deleted = Number(record.is_deleted) || 0;
  editVisible.value = true;
}

async function openLinkBusiness(record: DatabaseItem) {
  linkDatabaseName.value = record.database_name;
  linkModalOpen.value = true;
  linkLoading.value = true;
  selectedAppNames.value = [];
  businessOptions.value = [];
  try {
    const [bizRes, relRes] = await Promise.all([
      baseRequestClient.get('/v1/meta/business-info/list'),
      baseRequestClient.get('/v1/meta/database-business/list', {
        params: { exact_database_name: record.database_name },
      }),
    ]);
    const bizPayload = (bizRes as any)?.data ?? bizRes;
    const bizList = Array.isArray(bizPayload?.data) ? bizPayload.data : [];
    businessOptions.value = bizList
      .map((x: { app_name?: string }) => String(x.app_name || '').trim())
      .filter(Boolean)
      .map((name: string) => ({ label: name, value: name }));
    const relPayload = (relRes as any)?.data ?? relRes;
    const relList = Array.isArray(relPayload?.data) ? relPayload.data : [];
    selectedAppNames.value = relList
      .map((x: { app_name?: string }) => String(x.app_name || '').trim())
      .filter(Boolean);
  } catch (error: any) {
    message.error(error?.message || '加载业务信息失败');
    linkModalOpen.value = false;
  } finally {
    linkLoading.value = false;
  }
}

async function handleLinkSubmit() {
  linkSaving.value = true;
  try {
    const response = await baseRequestClient.post('/v1/meta/database-business/batch-sync', {
      database_name: linkDatabaseName.value,
      app_names: selectedAppNames.value,
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(payload?.msg || '保存关联失败');
      return;
    }
    message.success('业务关联已保存');
    linkModalOpen.value = false;
  } catch (error: any) {
    message.error(error?.message || '保存关联失败');
  } finally {
    linkSaving.value = false;
  }
}

async function handleUpdateSubmit() {
  if (!editingId.value) return;
  editLoading.value = true;
  try {
    const response = await baseRequestClient.put('/v1/meta/database/list', {
      ...editForm,
      id: editingId.value,
      is_deleted: Number(editForm.is_deleted) || 0,
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(payload?.msg || '修改失败');
      return;
    }
    message.success('修改成功');
    editVisible.value = false;
    fetchDatabases();
  } catch (error: any) {
    message.error(error?.message || '修改失败');
  } finally {
    editLoading.value = false;
  }
}

onMounted(fetchDatabases);
</script>

<template>
  <div class="p-5">
    <Card title="数据库列表">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="数据库名" class="query-item">
            <Input
              v-model:value="queryForm.database_name"
              placeholder="请输入数据库名"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="数据库类型" class="query-item">
            <Input
              v-model:value="queryForm.datasource_type"
              placeholder="请输入数据库类型"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属主机" class="query-item">
            <Input
              v-model:value="queryForm.host"
              placeholder="请输入所属主机"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="所属端口" class="query-item">
            <Input
              v-model:value="queryForm.port"
              placeholder="请输入所属端口"
              allow-clear
              class="query-control"
            />
          </Form.Item>
          <Form.Item label="是否删除" class="query-item">
            <Select v-model:value="queryForm.is_deleted" allow-clear class="query-control">
              <Select.Option value="0">否</Select.Option>
              <Select.Option value="1">是</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: DatabaseItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'is_deleted'">
            {{ deletedText(record.is_deleted) }}
          </template>
          <template v-else-if="column.key === 'option'">
            <Space size="small">
              <a @click="openEdit(record as DatabaseItem)">编辑</a>
              <a @click="openLinkBusiness(record as DatabaseItem)">关联业务</a>
            </Space>
          </template>
          <template v-else-if="column.key === 'gmt_created'">
            {{ formatDate(record.gmt_created) }}
          </template>
          <template v-else-if="column.key === 'gmt_updated'">
            {{ formatDate(record.gmt_updated) }}
          </template>
        </template>
      </Table>

      <Modal
        v-model:open="linkModalOpen"
        title="关联业务信息"
        :confirm-loading="linkSaving"
        :ok-button-props="{ disabled: linkLoading }"
        width="560px"
        destroy-on-close
        @ok="handleLinkSubmit"
      >
        <div v-if="linkLoading" class="py-8 text-center text-gray-500">加载中…</div>
        <template v-else>
          <p class="mb-3 text-sm text-gray-500">
            数据库：<strong>{{ linkDatabaseName }}</strong>
          </p>
          <p class="mb-2 text-sm text-gray-500">
            多选应用名称，保存后将写入「库表业务关联」表；取消勾选可解除关联。
          </p>
          <Select
            v-model:value="selectedAppNames"
            mode="multiple"
            allow-clear
            show-search
            placeholder="请选择要关联的业务（应用名称）"
            class="w-full"
            :options="businessOptions"
            :filter-option="
              (input: string, option: any) =>
                (option?.label ?? '')
                  .toString()
                  .toLowerCase()
                  .includes(input.toLowerCase())
            "
            :max-tag-count="8"
          />
          <p v-if="businessOptions.length === 0" class="mt-2 text-sm text-amber-600">
            暂无业务信息，请先在「数据字典 → 业务信息」中维护应用。
          </p>
        </template>
      </Modal>

      <Modal
        v-model:open="editVisible"
        title="编辑数据库信息"
        :confirm-loading="editLoading"
        @ok="handleUpdateSubmit"
      >
        <Form layout="vertical">
          <Form.Item label="数据库别名">
            <Input v-model:value="editForm.alias_name" />
          </Form.Item>
          <Form.Item label="数据库运维负责人">
            <Input v-model:value="editForm.ops_owner" placeholder="可选" />
          </Form.Item>
          <Form.Item label="运维负责人电话">
            <Input v-model:value="editForm.ops_owner_phone" placeholder="可选" />
          </Form.Item>
          <Form.Item label="是否删除">
            <Select v-model:value="editForm.is_deleted">
              <Select.Option :value="0">否</Select.Option>
              <Select.Option :value="1">是</Select.Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  </div>
</template>

<style scoped>
.query-grid {
  column-gap: 16px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  row-gap: 12px;
}

:deep(.query-item) {
  margin-bottom: 0;
  min-width: 0;
}

:deep(.query-item .ant-form-item-row) {
  align-items: center;
  display: flex;
}

:deep(.query-item .ant-form-item-label) {
  flex: 0 0 5.5rem;
  max-width: 7rem;
  padding-right: 8px;
  text-align: right;
}

:deep(.query-item .ant-form-item-control) {
  flex: 1;
  min-width: 0;
}

:deep(.query-control) {
  max-width: 100%;
  min-width: 0;
  width: 100%;
}

.query-actions {
  display: flex;
  justify-content: flex-end;
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

  .query-actions {
    justify-content: flex-start;
  }
}
</style>
