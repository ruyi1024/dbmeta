import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'Categorization',
    path: '/categorization',
    component: () => import('#/views/categorization/index.vue'),
    meta: {
      icon: 'lucide:layers',
      /** 暂时隐藏，恢复时去掉 */
      hideInMenu: true,
      order: 3.45,
      title: $t('menu.categorization.title'),
    },
  },
];

export default routes;
