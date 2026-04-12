import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Categorization',
    path: '/categorization',
    component: () => import('#/views/categorization/index.vue'),
    meta: {
      icon: 'lucide:layers',
      order: 3.45,
      title: '数据分类',
    },
  },
];

export default routes;
