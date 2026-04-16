import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'Quality',
    path: '/quality',
    redirect: '/quality/dashboard',
    meta: {
      icon: 'lucide:shield-check',
      order: 3,
      title: $t('menu.quality.title'),
    },
    children: [
      {
        name: 'QualityDashboard',
        path: '/quality/dashboard',
        component: () => import('#/views/quality/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: $t('menu.quality.overview'),
        },
      },
      {
        name: 'QualityIssues',
        path: '/quality/issues',
        component: () => import('#/views/quality/issues/index.vue'),
        meta: {
          icon: 'lucide:triangle-alert',
          title: $t('menu.quality.issues'),
        },
      },
      {
        name: 'QualityRules',
        path: '/quality/rules',
        component: () => import('#/views/quality/rules/index.vue'),
        meta: {
          icon: 'lucide:settings-2',
          title: $t('menu.quality.rules'),
        },
      },
      {
        name: 'QualityTasks',
        path: '/quality/tasks',
        component: () => import('#/views/quality/tasks/index.vue'),
        meta: {
          icon: 'lucide:list-checks',
          title: $t('menu.quality.tasks'),
        },
      },
    ],
  },
];

export default routes;
