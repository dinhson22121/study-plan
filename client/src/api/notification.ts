import { api } from '@/api/client';
import type { NotificationItem, NotificationPreference } from '@/api/types';

export const notificationApi = {
  registerDevice: (token: string, platform = 'android') =>
    api.post<{ ok: boolean }>('/devices/token', { token, platform }),
  deleteDevice: (token: string) =>
    api.delete<{ ok: boolean }>('/devices/token', { data: { token } }),
  listPreferences: () =>
    api.get<NotificationPreference[]>('/notifications/preferences'),
  setPreference: (type: string, enabled: boolean) =>
    api.put<NotificationPreference>(`/notifications/preferences/${type}`, { enabled }),
  history: () => api.get<NotificationItem[]>('/notifications/history'),
};
