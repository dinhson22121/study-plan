import { api } from '@/api/client';
import type { Analytics, ProgressOverview, TopicProgress, WeakTopic } from '@/api/types';

export const progressApi = {
  overview: () => api.get<ProgressOverview>('/progress'),
  topics: () => api.get<TopicProgress[]>('/progress/topics'),
};

export const analyticsApi = {
  me: () => api.get<Analytics>('/analytics/me'),
  weakTopics: (limit = 5) =>
    api.get<WeakTopic[]>(`/analytics/me/weak-topics?limit=${limit}`),
};
