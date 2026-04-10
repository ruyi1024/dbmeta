import { defineOverridesPreferences } from '@vben/preferences';

/**
 * @description 项目配置文件
 * 只需要覆盖项目中的一部分配置，不需要的配置不用覆盖，会自动使用默认配置
 * !!! 更改配置后请清空缓存，否则可能不生效
 */
export const overridesPreferences = defineOverridesPreferences({
  // overrides
  app: {
    defaultHomePath: '/analytics',
    name: import.meta.env.VITE_APP_TITLE,
    /** 全屏水印（与旧版数据查询页一致，可在偏好设置中关闭或自定义文案） */
    watermark: true,
    watermarkContent: '',
  },
});
