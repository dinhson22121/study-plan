import { api } from '@/api/client';
import type { StudyPlan } from '@/api/types';

export const studyPlanApi = {
  generate: () => api.post<StudyPlan>('/studyplans/generate', {}),
  list: () => api.get<StudyPlan[]>('/studyplans'),
  get: (id: string) => api.get<StudyPlan>(`/studyplans/${id}`),
};
