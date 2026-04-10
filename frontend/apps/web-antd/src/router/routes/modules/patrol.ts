import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataPatrol',
    path: '/patrol',
    redirect: '/patrol/config',
    meta: {
      icon: 'lucide:scan-line',
      order: 4.6,
      title: '数据巡检',
    },
    children: [
      {
        name: 'DataPatrolIndex',
        path: '/patrol/index',
        component: () => import('#/views/patrol/index.vue'),
        meta: {
          icon: 'lucide:radar',
          title: '数据巡检',
        },
      },
      {
        name: 'DataPatrolConfig',
        path: '/patrol/config',
        component: () => import('#/views/patrol/config/index.vue'),
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: '巡检配置',
        },
      },
    ],
  },
];

export default routes;
