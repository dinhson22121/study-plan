import { createNativeStackNavigator } from '@react-navigation/native-stack';

import OnboardingScreen from '@/screens/onboarding/OnboardingScreen';
import DiagnosticScreen from '@/screens/onboarding/DiagnosticScreen';
import DiagnosticResultScreen from '@/screens/onboarding/DiagnosticResultScreen';
import type { OnboardingStackParamList } from '@/navigation/types';

const Stack = createNativeStackNavigator<OnboardingStackParamList>();

export function OnboardingStack() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false }}>
      <Stack.Screen name="Onboarding" component={OnboardingScreen} />
      <Stack.Screen name="Diagnostic" component={DiagnosticScreen} />
      <Stack.Screen name="DiagnosticResult" component={DiagnosticResultScreen} />
    </Stack.Navigator>
  );
}
