import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { defineConfig } from '@vben/vite-config';
import { loadEnv } from 'vite';

const dir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig(async () => {
  const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development';
  const env = loadEnv(mode, dir);
  // 默认代理到开源后端 8086；联调企业版请在 .env.development 中设置 VITE_PROXY_TARGET（如 http://127.0.0.1:8088）
  const proxyTarget =
    env.VITE_PROXY_TARGET?.trim() || 'http://127.0.0.1:8086';

  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            target: proxyTarget,
            ws: true,
          },
        },
      },
    },
  };
});
