import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Query',
    path: '/query',
    component: () => import('#/views/meta/query/index.vue'),
    meta: {
      icon: 'lucide:search-code',
      order: 0,
      title: '数据查询',
    },
  },
];

export default routes;
