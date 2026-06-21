import { createNativeStackNavigator } from '@react-navigation/native-stack';

import SplashScreen from '@/screens/auth/SplashScreen';
import LoginScreen from '@/screens/auth/LoginScreen';
import type { AuthStackParamList } from '@/navigation/types';

const Stack = createNativeStackNavigator<AuthStackParamList>();

export function AuthStack() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false }}>
      <Stack.Screen name="Splash" component={SplashScreen} />
      <Stack.Screen name="Login" component={LoginScreen} />
    </Stack.Navigator>
  );
}
