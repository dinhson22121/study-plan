import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from "axios";
import { env } from "@/lib/env";
import { tokenStore } from "@/auth/tokenStore";
import type { Envelope, Meta, TokenPair } from "./types";

export class ApiError extends Error {
  code: string;
  status?: number;
  constructor(code: string, message: string, status?: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

let onUnauthorized: () => void = () => {};
export function setOnUnauthorized(fn: () => void): void {
  onUnauthorized = fn;
}

export const api = axios.create({ baseURL: env.apiBaseUrl });

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStore.access;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

let refreshing: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  const rt = tokenStore.refresh;
  if (!rt) return null;
  try {
    const res = await axios.post<Envelope<TokenPair>>(`${env.apiBaseUrl}/auth/refresh`, {
      refresh_token: rt,
    });
    if (!res.data.success) return null;
    tokenStore.set(res.data.data.access_token, res.data.data.refresh_token);
    return res.data.data.access_token;
  } catch {
    return null;
  }
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined;
    const status = error.response?.status as number | undefined;

    if (status === 401 && original && !original._retry && tokenStore.refresh) {
      original._retry = true;
      // Single-flight: concurrent 401s share one in-flight refresh promise.
      if (!refreshing) {
        refreshing = refreshAccessToken().finally(() => {
          refreshing = null;
        });
      }
      const newToken = await refreshing;
      if (newToken) {
        original.headers.Authorization = `Bearer ${newToken}`;
        return api(original);
      }
      tokenStore.clear();
      onUnauthorized();
    }
    return Promise.reject(toApiError(error));
  },
);

function toApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const body = error.response?.data as Envelope<unknown> | undefined;
    if (body?.error) {
      return new ApiError(body.error.code, body.error.message, error.response?.status);
    }
    return new ApiError("NETWORK_ERROR", error.message, error.response?.status);
  }
  return new ApiError("UNKNOWN", "unexpected error");
}

export async function unwrap<T>(p: Promise<AxiosResponse<Envelope<T>>>): Promise<T> {
  const res = await p;
  if (!res.data.success) {
    throw new ApiError(res.data.error?.code ?? "ERROR", res.data.error?.message ?? "request failed");
  }
  return res.data.data;
}

export async function unwrapList<T>(
  p: Promise<AxiosResponse<Envelope<T>>>,
): Promise<{ data: T; meta?: Meta }> {
  const res = await p;
  if (!res.data.success) {
    throw new ApiError(res.data.error?.code ?? "ERROR", res.data.error?.message ?? "request failed");
  }
  return { data: res.data.data, meta: res.data.meta };
}
