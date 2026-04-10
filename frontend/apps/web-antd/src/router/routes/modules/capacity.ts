import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataCapacity',
    path: '/capacity',
    redirect: '/capacity/dashboard',
    meta: {
      icon: 'lucide:hard-drive',
      order: 2,
      title: '数据容量',
    },
    children: [
      {
        name: 'DataCapacityDashboard',
        path: '/capacity/dashboard',
        component: () => import('#/views/capacity/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: '数据容量概览',
        },
      },
      {
        name: 'DataCapacityDatabaseQuery',
        path: '/capacity/database-query',
        component: () => import('#/views/capacity/database-query/index.vue'),
        meta: {
          icon: 'lucide:search',
          title: '数据库容量查询',
        },
      },
      {
        name: 'DataCapacityTableQuery',
        path: '/capacity/table-query',
        component: () => import('#/views/capacity/table-query/index.vue'),
        meta: {
          icon: 'lucide:table-2',
          title: '数据表容量查询',
        },
      },
    ],
  },
];

export default routes;
