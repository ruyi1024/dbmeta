<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { FileTextOutlined } from '@ant-design/icons-vue';
import DOMPurify from 'dompurify';
import MarkdownIt from 'markdown-it';

import { Page } from '@vben/common-ui';

import { Button, Drawer, Empty, Input, Spin, Tag, Tooltip, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';

interface InsightReport {
  created_at?: string;
  id?: number;
  report_content?: string;
  task_name?: string;
}

function formatTime(v?: string) {
  if (!v) return '-';
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? v : d.toLocaleString('zh-CN', { hour12: false });
}

const loading = ref(false);
const reports = ref<InsightReport[]>([]);
const keyword = ref('');
const detailOpen = ref(false);
const currentReport = ref<InsightReport | null>(null);
const pager = reactive({
  page: 1,
  pageSize: 200,
});

const filteredReports = computed(() => {
  const k = keyword.value.trim();
  if (!k) return reports.value;
  return reports.value.filter((item) => String(item.task_name ?? '').includes(k));
});

async function fetchReports() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/task/analysis/logs', {
      params: {
        currentPage: pager.page,
        pageSize: pager.pageSize,
      },
    });
    const payload = (response as any)?.data ?? response;
    if (payload?.success === false) {
      message.error(String(payload?.msg ?? '加载报告失败'));
      reports.value = [];
      return;
    }
    const list = Array.isArray(payload?.data) ? (payload.data as InsightReport[]) : [];
    reports.value = list
      .filter((item) => String(item.report_content ?? '').trim().length > 0)
      .map((item) => ({
        created_at: item.created_at,
        id: item.id,
        report_content: item.report_content,
        task_name: item.task_name || '未命名洞察报告',
      }));
  } catch (e: unknown) {
    reports.value = [];
    message.error((e as Error)?.message || '加载报告失败');
  } finally {
    loading.value = false;
  }
}

function openDetail(report: InsightReport) {
  currentReport.value = report;
  detailOpen.value = true;
}

const md = new MarkdownIt({
  breaks: true,
  html: false,
  linkify: true,
  typographer: true,
});

const markdownHtml = computed(() => {
  const raw = currentReport.value?.report_content || '暂无内容';
  const rendered = md.render(raw);
  return DOMPurify.sanitize(rendered);
});

onMounted(() => {
  void fetchReports();
});
</script>

<template>
  <Page auto-content-height description="查看已生成的数据洞察报告，按书架样式浏览。">
    <div class="mb-4 flex items-center justify-between gap-3">
      <Input
        v-model:value="keyword"
        allow-clear
        class="max-w-[320px]"
        placeholder="按报告名称检索"
      />
      <Button @click="fetchReports">刷新</Button>
    </div>

    <Spin :spinning="loading">
      <Empty v-if="filteredReports.length === 0" description="暂无已生成的洞察报告" />
      <div v-else class="bookshelf">
        <div
          v-for="report in filteredReports"
          :key="report.id"
          class="book-card"
          @click="openDetail(report)"
        >
          <div class="book-content">
            <Tooltip :title="report.task_name">
              <div class="book-title-wrap">
                <FileTextOutlined class="book-title-icon" />
                <div class="book-title">{{ report.task_name }}</div>
              </div>
            </Tooltip>
            <div class="book-meta">
              <Tag color="blue">洞察报告</Tag>
              <span>{{ formatTime(report.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>
    </Spin>

    <Drawer
      v-model:open="detailOpen"
      :title="currentReport?.task_name || '洞察报告'"
      width="62%"
      placement="right"
      destroy-on-close
    >
      <div class="detail-date">生成时间：{{ formatTime(currentReport?.created_at) }}</div>
      <div class="detail-content markdown-body" v-html="markdownHtml" />
    </Drawer>
  </Page>
</template>

<style scoped>
.bookshelf {
  background:
    radial-gradient(circle at 20% 10%, rgb(59 130 246 / 8%), transparent 35%),
    radial-gradient(circle at 80% 90%, rgb(16 185 129 / 8%), transparent 35%),
    linear-gradient(180deg, hsl(var(--background)) 0%, hsl(var(--muted) / 35%) 100%);
  border: 1px solid hsl(var(--border));
  border-radius: 14px;
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  padding: 16px;
  position: relative;
}

.bookshelf::before {
  background: repeating-linear-gradient(
    90deg,
    transparent,
    transparent 22px,
    rgb(148 163 184 / 6%) 22px,
    rgb(148 163 184 / 6%) 23px
  );
  border-radius: 14px;
  content: '';
  inset: 0;
  pointer-events: none;
  position: absolute;
}

.bookshelf::after {
  background: linear-gradient(180deg, rgb(255 255 255 / 18%), transparent 24%, transparent 76%, rgb(15 23 42 / 8%));
  border-radius: 14px;
  content: '';
  inset: 0;
  pointer-events: none;
  position: absolute;
}

.book-card {
  background: linear-gradient(165deg, #1e293b 0%, #2a3447 52%, #243244 100%);
  border: 1px solid rgb(148 163 184 / 20%);
  border-radius: 12px;
  box-shadow: 0 8px 16px rgb(15 23 42 / 16%);
  cursor: pointer;
  display: flex;
  min-height: 280px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.book-card:hover {
  border-color: rgb(125 180 255 / 45%);
  box-shadow: 0 12px 20px rgb(15 23 42 / 22%);
  transform: translateY(-2px);
}

.book-content {
  display: flex;
  flex: 1;
  flex-direction: column;
  justify-content: space-between;
  padding: 14px 16px 12px;
}

.book-title {
  color: #e5e7eb;
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
  max-height: 64px;
  overflow: hidden;
}

.book-title-wrap {
  align-items: flex-start;
  display: flex;
  gap: 8px;
}

.book-title-icon {
  color: #60a5fa;
  font-size: 15px;
  margin-top: 2px;
}

.book-meta {
  align-items: center;
  color: #94a3b8;
  display: flex;
  font-size: 12px;
  gap: 8px;
}

.detail-date {
  color: #6b7280;
  font-size: 13px;
  margin-bottom: 12px;
}

.detail-content {
  background: #fafafa;
  border: 1px solid #eee;
  border-radius: 8px;
  max-height: 62vh;
  overflow: auto;
  padding: 14px;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  color: hsl(var(--foreground));
  margin: 10px 0;
}

.markdown-body :deep(p) {
  line-height: 1.8;
  margin: 10px 0;
}

.markdown-body :deep(ul) {
  margin: 10px 0;
  padding-left: 22px;
}

.markdown-body :deep(code) {
  background: hsl(var(--muted));
  border-radius: 4px;
  padding: 1px 6px;
}

.markdown-body :deep(pre) {
  background: #0f172a;
  border-radius: 8px;
  color: #e2e8f0;
  overflow: auto;
  padding: 12px;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
}
</style>
