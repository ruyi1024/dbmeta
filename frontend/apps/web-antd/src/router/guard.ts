import type { Router } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { startProgress, stopProgress } from '@vben/utils';

import { accessRoutes, coreRouteNames } from '#/router/routes';
import { useAuthStore, useEditionStore } from '#/store';

import { generateAccess } from './access';

function cloneRoutesWithCommercialMenuVisibility(
  routes: RouteRecordRaw[],
  showCommercialMenus: boolean,
): RouteRecordRaw[] {
  return routes.map((route) => {
    const cloned: RouteRecordRaw = {
      ...route,
      meta: route.meta ? { ...route.meta } : undefined,
    };

    if (route.children?.length) {
      cloned.children = cloneRoutesWithCommercialMenuVisibility(
        route.children as RouteRecordRaw[],
        showCommercialMenus,
      );
    }

    if (!showCommercialMenus && cloned.meta?.commercialOnly) {
      cloned.meta = { ...cloned.meta, hideInMenu: true };
      if (cloned.children?.length) {
        cloned.children = cloned.children.map((child) => ({
          ...child,
          meta: { ...(child.meta ?? {}), hideInMenu: true },
        }));
      }
    }

    return cloned;
  });
}

/**
 * 通用守卫配置
 * @param router
 */
function setupCommonGuard(router: Router) {
  // 记录已经加载的页面
  const loadedPaths = new Set<string>();

  router.beforeEach((to) => {
    to.meta.loaded = loadedPaths.has(to.path);

    // 页面加载进度条
    if (!to.meta.loaded && preferences.transition.progress) {
      startProgress();
    }
    return true;
  });

  router.afterEach((to) => {
    // 记录页面是否加载,如果已经加载，后续的页面切换动画等效果不在重复执行

    loadedPaths.add(to.path);

    // 关闭页面加载进度条
    if (preferences.transition.progress) {
      stopProgress();
    }
  });
}

/**
 * 权限访问守卫配置
 * @param router
 */
function setupAccessGuard(router: Router) {
  router.beforeEach(async (to, from) => {
    const accessStore = useAccessStore();
    const userStore = useUserStore();
    const authStore = useAuthStore();

    // 基本路由，这些路由不需要进入权限拦截
    if (coreRouteNames.includes(to.name as string)) {
      if (to.path === LOGIN_PATH && accessStore.accessToken) {
        return decodeURIComponent(
          (to.query?.redirect as string) ||
            userStore.userInfo?.homePath ||
            preferences.app.defaultHomePath,
        );
      }
      return true;
    }

    // accessToken 检查
    if (!accessStore.accessToken) {
      useEditionStore().reset();
      // 明确声明忽略权限访问权限，则可以访问
      if (to.meta.ignoreAccess) {
        return true;
      }

      // 没有访问权限，跳转登录页面
      if (to.fullPath !== LOGIN_PATH) {
        return {
          path: LOGIN_PATH,
          // 如不需要，直接删除 query
          query:
            to.fullPath === preferences.app.defaultHomePath
              ? {}
              : { redirect: encodeURIComponent(to.fullPath) },
          // 携带当前跳转的页面，登录后重新跳转该页面
          replace: true,
        };
      }
      return to;
    }

    const editionStore = useEditionStore();
    await editionStore.ensureLoaded();

    // 数据洞察、审计中心、数据安全等商业功能：菜单始终展示；未加载企业模块时进入路由则跳转升级提示页
    if (
      (to.path.startsWith('/Insight') ||
        to.path === '/audit' ||
        to.path.startsWith('/audit/') ||
        to.path === '/security' ||
        to.path.startsWith('/security/')) &&
      !editionStore.isCommercial
    ) {
      return { name: 'CommercialUpgrade', replace: true };
    }

    // 是否已经生成过动态路由
    if (accessStore.isAccessChecked) {
      return true;
    }

    // 生成路由表
    // 当前登录用户拥有的角色标识列表
    const userInfo = userStore.userInfo || (await authStore.fetchUserInfo());
    const userRoles = userInfo.roles ?? [];

    // 默认隐藏企业版菜单；仅在 /v1/edition 返回 commercial=true 时展示
    const accessRoutesByEdition = cloneRoutesWithCommercialMenuVisibility(
      accessRoutes,
      editionStore.isCommercial,
    );

    // 生成菜单和路由
    const { accessibleMenus, accessibleRoutes } = await generateAccess({
      roles: userRoles,
      router,
      // 则会在菜单中显示，但是访问会被重定向到403
      routes: accessRoutesByEdition,
    });

    // 保存菜单信息和路由信息
    accessStore.setAccessMenus(accessibleMenus);
    accessStore.setAccessRoutes(accessibleRoutes);
    accessStore.setIsAccessChecked(true);
    const redirectPath = (from.query.redirect ??
      (to.path === preferences.app.defaultHomePath
        ? userInfo.homePath || preferences.app.defaultHomePath
        : to.fullPath)) as string;

    return {
      ...router.resolve(decodeURIComponent(redirectPath)),
      replace: true,
    };
  });
}

/**
 * 项目守卫配置
 * @param router
 */
function createRouterGuard(router: Router) {
  /** 通用 */
  setupCommonGuard(router);
  /** 权限访问 */
  setupAccessGuard(router);
}

export { createRouterGuard };
