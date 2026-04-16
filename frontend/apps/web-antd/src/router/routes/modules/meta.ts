import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'Meta',
    path: '/meta',
    redirect: '/meta/dashboard',
    meta: {
      icon: 'lucide:database',
      title: $t('menu.dictionary.title'),
      order: 1,
    },
    children: [
      {
        name: 'MetaDashboard',
        path: '/meta/dashboard',
        component: () => import('#/views/meta/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: $t('menu.dictionary.overview'),
        },
      },
      {
        name: 'MetaQuality',
        path: '/meta/quality',
        component: () => import('#/views/meta/quality/index.vue'),
        meta: {
          icon: 'lucide:badge-check',
          title: $t('menu.dictionary.quality'),
        },
      },
      {
        name: 'MetaInstance',
        path: '/meta/instance',
        component: () => import('#/views/meta/instance/index.vue'),
        meta: {
          icon: 'lucide:server',
          title: $t('menu.dictionary.instance'),
        },
      },
      {
        name: 'MetaDatabase',
        path: '/meta/database',
        component: () => import('#/views/meta/database/index.vue'),
        meta: {
          icon: 'lucide:database-zap',
          title: $t('menu.dictionary.database'),
        },
      },
      {
        name: 'MetaBusinessInfo',
        path: '/meta/business-info',
        component: () => import('#/views/meta/business-info/index.vue'),
        meta: {
          icon: 'lucide:building-2',
          title: $t('menu.dictionary.businessInfo'),
        },
      },
      {
        name: 'MetaDatabaseBusiness',
        path: '/meta/database-business',
        component: () => import('#/views/meta/database-business/index.vue'),
        meta: {
          icon: 'lucide:link-2',
          title: $t('menu.dictionary.databaseBusiness'),
        },
      },
      {
        name: 'MetaTable',
        path: '/meta/table',
        component: () => import('#/views/meta/table/index.vue'),
        meta: {
          icon: 'lucide:table',
          title: $t('menu.dictionary.table'),
        },
      },
      {
        name: 'MetaColumn',
        path: '/meta/column',
        component: () => import('#/views/meta/column/index.vue'),
        meta: {
          icon: 'lucide:columns-3',
          title: $t('menu.dictionary.column'),
        },
      },
    ],
  },
];

export default routes;
