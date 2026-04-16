import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataGrading',
    path: '/grading',
    redirect: '/grading/grade-dict',
    meta: {
      icon: 'lucide:gauge',
      /** 暂时隐藏，恢复时去掉 */
      hideInMenu: true,
      order: 3.4,
      title: $t('menu.grading.title'),
    },
    children: [
      {
        name: 'GradingGradeDict',
        path: '/grading/grade-dict',
        component: () => import('#/views/grading/grade-dict/index.vue'),
        meta: {
          icon: 'lucide:book-marked',
          title: $t('menu.grading.gradeDict'),
        },
      },
      {
        name: 'GradingAsset',
        path: '/grading/asset',
        component: () => import('#/views/grading/asset-grade/index.vue'),
        meta: {
          icon: 'lucide:table-properties',
          title: $t('menu.grading.asset'),
        },
      },
      {
        name: 'GradingLog',
        path: '/grading/log',
        component: () => import('#/views/grading/grade-log/index.vue'),
        meta: {
          icon: 'lucide:history',
          title: $t('menu.grading.log'),
        },
      },
    ],
  },
];

export default routes;
