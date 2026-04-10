import CryptoJS from 'crypto-js';

import { baseRequestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    msg?: string;
    successLogin?: boolean;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  const secret = CryptoJS.enc.Utf8.parse('1234567890abcdef');
  const encrypted = CryptoJS.AES.encrypt(
    JSON.stringify({ ...data, type: 'account' }),
    secret,
    {
      iv: secret,
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7,
    },
  );
  const payload = encrypted.ciphertext.toString(CryptoJS.enc.Hex);

  return baseRequestClient.post<AuthApi.LoginResult>(
    '/v1/login/account',
    payload,
    {
      headers: {
        'Content-Type': 'text/plain;charset=UTF-8',
      },
      withCredentials: true,
    },
  );
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  return baseRequestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh', {
    withCredentials: true,
  });
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.get('/v1/login/outLogin', {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
  // 当前后端未提供权限码接口，先返回空数组以打通登录链路。
  return [];
}
