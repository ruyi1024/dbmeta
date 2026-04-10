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
  { title: '规则名称', dataIndex: 'ruleName', key: 'ruleName', width: 200 },
  { title: '规则类型', dataIndex: 'ruleType', key: 'ruleType', width: 120 },
  { title: '规则描述', dataIndex: 'ruleDesc', key: 'ruleDesc' },
  {
    title: '阈值',
    dataIndex: 'threshold',
    key: 'threshold',
    width: 100,
    customRender: ({ record }) => `${record.threshold}%`,
  },
  { title: '严重程度', dataIndex: 'severity', key: 'severity', width: 100 },
  { title: '是否启用', dataIndex: 'enabled', key: 'enabled', width: 110 },
  { title: '创建人', dataIndex: 'createdBy', key: 'createdBy', width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '操作', key: 'option', width: 160, fixed: 'right' },
];

function severityTag(severity: string) {
  if (severity === 'high') return { color: 'red', text: '高' };
  if (severity === 'medium') return { color: 'orange', text: '中' };
  return { color: 'blue', text: '低' };
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
    message.error(error?.message || '获取规则失败');
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
  isEdit.value = false;
  resetFormModel();
  modalVisible.value = true;
}

function openEdit(record: RuleListItem) {
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
  if (!formModel.ruleName) {
    message.warning('请输入规则名称');
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
      message.error(result?.msg || '保存失败');
      return;
    }
    message.success(isEdit.value ? '更新成功' : '创建成功');
    modalVisible.value = false;
    fetchRules();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function handleDelete(id: number) {
  try {
    const response = await baseRequestClient.delete(`/v1/dataquality/rules/${id}`);
    const result = (response as any)?.data ?? response;
    if (result?.code && result.code !== 200) {
      message.error(result?.msg || '删除失败');
      return;
    }
    message.success('删除成功');
    fetchRules();
  } catch (error: any) {
    message.error(error?.message || '删除失败');
  }
}

async function handleToggleEnabled(record: RuleListItem, enabled: boolean) {
  try {
    const response = await baseRequestClient.put('/v1/dataquality/rules', {
      ...record,
      enabled: enabled ? 1 : 0,
      id: record.id,
    });
    const result = (response as any)?.data ?? response;
    if (result?.code && result.code !== 200) {
      message.error(result?.msg || '操作失败');
      return;
    }
    message.success(enabled ? '已启用' : '已禁用');
    fetchRules();
  } catch (error: any) {
    message.error(error?.message || '操作失败');
  }
}

onMounted(fetchRules);
</script>

<template>
  <div class="p-5">
    <Card title="质量规则配置">
      <Form class="mb-4">
        <div class="query-grid">
          <Form.Item label="规则类型" class="query-item">
            <Select v-model:value="queryForm.ruleType" allow-clear class="query-control">
              <Select.Option value="完整性">完整性</Select.Option>
              <Select.Option value="准确性">准确性</Select.Option>
              <Select.Option value="唯一性">唯一性</Select.Option>
              <Select.Option value="一致性">一致性</Select.Option>
              <Select.Option value="及时性">及时性</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="是否启用" class="query-item">
            <Select v-model:value="queryForm.enabled" allow-clear class="query-control">
              <Select.Option :value="1">启用</Select.Option>
              <Select.Option :value="0">禁用</Select.Option>
            </Select>
          </Form.Item>
        </div>
        <div class="query-actions">
          <Space>
            <Button type="primary" @click="openCreate">新建规则</Button>
            <Button type="primary" ghost @click="handleSearch">查询</Button>
            <Button @click="handleReset">重置</Button>
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
          <template v-if="column.key === 'severity'">
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
              <Button type="link" size="small" @click="openEdit(record)">编辑</Button>
              <Popconfirm title="确定要删除这条规则吗？" @confirm="handleDelete(record.id)">
                <Button type="link" danger size="small">删除</Button>
              </Popconfirm>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="modalVisible"
      :title="isEdit ? '编辑规则' : '新建规则'"
      :confirm-loading="saving"
      width="720px"
      @ok="submitForm"
    >
      <Form layout="vertical">
        <Form.Item label="规则名称" required>
          <Input v-model:value="formModel.ruleName" placeholder="请输入规则名称" />
        </Form.Item>
        <Form.Item label="规则类型" required>
          <Select v-model:value="formModel.ruleType">
            <Select.Option value="完整性">完整性</Select.Option>
            <Select.Option value="准确性">准确性</Select.Option>
            <Select.Option value="唯一性">唯一性</Select.Option>
            <Select.Option value="一致性">一致性</Select.Option>
            <Select.Option value="及时性">及时性</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="规则描述">
          <Input.TextArea v-model:value="formModel.ruleDesc" :rows="3" />
        </Form.Item>
        <Form.Item label="规则配置">
          <Input.TextArea
            v-model:value="formModel.ruleConfig"
            :rows="4"
            placeholder='JSON格式，如: {"field":"email","pattern":"email"}'
          />
        </Form.Item>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <Form.Item label="阈值(%)">
            <InputNumber v-model:value="formModel.threshold" :max="100" :min="0" class="w-full" />
          </Form.Item>
          <Form.Item label="严重程度">
            <Select v-model:value="formModel.severity">
              <Select.Option value="high">高</Select.Option>
              <Select.Option value="medium">中</Select.Option>
              <Select.Option value="low">低</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="是否启用">
            <Select v-model:value="formModel.enabled">
              <Select.Option :value="1">启用</Select.Option>
              <Select.Option :value="0">禁用</Select.Option>
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
