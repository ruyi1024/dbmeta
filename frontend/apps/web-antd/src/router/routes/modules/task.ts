import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'TaskPlan',
    path: '/task',
    component: () => import('#/views/task/plan/index.vue'),
    meta: {
      icon: 'lucide:list-todo',
      order: 5,
      title: $t('menu.task.title'),
    },
  },
];

export default routes;
