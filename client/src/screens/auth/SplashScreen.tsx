// SplashScreen — brand hero, auto-advances to Login.
import { useEffect, useRef } from 'react';
import { View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { AppText, Button } from '@/components/ui';
import { colors, spacing } from '@/theme/tokens';
import type { AuthStackParamList } from '@/navigation/types';

const AUTO_ADVANCE_MS = 1500;

type Props = NativeStackScreenProps<AuthStackParamList, 'Splash'>;

export default function SplashScreen({ navigation }: Props) {
  const navigatedRef = useRef(false);

  const goToLogin = () => {
    if (navigatedRef.current) {
      return;
    }
    navigatedRef.current = true;
    navigation.replace('Login');
  };

  useEffect(() => {
    const timer = setTimeout(goToLogin, AUTO_ADVANCE_MS);
    return () => clearTimeout(timer);
    // navigation is stable for the lifetime of the screen.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: colors.brandPrimary }} edges={['top', 'bottom']}>
      <View
        style={{
          flex: 1,
          padding: spacing.xl,
          alignItems: 'center',
          justifyContent: 'center',
          gap: spacing.md,
        }}
      >
        <AppText variant="hero" weight="bold" display center color={colors.textOnBrand}>
          AI Study Coach
        </AppText>
        <AppText variant="bodyLarge" weight="medium" center color={colors.surfaceSoft}>
          Học đúng trọng tâm, tiến bộ mỗi ngày
        </AppText>
      </View>

      <View style={{ padding: spacing.xl, gap: spacing.md }}>
        <Button
          title="Bắt đầu"
          variant="secondary"
          onPress={goToLogin}
          testID="splash-start"
        />
      </View>
    </SafeAreaView>
  );
}
