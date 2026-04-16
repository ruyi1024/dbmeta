import type { RouteRecordRaw } from 'vue-router';

import { commercialComponent } from '#/router/utils/commercial-component';
import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'AuditCenter',
    path: '/audit',
    redirect: '/audit/query-audit',
    meta: {
      commercialOnly: true,
      icon: 'lucide:shield-check',
      order: 4.55,
      title: $t('menu.audit.title'),
    },
    children: [
      {
        name: 'AuditCenterQueryAudit',
        path: '/audit/query-audit',
        component: commercialComponent(
          () => import('#/views/security/audit/index.vue'),
        ),
        meta: {
          icon: 'lucide:clipboard-list',
          title: $t('menu.audit.queryAudit'),
        },
      },
    ],
  },
];

export default routes;
