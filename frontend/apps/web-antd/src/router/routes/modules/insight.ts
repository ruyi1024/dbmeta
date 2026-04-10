import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Insight',
    path: '/Insight',
    redirect: '/Insight/management',
    meta: {
      icon: 'lucide:chart-line',
      order: 4.5,
      title: '数据洞察',
    },
    children: [
      {
        name: 'InsightIndex',
        path: '/Insight/index',
        component: () => import('#/views/insight/index.vue'),
        meta: {
          icon: 'lucide:lightbulb',
          title: '洞察报告',
        },
      },
      {
        name: 'InsightManagement',
        path: '/Insight/management',
        component: () => import('#/views/insight/management/index.vue'),
        meta: {
          icon: 'lucide:bot-message-square',
          title: '洞察管理',
        },
      },
    ],
  },
];

export default routes;
