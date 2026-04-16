<script lang="ts" setup>
import type { Dayjs } from 'dayjs';

import {
  Button,
  Card,
  DatePicker,
  Drawer,
  Select,
  Space,
  Table,
  Tag,
  type TableColumnsType,
} from 'ant-design-vue';
import { message } from 'ant-design-vue';
import { onMounted, reactive, ref } from 'vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

interface AlarmOption {
  alarm_name: string;
  id: number;
}

interface PatrolReportRow {
  alarm_id: number;
  alarm_name: string;
  complete_time?: string;
  created_at: string;
  data_count: number;
  email_sent: boolean;
  error_message?: string;
  id: number;
  rule_matched: boolean;
  start_time: string;
  status: string;
}

const loading = ref(false);
const reports = ref<PatrolReportRow[]>([]);
const alarmOptions = ref<AlarmOption[]>([]);

const queryForm = reactive({
  alarmId: undefined as number | undefined,
  endDate: undefined as Dayjs | undefined,
  startDate: undefined as Dayjs | undefined,
  status: undefined as string | undefined,
});

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

const reportDrawerOpen = ref(false);
const reportDetailLoading = ref(false);
const currentReportHtml = ref('');
const currentReportMeta = ref<Record<string, any> | null>(null);

const columns: TableColumnsType<PatrolReportRow> = [
  { title: $t('page.patrolReport.columns.id'), dataIndex: 'id', key: 'id', width: 88 },
  { title: $t('page.patrolReport.columns.alarmName'), dataIndex: 'alarm_name', key: 'alarm_name', ellipsis: true, width: 180 },
  { title: $t('page.patrolReport.columns.status'), dataIndex: 'status', key: 'status', width: 120 },
  { title: $t('page.patrolReport.columns.dataCount'), dataIndex: 'data_count', key: 'data_count', width: 100 },
  { title: $t('page.patrolReport.columns.ruleMatched'), dataIndex: 'rule_matched', key: 'rule_matched', width: 110 },
  { title: $t('page.patrolReport.columns.emailSent'), dataIndex: 'email_sent', key: 'email_sent', width: 110 },
  { title: $t('page.patrolReport.columns.startTime'), dataIndex: 'start_time', key: 'start_time', width: 180 },
  { title: $t('page.patrolReport.columns.completeTime'), dataIndex: 'complete_time', key: 'complete_time', width: 180 },
  { title: $t('page.patrolReport.columns.errorMessage'), dataIndex: 'error_message', key: 'error_message', ellipsis: true },
  { title: $t('page.patrolReport.columns.action'), key: 'action', fixed: 'right', width: 110 },
];

function statusTag(status: string) {
  if (status === 'triggered') return { color: 'orange', text: $t('page.patrolReport.status.triggered') };
  if (status === 'success') return { color: 'green', text: $t('page.patrolReport.status.success') };
  if (status === 'failed') return { color: 'red', text: $t('page.patrolReport.status.failed') };
  return { color: 'blue', text: $t('page.patrolReport.status.running') };
}

function boolTag(v: boolean, yesText: string, noText: string) {
  return v
    ? { color: 'green', text: yesText }
    : { color: 'default', text: noText };
}

async function fetchAlarmOptions() {
  try {
    const response = await baseRequestClient.get('/v1/data/alarm/list', {
      params: {
        currentPage: 1,
        pageSize: 200,
      },
    });
    const payload = (response as any)?.data ?? response;
    const list = payload?.data ?? [];
    alarmOptions.value = Array.isArray(list)
      ? list.map((item: any) => ({
          id: Number(item.id) || 0,
          alarm_name: String(item.alarm_name || ''),
        }))
      : [];
  } catch {
    alarmOptions.value = [];
  }
}

async function fetchReports() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/data/alarm/logs', {
      params: {
        alarm_id: queryForm.alarmId,
        currentPage: pagination.current,
        end_date: queryForm.endDate?.format('YYYY-MM-DD'),
        pageSize: pagination.pageSize,
        start_date: queryForm.startDate?.format('YYYY-MM-DD'),
        status: queryForm.status,
      },
    });
    const payload = (response as any)?.data ?? response;
    const list = payload?.data ?? [];
    reports.value = Array.isArray(list)
      ? list.map((item: any) => ({
          alarm_id: Number(item.alarm_id) || 0,
          alarm_name: String(item.alarm_name || ''),
          complete_time: item.complete_time || '',
          created_at: item.created_at || '',
          data_count: Number(item.data_count) || 0,
          email_sent: Boolean(item.email_sent),
          error_message: item.error_message || '',
          id: Number(item.id) || 0,
          rule_matched: Boolean(item.rule_matched),
          start_time: item.start_time || '',
          status: String(item.status || ''),
        }))
      : [];
    pagination.total = Number(payload?.total) || reports.value.length;
  } catch (error: any) {
    message.error(error?.message || $t('page.patrolReport.message.queryFailed'));
  } finally {
    loading.value = false;
  }
}

function onSearch() {
  pagination.current = 1;
  fetchReports();
}

function onReset() {
  queryForm.alarmId = undefined;
  queryForm.status = undefined;
  queryForm.startDate = undefined;
  queryForm.endDate = undefined;
  pagination.current = 1;
  fetchReports();
}

function onTableChange(page: any) {
  pagination.current = page.current;
  pagination.pageSize = page.pageSize;
  fetchReports();
}

async function viewReport(record: PatrolReportRow) {
  reportDrawerOpen.value = true;
  reportDetailLoading.value = true;
  currentReportHtml.value = '';
  currentReportMeta.value = null;
  try {
    const response = await baseRequestClient.get(`/v1/data/alarm/report/${record.id}`);
    const payload = (response as any)?.data ?? response;
    const reportData = payload?.data ?? {};
    currentReportMeta.value = reportData;
    currentReportHtml.value = reportData?.report_html || `<html><body><p>${$t('page.patrolReport.message.emptyReport')}</p></body></html>`;
  } catch (error: any) {
    message.error(error?.message || $t('page.patrolReport.message.loadFailed'));
    currentReportHtml.value = `<html><body><p>${$t('page.patrolReport.message.loadFailed')}</p></body></html>`;
  } finally {
    reportDetailLoading.value = false;
  }
}

onMounted(async () => {
  await fetchAlarmOptions();
  await fetchReports();
});
</script>

<template>
  <div class="p-5">
    <Card :title="$t('page.patrolReport.title')">
      <div class="mb-4 flex flex-wrap items-center gap-3">
        <Select
          v-model:value="queryForm.alarmId"
          allow-clear
          class="w-[220px]"
          :placeholder="$t('page.patrolReport.placeholder.alarm')"
        >
          <Select.Option
            v-for="item in alarmOptions"
            :key="item.id"
            :value="item.id"
          >
            {{ item.alarm_name }}
          </Select.Option>
        </Select>

        <Select
          v-model:value="queryForm.status"
          allow-clear
          class="w-[160px]"
          :placeholder="$t('page.patrolReport.placeholder.status')"
        >
          <Select.Option value="running">{{ $t('page.patrolReport.status.running') }}</Select.Option>
          <Select.Option value="success">{{ $t('page.patrolReport.status.success') }}</Select.Option>
          <Select.Option value="triggered">{{ $t('page.patrolReport.status.triggered') }}</Select.Option>
          <Select.Option value="failed">{{ $t('page.patrolReport.status.failed') }}</Select.Option>
        </Select>

        <DatePicker
          v-model:value="queryForm.startDate"
          allow-clear
          class="w-[180px]"
          :placeholder="$t('page.patrolReport.placeholder.startDate')"
        />
        <DatePicker
          v-model:value="queryForm.endDate"
          allow-clear
          class="w-[180px]"
          :placeholder="$t('page.patrolReport.placeholder.endDate')"
        />

        <Space>
          <Button type="primary" @click="onSearch">{{ $t('page.common.search') }}</Button>
          <Button @click="onReset">{{ $t('page.common.reset') }}</Button>
        </Space>
      </div>

      <Table
        :columns="columns"
        :data-source="reports"
        :loading="loading"
        :pagination="pagination"
        :row-key="(record: PatrolReportRow) => record.id"
        bordered
        size="middle"
        @change="onTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <Tag :color="statusTag(record.status).color">{{ statusTag(record.status).text }}</Tag>
          </template>
          <template v-else-if="column.key === 'rule_matched'">
            <Tag :color="boolTag(record.rule_matched, $t('page.patrolReport.match.hit'), $t('page.patrolReport.match.miss')).color">
              {{ boolTag(record.rule_matched, $t('page.patrolReport.match.hit'), $t('page.patrolReport.match.miss')).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'email_sent'">
            <Tag :color="boolTag(record.email_sent, $t('page.patrolReport.email.sent'), $t('page.patrolReport.email.unsent')).color">
              {{ boolTag(record.email_sent, $t('page.patrolReport.email.sent'), $t('page.patrolReport.email.unsent')).text }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'complete_time'">
            {{ record.complete_time || '-' }}
          </template>
          <template v-else-if="column.key === 'error_message'">
            <span class="inline-block max-w-[280px] truncate">{{ record.error_message || '-' }}</span>
          </template>
          <template v-else-if="column.key === 'action'">
            <Button type="link" size="small" @click="viewReport(record)">{{ $t('page.patrolReport.action.viewOnline') }}</Button>
          </template>
        </template>
      </Table>
    </Card>

    <Drawer
      v-model:open="reportDrawerOpen"
      :title="`${$t('page.patrolReport.drawer.title')} #${currentReportMeta?.id || ''}`"
      :width="960"
      destroy-on-close
    >
      <div class="mb-3 rounded bg-muted px-3 py-2 text-xs text-muted-foreground">
        <span class="mr-4">{{ $t('page.patrolReport.drawer.task') }}: {{ currentReportMeta?.alarm_name || '-' }}</span>
        <span class="mr-4">{{ $t('page.patrolReport.drawer.status') }}: {{ currentReportMeta?.status || '-' }}</span>
        <span class="mr-4">{{ $t('page.patrolReport.drawer.start') }}: {{ currentReportMeta?.start_time || '-' }}</span>
        <span>{{ $t('page.patrolReport.drawer.complete') }}: {{ currentReportMeta?.complete_time || '-' }}</span>
      </div>
      <div v-if="reportDetailLoading" class="py-10 text-center text-muted-foreground">{{ $t('page.patrolReport.drawer.loading') }}</div>
      <iframe
        v-else
        :srcdoc="currentReportHtml"
        class="h-[72vh] w-full rounded border border-border"
        sandbox="allow-same-origin"
      />
    </Drawer>
  </div>
</template>
