<script lang="ts" setup>
import type {
  WorkbenchProjectItem,
  WorkbenchQuickNavItem,
  WorkbenchTodoItem,
} from '@vben/common-ui';
import type { EchartsUIType } from '@vben/plugins/echarts';

import { computed, onActivated, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import dayjs from 'dayjs';
import {
  AnalysisChartCard,
  WorkbenchHeader,
  WorkbenchProject,
  WorkbenchQuickNav,
  WorkbenchTodo,
} from '@vben/common-ui';
import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

const userStore = useUserStore();
const now = dayjs();
const CACHE_TTL_MS = 60 * 1000;
const SUMMARY_CACHE_KEY = 'workspace-summary-cache';

interface WorkspaceSummary {
  totalDatabases: number;
  totalDataSize: string;
  totalRows: number;
  totalTables: number;
}

interface TrendPoint {
  x?: string;
  y?: number;
}

interface CacheWrapper<T> {
  data: T;
  expiresAt: number;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') return {};
  const r = response as Record<string, unknown>;
  if ('data' in r && r.data !== undefined && typeof r.data === 'object' && 'status' in r) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r;
}

const summary = reactive<WorkspaceSummary>({
  totalDatabases: 0,
  totalDataSize: '0 B',
  totalRows: 0,
  totalTables: 0,
});
const task24hTotal = ref(0);
const task24hTrend = ref<TrendPoint[]>([]);
const taskChartLoading = ref(true);
let summaryDelayTimer: ReturnType<typeof setTimeout> | undefined;
let taskChartDelayTimer: ReturnType<typeof setTimeout> | undefined;

const task24hChartRef = ref<EchartsUIType>();
const { renderEcharts: renderTask24hChart } = useEcharts(task24hChartRef);

const headerStats = computed(() => [
  {
    label: $t('page.capacity.dashboard.stat.databaseCount'),
    value: summary.totalDatabases.toLocaleString(),
  },
  {
    label: $t('page.capacity.dashboard.stat.tableCount'),
    value: summary.totalTables.toLocaleString(),
  },
  {
    label: $t('page.capacity.dashboard.stat.totalDataSize'),
    value: summary.totalDataSize || '0 B',
  },
  {
    label: $t('page.capacity.dashboard.stat.totalRows'),
    value: summary.totalRows.toLocaleString(),
  },
]);

const greeting = computed(() => {
  const hour = dayjs().hour();
  if (hour < 12) return $t('page.workspace.greeting.morning');
  if (hour < 18) return $t('page.workspace.greeting.afternoon');
  return $t('page.workspace.greeting.evening');
});

const headerDescription = computed(() => {
  return `${now.format('YYYY-MM-DD')} · ${$t('page.workspace.description.focus')}`;
});

const projectItems: WorkbenchProjectItem[] = [
  {
    color: '#6366f1',
    content: $t('page.workspace.cards.metaDashboard.desc'),
    date: now.format('YYYY-MM-DD'),
    group: $t('page.workspace.cards.group.core'),
    icon: 'lucide:database',
    title: $t('page.workspace.cards.metaDashboard.title'),
    url: '/meta/dashboard',
  },
  {
    color: '#10b981',
    content: $t('page.workspace.cards.qualityRules.desc'),
    date: now.format('YYYY-MM-DD'),
    group: $t('page.workspace.cards.group.quality'),
    icon: 'lucide:shield-check',
    title: $t('page.workspace.cards.qualityRules.title'),
    url: '/quality/rules',
  },
  {
    color: '#f59e0b',
    content: $t('page.workspace.cards.taskPlan.desc'),
    date: now.format('YYYY-MM-DD'),
    group: $t('page.workspace.cards.group.ops'),
    icon: 'lucide:list-checks',
    title: $t('page.workspace.cards.taskPlan.title'),
    url: '/task',
  },
  {
    color: '#ef4444',
    content: $t('page.workspace.cards.notice.desc'),
    date: now.format('YYYY-MM-DD'),
    group: $t('page.workspace.cards.group.config'),
    icon: 'lucide:messages-square',
    title: $t('page.workspace.cards.notice.title'),
    url: '/setting/notice',
  },
  {
    color: '#14b8a6',
    content: $t('page.workspace.cards.datasource.desc'),
    date: now.format('YYYY-MM-DD'),
    group: $t('page.workspace.cards.group.config'),
    icon: 'lucide:server-cog',
    title: $t('page.workspace.cards.datasource.title'),
    url: '/setting/datasource',
  },
  {
    color: '#3b82f6',
    content: $t('page.workspace.cards.aiModels.desc'),
    date: now.format('YYYY-MM-DD'),
    group: $t('page.workspace.cards.group.ai'),
    icon: 'lucide:bot',
    title: $t('page.workspace.cards.aiModels.title'),
    url: '/setting/ai_models',
  },
];

const quickNavItems: WorkbenchQuickNavItem[] = [
  {
    color: '#1fdaca',
    icon: 'lucide:layout-dashboard',
    title: $t('page.workspace.quickNav.databaseQuery'),
    url: '/capacity/database-query',
  },
  {
    color: '#bf0c2c',
    icon: 'lucide:scale',
    title: $t('page.workspace.quickNav.capacity'),
    url: '/capacity/dashboard',
  },
  {
    color: '#e18525',
    icon: 'lucide:badge-check',
    title: $t('page.workspace.quickNav.quality'),
    url: '/quality/dashboard',
  },
  {
    color: '#3fb27f',
    icon: 'lucide:search-check',
    title: $t('page.workspace.quickNav.query'),
    url: '/query',
  },
  {
    color: '#4daf1bc9',
    icon: 'lucide:users',
    title: $t('page.workspace.quickNav.users'),
    url: '/users/manager',
  },
  {
    color: '#00d8ff',
    icon: 'lucide:settings',
    title: $t('page.workspace.quickNav.env'),
    url: '/setting/env',
  },
];

const todoItems = ref<WorkbenchTodoItem[]>([
  {
    completed: false,
    content: $t('page.workspace.todo.datasourceCheck.content'),
    date: now.format('YYYY-MM-DD HH:mm:ss'),
    title: $t('page.workspace.todo.datasourceCheck.title'),
  },
  {
    completed: false,
    content: $t('page.workspace.todo.qualityTask.content'),
    date: now.format('YYYY-MM-DD HH:mm:ss'),
    title: $t('page.workspace.todo.qualityTask.title'),
  },
  {
    completed: false,
    content: $t('page.workspace.todo.aiModel.content'),
    date: now.format('YYYY-MM-DD HH:mm:ss'),
    title: $t('page.workspace.todo.aiModel.title'),
  },
  {
    completed: false,
    content: $t('page.workspace.todo.notice.content'),
    date: now.format('YYYY-MM-DD HH:mm:ss'),
    title: $t('page.workspace.todo.notice.title'),
  },
]);

const router = useRouter();

function navTo(nav: WorkbenchProjectItem | WorkbenchQuickNavItem) {
  if (nav.url?.startsWith('http')) {
    openWindow(nav.url);
    return;
  }
  if (nav.url?.startsWith('/')) {
    router.push(nav.url).catch((error) => {
      console.error('Navigation failed:', error);
    });
  } else {
    console.warn(`Unknown URL for navigation item: ${nav.title} -> ${nav.url}`);
  }
}

function loadCache<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return null;
    const wrapped = JSON.parse(raw) as CacheWrapper<T>;
    if (!wrapped?.expiresAt || Date.now() > wrapped.expiresAt) return null;
    return wrapped.data;
  } catch {
    return null;
  }
}

function saveCache<T>(key: string, data: T) {
  try {
    const wrapped: CacheWrapper<T> = {
      data,
      expiresAt: Date.now() + CACHE_TTL_MS,
    };
    localStorage.setItem(key, JSON.stringify(wrapped));
  } catch {
    // ignore cache failure
  }
}

async function hydrateFromCache() {
  const summaryCached = loadCache<WorkspaceSummary>(SUMMARY_CACHE_KEY);
  if (summaryCached) {
    Object.assign(summary, summaryCached);
  }
}

async function fetchSummaryStats() {
  try {
    const response = await baseRequestClient.get('/v1/pumpkin/capacity/stats');
    const body = extractApiBody(response);
    const data = (body.data ?? body) as Record<string, unknown>;
    summary.totalDatabases = Number(data.totalDatabases) || 0;
    summary.totalTables = Number(data.totalTables) || 0;
    summary.totalDataSize = String(data.totalDataSize ?? '0 B');
    summary.totalRows = Number(data.totalRows) || 0;
    saveCache(SUMMARY_CACHE_KEY, {
      totalDatabases: summary.totalDatabases,
      totalTables: summary.totalTables,
      totalDataSize: summary.totalDataSize,
      totalRows: summary.totalRows,
    });
  } catch {
    summary.totalDatabases = 0;
    summary.totalTables = 0;
    summary.totalDataSize = '0 B';
    summary.totalRows = 0;
  }
}

function renderTask24hOverview() {
  const xs = task24hTrend.value.map((i) => String(i.x ?? ''));
  const ys = task24hTrend.value.map((i) => Number(i.y) || 0);

  renderTask24hChart({
    grid: { bottom: 32, left: 36, right: 16, top: 18 },
    series: [
      {
        areaStyle: { color: 'rgba(22,119,255,0.12)' },
        data: ys,
        lineStyle: { color: '#1677ff', width: 2 },
        smooth: true,
        symbol: 'none',
        type: 'line',
      },
    ],
    tooltip: { trigger: 'axis' },
    xAxis: {
      axisLabel: {
        color: '#8c8c8c',
        fontSize: 10,
        formatter: (v: string) => (v.includes(' ') ? (v.split(' ')[1] ?? v) : v),
      },
      axisTick: { show: false },
      data: xs,
      type: 'category',
    },
    yAxis: {
      minInterval: 1,
      splitLine: { lineStyle: { color: '#f0f0f0' } },
      type: 'value',
    },
  });
}

async function fetchTask24hStats() {
  try {
    const response = await baseRequestClient.get('/v1/task/today/stats');
    const body = extractApiBody(response);
    task24hTotal.value = Number(body.hour24Total) || 0;
    task24hTrend.value = Array.isArray(body.hour24Trend) ? (body.hour24Trend as TrendPoint[]) : [];
  } catch {
    task24hTotal.value = 0;
    task24hTrend.value = [];
  } finally {
    taskChartLoading.value = false;
    renderTask24hOverview();
  }
}

function scheduleSummaryStats() {
  summaryDelayTimer = setTimeout(() => {
    void fetchSummaryStats();
  }, 120);
}

function scheduleTask24hChart() {
  taskChartLoading.value = true;
  taskChartDelayTimer = setTimeout(() => {
    void fetchTask24hStats();
  }, 250);
}

function clearAllTimers() {
  if (summaryDelayTimer) {
    clearTimeout(summaryDelayTimer);
    summaryDelayTimer = undefined;
  }
  if (taskChartDelayTimer) {
    clearTimeout(taskChartDelayTimer);
    taskChartDelayTimer = undefined;
  }
}

function triggerBackgroundRefresh() {
  clearAllTimers();
  scheduleSummaryStats();
  scheduleTask24hChart();
}

onMounted(() => {
  void hydrateFromCache();
  triggerBackgroundRefresh();
});

onActivated(() => {
  triggerBackgroundRefresh();
});

onUnmounted(() => {
  clearAllTimers();
});
</script>

<template>
  <div class="p-5">
    <WorkbenchHeader
      :avatar="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
    >
      <template #title>
        {{ greeting }}, {{ userStore.userInfo?.realName || $t('page.workspace.userFallback') }}
      </template>
      <template #description>
        {{ headerDescription }}
      </template>
      <template #stats>
        <div
          v-for="(item, index) in headerStats"
          :key="item.label"
          :class="index === 0 ? 'mr-4 flex flex-col justify-center text-right md:mr-10' : 'mr-4 ml-8 flex flex-col justify-center text-right md:mr-10 md:ml-12'"
        >
          <span class="text-foreground/80">{{ item.label }}</span>
          <span class="text-2xl">{{ item.value }}</span>
        </div>
      </template>
    </WorkbenchHeader>

    <div class="mt-5 flex flex-col lg:flex-row">
      <div class="mr-4 w-full lg:w-3/5">
        <WorkbenchProject :items="projectItems" :title="$t('page.workspace.section.commonFunctions')" @click="navTo" />
        <AnalysisChartCard class="mt-5" :title="$t('page.workspace.chart.taskTitle')">
          <template #extra>
            <span class="text-xs text-foreground/70">
              {{ $t('page.workspace.chart.task.hour24Total') }}: {{ task24hTotal }}
            </span>
          </template>
          <div v-if="taskChartLoading" class="h-[240px] animate-pulse rounded bg-muted/40" />
          <EchartsUI v-else ref="task24hChartRef" class="h-[240px]" />
        </AnalysisChartCard>
      </div>
      <div class="w-full lg:w-2/5">
        <WorkbenchQuickNav
          :items="quickNavItems"
          class="mt-5 lg:mt-0"
          :title="$t('page.workspace.section.quickNav')"
          @click="navTo"
        />
        <WorkbenchTodo :items="todoItems" class="mt-5" :title="$t('page.workspace.section.todo')" />
      </div>
    </div>
  </div>
</template>
