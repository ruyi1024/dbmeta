import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Setting',
    path: '/setting',
    redirect: '/setting/index',
    meta: {
      icon: 'lucide:settings',
      order: 6,
      title: '设置',
    },
    children: [
      {
        name: 'SettingIndex',
        path: '/setting/index',
        component: () => import('#/views/config-center/index.vue'),
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: '设置',
        },
      },
      {
        name: 'SettingIdc',
        path: '/setting/idc',
        component: () => import('#/views/config/idc/index.vue'),
        meta: {
          icon: 'lucide:building-2',
          title: '机房',
        },
      },
      {
        name: 'SettingEnv',
        path: '/setting/env',
        component: () => import('#/views/config/env/index.vue'),
        meta: {
          icon: 'lucide:layers',
          title: '环境',
        },
      },
      {
        name: 'SettingDatasource',
        path: '/setting/datasource',
        component: () => import('#/views/config/datasource/index.vue'),
        meta: {
          icon: 'lucide:database',
          title: '数据源',
        },
      },
      {
        name: 'SettingAiModels',
        path: '/setting/ai_models',
        component: () => import('#/views/config/ai-models/index.vue'),
        meta: {
          icon: 'lucide:bot',
          title: '模型',
        },
      },
    ],
  },
];

export default routes;
