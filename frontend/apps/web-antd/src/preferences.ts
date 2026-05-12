import { defineOverridesPreferences } from '@vben/preferences';

/**
 * @description 项目配置文件
 * 只需要覆盖项目中的一部分配置，不需要的配置不用覆盖，会自动使用默认配置
 * !!! 更改配置后请清空缓存，否则可能不生效
 */
export const overridesPreferences = defineOverridesPreferences({
  // overrides
  app: {
    defaultHomePath: '/workspace',
    layout: 'header-mixed-nav',
    name: import.meta.env.VITE_APP_TITLE,
    /** 水印改为仅企业版生效，开源版默认关闭 */
    watermark: false,
    watermarkContent: '',
    
  },
  "sidebar": {
    "autoActivateChild": true,
    "width": 228
  },
  "theme": {
    "builtinType": "deep-blue",
    "colorPrimary": "hsl(211 91% 39%)",
    "radius": "0.25"
  },
  "widget": {
    "globalSearch": false,
    "languageToggle": true,
    "themeToggle": true
  }
});
