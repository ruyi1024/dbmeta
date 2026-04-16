import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    name: 'Setting',
    path: '/setting',
    redirect: '/setting/idc',
    meta: {
      icon: 'lucide:settings',
      order: 6,
      title: $t('menu.setting.title'),
    },
    children: [
      {
        name: 'SettingIndex',
        path: '/setting/index',
        component: () => import('#/views/config-center/index.vue'),
        meta: {
          hideInMenu: true,
          icon: 'lucide:sliders-horizontal',
          title: $t('menu.setting.title'),
        },
      },
      {
        name: 'SettingIdc',
        path: '/setting/idc',
        component: () => import('#/views/config/idc/index.vue'),
        meta: {
          icon: 'lucide:building-2',
          title: $t('menu.setting.idc'),
        },
      },
      {
        name: 'SettingEnv',
        path: '/setting/env',
        component: () => import('#/views/config/env/index.vue'),
        meta: {
          icon: 'lucide:layers',
          title: $t('menu.setting.env'),
        },
      },
      {
        name: 'SettingDatasource',
        path: '/setting/datasource',
        component: () => import('#/views/config/datasource/index.vue'),
        meta: {
          icon: 'lucide:database',
          title: $t('menu.setting.datasource'),
        },
      },
      {
        name: 'SettingAiModels',
        path: '/setting/ai_models',
        component: () => import('#/views/config/ai-models/index.vue'),
        meta: {
          icon: 'lucide:bot',
          title: $t('menu.setting.models'),
        },
      },
      {
        name: 'SettingNotice',
        path: '/setting/notice',
        component: () => import('#/views/config/notice/index.vue'),
        meta: {
          icon: 'lucide:messages-square',
          title: $t('menu.setting.notice'),
        },
      },
    ],
  },
];

export default routes;
