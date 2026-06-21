import { api } from '@/api/client';
import type { Chapter, Subject, Topic } from '@/api/types';

export const curriculumApi = {
  listSubjects: () => api.get<Subject[]>('/curriculum/subjects'),
  listChapters: (subjectId: string) =>
    api.get<Chapter[]>(`/curriculum/subjects/${subjectId}/chapters`),
  listTopics: (chapterId: string) =>
    api.get<Topic[]>(`/curriculum/chapters/${chapterId}/topics`),
  getTopic: (topicId: string) => api.get<Topic>(`/curriculum/topics/${topicId}`),
};
