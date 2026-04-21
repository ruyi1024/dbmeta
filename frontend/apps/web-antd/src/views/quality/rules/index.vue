<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import dayjs from 'dayjs';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';
import { checkPermission } from '#/utils/check-permission';

interface RuleListItem {
  createdAt: string;
  createdBy?: string;
  enabled: number;
  id: number;
  ruleConfig: string;
  ruleDesc: string;
  ruleName: string;
  ruleType: string;
  severity: string;
  threshold: number;
  updatedAt: string;
}

const loading = ref(false);
const saving = ref(false);
const modalVisible = ref(false);
const isEdit = ref(false);

const dataSource = ref<RuleListItem[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const queryForm = reactive({
  enabled: undefined as number | undefined,
  ruleType: undefined as string | undefined,
});

const formModel = reactive<Partial<RuleListItem>>({
  createdBy: 'admin',
  enabled: 1,
  ruleConfig: '',
  ruleDesc: '',
  ruleName: '',
  ruleType: '完整性',
  severity: 'medium',
  threshold: 80,
});

const columns: TableColumnsType<RuleListItem> = [
  { title: $t('page.qualityRules.columns.ruleName'), dataIndex: 'ruleName', key: 'ruleName', width: 200 },
  { title: $t('page.qualityRules.columns.ruleType'), dataIndex: 'ruleType', key: 'ruleType', width: 120 },
  { title: $t('page.qualityRules.columns.ruleDesc'), dataIndex: 'ruleDesc', key: 'ruleDesc' },
  {
    title: $t('page.qualityRules.columns.threshold'),
    dataIndex: 'threshold',
    key: 'threshold',
    width: 100,
    customRender: ({ record }) => `${record.threshold}%`,
  },
  { title: $t('page.qualityRules.columns.severity'), dataIndex: 'severity', key: 'severity', width: 100 },
  { title: $t('page.qualityRules.columns.enabled'), dataIndex: 'enabled', key: 'enabled', width: 110 },
  { title: $t('page.qualityRules.columns.createdBy'), dataIndex: 'createdBy', key: 'createdBy', width: 100 },
  { title: $t('page.qualityRules.columns.createdAt'), dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: $t('page.qualityRules.columns.operation'), key: 'option', width: 160, fixed: 'right' },
];

function severityTag(severity: string) {
  if (severity === 'high') return { color: 'red', text: $t('page.qualityIssues.level.high') };
  if (severity === 'medium') return { color: 'orange', text: $t('page.qualityIssues.level.medium') };
  return { color: 'blue', text: $t('page.qualityIssues.level.low') };
}

function ruleTypeLabel(ruleType: string) {
  switch (ruleType) {
    case '完整性':
      return $t('page.qualityIssues.issueType.completeness');
    case '准确性':
      return $t('page.qualityIssues.issueType.accuracy');
    case '唯一性':
      return $t('page.qualityIssues.issueType.uniqueness');
    case '一致性':
      return $t('page.qualityIssues.issueType.consistency');
    case '及时性':
      return $t('page.qualityIssues.issueType.timeliness');
    default:
      return ruleType;
  }
}

function formatDate(value?: string) {
  if (!value) return '-';
  return dayjs(value).isValid() ? dayjs(value).format('YYYY-MM-DD HH:mm:ss') : value;
}

async function fetchRules() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/dataquality/rules', {
      params: {
        ...queryForm,
        current: pagination.current,
        pageSize: pagination.pageSize,
      },
    });
    const payload = (response as any)?.data ?? response;
    const list = payload?.data?.list ?? payload?.list ?? [];
    const total = payload?.data?.total ?? payload?.total ?? 0;
    dataSource.value = Array.isArray(list) ? list : [];
    pagination.total = Number(total) || dataSource.value.length;
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityRules.message.fetchFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchRules();
}

function handleReset() {
  queryForm.ruleType = undefined;
  queryForm.enabled = undefined;
  pagination.current = 1;
  fetchRules();
}

function handleTableChange(page: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  fetchRules();
}

function resetFormModel() {
  formModel.id = undefined;
  formModel.ruleName = '';
  formModel.ruleType = '完整性';
  formModel.ruleDesc = '';
  formModel.ruleConfig = '';
  formModel.threshold = 80;
  formModel.severity = 'medium';
  formModel.enabled = 1;
  formModel.createdBy = 'admin';
}

function openCreate() {
  if (!checkPermission()) return;
  isEdit.value = false;
  resetFormModel();
  modalVisible.value = true;
}

function openEdit(record: RuleListItem) {
  if (!checkPermission()) return;
  isEdit.value = true;
  formModel.id = record.id;
  formModel.ruleName = record.ruleName;
  formModel.ruleType = record.ruleType;
  formModel.ruleDesc = record.ruleDesc;
  formModel.ruleConfig = record.ruleConfig;
  formModel.threshold = record.threshold;
  formModel.severity = record.severity;
  formModel.enabled = record.enabled;
  formModel.createdBy = record.createdBy || 'admin';
  modalVisible.value = true;
}

async function submitForm() {
  if (!checkPermission()) return;
  if (!formModel.ruleName) {
    message.warning($t('page.qualityRules.message.ruleNameRequired'));
    return;
  }
  saving.value = true;
  try {
    const payload = {
      ...formModel,
      enabled: Number(formModel.enabled) || 0,
      threshold: Number(formModel.threshold) || 0,
    };
    const response = isEdit.value
      ? await baseRequestClient.put('/v1/dataquality/rules', payload)
      : await baseRequestClient.post('/v1/dataquality/rules', payload);
    const result = (response as any)?.data ?? response;
    if (result?.code && result.code !== 200) {
      message.error(result?.msg || $t('page.qualityRules.message.saveFailed'));
      return;
    }
    message.success(isEdit.value ? $t('page.qualityRules.message.updateSuccess') : $t('page.qualityRules.message.createSuccess'));
    modalVisible.value = false;
    fetchRules();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityRules.message.saveFailed'));
  } finally {
    saving.value = false;
  }
}

async function handleDelete(id: number) {
  if (!checkPermission()) return;
  try {
    const response = await baseRequestClient.delete(`/v1/dataquality/rules/${id}`);
    const result = (response as any)?.data ?? response;
    if (result?.code && result.code !== 200) {
      message.error(result?.msg || $t('page.qualityRules.message.deleteFailed'));
      return;
    }
    message.success($t('page.qualityRules.message.deleteSuccess'));
    fetchRules();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityRules.message.deleteFailed'));
  }
}

async function handleToggleEnabled(record: RuleListItem, enabled: boolean) {
  if (!checkPermission()) return;
  try {
    const response = await baseRequestClient.put('/v1/dataquality/rules', {
      ...record,
      enabled: enabled ? 1 : 0,
      id: record.id,
    });
    const result = (response as any)?.data ?? response;
    if (result?.code && result.code !== 200) {
      message.error(result?.msg || $t('page.qualityRules.message.toggleFailed'));
      return;
    }
    message.success(enabled ? $t('page.qualityRules.message.enabledOn') : $t('page.qualityRules.message.enabledOff'));
    fetchRules();
  } catch (error: any) {
    message.error(error?.message || $t('page.qualityRules.message.toggleFailed'));
  }
}

onMounted(fetchRules);
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.qualityRules.title')">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item :label="$t('page.qualityRules.form.ruleType')" class="query-item">
            <Select v-model:value="queryForm.ruleType" allow-clear class="query-control">
              <Select.Option value="完整性">{{ $t('page.qualityIssues.issueType.completeness') }}</Select.Option>
              <Select.Option value="准确性">{{ $t('page.qualityIssues.issueType.accuracy') }}</Select.Option>
              <Select.Option value="唯一性">{{ $t('page.qualityIssues.issueType.uniqueness') }}</Select.Option>
              <Select.Option value="一致性">{{ $t('page.qualityIssues.issueType.consistency') }}</Select.Option>
              <Select.Option value="及时性">{{ $t('page.qualityIssues.issueType.timeliness') }}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item :label="$t('page.qualityRules.form.enabled')" class="query-item">
            <Select v-model:value="queryForm.enabled" allow-clear class="query-control">
              <Select.Option :value="1">{{ $t('page.qualityRules.enabled.on') }}</Select.Option>
              <Select.Option :value="0">{{ $t('page.qualityRules.enabled.off') }}</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="openCreate">{{ $t('page.qualityRules.action.newRule') }}</Button>
            <Button type="primary" ghost @click="handleSearch">{{ $t('page.common.search') }}</Button>
            <Button @click="handleReset">{{ $t('page.common.reset') }}</Button>
          </Space>
        </div>
      </Form>

      <Table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: RuleListItem) => record.id"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'ruleType'">
            {{ ruleTypeLabel(record.ruleType) }}
          </template>
          <template v-else-if="column.key === 'severity'">
            <Tag :color="severityTag(record.severity).color">{{ severityTag(record.severity).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'enabled'">
            <Switch
              :checked="record.enabled === 1"
              @change="(checked) => handleToggleEnabled(record, checked)"
            />
          </template>
          <template v-else-if="column.key === 'createdAt'">
            {{ formatDate(record.createdAt) }}
          </template>
          <template v-else-if="column.key === 'option'">
            <Space>
              <Button type="link" size="small" @click="openEdit(record)">{{ $t('page.common.edit') }}</Button>
              <Popconfirm :title="$t('page.qualityRules.confirmDelete')" @confirm="handleDelete(record.id)">
                <Button type="link" danger size="small">{{ $t('page.common.delete') }}</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalVisible"
      :title="isEdit ? $t('page.qualityRules.modal.editTitle') : $t('page.qualityRules.modal.createTitle')"
      :confirm-loading="saving"
      width="720px"
      @ok="submitForm"
    >
      <Form layout="vertical">
        <Form.Item :label="$t('page.qualityRules.form.ruleName')" required>
          <Input v-model:value="formModel.ruleName" :placeholder="$t('page.qualityRules.placeholder.ruleName')" />
        </Form.Item>
        <Form.Item :label="$t('page.qualityRules.form.ruleType')" required>
          <Select v-model:value="formModel.ruleType">
            <Select.Option value="完整性">{{ $t('page.qualityIssues.issueType.completeness') }}</Select.Option>
            <Select.Option value="准确性">{{ $t('page.qualityIssues.issueType.accuracy') }}</Select.Option>
            <Select.Option value="唯一性">{{ $t('page.qualityIssues.issueType.uniqueness') }}</Select.Option>
            <Select.Option value="一致性">{{ $t('page.qualityIssues.issueType.consistency') }}</Select.Option>
            <Select.Option value="及时性">{{ $t('page.qualityIssues.issueType.timeliness') }}</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item :label="$t('page.qualityRules.form.ruleDesc')">
          <Input.TextArea v-model:value="formModel.ruleDesc" :rows="3" />
        </Form.Item>
        <Form.Item :label="$t('page.qualityRules.form.ruleConfig')">
          <Input.TextArea
            v-model:value="formModel.ruleConfig"
            :rows="4"
            :placeholder="$t('page.qualityRules.placeholder.ruleConfig')"
          />
        </Form.Item>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <Form.Item :label="$t('page.qualityRules.form.threshold')">
            <InputNumber v-model:value="formModel.threshold" :max="100" :min="0" class="w-full" />
          </Form.Item>
          <Form.Item :label="$t('page.qualityRules.form.severity')">
            <Select v-model:value="formModel.severity">
              <Select.Option value="high">{{ $t('page.qualityIssues.level.high') }}</Select.Option>
              <Select.Option value="medium">{{ $t('page.qualityIssues.level.medium') }}</Select.Option>
              <Select.Option value="low">{{ $t('page.qualityIssues.level.low') }}</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item :label="$t('page.qualityRules.form.enabled')">
            <Select v-model:value="formModel.enabled">
              <Select.Option :value="1">{{ $t('page.qualityRules.enabled.on') }}</Select.Option>
              <Select.Option :value="0">{{ $t('page.qualityRules.enabled.off') }}</Select.Option>
            </Select>
          </Form.Item>
        </div>
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
