import { api } from '@/api/client';
import type { Quiz, QuizResult } from '@/api/types';

export const quizApi = {
  start: (topicId?: string) => api.post<Quiz>('/quizzes', { topic_id: topicId }),
  submit: (
    quizId: string,
    answers: { question_id: string; selected_option_id: string }[],
  ) => api.post<QuizResult>(`/quizzes/${quizId}/submit`, { answers }),
  getResult: (quizId: string) => api.get<QuizResult>(`/quizzes/${quizId}`),
  list: () => api.get<Quiz[]>('/quizzes'),
};
