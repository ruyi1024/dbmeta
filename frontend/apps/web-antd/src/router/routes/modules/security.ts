import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'DataSecurity',
    path: '/security',
    redirect: '/security/dashboard',
    meta: {
      icon: 'lucide:shield',
      order: 4,
      title: '数据安全',
    },
    children: [
      {
        name: 'DataSecurityDashboard',
        path: '/security/dashboard',
        component: () => import('#/views/security/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: '数据安全大盘',
        },
      },
      {
        name: 'DataSecurityQueryAudit',
        path: '/security/audit',
        component: () => import('#/views/security/audit/index.vue'),
        meta: {
          icon: 'lucide:clipboard-list',
          title: '数据查询审计',
        },
      },
      {
        name: 'DataSecurityQueryPrivilege',
        path: '/security/privilege',
        component: () => import('#/views/security/privilege/index.vue'),
        meta: {
          icon: 'lucide:key-round',
          title: '数据查询授权',
        },
      },
      {
        name: 'DataSecuritySensitiveInventory',
        path: '/security/sensitive-inventory',
        component: () => import('#/views/security/sensitive-inventory/index.vue'),
        meta: {
          icon: 'lucide:scan-search',
          title: '敏感信息盘点',
        },
      },
      {
        name: 'DataSecuritySensitiveRule',
        path: '/security/sensitive-rule',
        component: () => import('#/views/security/sensitive-rule/index.vue'),
        meta: {
          icon: 'lucide:list-filter',
          title: '敏感信息探测规则',
        },
      },
    ],
  },
];

export default routes;
