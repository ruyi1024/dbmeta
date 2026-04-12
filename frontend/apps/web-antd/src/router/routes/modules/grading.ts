import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataGrading',
    path: '/grading',
    redirect: '/grading/grade-dict',
    meta: {
      icon: 'lucide:gauge',
      order: 3.4,
      title: '数据分级',
    },
    children: [
      {
        name: 'GradingGradeDict',
        path: '/grading/grade-dict',
        component: () => import('#/views/grading/grade-dict/index.vue'),
        meta: {
          icon: 'lucide:book-marked',
          title: '分级字典',
        },
      },
      {
        name: 'GradingAsset',
        path: '/grading/asset',
        component: () => import('#/views/grading/asset-grade/index.vue'),
        meta: {
          icon: 'lucide:table-properties',
          title: '资产分级',
        },
      },
      {
        name: 'GradingLog',
        path: '/grading/log',
        component: () => import('#/views/grading/grade-log/index.vue'),
        meta: {
          icon: 'lucide:history',
          title: '变更记录',
        },
      },
    ],
  },
];

export default routes;
