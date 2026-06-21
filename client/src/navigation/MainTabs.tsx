import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Text } from 'react-native';

import DashboardScreen from '@/screens/dashboard/DashboardScreen';
import RoadmapScreen from '@/screens/roadmap/RoadmapScreen';
import QuizScreen from '@/screens/quiz/QuizScreen';
import ProgressScreen from '@/screens/progress/ProgressScreen';
import ProfileScreen from '@/screens/profile/ProfileScreen';
import { colors, fontFamily, fontSize } from '@/theme/tokens';
import type { TabsParamList } from '@/navigation/types';

const Tab = createBottomTabNavigator<TabsParamList>();

const icons: Record<keyof TabsParamList, string> = {
  Dashboard: '🏠',
  Roadmap: '🗺️',
  QuizTab: '📝',
  Progress: '📈',
  Profile: '👤',
};

const labels: Record<keyof TabsParamList, string> = {
  Dashboard: 'Trang chủ',
  Roadmap: 'Lộ trình',
  QuizTab: 'Quiz',
  Progress: 'Tiến độ',
  Profile: 'Cá nhân',
};

export function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarActiveTintColor: colors.brandPrimary,
        tabBarInactiveTintColor: colors.textMuted,
        tabBarLabelStyle: { fontFamily: fontFamily.uiMedium, fontSize: fontSize.caption },
        tabBarStyle: { backgroundColor: colors.surfaceCard, borderTopColor: colors.borderSoft },
        tabBarIcon: ({ color }) => (
          <Text style={{ fontSize: 18, color }}>{icons[route.name]}</Text>
        ),
        tabBarLabel: labels[route.name],
      })}
    >
      <Tab.Screen name="Dashboard" component={DashboardScreen} />
      <Tab.Screen name="Roadmap" component={RoadmapScreen} />
      <Tab.Screen name="QuizTab" component={QuizScreen} />
      <Tab.Screen name="Progress" component={ProgressScreen} />
      <Tab.Screen name="Profile" component={ProfileScreen} />
    </Tab.Navigator>
  );
}
