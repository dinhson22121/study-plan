import axios, {
  AxiosError,
  AxiosInstance,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from 'axios';

import { env } from '@/lib/env';
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  saveTokens,
} from '@/lib/tokenStore';
import type { Envelope, TokenPairDTO } from '@/api/types';

export class ApiError extends Error {
  code: string;
  status: number;
  constructor(code: string, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

// Called when refresh fails — the auth store wires this to force a logout.
let onUnauthorized: (() => void) | null = null;
export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn;
}

const raw: AxiosInstance = axios.create({
  baseURL: env.apiBaseUrl,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

raw.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  const token = await getAccessToken();
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  return config;
});

// Single-flight refresh: concurrent 401s share one refresh round-trip.
let refreshing: Promise<boolean> | null = null;

async function doRefresh(): Promise<boolean> {
  const refreshToken = await getRefreshToken();
  if (!refreshToken) return false;
  try {
    // Bare axios (no interceptors) to avoid recursion.
    const res = await axios.post<Envelope<TokenPairDTO>>(
      `${env.apiBaseUrl}/auth/refresh`,
      { refresh_token: refreshToken },
      { headers: { 'Content-Type': 'application/json' } },
    );
    const pair = res.data?.data;
    if (!pair) return false;
    await saveTokens({
      accessToken: pair.access_token,
      refreshToken: pair.refresh_token,
    });
    return true;
  } catch {
    return false;
  }
}

raw.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<Envelope<unknown>>) => {
    const original = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined;
    const status = error.response?.status ?? 0;

    if (status === 401 && original && !original._retry) {
      original._retry = true;
      if (!refreshing) {
        refreshing = doRefresh().finally(() => {
          refreshing = null;
        });
      }
      const ok = await refreshing;
      if (ok) {
        const token = await getAccessToken();
        if (token) original.headers.set('Authorization', `Bearer ${token}`);
        return raw(original);
      }
      await clearTokens();
      onUnauthorized?.();
    }

    const body = error.response?.data;
    if (body && typeof body === 'object' && 'error' in body && body.error) {
      throw new ApiError(body.error.code, body.error.message, status);
    }
    throw new ApiError('NETWORK_ERROR', error.message ?? 'Network error', status);
  },
);

// Unwraps the {success,data,error} envelope and returns the data payload.
async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const res = await raw.request<Envelope<T>>(config);
  if (res.data?.success === false || res.data?.data === undefined) {
    const e = res.data?.error;
    throw new ApiError(e?.code ?? 'UNKNOWN', e?.message ?? 'Unknown error', res.status);
  }
  return res.data.data as T;
}

// Same as request but also returns pagination meta.
async function requestWithMeta<T>(
  config: AxiosRequestConfig,
): Promise<{ data: T; meta?: Envelope<T>['meta'] }> {
  const res = await raw.request<Envelope<T>>(config);
  if (res.data?.success === false) {
    const e = res.data?.error;
    throw new ApiError(e?.code ?? 'UNKNOWN', e?.message ?? 'Unknown error', res.status);
  }
  return { data: res.data.data as T, meta: res.data.meta };
}

export const api = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    request<T>({ ...config, method: 'GET', url }),
  post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    request<T>({ ...config, method: 'POST', url, data }),
  put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    request<T>({ ...config, method: 'PUT', url, data }),
  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    request<T>({ ...config, method: 'DELETE', url }),
  getList: requestWithMeta,
  raw,
};
