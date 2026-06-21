import { api } from '@/api/client';
import type { Goal } from '@/api/types';

export const goalApi = {
  get: () => api.get<Goal>('/goals'),
  set: (goal: Goal) => api.put<Goal>('/goals', goal),
};
