<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from 'ant-design-vue';
import type { TableColumnsType } from 'ant-design-vue';

import { Page } from '@vben/common-ui';

import { baseRequestClient } from '#/api/request';

defineOptions({ name: 'GradingGradeDict' });

interface GradeRow {
  id: number;
  gradeCode: string;
  gradeName: string;
  levelOrder: number;
  description?: string;
  standardRef?: string;
  enable: number;
  gmtCreated?: string;
  gmtUpdated?: string;
}

const loading = ref(false);
const dataSource = ref<GradeRow[]>([]);

const editOpen = ref(false);
const saving = ref(false);
const editForm = reactive({
  id: 0,
  description: '',
  standardRef: '',
  enable: 1 as number,
});

function parseGrades(response: unknown): GradeRow[] {
  const body = (response as any)?.data ?? response;
  const raw = body?.data ?? body;
  return Array.isArray(raw) ? (raw as GradeRow[]) : [];
}

async function fetchGrades() {
  loading.value = true;
  try {
    const res = await baseRequestClient.get('/v1/grading/grades');
    dataSource.value = parseGrades(res);
  } catch (e: unknown) {
    message.error((e as Error)?.message || '加载分级字典失败');
    dataSource.value = [];
  } finally {
    loading.value = false;
  }
}

function openEdit(record: GradeRow) {
  editForm.id = record.id;
  editForm.description = record.description ?? '';
  editForm.standardRef = record.standardRef ?? '';
  editForm.enable = Number(record.enable) === 1 ? 1 : 0;
  editOpen.value = true;
}

async function saveEdit() {
  saving.value = true;
  try {
    await baseRequestClient.put('/v1/grading/grades', {
      id: editForm.id,
      description: editForm.description,
      standardRef: editForm.standardRef,
      enable: editForm.enable,
    });
    message.success('已保存');
    editOpen.value = false;
    await fetchGrades();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

async function onEnableChange(record: GradeRow, checked: boolean) {
  const next = checked ? 1 : 0;
  try {
    await baseRequestClient.put('/v1/grading/grades', {
      id: record.id,
      enable: next,
      description: record.description ?? '',
      standardRef: record.standardRef ?? '',
    });
    message.success('已更新');
    await fetchGrades();
  } catch (e: unknown) {
    message.error((e as Error)?.message || '更新失败');
    await fetchGrades();
  }
}

const columns: TableColumnsType<GradeRow> = [
  { title: '编码', dataIndex: 'gradeCode', key: 'gradeCode', width: 120 },
  { title: '名称', dataIndex: 'gradeName', key: 'gradeName', width: 120 },
  { title: '级别序', dataIndex: 'levelOrder', key: 'levelOrder', width: 88 },
  { title: '说明', dataIndex: 'description', key: 'description', ellipsis: true },
  { title: '依据说明', dataIndex: 'standardRef', key: 'standardRef', width: 160, ellipsis: true },
  { title: '启用', key: 'enable', width: 100 },
  { title: '操作', key: 'action', width: 100, fixed: 'right' },
];

onMounted(() => {
  void fetchGrades();
});
</script>

<template>
  <Page auto-content-height description="维护 GB 通用三类分级（一般 / 重要 / 核心）字典，可调整启用状态与说明。">
    <Card title="分级字典">
      <Space class="mb-3">
        <Button @click="fetchGrades">刷新</Button>
      </Space>
      <Table
        row-key="id"
        :loading="loading"
        :columns="columns"
        :data-source="dataSource"
        :pagination="false"
        bordered
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'enable'">
            <Switch
              :checked="Number(record.enable) === 1"
              @change="(c: any) => onEnableChange(record as GradeRow, !!c)"
            />
          </template>
          <template v-else-if="column.key === 'action'">
            <Button size="small" type="link" @click="openEdit(record as GradeRow)">编辑</Button>
          </template>
          <template v-else-if="column.key === 'gradeName'">
            <Tag color="blue">{{ record.gradeName }}</Tag>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="editOpen"
      title="编辑分级说明"
      :confirm-loading="saving"
      destroy-on-close
      @ok="saveEdit"
    >
      <Form layout="vertical">
        <Form.Item label="说明">
          <Input.TextArea v-model:value="editForm.description" :rows="4" placeholder="分级含义说明" />
        </Form.Item>
        <Form.Item label="依据说明">
          <Input v-model:value="editForm.standardRef" placeholder="标准/版本等可追溯信息" />
        </Form.Item>
        <Form.Item label="启用">
          <Switch :checked="editForm.enable === 1" @change="(c: any) => (editForm.enable = c ? 1 : 0)" />
        </Form.Item>
      </Form>
    </Modal>
  </Page>
</template>
