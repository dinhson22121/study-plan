import { api } from '@/api/client';
import type { Lesson } from '@/api/types';

export const contentApi = {
  listByTopic: (topicId: string) =>
    api.get<Lesson[]>(`/topics/${topicId}/lessons`),
  get: (lessonId: string) => api.get<Lesson>(`/lessons/${lessonId}`),
};
