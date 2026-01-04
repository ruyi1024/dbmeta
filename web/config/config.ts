// https://umijs.org/config/
import { defineConfig } from 'umi';
import { join } from 'path';

import defaultSettings from './defaultSettings';
import proxy from './proxy';
import routes from './routes';

//import aliyunTheme from '@ant-design/aliyun-theme';

const { REACT_APP_ENV } = process.env;

export default defineConfig({
  hash: false,
  publicPath: '/public/static/',
  history: { type: 'hash' },
  antd: {
    // antd5 使用 CSS-in-JS，不需要导入 less 文件
    //import: false,
  },
  dva: {
    hmr: true,
  },
  layout: {
    // https://umijs.org/zh-CN/plugins/plugin-layout
    locale: true,
    siderWidth: 208,
    ...defaultSettings,
  },
  // https://umijs.org/zh-CN/plugins/plugin-locale
  locale: {
    // default zh-CN
    default: 'zh-CN',
    antd: true,
    // default true, when it is true, will use `navigator.language` overwrite default
    baseNavigator: true,
  },
  dynamicImport: {
    loading: '@/components/PageLoading',
  },
  targets: {
    ie: 11,
  },
  // umi routes: https://umijs.org/docs/routing
  routes,
  access: {},
  //Theme for antd: https://ant.design/docs/react/customize-theme-cn
  theme: {
    // antd5 兼容变量 - 通过 theme 配置 less 变量
    '@primary-color': '#1a365d',
    '@primary-1': '#e6f0ff', // 最浅色，用于 hover 背景
    '@primary-6': '#1a365d', // 主色
    '@link-color': '#2c5282',
    '@success-color': '#38a169',
    '@warning-color': '#d69e2e',
    '@error-color': '#e53e3e',
    '@red-6': '#f5222d', // 红色，用于上升趋势
    '@green-6': '#52c41a', // 绿色，用于下降趋势
    '@heading-color': '#2d3748',
    '@text-color': '#4a5568',
    '@text-color-secondary': '#718096',
    '@disabled-color': '#a0aec0',
    '@border-radius-base': '4px',
    '@border-color-base': '#e2e8f0',
    '@border-color-split': '#e8e8e8',
    '@box-shadow-base': '0 2px 8px rgba(0, 0, 0, 0.15)',
    '@shadow-1-down': '0 2px 8px rgba(0, 0, 0, 0.15)',
    '@background-color-base': '#f5f5f5',
    '@component-background': '#ffffff',
    '@input-bg': '#ffffff',
    '@popover-bg': '#ffffff',
    '@card-shadow': '0 2px 8px rgba(0, 0, 0, 0.15)',
    '@font-size-base': '14px',
    '@line-height-base': '1.5', // 行高
    '@screen-xs': '480px',
    '@screen-sm': '576px',
    '@screen-md': '768px',
    '@screen-md-min': '768px',
    '@screen-lg': '992px',
    '@screen-xl': '1200px',
    '@screen-xxl': '1600px',
    '@ant-prefix': 'ant',
  },
  // esbuild is father build tools
  // https://umijs.org/plugins/plugin-esbuild
  esbuild: {},
  title: false,
  ignoreMomentLocale: true,
  proxy: proxy[REACT_APP_ENV || 'dev'],
  manifest: {
    basePath: '/',
  },
  // Fast Refresh 热更新
  fastRefresh: {},
  openAPI: [
    {
      requestLibPath: "import { request } from 'umi'",
      // 或者使用在线的版本
      // schemaPath: "https://gw.alipayobjects.com/os/antfincdn/M%24jrzTTYJN/oneapi.json"
      schemaPath: join(__dirname, 'oneapi.json'),
      mock: false,
    },
    {
      requestLibPath: "import { request } from 'umi'",
      schemaPath: 'https://gw.alipayobjects.com/os/antfincdn/CA1dOm%2631B/openapi.json',
      projectName: 'swagger',
    },
  ],
  nodeModulesTransform: { type: 'none' },
  //mfsu: {}, // 禁用MFSU以避免兼容性问题
  webpack5: {},
  exportStatic: {},
  // 配置 webpack 别名，避免导入不存在的 antd less 文件
  chainWebpack(config: any) {
    const emptyLessPath = join(__dirname, 'empty.less');
    // 将 antd/dist/antd.less 重定向到空文件（使用绝对路径）
    config.resolve.alias.set('antd/dist/antd.less', emptyLessPath);
    // 将 antd/es/auto-complete/style 重定向到空文件
    config.resolve.alias.set('antd/es/auto-complete/style', emptyLessPath);
    config.resolve.alias.set('antd/lib/auto-complete/style', emptyLessPath);
    // 使用 webpack NormalModuleReplacementPlugin 处理所有 antd 组件样式导入
    const webpack = require('webpack');
    config.plugin('antd-style-replace').use(webpack.NormalModuleReplacementPlugin, [
      /^antd\/(es|lib)\/[\w-]+\/style$/,
      emptyLessPath,
    ]);
    // 也处理 antd/dist/antd.less
    config.plugin('antd-dist-less-replace').use(webpack.NormalModuleReplacementPlugin, [
      /^antd\/dist\/antd\.less$/,
      emptyLessPath,
    ]);
  },
});
