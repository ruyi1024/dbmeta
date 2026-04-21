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
  Tag,
  message,
} from 'ant-design-vue';
import type { TableColumnsType } from 'ant-design-vue';
import type { TablePaginationConfig } from 'ant-design-vue/es/table/interface';

import { Page } from '@vben/common-ui';

import { baseRequestClient } from '#/api/request';
import { checkPermission } from '#/utils/check-permission';

defineOptions({ name: 'GradingAsset' });

interface GradeOpt {
  id: number;
  gradeName: string;
  enable?: number;
}

interface DsOpt {
  id?: number;
  name?: string;
  host?: string;
  port?: string;
}

interface AssetRow {
  id: number;
  datasourceId: number;
  databaseName: string;
  tableName: string;
  columnName: string;
  gradeId: number;
  gradeName?: string;
  assignSource?: string;
  remark?: string;
  gmtUpdated?: string;
}

const loading = ref(false);
const dataSource = ref<AssetRow[]>([]);
const dsOptions = ref<DsOpt[]>([]);
const gradeOptions = ref<GradeOpt[]>([]);

const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (t: number) => `共 ${t} 条`,
});

const queryForm = reactive({
  datasourceId: undefined as number | undefined,
  databaseName: '',
  tableName: '',
  gradeId: undefined as number | undefined,
});

const modalOpen = ref(false);
const saving = ref(false);
const isEdit = ref(false);
const formModel = reactive({
  id: 0,
  datasourceId: undefined as number | undefined,
  databaseName: '',
  tableName: '',
  columnName: '',
  gradeId: undefined as number | undefined,
  assignSource: 'manual',
  remark: '',
});

function dsLabel(d: DsOpt) {
  const h = d.host && d.port ? `${d.host}:${d.port}` : '';
  return `${d.name ?? '-'}${h ? ` (${h})` : ''}`;
}

function parsePage(response: unknown) {
  const body = (response as any)?.data ?? response;
  const inner = body?.data ?? body;
  const list = inner?.list ?? [];
  const total = inner?.total ?? 0;
  return {
    list: Array.isArray(list) ? list : [],
    total: Number(total) || 0,
  };
}

async function fetchGrades() {
  const res = await baseRequestClient.get('/v1/grading/grades');
  const body = (res as any)?.data ?? res;
  const raw = body?.data ?? body;
  const arr = Array.isArray(raw) ? raw : [];
  gradeOptions.value = arr
    .filter((g: any) => Number(g.enable) === 1)
    .map((g: any) => ({ id: g.id, gradeName: g.gradeName }));
}

async function fetchDatasources() {
  const res = await baseRequestClient.get('/v1/datasource/list');
  const body = (res as any)?.data ?? res;
  const list = body?.data ?? body;
  dsOptions.value = Array.isArray(list) ? list : [];
}

async function fetchAssets() {
  loading.value = true;
  try {
    const res = await baseRequestClient.get('/v1/grading/assets', {
      params: {
        current: pagination.current,
        pageSize: pagination.pageSize,
        datasourceId: queryForm.datasourceId,
        databaseName: queryForm.databaseName || undefined,
        tableName: queryForm.tableName || undefined,
        gradeId: queryForm.gradeId,
      },
    });
    const { list, total } = parsePage(res);
    dataSource.value = list as AssetRow[];
    pagination.total = total;
  } catch (e: unknown) {
    message.error((e as Error)?.message || '加载失败');
    dataSource.value = [];
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  if (!checkPermission()) return;
  isEdit.value = false;
  formModel.id = 0;
  formModel.datasourceId = undefined;
  formModel.databaseName = '';
  formModel.tableName = '';
  formModel.columnName = '';
  formModel.gradeId = gradeOptions.value[0]?.id;
  formModel.assignSource = 'manual';
  formModel.remark = '';
  modalOpen.value = true;
}

function openEdit(row: AssetRow) {
  if (!checkPermission()) return;
  isEdit.value = true;
  formModel.id = row.id;
  formModel.datasourceId = row.datasourceId;
  formModel.databaseName = row.databaseName ?? '';
  formModel.tableName = row.tableName ?? '';
  formModel.columnName = row.columnName ?? '';
  formModel.gradeId = row.gradeId;
  formModel.assignSource = row.assignSource ?? 'manual';
  formModel.remark = row.remark ?? '';
  modalOpen.value = true;
}

async function saveAsset() {
  if (!checkPermission()) return;
  if (!formModel.datasourceId || !formModel.tableName?.trim() || !formModel.gradeId) {
    message.warning('请填写数据源、表名与分级');
    return;
  }
  saving.value = true;
  try {
    const payload = {
      datasourceId: formModel.datasourceId,
      databaseName: formModel.databaseName?.trim() ?? '',
      tableName: formModel.tableName.trim(),
      columnName: formModel.columnName?.trim() ?? '',
      gradeId: formModel.gradeId,
      assignSource: formModel.assignSource || 'manual',
      remark: formModel.remark?.trim() ?? '',
    };
    if (isEdit.value) {
      await baseRequestClient.put('/v1/grading/assets', { id: formModel.id, ...payload });
      message.success('已更新');
    } else {
      await baseRequestClient.post('/v1/grading/assets', payload);
      message.success('已创建');
    }
    modalOpen.value = false;
    await fetchAssets();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function removeRow(id: number) {
  if (!checkPermission()) return;
  try {
    await baseRequestClient.delete(`/v1/grading/assets/${id}`);
    message.success('已删除');
    await fetchAssets();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '删除失败');
  }
}

function handleSearch() {
  pagination.current = 1;
  void fetchAssets();
}

function handleTableChange(p: TablePaginationConfig) {
  pagination.current = p.current ?? 1;
  pagination.pageSize = p.pageSize ?? 10;
  void fetchAssets();
}

const columns: TableColumnsType<AssetRow> = [
  { title: '数据源', key: 'ds', width: 200, ellipsis: true },
  { title: '库名', dataIndex: 'databaseName', key: 'databaseName', width: 140, ellipsis: true },
  { title: '表名', dataIndex: 'tableName', key: 'tableName', width: 160 },
  { title: '列名', dataIndex: 'columnName', key: 'columnName', width: 140 },
  { title: '分级', key: 'grade', width: 120 },
  { title: '来源', dataIndex: 'assignSource', key: 'assignSource', width: 100 },
  { title: '更新时间', dataIndex: 'gmtUpdated', key: 'gmtUpdated', width: 170 },
  { title: '操作', key: 'action', width: 140, fixed: 'right' },
];

function dsNameById(id: number) {
  const d = dsOptions.value.find((x) => x.id === id);
  return d ? dsLabel(d) : String(id);
}

onMounted(async () => {
  await fetchDatasources();
  await fetchGrades();
  await fetchAssets();
});
</script>

<template>
  <Page auto-content-height description="为库表/字段标注安全分级；列名为空表示整表默认分级。">
    <Card title="资产分级">
      <div class="asset-grade-toolbar">
        <Form layout="inline" class="flex flex-wrap gap-x-3 gap-y-2">
        <Form.Item label="数据源">
          <Select
            v-model:value="queryForm.datasourceId"
            allow-clear
            placeholder="全部"
            style="width: 200px"
            :options="dsOptions.map((d) => ({ label: dsLabel(d), value: d.id }))"
          />
        </Form.Item>
        <Form.Item label="库名">
          <Input v-model:value="queryForm.databaseName" allow-clear placeholder="模糊" style="width: 140px" />
        </Form.Item>
        <Form.Item label="表名">
          <Input v-model:value="queryForm.tableName" allow-clear placeholder="模糊" style="width: 140px" />
        </Form.Item>
        <Form.Item label="分级">
          <Select
            v-model:value="queryForm.gradeId"
            allow-clear
            placeholder="全部"
            style="width: 160px"
            :options="gradeOptions.map((g) => ({ label: g.gradeName, value: g.id }))"
          />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" @click="handleSearch">查询</Button>
            <Button @click="fetchAssets">刷新</Button>
            <Button type="primary" @click="openCreate">新建</Button>
          </Space>
        </Form.Item>
        </Form>
      </div>

      <Table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="pagination"
        bordered
        size="small"
        :scroll="{ x: 1100 }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'ds'">
            {{ dsNameById(record.datasourceId) }}
          </template>
          <template v-else-if="column.key === 'grade'">
            <Tag color="processing">{{ record.gradeName || record.gradeId }}</Tag>
          </template>
          <template v-else-if="column.key === 'columnName'">
            <span>{{ record.columnName || '（表级默认）' }}</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Button size="small" type="link" @click="openEdit(record as AssetRow)">编辑</Button>
            <Popconfirm title="确认删除该分级标注？" @confirm="removeRow(record.id)">
              <Button size="small" type="link" danger>删除</Button>
            </Popconfirm>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalOpen"
      :title="isEdit ? '编辑资产分级' : '新建资产分级'"
      :confirm-loading="saving"
      width="560px"
      destroy-on-close
      @ok="saveAsset"
    >
      <Form layout="vertical">
        <Form.Item label="数据源" required>
          <Select
            v-model:value="formModel.datasourceId"
            show-search
            option-filter-prop="label"
            placeholder="请选择"
            :options="dsOptions.map((d) => ({ label: dsLabel(d), value: d.id }))"
            style="width: 100%"
          />
        </Form.Item>
        <Form.Item label="库名">
          <Input v-model:value="formModel.databaseName" placeholder="可空" />
        </Form.Item>
        <Form.Item label="表名" required>
          <Input v-model:value="formModel.tableName" placeholder="表名" />
        </Form.Item>
        <Form.Item label="列名">
          <Input v-model:value="formModel.columnName" placeholder="留空表示整表默认分级" />
        </Form.Item>
        <Form.Item label="分级" required>
          <Select
            v-model:value="formModel.gradeId"
            placeholder="请选择"
            :options="gradeOptions.map((g) => ({ label: g.gradeName, value: g.id }))"
            style="width: 100%"
          />
        </Form.Item>
        <Form.Item label="来源">
          <Input v-model:value="formModel.assignSource" placeholder="manual / rule / import" />
        </Form.Item>
        <Form.Item label="备注">
          <Input.TextArea v-model:value="formModel.remark" :rows="2" />
        </Form.Item>
      </Form>
    </Modal>
  </Page>
</template>

<style scoped>
.asset-grade-toolbar {
  border-bottom: 1px solid hsl(var(--border));
  margin-bottom: 16px;
  padding-bottom: 16px;
}

.asset-grade-toolbar :deep(.ant-form-item) {
  margin-bottom: 8px;
}
</style>
