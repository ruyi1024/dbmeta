import type { RouteRecordRaw } from 'vue-router';

import { commercialComponent } from '#/router/utils/commercial-component';
import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataSecurity',
    path: '/security',
    redirect: '/security/dashboard',
    meta: {
      commercialOnly: true,
      icon: 'lucide:shield',
      order: 4,
      title: $t('menu.security.title'),
    },
    children: [
      {
        name: 'DataSecurityDashboard',
        path: '/security/dashboard',
        component: commercialComponent(
          () => import('#/views/security/dashboard/index.vue'),
        ),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: $t('menu.security.dashboard'),
        },
      },
      {
        name: 'DataSecurityQueryPrivilege',
        path: '/security/privilege',
        component: commercialComponent(
          () => import('#/views/security/privilege/index.vue'),
        ),
        meta: {
          icon: 'lucide:key-round',
          title: $t('menu.security.privilege'),
        },
      },
      {
        name: 'DataSecuritySensitiveInventory',
        path: '/security/sensitive-inventory',
        component: commercialComponent(
          () => import('#/views/security/sensitive-inventory/index.vue'),
        ),
        meta: {
          icon: 'lucide:scan-search',
          title: $t('menu.security.sensitiveInventory'),
        },
      },
      {
        name: 'DataSecuritySensitiveRule',
        path: '/security/sensitive-rule',
        component: commercialComponent(
          () => import('#/views/security/sensitive-rule/index.vue'),
        ),
        meta: {
          icon: 'lucide:list-filter',
          title: $t('menu.security.sensitiveRule'),
        },
      },
    ],
  },
];

export default routes;
