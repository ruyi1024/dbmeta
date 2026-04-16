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
import { $t } from '#/locales';

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
  { label: $t('page.securityPrivilege.permission.select'), value: 'select' },
  { label: $t('page.securityPrivilege.permission.insert'), value: 'insert' },
  { label: $t('page.securityPrivilege.permission.update'), value: 'update' },
  { label: $t('page.securityPrivilege.permission.delete'), value: 'delete' },
  { label: $t('page.securityPrivilege.permission.create'), value: 'create' },
  { label: $t('page.securityPrivilege.permission.alter'), value: 'alter' },
];

const checkBoxOptionsTable = [
  { label: $t('page.securityPrivilege.permission.select'), value: 'select' },
  { label: $t('page.securityPrivilege.permission.insert'), value: 'insert' },
  { label: $t('page.securityPrivilege.permission.update'), value: 'update' },
  { label: $t('page.securityPrivilege.permission.delete'), value: 'delete' },
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
  { value: '7', label: $t('page.securityPrivilege.expire.7d') },
  { value: '31', label: $t('page.securityPrivilege.expire.1m') },
  { value: '92', label: $t('page.securityPrivilege.expire.3m') },
  { value: '183', label: $t('page.securityPrivilege.expire.6m') },
  { value: '365', label: $t('page.securityPrivilege.expire.1y') },
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
    message.warning($t('page.securityPrivilege.message.selectUser'));
    return;
  }
  if (!formState.type) {
    message.warning($t('page.securityPrivilege.message.selectDatasourceType'));
    return;
  }
  if (!formState.datasource) {
    message.warning($t('page.securityPrivilege.message.selectDatasource'));
    return;
  }

  if (showGrantDetail.value) {
    if (!formState.grant_type) {
      message.warning($t('page.securityPrivilege.message.selectGrantType'));
      return;
    }
    if (!formState.database) {
      message.warning($t('page.securityPrivilege.message.selectDatabase'));
      return;
    }
    if (!formState.privileges?.length) {
      message.warning($t('page.securityPrivilege.message.selectPermission'));
      return;
    }
    if (formState.grant_type === 'table' && transferTargetKeys.value.length === 0) {
      message.warning($t('page.securityPrivilege.message.selectTable'));
      return;
    }
  }

  if (!formState.expire_day) {
    message.warning($t('page.securityPrivilege.message.selectExpireDay'));
    return;
  }
  if (!formState.reason?.trim()) {
    message.warning($t('page.securityPrivilege.message.fillReason'));
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
      message.success($t('page.securityPrivilege.message.grantSuccess'));
    } else {
      message.error(`${$t('page.securityPrivilege.message.grantFailed')}${String(body.msg ?? '')}`);
    }
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.securityPrivilege.message.grantFailed'));
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
  if (v === 'database') return $t('page.securityPrivilege.grantType.database');
  if (v === 'table') return $t('page.securityPrivilege.grantType.table');
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
    message.error($t('page.securityPrivilege.message.loadListFailed'));
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
  { title: $t('page.securityPrivilege.columns.username'), dataIndex: 'username', key: 'username', width: 110, ellipsis: true },
  { title: $t('page.securityPrivilege.columns.datasourceType'), dataIndex: 'datasource_type', key: 'datasource_type', width: 100 },
  { title: $t('page.securityPrivilege.columns.datasource'), dataIndex: 'datasource', key: 'datasource', width: 180, ellipsis: true },
  {
    title: $t('page.securityPrivilege.columns.grantType'),
    dataIndex: 'grant_type',
    key: 'grant_type',
    width: 88,
  },
  { title: $t('page.securityPrivilege.columns.databaseName'), dataIndex: 'database_name', key: 'database_name', width: 120, ellipsis: true },
  { title: $t('page.securityPrivilege.columns.tableName'), dataIndex: 'table_name', key: 'table_name', width: 120, ellipsis: true },
  { title: $t('page.securityPrivilege.columns.select'), dataIndex: 'do_select', key: 'do_select', width: 56, align: 'center' },
  { title: $t('page.securityPrivilege.columns.insert'), dataIndex: 'do_insert', key: 'do_insert', width: 56, align: 'center' },
  { title: $t('page.securityPrivilege.columns.update'), dataIndex: 'do_update', key: 'do_update', width: 56, align: 'center' },
  { title: $t('page.securityPrivilege.columns.delete'), dataIndex: 'do_delete', key: 'do_delete', width: 56, align: 'center' },
  { title: $t('page.securityPrivilege.columns.create'), dataIndex: 'do_create', key: 'do_create', width: 72, align: 'center' },
  { title: $t('page.securityPrivilege.columns.alter'), dataIndex: 'do_alter', key: 'do_alter', width: 72, align: 'center' },
  { title: $t('page.securityPrivilege.columns.maxSelect'), dataIndex: 'max_select', key: 'max_select', width: 80 },
  { title: $t('page.securityPrivilege.columns.maxUpdate'), dataIndex: 'max_update', key: 'max_update', width: 80 },
  { title: $t('page.securityPrivilege.columns.maxDelete'), dataIndex: 'max_delete', key: 'max_delete', width: 80 },
  {
    title: $t('page.securityPrivilege.columns.enable'),
    dataIndex: 'enable',
    key: 'enable',
    width: 72,
  },
  { title: $t('page.securityPrivilege.columns.expireDate'), dataIndex: 'expire_date', key: 'expire_date', width: 110 },
  { title: $t('page.securityPrivilege.columns.createdAt'), dataIndex: 'gmt_created', key: 'gmt_created', width: 170 },
  { title: $t('page.securityPrivilege.columns.reason'), dataIndex: 'reason', key: 'reason', ellipsis: true, width: 160 },
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
        <Tabs.TabPane key="grant" :tab="$t('page.securityPrivilege.tabs.grant')">
          <Form
            :label-col="{ span: 4 }"
            :wrapper-col="{ span: 18 }"
            class="max-w-4xl"
            :model="formState"
            @submit.prevent
          >
        <Form.Item :label="$t('page.securityPrivilege.form.username')" required>
          <Select
            v-model:value="formState.username"
            show-search
            :placeholder="$t('page.securityPrivilege.placeholder.selectUser')"
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

        <Form.Item :label="$t('page.securityPrivilege.form.type')" required>
          <Select
            v-model:value="formState.type"
            show-search
            :placeholder="$t('page.securityPrivilege.placeholder.select')"
            class="w-full max-w-md"
            :options="typeList.map((t) => ({ value: t.name, label: t.name }))"
            @change="onTypeChange"
          />
        </Form.Item>

        <Form.Item :label="$t('page.securityPrivilege.form.datasource')" required>
          <Select
            v-model:value="formState.datasource"
            show-search
            :placeholder="$t('page.securityPrivilege.placeholder.selectDatasource')"
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
          <Form.Item :label="$t('page.securityPrivilege.form.grantType')" required>
            <Select
              v-model:value="formState.grant_type"
              :placeholder="$t('page.securityPrivilege.placeholder.select')"
              class="w-52"
              :options="[
                { value: 'database', label: $t('page.securityPrivilege.grantType.databaseFull') },
                { value: 'table', label: $t('page.securityPrivilege.grantType.tableFull') },
              ]"
            />
          </Form.Item>

          <Form.Item :label="$t('page.securityPrivilege.form.database')" required>
            <Select
              v-model:value="formState.database"
              show-search
              :placeholder="$t('page.securityPrivilege.placeholder.select')"
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

          <Form.Item v-if="formState.grant_type === 'table'" :label="$t('page.securityPrivilege.form.tables')" required>
            <Transfer
              v-model:target-keys="transferTargetKeys"
              :data-source="transferDataSource"
              show-search
              :titles="[$t('page.securityPrivilege.transfer.source'), $t('page.securityPrivilege.transfer.target')]"
              :list-style="{ width: '320px', height: '300px' }"
              :render="(item: any) => item.title"
            />
          </Form.Item>

          <Form.Item :label="$t('page.securityPrivilege.form.privileges')" required>
            <Checkbox.Group v-model:value="formState.privileges" :options="privilegeOptions" />
          </Form.Item>

          <Form.Item :label="$t('page.securityPrivilege.form.maxSelect')" required>
            <Select v-model:value="formState.max_select" class="w-32" :options="maxQueryNumberOptions" />
          </Form.Item>

          <Form.Item :label="$t('page.securityPrivilege.form.maxUpdate')" required>
            <Select v-model:value="formState.max_update" class="w-32" :options="maxQueryNumberOptions" />
          </Form.Item>

          <Form.Item :label="$t('page.securityPrivilege.form.maxDelete')" required>
            <Select v-model:value="formState.max_delete" class="w-32" :options="maxQueryNumberOptions" />
          </Form.Item>
        </template>

        <Form.Item :label="$t('page.securityPrivilege.form.expireDay')" required>
          <Select v-model:value="formState.expire_day" class="w-32" :options="expireDayOptions" />
        </Form.Item>

        <Form.Item :label="$t('page.securityPrivilege.form.reason')" required>
          <Input.TextArea
            v-model:value="formState.reason"
            :rows="4"
            :maxlength="100"
            show-count
            :placeholder="$t('page.securityPrivilege.placeholder.reason')"
            class="max-w-xl"
          />
        </Form.Item>

        <Form.Item :wrapper-col="{ offset: 4, span: 18 }">
          <Button type="primary" html-type="button" :loading="loading" @click="handleSubmit">
            {{ $t('page.securityPrivilege.action.submitGrant') }}
          </Button>
        </Form.Item>
      </Form>
        </Tabs.TabPane>

        <Tabs.TabPane key="list" :tab="$t('page.securityPrivilege.tabs.list')">
          <Card
            :title="$t('page.securityPrivilege.queryCardTitle')"
            class="list-query-toolbar mb-5"
            size="small"
            :bordered="true"
            :body-style="{ padding: '16px 20px' }"
          >
            <Form layout="inline" class="list-query-form flex flex-wrap items-end gap-x-4 gap-y-3">
              <Form.Item :label="$t('page.securityPrivilege.query.username')" class="mb-0">
                <Input
                  v-model:value="listQuery.username"
                  allow-clear
                  :placeholder="$t('page.securityPrivilege.placeholder.username')"
                  style="width: 140px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item :label="$t('page.securityPrivilege.query.datasourceType')" class="mb-0">
                <Input
                  v-model:value="listQuery.datasource_type"
                  allow-clear
                  :placeholder="$t('page.securityPrivilege.placeholder.datasourceType')"
                  style="width: 120px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item :label="$t('page.securityPrivilege.query.grantType')" class="mb-0">
                <Select
                  v-model:value="listQuery.grant_type"
                  allow-clear
                  :placeholder="$t('page.securityPrivilege.placeholder.all')"
                  style="width: 110px"
                  :options="[
                    { value: 'database', label: $t('page.securityPrivilege.grantType.database') },
                    { value: 'table', label: $t('page.securityPrivilege.grantType.table') },
                  ]"
                />
              </Form.Item>
              <Form.Item :label="$t('page.securityPrivilege.query.databaseName')" class="mb-0">
                <Input
                  v-model:value="listQuery.database_name"
                  allow-clear
                  :placeholder="$t('page.securityPrivilege.placeholder.fuzzy')"
                  style="width: 140px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item :label="$t('page.securityPrivilege.query.tableName')" class="mb-0">
                <Input
                  v-model:value="listQuery.table_name"
                  allow-clear
                  :placeholder="$t('page.securityPrivilege.placeholder.fuzzy')"
                  style="width: 140px"
                  @press-enter="fetchPrivilegeList"
                />
              </Form.Item>
              <Form.Item class="mb-0">
                <Space>
                  <Button type="primary" @click="fetchPrivilegeList">{{ $t('page.common.search') }}</Button>
                  <Button @click="resetListQuery">{{ $t('page.common.reset') }}</Button>
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
            :pagination="{ pageSize: 20, showSizeChanger: true, showTotal: (t) => `${$t('page.common.total')} ${t} ${$t('page.common.records')}` }"
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
                  {{ Number(text) === 1 ? $t('page.securityPrivilege.yes') : '' }}
                </Tag>
              </template>
              <template v-else-if="column.key === 'enable'">
                <Tag :color="Number(text) === 1 ? 'success' : 'default'">
                  {{ Number(text) === 1 ? $t('page.securityPrivilege.enable.normal') : $t('page.securityPrivilege.enable.forbidden') }}
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
