<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

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
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString('zh-CN', { hour12: false });
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
    message.error((e as Error)?.message || '加载环境列表失败');
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
    message.warning('请填写环境标识');
    return Promise.reject();
  }
  if (!formModel.env_name?.trim()) {
    message.warning('请填写环境名称');
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
        message.error(String(body.msg ?? '新增失败'));
        throw new Error('biz');
      }
      message.success('新增成功');
    } else {
      const response = await baseRequestClient.put('/v1/datasource_env/list', { ...payload, id: formModel.id });
      const body = extractApiBody(response);
      if (body.success !== true) {
        message.error(String(body.msg ?? '修改失败'));
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

async function handleDelete(record: EnvRow) {
  if (record.id === undefined) return;
  try {
    const response = await baseRequestClient.delete('/v1/datasource_env/list', {
      data: { id: record.id },
    } as any);
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? '删除失败'));
      return;
    }
    message.success('删除成功');
    void fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '删除失败');
  }
}

const columns: TableColumnsType<EnvRow> = [
  { title: '环境标识', dataIndex: 'env_key', key: 'env_key', width: 180 },
  { title: '环境名称', dataIndex: 'env_name', key: 'env_name', width: 180 },
  { title: '环境描述', dataIndex: 'description', key: 'description' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', width: 180 },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', width: 180 },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card title="环境设置">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="环境标识" class="query-item">
            <Input v-model:value="searchForm.env_key" allow-clear class="query-control" placeholder="请输入环境标识" @press-enter="handleSearch" />
          </Form.Item>
          <Form.Item label="环境名称" class="query-item">
            <Input v-model:value="searchForm.env_name" allow-clear class="query-control" placeholder="请输入环境名称" @press-enter="handleSearch" />
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
        :row-key="(record: EnvRow, index: number) => record.id ?? `env-${pagination.current}-${index}`"
        :scroll="{ x: 1000 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'gmt_created'">{{ formatTime(record.gmt_created) }}</template>
          <template v-else-if="column.key === 'gmt_updated'">{{ formatTime(record.gmt_updated) }}</template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">修改</Button>
              <Popconfirm title="确认删除该环境？删除后不可恢复。" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal v-model:open="modalOpen" :title="modalMode === 'create' ? '新建环境' : '修改环境'" :confirm-loading="saving" width="640px" destroy-on-close @ok="submitModal">
      <Form layout="vertical" class="mt-2">
        <Form.Item label="环境标识" required>
          <Input v-model:value="formModel.env_key" placeholder="如 prod / test" :disabled="modalMode === 'edit'" />
        </Form.Item>
        <Form.Item label="环境名称" required>
          <Input v-model:value="formModel.env_name" placeholder="请输入环境名称" />
        </Form.Item>
        <Form.Item label="环境描述">
          <Input.TextArea v-model:value="formModel.description" :rows="4" placeholder="请输入环境描述" />
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
