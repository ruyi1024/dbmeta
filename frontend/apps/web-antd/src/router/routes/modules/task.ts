import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'TaskManagement',
    path: '/task',
    redirect: '/task/plan',
    meta: {
      icon: 'lucide:calendar-clock',
      order: 5,
      title: '任务',
    },
    children: [
      {
        name: 'TaskPlan',
        path: '/task/plan',
        component: () => import('#/views/task/plan/index.vue'),
        meta: {
          icon: 'lucide:list-todo',
          title: '计划任务',
        },
      },
    ],
  },
];

export default routes;
