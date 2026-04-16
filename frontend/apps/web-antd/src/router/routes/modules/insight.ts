import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';
import { commercialComponent } from '#/router/utils/commercial-component';

const routes: RouteRecordRaw[] = [
  {
    name: 'Insight',
    path: '/Insight',
    redirect: '/Insight/management',
    meta: {
      commercialOnly: true,
      icon: 'lucide:chart-line',
      order: 4.5,
      title: $t('menu.insight.title'),
    },
    children: [
      {
        name: 'InsightIndex',
        path: '/Insight/index',
        component: commercialComponent(() => import('#/views/insight/index.vue')),
        meta: {
          icon: 'lucide:lightbulb',
          title: $t('menu.insight.report'),
        },
      },
      {
        name: 'InsightManagement',
        path: '/Insight/management',
        component: commercialComponent(
          () => import('#/views/insight/management/index.vue'),
        ),
        meta: {
          icon: 'lucide:bot-message-square',
          title: $t('menu.insight.management'),
        },
      },
    ],
  },
];

export default routes;
