<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

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
  const d = new Date(v);
  return Number.isNaN(d.getTime())
    ? v
    : d.toLocaleString('zh-CN', { hour12: false });
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
    message.error((e as Error)?.message || '加载机房列表失败');
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
    message.warning('请填写机房标识');
    return Promise.reject();
  }
  if (!formModel.idc_name?.trim()) {
    message.warning('请填写机房名');
    return Promise.reject();
  }
  if (!formModel.city?.trim()) {
    message.warning('请填写所在城市');
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
        message.error(String(body.msg ?? '新增失败'));
        throw new Error('biz');
      }
      message.success('新增成功');
    } else {
      const response = await baseRequestClient.put('/v1/datasource_idc/list', {
        ...payload,
        id: formModel.id,
      });
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

async function handleDelete(record: IdcRow) {
  if (record.id === undefined) return;
  try {
    const response = await baseRequestClient.delete('/v1/datasource_idc/list', {
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

const columns: TableColumnsType<IdcRow> = [
  { title: '机房标识', dataIndex: 'idc_key', key: 'idc_key', width: 160 },
  { title: '机房名', dataIndex: 'idc_name', key: 'idc_name', width: 180 },
  { title: '所在城市', dataIndex: 'city', key: 'city', width: 150 },
  { title: '机房备注', dataIndex: 'description', key: 'description' },
  { title: '创建时间', dataIndex: 'gmt_created', key: 'gmt_created', width: 180 },
  { title: '修改时间', dataIndex: 'gmt_updated', key: 'gmt_updated', width: 180 },
  { title: '操作', key: 'action', fixed: 'right', width: 140 },
];

onMounted(() => {
  void fetchList();
});
</script>

<template>
  <div class="p-5">
    <Card title="机房配置">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="机房标识" class="query-item">
            <Input
              v-model:value="searchForm.idc_key"
              allow-clear
              class="query-control"
              placeholder="请输入机房标识"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item label="机房名" class="query-item">
            <Input
              v-model:value="searchForm.idc_name"
              allow-clear
              class="query-control"
              placeholder="请输入机房名"
              @press-enter="handleSearch"
            />
          </Form.Item>
          <Form.Item label="所在城市" class="query-item">
            <Input
              v-model:value="searchForm.city"
              allow-clear
              class="query-control"
              placeholder="请输入所在城市"
              @press-enter="handleSearch"
            />
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
              <Button type="link" size="small" @click="openEdit(record)">修改</Button>
              <Popconfirm
                title="确认删除该机房？删除后不可恢复。"
                placement="left"
                @confirm="handleDelete(record)"
              >
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? '新建机房' : '修改机房'"
      :confirm-loading="saving"
      width="640px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <Form.Item label="机房标识" required>
          <Input
            v-model:value="formModel.idc_key"
            placeholder="如 cn-hz-a"
            :disabled="modalMode === 'edit'"
          />
        </Form.Item>
        <Form.Item label="机房名" required>
          <Input v-model:value="formModel.idc_name" placeholder="请输入机房名" />
        </Form.Item>
        <Form.Item label="所在城市" required>
          <Input v-model:value="formModel.city" placeholder="请输入所在城市" />
        </Form.Item>
        <Form.Item label="机房备注">
          <Input.TextArea v-model:value="formModel.description" :rows="4" placeholder="请输入机房备注" />
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
