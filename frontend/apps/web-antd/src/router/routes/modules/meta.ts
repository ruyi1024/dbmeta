import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Meta',
    path: '/meta',
    redirect: '/meta/dashboard',
    meta: {
      icon: 'lucide:database',
      title: '数据字典',
      order: 1,
    },
    children: [
      {
        name: 'MetaDashboard',
        path: '/meta/dashboard',
        component: () => import('#/views/meta/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: '元数据概览',
        },
      },
      {
        name: 'MetaQuality',
        path: '/meta/quality',
        component: () => import('#/views/meta/quality/index.vue'),
        meta: {
          icon: 'lucide:badge-check',
          title: '元数据质量',
        },
      },
      {
        name: 'MetaInstance',
        path: '/meta/instance',
        component: () => import('#/views/meta/instance/index.vue'),
        meta: {
          icon: 'lucide:server',
          title: '实例信息',
        },
      },
      {
        name: 'MetaDatabase',
        path: '/meta/database',
        component: () => import('#/views/meta/database/index.vue'),
        meta: {
          icon: 'lucide:database-zap',
          title: '数据库信息',
        },
      },
      {
        name: 'MetaBusinessInfo',
        path: '/meta/business-info',
        component: () => import('#/views/meta/business-info/index.vue'),
        meta: {
          icon: 'lucide:building-2',
          title: '业务信息',
        },
      },
      {
        name: 'MetaDatabaseBusiness',
        path: '/meta/database-business',
        component: () => import('#/views/meta/database-business/index.vue'),
        meta: {
          icon: 'lucide:link-2',
          title: '数据库业务信息',
        },
      },
      {
        name: 'MetaTable',
        path: '/meta/table',
        component: () => import('#/views/meta/table/index.vue'),
        meta: {
          icon: 'lucide:table',
          title: '数据表信息',
        },
      },
      {
        name: 'MetaColumn',
        path: '/meta/column',
        component: () => import('#/views/meta/column/index.vue'),
        meta: {
          icon: 'lucide:columns-3',
          title: '数据字段信息',
        },
      },
    ],
  },
];

export default routes;
