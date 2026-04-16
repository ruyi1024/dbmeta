import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'Query',
    path: '/query',
    component: () => import('#/views/meta/query/index.vue'),
    meta: {
      icon: 'lucide:search-code',
      order: 0,
      title: $t('menu.query.title'),
    },
  },
];

export default routes;
