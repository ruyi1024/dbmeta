import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataPatrol',
    path: '/patrol',
    redirect: '/patrol/config',
    meta: {
      icon: 'lucide:scan-line',
      order: 3.2,
      title: $t('menu.patrol.title'),
    },
    children: [
      {
        name: 'DataPatrolIndex',
        path: '/patrol/index',
        component: () => import('#/views/patrol/index.vue'),
        meta: {
          icon: 'lucide:radar',
          title: $t('menu.patrol.report'),
        },
      },
      {
        name: 'DataPatrolConfig',
        path: '/patrol/config',
        component: () => import('#/views/patrol/config/index.vue'),
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: $t('menu.patrol.config'),
        },
      },
    ],
  },
];

export default routes;
