import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'Users',
    path: '/users',
    redirect: '/users/manager',
    meta: {
      icon: 'lucide:users',
      order: 7,
      title: $t('menu.users.title'),
    },
    children: [
      {
        name: 'UsersManager',
        path: '/users/manager',
        component: () => import('#/views/users/manager/index.vue'),
        meta: {
          icon: 'lucide:user-cog',
          title: $t('menu.users.manager'),
        },
      },
    ],
  },
];

export default routes;
