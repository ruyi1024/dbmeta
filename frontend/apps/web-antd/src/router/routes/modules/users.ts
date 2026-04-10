import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Users',
    path: '/users',
    redirect: '/users/manager',
    meta: {
      icon: 'lucide:users',
      order: 7,
      title: '用户',
    },
    children: [
      {
        name: 'UsersManager',
        path: '/users/manager',
        component: () => import('#/views/users/manager/index.vue'),
        meta: {
          icon: 'lucide:user-cog',
          title: '用户管理',
        },
      },
    ],
  },
];

export default routes;
