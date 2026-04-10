<script lang="ts" setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';

import type { TableColumnsType } from 'ant-design-vue';

import {
  Button,
  Card,
  Checkbox,
  Form,
  Input,
  message,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  Transfer,
} from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'DataSecurityQueryPrivilege' });

interface PrivilegeRow {
  id?: number;
  username?: string;
  datasource_type?: string;
  datasource?: string;
  grant_type?: string;
  database_name?: string;
  table_name?: string;
  do_select?: number;
  do_insert?: number;
  do_update?: number;
  do_delete?: number;
  do_create?: number;
  do_alter?: number;
  max_select?: number;
  max_update?: number;
  max_delete?: number;
  enable?: number;
  expire_date?: string;
  gmt_created?: string;
  reason?: string;
}

const privilegeTab = ref('grant');

const listLoading = ref(false);
const privilegeList = ref<PrivilegeRow[]>([]);
const listQuery = reactive({
  username: '',
  datasource_type: '',
  grant_type: undefined as string | undefined,
  database_name: '',
  table_name: '',
});

const checkBoxOptionsDatabase = [
  { label: '查询数据', value: 'select' },
  { label: '写入数据', value: 'insert' },
  { label: '更新数据', value: 'update' },
  { label: '删除数据', value: 'delete' },
  { label: '创建结构', value: 'create' },
  { label: '修改结构', value: 'alter' },
];

const checkBoxOptionsTable = [
  { label: '查询数据', value: 'select' },
  { label: '写入数据', value: 'insert' },
  { label: '更新数据', value: 'update' },
  { label: '删除数据', value: 'delete' },
];

const maxQueryNumberOptions = [
  { value: '100', label: '100' },
  { value: '300', label: '300' },
  { value: '500', label: '500' },
  { value: '1000', label: '1000' },
  { value: '5000', label: '5000' },
  { value: '10000', label: '10000' },
];

const expireDayOptions = [
  { value: '7', label: '7天' },
  { value: '31', label: '1月' },
  { value: '92', label: '3月' },
  { value: '183', label: '6月' },
  { value: '365', label: '1年' },
];

const GRANT_TYPES_WITH_DETAIL = [
  'MySQL',
  'Oracle',
  'PostgreSQL',
  'SQLServer',
  'ClickHouse',
  'TiDB',
  'Doris',
  'MongoDB',
];

function unwrapPayload(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') {
    return {};
  }
  const r = response as Record<string, unknown>;
  const httpBody =
    'status' in r && typeof r.status === 'number' && r.data !== undefined ? r.data : r;
  return typeof httpBody === 'object' && httpBody !== null
    ? (httpBody as Record<string, unknown>)
    : {};
}

/** 解析 POST/GET 返回值：兼容 AxiosResponse 与已解包的业务 JSON */
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
  return r as Record<string, unknown>;
}

const loading = ref(false);
const userList = ref<{ username?: string; chineseName?: string }[]>([]);
const typeList = ref<{ name?: string }[]>([]);
const datasourceList = ref<any[]>([]);
const databaseList = ref<any[]>([]);
const tableList = ref<{ table_name?: string }[]>([]);

const transferTargetKeys = ref<string[]>([]);

const formState = reactive({
  username: undefined as string | undefined,
  type: undefined as string | undefined,
  datasource: undefined as string | undefined,
  grant_type: undefined as 'database' | 'table' | undefined,
  database: undefined as string | undefined,
  privileges: ['select'] as string[],
  max_select: '500',
  max_update: '100',
  max_delete: '100',
  expire_day: '7',
  reason: '',
});

const showGrantDetail = computed(() =>
  formState.type ? GRANT_TYPES_WITH_DETAIL.includes(formState.type) : false,
);

const privilegeOptions = computed(() =>
  formState.grant_type === 'database' ? checkBoxOptionsDatabase : checkBoxOptionsTable,
);

const transferDataSource = computed(() =>
  tableList.value.map((t) => ({
    key: String(t.table_name ?? ''),
    title: String(t.table_name ?? ''),
  })),
);

watch(
  () => formState.grant_type,
  (v) => {
    if (v !== 'table') {
      transferTargetKeys.value = [];
    }
    if (v === 'table') {
      const allowed = new Set(['select', 'insert', 'update', 'delete']);
      formState.privileges = (formState.privileges || []).filter((p) => allowed.has(p));
      if (!formState.privileges.length) {
        formState.privileges = ['select'];
      }
    }
  },
);

async function loadUsers() {
  try {
    const response = await baseRequestClient.get('/v1/users/manager/lists', {
      params: { offset: 0, limit: 100 },
    });
    const payload = unwrapPayload(response);
    const raw = payload?.data ?? payload;
    userList.value = Array.isArray(raw) ? raw : [];
  } catch {
    userList.value = [];
  }
}

async function loadDatasourceTypes() {
  try {
    const response = await baseRequestClient.get('/v1/datasource_type/list', {
      params: { enable: 1 },
    });
    const payload = unwrapPayload(response);
    const raw = payload?.data ?? payload;
    typeList.value = Array.isArray(raw) ? raw : [];
  } catch {
    typeList.value = [];
  }
}

async function onTypeChange() {
  formState.datasource = undefined;
  formState.database = undefined;
  formState.grant_type = undefined;
  databaseList.value = [];
  tableList.value = [];
  transferTargetKeys.value = [];
  if (!formState.type) {
    datasourceList.value = [];
    return;
  }
  try {
    const response = await baseRequestClient.get('/v1/datasource/list', {
      params: { type: formState.type },
    });
    const payload = unwrapPayload(response);
    const raw = payload?.data ?? payload;
    datasourceList.value = Array.isArray(raw) ? raw : [];
  } catch {
    datasourceList.value = [];
  }
}

async function onDatasourceChange() {
  formState.database = undefined;
  formState.grant_type = undefined;
  tableList.value = [];
  transferTargetKeys.value = [];
  databaseList.value = [];
  if (!formState.datasource || !formState.type) {
    return;
  }
  try {
    const response = await baseRequestClient.get('/v1/query/database', {
      params: {
        datasource: formState.datasource,
        type: formState.type,
      },
    });
    const payload = unwrapPayload(response);
    const raw = payload?.data ?? payload;
    databaseList.value = Array.isArray(raw) ? raw : [];
  } catch {
    databaseList.value = [];
  }
}

async function onDatabaseChange() {
  tableList.value = [];
  transferTargetKeys.value = [];
  if (!formState.database || !formState.datasource || !formState.type) {
    return;
  }
  try {
    const response = await baseRequestClient.get('/v1/query/table', {
      params: {
        datasource: formState.datasource,
        database: formState.database,
        type: formState.type,
      },
    });
    const payload = unwrapPayload(response);
    const raw = payload?.data ?? payload;
    tableList.value = Array.isArray(raw) ? raw : [];
  } catch {
    tableList.value = [];
  }
}

async function handleSubmit() {
  if (!formState.username) {
    message.warning('请选择授权用户');
    return;
  }
  if (!formState.type) {
    message.warning('请选择授权数据源类型');
    return;
  }
  if (!formState.datasource) {
    message.warning('请选择数据源');
    return;
  }

  if (showGrantDetail.value) {
    if (!formState.grant_type) {
      message.warning('请选择授权范围');
      return;
    }
    if (!formState.database) {
      message.warning('请选择授权数据库');
      return;
    }
    if (!formState.privileges?.length) {
      message.warning('请选择授权权限');
      return;
    }
    if (formState.grant_type === 'table' && transferTargetKeys.value.length === 0) {
      message.warning('请选择授权数据表');
      return;
    }
  }

  if (!formState.expire_day) {
    message.warning('请选择有效期限');
    return;
  }
  if (!formState.reason?.trim()) {
    message.warning('请填写授权原因');
    return;
  }

  const tablesJoined =
    showGrantDetail.value && formState.grant_type === 'table'
      ? transferTargetKeys.value.join(';')
      : '';

  loading.value = true;
  try {
    const response = await baseRequestClient.post('/v1/privilege/grant', {
      username: formState.username,
      type: formState.type,
      datasource: formState.datasource,
      grant_type: formState.grant_type ?? '',
      database: formState.database ?? '',
      tables: tablesJoined,
      privileges: (formState.privileges || []).join(';'),
      max_select: formState.max_select,
      max_update: formState.max_update,
      max_delete: formState.max_delete,
      expire_day: formState.expire_day,
      reason: formState.reason.trim(),
      enable: '1',
    });
    const body = extractApiBody(response);
    if (body.success === true) {
      message.success('授权成功');
    } else {
      message.error(`执行授权失败：${String(body.msg ?? '')}`);
    }
  } catch (e: unknown) {
    message.error((e as Error)?.message || '执行授权失败');
  } finally {
    loading.value = false;
  }
}

function formatDateTime(v?: string) {
  if (!v) return '-';
  return dayjs(v).isValid() ? dayjs(v).format('YYYY-MM-DD HH:mm:ss') : v;
}

function formatDateOnly(v?: string) {
  if (!v) return '-';
  return dayjs(v).isValid() ? dayjs(v).format('YYYY-MM-DD') : v;
}

function grantTypeLabel(v?: string) {
  if (v === 'database') return '整库';
  if (v === 'table') return '按表';
  return v || '-';
}

async function fetchPrivilegeList() {
  listLoading.value = true;
  try {
    const params: Record<string, string> = {};
    if (listQuery.username.trim()) {
      params.username = listQuery.username.trim();
    }
    if (listQuery.datasource_type.trim()) {
      params.datasource_type = listQuery.datasource_type.trim();
    }
    if (listQuery.grant_type) {
      params.grant_type = listQuery.grant_type;
    }
    if (listQuery.database_name.trim()) {
      params.database_name = listQuery.database_name.trim();
    }
    if (listQuery.table_name.trim()) {
      params.table_name = listQuery.table_name.trim();
    }
    const response = await baseRequestClient.get('/v1/privilege/list', { params });
    const payload = unwrapPayload(response);
    const raw = payload?.data;
    privilegeList.value = Array.isArray(raw) ? (raw as PrivilegeRow[]) : [];
  } catch {
    privilegeList.value = [];
    message.error('加载授权列表失败');
  } finally {
    listLoading.value = false;
  }
}

function resetListQuery() {
  listQuery.username = '';
  listQuery.datasource_type = '';
  listQuery.grant_type = undefined;
  listQuery.database_name = '';
  listQuery.table_name = '';
  void fetchPrivilegeList();
}

function onPrivilegeTabChange(key: string | number) {
  if (String(key) === 'list') {
    void fetchPrivilegeList();
  }
}

const listColumns: TableColumnsType<PrivilegeRow> = [
  { title: '申请账号', dataIndex: 'username', key: 'username', width: 110, ellipsis: true },
  { title: '数据源类型', dataIndex: 'datasource_type', key: 'datasource_type', width: 100 },
  { title: '数据源', dataIndex: 'datasource', key: 'datasource', width: 180, ellipsis: true },
  {
    title: '授权方式',
    dataIndex: 'grant_type',
    key: 'grant_type',
    width: 88,
  },
  { title: '数据库', dataIndex: 'database_name', key: 'database_name', width: 120, ellipsis: true },
  { title: '数据表', dataIndex: 'table_name', key: 'table_name', width: 120, ellipsis: true },
  { title: '查询', dataIndex: 'do_select', key: 'do_select', width: 56, align: 'center' },
  { title: '插入', dataIndex: 'do_insert', key: 'do_insert', width: 56, align: 'center' },
  { title: '更新', dataIndex: 'do_update', key: 'do_update', width: 56, align: 'center' },
  { title: '删除', dataIndex: 'do_delete', key: 'do_delete', width: 56, align: 'center' },
  { title: '结构创建', dataIndex: 'do_create', key: 'do_create', width: 72, align: 'center' },
  { title: '结构变更', dataIndex: 'do_alter', key: 'do_alter', width: 72, align: 'center' },
  { title: '查询上限', dataIndex: 'max_select', key: 'max_select', width: 80 },
  { title: '更新上限', dataIndex: 'max_update', key: 'max_update', width: 80 },
  { title: '删除上限', dataIndex: 'max_delete', key: 'max_delete', width: 80 },
  {
    title: '状态',
    dataIndex: 'enable',
    key: 'enable',
    width: 72,
  },
  { title: '到期日期', dataIndex: 'expire_date', key: 'expire_date', width: 110 },
  { title: '授权日期', dataIndex: 'gmt_created', key: 'gmt_created', width: 170 },
  { title: '授权原因', dataIndex: 'reason', key: 'reason', ellipsis: true, width: 160 },
];

onMounted(() => {
  void loadUsers();
  void loadDatasourceTypes();
});
</script>

<template>
  <div class="p-5">
    <Card size="small">
      <Tabs v-model:active-key="privilegeTab" @change="onPrivilegeTabChange">
        <Tabs.TabPane key="grant" tab="数据查询授权">
          <Form
            :label-col="{ span: 4 }"
            :wrapper-col="{ span: 18 }"
            class="max-w-4xl"
            :model="formState"
            @submit.prevent
          >
        <Form.Item label="授权用户" required>
          <Select
            v-model:value="formState.username"
            show-search
            placeholder="请选择用户"
            class="w-full max-w-md"
            :options="
              userList.map((u) => ({
                value: u.username,
                label: u.chineseName || u.username || '',
              }))
            "
            :filter-option="
              (input: string, option: any) =>
                String(option?.label ?? '')
                  .toLowerCase()
                  .includes(input.toLowerCase())
            "
          />
        </Form.Item>

        <Form.Item label="授权数据源类型" required>
          <Select
            v-model:value="formState.type"
            show-search
            placeholder="请选择"
            class="w-full max-w-md"
            :options="typeList.map((t) => ({ value: t.name, label: t.name }))"
            @change="onTypeChange"
          />
        </Form.Item>

        <Form.Item label="选择数据源" required>
          <Select
            v-model:value="formState.datasource"
            show-search
            placeholder="请选择数据源"
            class="w-full max-w-md"
            :options="
              datasourceList.map((item) => ({
                value: `${item.host}:${item.port}`,
                label: `${item.name}[${item.host}:${item.port}]`,
              }))
            "
            @change="onDatasourceChange"
          />
        </Form.Item>

        <template v-if="showGrantDetail">
          <Form.Item label="授权范围" required>
            <Select
              v-model:value="formState.grant_type"
              placeholder="请选择"
              class="w-52"
              :options="[
                { value: 'database', label: '整库授权' },
                { value: 'table', label: '按表授权' },
              ]"
            />
          </Form.Item>

          <Form.Item label="授权数据库" required>
            <Select
              v-model:value="formState.database"
              show-search
              placeholder="请选择"
              class="w-full max-w-md"
              :options="
                databaseList.map((item) => ({
                  value: item.database_name,
                  label: item.database_name,
                }))
              "
              @change="onDatabaseChange"
            />
          </Form.Item>

          <Form.Item v-if="formState.grant_type === 'table'" label="授权数据表" required>
            <Transfer
              v-model:target-keys="transferTargetKeys"
              :data-source="transferDataSource"
              show-search
              :titles="['数据表', '授权表']"
              :list-style="{ width: '320px', height: '300px' }"
              :render="(item: any) => item.title"
            />
          </Form.Item>

          <Form.Item label="授权权限" required>
            <Checkbox.Group v-model:value="formState.privileges" :options="privilegeOptions" />
          </Form.Item>

          <Form.Item label="查询上限" required>
            <Select v-model:value="formState.max_select" class="w-32" :options="maxQueryNumberOptions" />
          </Form.Item>

          <Form.Item label="更新上限" required>
            <Select v-model:value="formState.max_update" class="w-32" :options="maxQueryNumberOptions" />
          </Form.Item>

          <Form.Item label="删除上限" required>
            <Select v-model:value="formState.max_delete" class="w-32" :options="maxQueryNumberOptions" />
          </Form.Item>
        </template>

        <Form.Item label="有效期限" required>
          <Select v-model:value="formState.expire_day" class="w-32" :options="expireDayOptions" />
        </Form.Item>

        <Form.Item label="授权原因" required>
          <Input.TextArea
            v-model:value="formState.reason"
            :rows="4"
            :maxlength="100"
            show-count
            placeholder="请填写授权原因"
            class="max-w-xl"
          />
        </Form.Item>

        <Form.Item :wrapper-col="{ offset: 4, span: 18 }">
          <Button type="primary" html-type="button" :loading="loading" @click="handleSubmit">
            执行授权
          </Button>
        </Form.Item>
      </Form>
        </Tabs.TabPane>

        <Tabs.TabPane key="list" tab="已授权限查询">
          <Card
            title="查询条件"
            class="list-query-toolbar mb-5"
            size="small"
            :bordered="true"
            :body-style="{ padding: '16px 20px' }"
          >
            <Form layout="inline" class="list-query-form flex flex-wrap items-end gap-x-4 gap-y-3">
              <Form.Item label="申请账号" class="mb-0">
                <Input
                  v-model:value="listQuery.username"
                  allow-clear
                  placeholder="用户名"
                  style="width: 140px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item label="数据源类型" class="mb-0">
                <Input
                  v-model:value="listQuery.datasource_type"
                  allow-clear
                  placeholder="如 MySQL"
                  style="width: 120px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item label="授权方式" class="mb-0">
                <Select
                  v-model:value="listQuery.grant_type"
                  allow-clear
                  placeholder="全部"
                  style="width: 110px"
                  :options="[
                    { value: 'database', label: '整库' },
                    { value: 'table', label: '按表' },
                  ]"
                />
              </Form.Item>
              <Form.Item label="数据库" class="mb-0">
                <Input
                  v-model:value="listQuery.database_name"
                  allow-clear
                  placeholder="模糊匹配"
                  style="width: 140px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item label="数据表" class="mb-0">
                <Input
                  v-model:value="listQuery.table_name"
                  allow-clear
                  placeholder="模糊匹配"
                  style="width: 140px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item class="mb-0">
                <Space>
                  <Button type="primary" @click="fetchPrivilegeList">查询</Button>
                  <Button @click="resetListQuery">重置</Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>

          <Table
            row-key="id"
            size="small"
            :loading="listLoading"
            :columns="listColumns"
            :data-source="privilegeList"
            :pagination="{ pageSize: 20, showSizeChanger: true, showTotal: (t) => `共 ${t} 条` }"
            :scroll="{ x: 2200 }"
          >
            <template #bodyCell="{ column, text, record }">
              <template v-if="column.key === 'grant_type'">
                {{ grantTypeLabel(record.grant_type) }}
              </template>
              <template
                v-else-if="
                  ['do_select', 'do_insert', 'do_update', 'do_delete', 'do_create', 'do_alter'].includes(
                    String(column.key),
                  )
                "
              >
                <Tag :color="Number(text) === 1 ? 'success' : 'default'">
                  {{ Number(text) === 1 ? '是' : '' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'enable'">
                <Tag :color="Number(text) === 1 ? 'success' : 'default'">
                  {{ Number(text) === 1 ? '正常' : '禁止' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'expire_date'">
                {{ formatDateOnly(record.expire_date) }}
              </template>
              <template v-else-if="column.key === 'gmt_created'">
                {{ formatDateTime(record.gmt_created) }}
              </template>
            </template>
          </Table>
        </Tabs.TabPane>
      </Tabs>
    </Card>
  </div>
</template>
