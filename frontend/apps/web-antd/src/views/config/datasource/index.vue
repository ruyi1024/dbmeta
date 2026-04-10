<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Badge, Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Table, Tooltip, message } from 'ant-design-vue';

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

function formatTime(v?: string) {
  if (!v) return '-';
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString('zh-CN', { hour12: false });
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

function boolOptions() {
  return [
    { value: 0, label: '否' },
    { value: 1, label: '是' },
  ];
}

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
    message.error((e as Error)?.message || '加载数据源失败');
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
  if (!formModel.name?.trim()) return '请填写数据源名称';
  if (!formModel.type?.trim()) return '请选择类型';
  if (!formModel.host?.trim()) return '请填写主机';
  if (!formModel.port?.trim()) return '请填写端口';
  if (!formModel.idc?.trim()) return '请选择机房';
  if (!formModel.env?.trim()) return '请选择环境';
  if (modalMode.value === 'create' && !formModel.pass?.trim()) return '新建时请填写密码';
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
      message.error(String(body.msg ?? '连接检查失败'));
      return;
    }
    message.success('连接检查成功');
  } catch (e: unknown) {
    message.error((e as Error)?.message || '连接检查失败');
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
        message.error(String(body.msg ?? '新增失败'));
        throw new Error('biz');
      }
      message.success('新增成功');
    } else {
      const response = await baseRequestClient.put('/v1/datasource/list', { ...payload, id: formModel.id });
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

async function handleDelete(record: DatasourceRow) {
  if (record.id === undefined) return;
  try {
    const response = await baseRequestClient.delete('/v1/datasource/list', {
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

const columns: TableColumnsType<DatasourceRow> = [
  { title: '数据源', dataIndex: 'name', key: 'name', width: 180 },
  { title: '类型', dataIndex: 'type', key: 'type', width: 130 },
  { title: '主机', dataIndex: 'host', key: 'host', width: 180 },
  { title: '端口', dataIndex: 'port', key: 'port', width: 90 },
  { title: '机房', dataIndex: 'idc', key: 'idc', width: 100 },
  { title: '环境', dataIndex: 'env', key: 'env', width: 100 },
  { title: '启用', dataIndex: 'enable', key: 'enable', width: 70 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '状态说明', dataIndex: 'status_text', key: 'status_text', width: 180 },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

onMounted(async () => {
  try {
    await loadOptions();
    await fetchList();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '初始化数据失败');
  }
});
</script>

<template>
  <div class="p-5">
    <Card title="数据源设置">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="数据源" class="query-item">
            <Input v-model:value="searchForm.name" allow-clear class="query-control" placeholder="请输入数据源名称" @press-enter="handleSearch" />
          </Form.Item>
          <Form.Item label="类型" class="query-item">
            <Select
              v-model:value="searchForm.type"
              allow-clear
              class="query-control"
              placeholder="请选择类型"
              :options="typeOptions.map((item) => ({ label: item.name, value: item.name }))"
            />
          </Form.Item>
          <Form.Item label="主机" class="query-item">
            <Input v-model:value="searchForm.host" allow-clear class="query-control" placeholder="请输入主机" @press-enter="handleSearch" />
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
        :row-key="(record: DatasourceRow, index: number) => record.id ?? `ds-${pagination.current}-${index}`"
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
              <Button type="link" size="small" @click="openEdit(record)">修改</Button>
              <Popconfirm title="确认删除该数据源？删除后不可恢复。" placement="left" @confirm="handleDelete(record)">
                <Button type="link" size="small" danger>删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="modalMode === 'create' ? '新建数据源' : '修改数据源'"
      :confirm-loading="saving"
      width="760px"
      destroy-on-close
      @ok="submitModal"
    >
      <Form layout="vertical" class="mt-2">
        <div class="form-grid">
          <Form.Item label="数据源名称" required>
            <Input v-model:value="formModel.name" placeholder="请输入数据源名称" />
          </Form.Item>
          <Form.Item label="类型" required>
            <Select
              v-model:value="formModel.type"
              placeholder="请选择类型"
              :options="typeOptions.map((item) => ({ label: item.name, value: item.name }))"
            />
          </Form.Item>
          <Form.Item label="主机" required>
            <Input v-model:value="formModel.host" placeholder="请输入主机" />
          </Form.Item>
          <Form.Item label="端口" required>
            <Input v-model:value="formModel.port" placeholder="请输入端口" />
          </Form.Item>
          <Form.Item label="用户">
            <Input v-model:value="formModel.user" placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item :label="modalMode === 'create' ? '密码' : '密码（留空将清空）'">
            <Input.Password v-model:value="formModel.pass" placeholder="请输入密码" />
          </Form.Item>
          <Form.Item label="DBID">
            <Input v-model:value="formModel.dbid" placeholder="Oracle 等场景可填写" />
          </Form.Item>
          <Form.Item label="机房" required>
            <Select
              v-model:value="formModel.idc"
              placeholder="请选择机房"
              :options="idcOptions.map((item) => ({ label: item.idc_name || item.idc_key, value: item.idc_key }))"
            />
          </Form.Item>
          <Form.Item label="环境" required>
            <Select
              v-model:value="formModel.env"
              placeholder="请选择环境"
              :options="envOptions.map((item) => ({ label: item.env_name || item.env_key, value: item.env_key }))"
            />
          </Form.Item>
          <Form.Item label="启用">
            <Select v-model:value="formModel.enable" :options="boolOptions()" />
          </Form.Item>
          <Form.Item label="查询开关">
            <Select v-model:value="formModel.execute_enable" :options="boolOptions()" />
          </Form.Item>
          <Form.Item label="元数据开关">
            <Select v-model:value="formModel.dbmeta_enable" :options="boolOptions()" />
          </Form.Item>
          <Form.Item label="探敏开关">
            <Select v-model:value="formModel.sensitive_enable" :options="boolOptions()" />
          </Form.Item>
          <Form.Item label="监控开关">
            <Select v-model:value="formModel.monitor_enable" :options="boolOptions()" />
          </Form.Item>
          <Form.Item label="告警开关">
            <Select v-model:value="formModel.alarm_enable" :options="boolOptions()" />
          </Form.Item>
        </div>
      </Form>
      <template #footer>
        <Space>
          <Button @click="modalOpen = false">取消</Button>
          <Button :loading="testing" @click="handleTestConnection">测试连接</Button>
          <Button type="primary" :loading="saving" @click="submitModal">保存</Button>
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
