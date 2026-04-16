<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { FileTextOutlined } from '@ant-design/icons-vue';
import DOMPurify from 'dompurify';
import MarkdownIt from 'markdown-it';

import { Page } from '@vben/common-ui';

import { Button, Drawer, Empty, Input, Spin, Tag, Tooltip, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

interface InsightReport {
  created_at?: string;
  /** 列表卡片摘要，由 report_content 生成 */
  description_preview?: string;
  id?: number;
  report_content?: string;
  task_name?: string;
}

/** public/report_bg 下的封面图，按报告稳定映射 */
const REPORT_BG_IMAGES = [
  '/report_bg/bg_01.png',
  '/report_bg/bg_02.png',
  '/report_bg/bg_03.png',
  '/report_bg/bg_04.png',
  '/report_bg/bg_05.png',
  '/report_bg/bg_06.png',
] as const;

function hashString(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = (Math.imul(31, h) + s.charCodeAt(i)) | 0;
  return h;
}

function reportBgUrl(report: InsightReport): string {
  const n = report.id ?? hashString(String(report.task_name ?? ''));
  const idx = Math.abs(n) % REPORT_BG_IMAGES.length;
  return REPORT_BG_IMAGES[idx];
}

/** 从 Markdown 正文提取卡片用摘要（标题下方展示） */
function plainTextPreview(raw: string, maxLen: number): string {
  const t = String(raw ?? '').trim();
  if (!t) return $t('page.insight.previewEmpty');
  let s = t
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/\s+/g, ' ')
    .trim();
  if (!s) return $t('page.insight.previewEmpty');
  if (s.length <= maxLen) return s;
  return `${s.slice(0, maxLen).replace(/\s+\S*$/, '')}…`;
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
  return reports.value.filter((item) => {
    const name = String(item.task_name ?? '');
    const desc = String(item.description_preview ?? plainTextPreview(String(item.report_content ?? ''), 500));
    return name.includes(k) || desc.includes(k);
  });
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
      message.error(String(payload?.msg ?? $t('page.insight.message.loadFailed')));
      reports.value = [];
      return;
    }
    const list = Array.isArray(payload?.data) ? (payload.data as InsightReport[]) : [];
    reports.value = list
      .filter((item) => String(item.report_content ?? '').trim().length > 0)
      .map((item) => {
        const report_content = item.report_content;
        return {
          created_at: item.created_at,
          description_preview: plainTextPreview(String(report_content ?? ''), 120),
          id: item.id,
          report_content,
          task_name: item.task_name || $t('page.insight.unnamedReport'),
        };
      });
  } catch (e: unknown) {
    reports.value = [];
    message.error((e as Error)?.message || $t('page.insight.message.loadFailed'));
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
  const raw = currentReport.value?.report_content || $t('page.insight.noContent');
  const rendered = md.render(raw);
  return DOMPurify.sanitize(rendered);
});

onMounted(() => {
  void fetchReports();
});
</script>

<template>
  <Page auto-content-height :description="$t('page.insight.description')">
    <div class="mb-4 flex items-center justify-between gap-3">
      <Input
        v-model:value="keyword"
        allow-clear
        class="max-w-[320px]"
        :placeholder="$t('page.insight.searchPlaceholder')"
      />
      <Button @click="fetchReports">{{ $t('page.common.refresh') }}</Button>
    </div>

    <Spin :spinning="loading">
      <Empty
        v-if="filteredReports.length === 0"
        :description="$t('page.insight.empty')"
      />
      <div v-else class="bookshelf">
        <div
          v-for="report in filteredReports"
          :key="report.id"
          class="book-card"
          :style="{ backgroundImage: `url(${reportBgUrl(report)})` }"
          @click="openDetail(report)"
        >
          <div class="book-content">
            <div class="book-text-block">
              <div class="book-title-row">
                <FileTextOutlined class="book-title-icon" />
                <Tooltip :title="report.task_name">
                  <div class="book-title">{{ report.task_name }}</div>
                </Tooltip>
              </div>
              <Tooltip :title="report.description_preview">
                <div class="book-desc">{{ report.description_preview }}</div>
              </Tooltip>
            </div>
            <div class="book-meta">
              <Tag color="blue">{{ $t('page.insight.reportTag') }}</Tag>
              <span>{{ formatTime(report.created_at) }}</span>
            </div>
          </div>
        </div>
      </div>
    </Spin>

    <Drawer
      v-model:open="detailOpen"
      :title="currentReport?.task_name || $t('page.insight.reportTag')"
      width="62%"
      placement="right"
      destroy-on-close
    >
      <div class="detail-date">
        {{ $t('page.insight.generatedAt') }}{{ formatTime(currentReport?.created_at) }}
      </div>
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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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
  background-color: #1e293b;
  background-position: center;
  background-repeat: no-repeat;
  background-size: cover;
  border: 1px solid rgb(148 163 184 / 20%);
  border-radius: 12px;
  box-shadow: 0 8px 16px rgb(15 23 42 / 16%);
  cursor: pointer;
  display: flex;
  min-height: 280px;
  overflow: hidden;
  position: relative;
  transition: all 0.2s ease;
}

.book-card::before {
  background: linear-gradient(
    165deg,
    rgb(15 23 42 / 58%) 0%,
    rgb(15 23 42 / 72%) 52%,
    rgb(15 23 42 / 78%) 100%
  );
  border-radius: inherit;
  content: '';
  inset: 0;
  pointer-events: none;
  position: absolute;
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
  position: relative;
  z-index: 1;
}

.book-text-block {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 10px;
  min-height: 0;
}

.book-title-row {
  align-items: flex-start;
  display: flex;
  gap: 8px;
}

.book-title {
  color: #f1f5f9;
  flex: 1;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.35;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.book-desc {
  color: rgb(203 213 225 / 92%);
  display: -webkit-box;
  font-size: 13px;
  font-weight: 400;
  -webkit-line-clamp: 4;
  line-clamp: 4;
  -webkit-box-orient: vertical;
  line-height: 1.55;
  overflow: hidden;
  padding-left: 23px;
  text-shadow: 0 1px 2px rgb(0 0 0 / 35%);
}

.book-title-icon {
  color: #60a5fa;
  flex-shrink: 0;
  font-size: 18px;
  margin-top: 3px;
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
