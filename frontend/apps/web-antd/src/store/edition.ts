import { computed, ref } from 'vue';

import { useAccessStore } from '@vben/stores';

import { baseRequestClient } from '#/api/request';
import { defineStore } from 'pinia';

/** 解析 /api/v1/edition 响应（兼容 axios 整包与已解包 body） */
function parseEditionCommercial(res: unknown): boolean {
  const r = res as {
    data?: { commercial?: boolean; data?: { commercial?: boolean }; success?: boolean };
    commercial?: boolean;
  };
  if (r?.data && typeof r.data === 'object' && 'commercial' in r.data) {
    return (r.data as { commercial?: boolean }).commercial === true;
  }
  const body = r?.data ?? r;
  const inner = (body as { data?: { commercial?: boolean } })?.data ?? body;
  return (inner as { commercial?: boolean })?.commercial === true;
}

/** 是否加载了企业扩展（与后端 /api/v1/edition 对齐） */
export const useEditionStore = defineStore('edition', () => {
  const commercial = ref<boolean | null>(null);

  const isCommercial = computed(() => commercial.value === true);

  function reset() {
    commercial.value = null;
  }

  async function ensureLoaded() {
    const accessStore = useAccessStore();
    if (!accessStore.accessToken) {
      commercial.value = null;
      return;
    }
    if (commercial.value !== null) {
      return;
    }
    try {
      const res = await baseRequestClient.get('/v1/edition');
      commercial.value = parseEditionCommercial(res);
    } catch {
      commercial.value = false;
    }
  }

  return { commercial, ensureLoaded, isCommercial, reset };
});
