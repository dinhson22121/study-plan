import { createNativeStackNavigator } from '@react-navigation/native-stack';

import { MainTabs } from '@/navigation/MainTabs';
import LessonScreen from '@/screens/lesson/LessonScreen';
import QuizScreen from '@/screens/quiz/QuizScreen';
import QuizResultScreen from '@/screens/quiz/QuizResultScreen';
import NotificationsScreen from '@/screens/notifications/NotificationsScreen';
import type { AppStackParamList } from '@/navigation/types';

const Stack = createNativeStackNavigator<AppStackParamList>();

export function AppStack() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false }}>
      <Stack.Screen name="Main" component={MainTabs} />
      <Stack.Screen name="Lesson" component={LessonScreen} />
      <Stack.Screen name="Quiz" component={QuizScreen} />
      <Stack.Screen name="QuizResult" component={QuizResultScreen} />
      <Stack.Screen
        name="Notifications"
        component={NotificationsScreen}
        options={{ presentation: 'modal' }}
      />
    </Stack.Navigator>
  );
}
