import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataCapacity',
    path: '/capacity',
    redirect: '/capacity/dashboard',
    meta: {
      icon: 'lucide:hard-drive',
      order: 2,
      title: $t('menu.capacity.title'),
    },
    children: [
      {
        name: 'DataCapacityDashboard',
        path: '/capacity/dashboard',
        component: () => import('#/views/capacity/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: $t('menu.capacity.overview'),
        },
      },
      {
        name: 'DataCapacityDatabaseQuery',
        path: '/capacity/database-query',
        component: () => import('#/views/capacity/database-query/index.vue'),
        meta: {
          icon: 'lucide:search',
          title: $t('menu.capacity.database'),
        },
      },
      {
        name: 'DataCapacityTableQuery',
        path: '/capacity/table-query',
        component: () => import('#/views/capacity/table-query/index.vue'),
        meta: {
          icon: 'lucide:table-2',
          title: $t('menu.capacity.table'),
        },
      },
    ],
  },
];

export default routes;
