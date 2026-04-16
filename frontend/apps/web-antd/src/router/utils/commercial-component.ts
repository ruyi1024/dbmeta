import type { RouteComponent } from 'vue-router';

type Loader = () => Promise<{ default: RouteComponent }>;

const placeholder: Loader = () => import('#/views/_core/commercial/upgrade.vue');

/**
 * 商业模块页面组件。
 * - 开发联调企业版：`.env.development` 中 VITE_COMMERCIAL_FULL_UI 非 false 时加载完整页面（默认）。
 * - 开源发行构建：`.env.production` 中 VITE_COMMERCIAL_FULL_UI=false，仅打包占位页，不依赖商业 views。
 */
export function commercialComponent(loader: Loader): Loader {
  if (import.meta.env.VITE_COMMERCIAL_FULL_UI === 'false') {
    return placeholder;
  }
  return loader;
}
