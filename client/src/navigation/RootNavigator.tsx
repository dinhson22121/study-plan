import { useEffect } from 'react';
import { ActivityIndicator, View } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';

import { AuthStack } from '@/navigation/AuthStack';
import { OnboardingStack } from '@/navigation/OnboardingStack';
import { AppStack } from '@/navigation/AppStack';
import { useAuthStore } from '@/store/authStore';
import { colors } from '@/theme/tokens';

export function RootNavigator() {
  const status = useAuthStore((s) => s.status);
  const onboarded = useAuthStore((s) => s.onboarded);
  const bootstrap = useAuthStore((s) => s.bootstrap);

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  if (status === 'loading') {
    return (
      <View
        style={{
          flex: 1,
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: colors.surfaceApp,
        }}
      >
        <ActivityIndicator size="large" color={colors.brandPrimary} />
      </View>
    );
  }

  return (
    <NavigationContainer>
      {status !== 'authenticated' ? (
        <AuthStack />
      ) : onboarded ? (
        <AppStack />
      ) : (
        <OnboardingStack />
      )}
    </NavigationContainer>
  );
}
