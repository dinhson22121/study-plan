// DTOs mirroring server/docs/API.md. snake_case matches the backend JSON.

export type Envelope<T> = {
  success: boolean;
  data?: T;
  error?: { code: string; message: string };
  meta?: { total: number; page: number; limit: number };
};

export type TokenPairDTO = {
  access_token: string;
  refresh_token: string;
  expires_at: string;
};

export type Role = 'STUDENT' | 'ADMIN';

export type User = {
  id: string;
  email: string;
  display_name?: string;
  role: Role;
};

export type Subject = {
  id: string;
  name: string;
  code?: string;
};

export type Chapter = {
  id: string;
  subject_id: string;
  name: string;
  order_index?: number;
};

export type Topic = {
  id: string;
  chapter_id: string;
  name: string;
  description?: string;
};

export type Lesson = {
  id: string;
  topic_id: string;
  title: string;
  body?: string;
  summary_points?: string[];
};

export type QuestionOption = {
  id: string;
  label: string;
  text: string;
};

export type Question = {
  id: string;
  topic_id?: string;
  stem: string;
  options: QuestionOption[];
  // Only present in admin/review contexts, never for students pre-submit:
  is_correct?: boolean;
  explanation?: string;
};

export type Goal = {
  goal_type: string;
  target_date?: string;
  daily_minutes?: number;
};

export type PlacementTest = {
  id: string;
  status: string;
  questions: Question[];
};

export type PlacementResult = {
  id: string;
  subject_id: string;
  subject_name?: string;
  score?: number;
  level?: string;
  summary?: string;
};

export type StudyPlan = {
  id: string;
  status: string;
  target_date?: string;
  milestones: Milestone[];
};

export type Milestone = {
  id: string;
  title: string;
  order_index: number;
  status: string;
  topic_ids?: string[];
};

export type Quiz = {
  id: string;
  status: string;
  questions: Question[];
};

export type QuizResultAnswer = {
  question_id: string;
  selected_option_id?: string;
  correct_option_id?: string;
  is_correct: boolean;
  explanation?: string;
};

export type QuizResult = {
  id: string;
  score: number;
  total: number;
  answers: QuizResultAnswer[];
};

export type ProgressOverview = {
  streak_days: number;
  weekly_progress_pct: number;
  weekly_minutes?: number;
  task_completion_pct?: number;
  fastest_improving_subject?: string;
};

export type TopicProgress = {
  topic_id: string;
  topic_name: string;
  completion_pct: number;
};

export type WeakTopic = {
  topic_id: string;
  topic_name: string;
  accuracy: number;
};

export type Analytics = {
  streak_days: number;
  weekly_minutes: number;
  weekly_progress_pct: number;
};

export type NotificationPreference = {
  type: string;
  enabled: boolean;
};

export type NotificationItem = {
  id: string;
  type: string;
  title: string;
  body: string;
  read: boolean;
  created_at: string;
};
