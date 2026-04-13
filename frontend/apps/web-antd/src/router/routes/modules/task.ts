import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'TaskPlan',
    path: '/task',
    component: () => import('#/views/task/plan/index.vue'),
    meta: {
      icon: 'lucide:list-todo',
      order: 5,
      title: '计划任务',
    },
  },
];

export default routes;
