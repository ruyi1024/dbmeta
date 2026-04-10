import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Quality',
    path: '/quality',
    redirect: '/quality/dashboard',
    meta: {
      icon: 'lucide:shield-check',
      order: 3,
      title: '数据质量',
    },
    children: [
      {
        name: 'QualityDashboard',
        path: '/quality/dashboard',
        component: () => import('#/views/quality/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: '数据质量概览',
        },
      },
      {
        name: 'QualityIssues',
        path: '/quality/issues',
        component: () => import('#/views/quality/issues/index.vue'),
        meta: {
          icon: 'lucide:triangle-alert',
          title: '数据质量问题',
        },
      },
      {
        name: 'QualityRules',
        path: '/quality/rules',
        component: () => import('#/views/quality/rules/index.vue'),
        meta: {
          icon: 'lucide:settings-2',
          title: '数据规则配置',
        },
      },
      {
        name: 'QualityTasks',
        path: '/quality/tasks',
        component: () => import('#/views/quality/tasks/index.vue'),
        meta: {
          icon: 'lucide:list-checks',
          title: '质量评估任务',
        },
      },
    ],
  },
];

export default routes;
