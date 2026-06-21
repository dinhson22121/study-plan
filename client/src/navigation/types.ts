import type { NavigatorScreenParams } from '@react-navigation/native';

export type AuthStackParamList = {
  Splash: undefined;
  Login: undefined;
};

export type OnboardingStackParamList = {
  Onboarding: undefined;
  Diagnostic: undefined;
  DiagnosticResult: { resultIds?: string[] } | undefined;
};

export type TabsParamList = {
  Dashboard: undefined;
  Roadmap: undefined;
  QuizTab: { topicId?: string } | undefined;
  Progress: undefined;
  Profile: undefined;
};

export type AppStackParamList = {
  Main: NavigatorScreenParams<TabsParamList> | undefined;
  Lesson: { lessonId?: string; topicId?: string };
  Quiz: { quizId?: string; topicId?: string } | undefined;
  QuizResult: { quizId: string };
  Notifications: undefined;
};
