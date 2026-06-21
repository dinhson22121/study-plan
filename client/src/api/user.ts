import { api } from '@/api/client';
import type { User } from '@/api/types';

export const userApi = {
  me: () => api.get<User>('/users/me'),
  updateMe: (patch: Partial<Pick<User, 'display_name'>>) =>
    api.put<User>('/users/me', patch),
};
