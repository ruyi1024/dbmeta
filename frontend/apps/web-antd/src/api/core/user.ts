import type { UserInfo } from '@vben/types';

import { baseRequestClient } from '#/api/request';

interface LegacyCurrentUserResponse {
  data?: {
    admin?: boolean;
    avatar?: string;
    chineseName?: string;
    id?: number;
    username?: string;
  };
  success?: boolean;
}

/**
 * 解析 /v1/currentUser：Go 返回 { success, data: user }；
 * baseRequestClient 未配置 responseReturn，结果为 AxiosResponse，用户对象在 response.data.data。
 */
function extractLegacyUser(raw: unknown): LegacyCurrentUserResponse['data'] {
  if (!raw || typeof raw !== 'object') {
    return {};
  }
  const r = raw as Record<string, unknown>;
  const httpBody =
    'status' in r && typeof (r as { status?: number }).status === 'number'
      ? (r as { data?: unknown }).data
      : r;
  const inner = (httpBody as Record<string, unknown> | undefined)?.data;
  if (inner && typeof inner === 'object') {
    return inner as LegacyCurrentUserResponse['data'];
  }
  return {};
}

/**
 * 获取用户信息
 */
export async function getUserInfoApi() {
  const response = await baseRequestClient.get<unknown>('/v1/currentUser');
  const legacyUser = extractLegacyUser(response);

  return {
    avatar: legacyUser.avatar || '',
    desc: '',
    homePath: '/analytics',
    realName: legacyUser.chineseName || legacyUser.username || '用户',
    roles: legacyUser.admin ? ['admin'] : ['user'],
    userId: String(legacyUser.id ?? ''),
    username: legacyUser.username || '',
  } as UserInfo;
}
