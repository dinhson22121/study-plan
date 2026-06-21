import { api } from '@/api/client';
import type { TokenPairDTO } from '@/api/types';

export const authApi = {
  register: (email: string, password: string) =>
    api.post<TokenPairDTO>('/auth/register', { email, password }),
  login: (email: string, password: string) =>
    api.post<TokenPairDTO>('/auth/login', { email, password }),
  refresh: (refreshToken: string) =>
    api.post<TokenPairDTO>('/auth/refresh', { refresh_token: refreshToken }),
  logout: (refreshToken: string) =>
    api.post<{ message: string }>('/auth/logout', { refresh_token: refreshToken }),
};
