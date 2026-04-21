import { useUserStore } from '@vben/stores';

import { message } from 'ant-design-vue';

export function checkPermission(tip = '没有操作权限，请联系管理员'): boolean {
  const userStore = useUserStore();
  const isAdmin = userStore.userInfo?.roles?.includes('admin') ?? false;
  if (!isAdmin) {
    message.warning(tip);
    return false;
  }
  return true;
}
