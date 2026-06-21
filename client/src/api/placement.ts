import { api } from '@/api/client';
import type { PlacementResult, PlacementTest } from '@/api/types';

export const placementApi = {
  startTest: (subjectId?: string) =>
    api.post<PlacementTest>('/placement/tests', { subject_id: subjectId }),
  submitTest: (
    testId: string,
    answers: { question_id: string; selected_option_id: string }[],
  ) => api.post<PlacementResult>(`/placement/tests/${testId}/submit`, { answers }),
  listResults: () => api.get<PlacementResult[]>('/placement/results'),
};
