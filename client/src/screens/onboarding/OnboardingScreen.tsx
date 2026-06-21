// OnboardingScreen — pick a study goal and daily study minutes.
import { useState } from 'react';
import { View } from 'react-native';
import { useMutation } from '@tanstack/react-query';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { AppText, Button, Chip, Screen } from '@/components/ui';
import { ApiError } from '@/api/client';
import { goalApi } from '@/api/goal';
import { colors, spacing } from '@/theme/tokens';
import type { OnboardingStackParamList } from '@/navigation/types';

type Props = NativeStackScreenProps<OnboardingStackParamList, 'Onboarding'>;

type GoalOption = { value: string; label: string };

const GOAL_OPTIONS: GoalOption[] = [
  { value: 'MATH_SCORE', label: 'Tăng điểm Toán' },
  { value: 'KEEP_STREAK', label: 'Giữ streak học đều' },
  { value: 'EXAM_PREP', label: 'Ôn thi tốt nghiệp' },
];

const MINUTES_OPTIONS: number[] = [20, 40, 60];

export default function OnboardingScreen({ navigation }: Props) {
  const [goalType, setGoalType] = useState<string | null>(null);
  const [dailyMinutes, setDailyMinutes] = useState<number | null>(null);

  // Persist the goal so the user counts as onboarded; navigate regardless of
  // persistence success so the demo flow is never blocked by the network.
  const mutation = useMutation({
    mutationFn: () =>
      goalApi.set({
        goal_type: goalType ?? GOAL_OPTIONS[0].value,
        daily_minutes: dailyMinutes ?? MINUTES_OPTIONS[0],
      }),
    onSettled: () => navigation.navigate('Diagnostic'),
  });

  const canContinue = goalType !== null && dailyMinutes !== null && !mutation.isPending;

  return (
    <Screen scroll>
      <View style={{ gap: spacing.xs, marginBottom: spacing.xxl }}>
        <AppText variant="hero" weight="bold" display>
          Mục tiêu của bạn
        </AppText>
        <AppText variant="body" color={colors.textSecondary}>
          Chúng tôi sẽ cá nhân hóa lộ trình theo lựa chọn này.
        </AppText>
      </View>

      <View style={{ gap: spacing.md, marginBottom: spacing.xxl }}>
        <AppText variant="bodyLarge" weight="semibold">
          Bạn muốn đạt điều gì?
        </AppText>
        <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm }}>
          {GOAL_OPTIONS.map((option) => (
            <Chip
              key={option.value}
              label={option.label}
              selected={goalType === option.value}
              onPress={() => setGoalType(option.value)}
              testID={`goal-${option.value}`}
            />
          ))}
        </View>
      </View>

      <View style={{ gap: spacing.md, marginBottom: spacing.xxl }}>
        <AppText variant="bodyLarge" weight="semibold">
          Thời lượng học mỗi ngày
        </AppText>
        <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm }}>
          {MINUTES_OPTIONS.map((minutes) => (
            <Chip
              key={minutes}
              label={`${minutes} phút`}
              selected={dailyMinutes === minutes}
              onPress={() => setDailyMinutes(minutes)}
              testID={`minutes-${minutes}`}
            />
          ))}
        </View>
      </View>

      {mutation.isError && mutation.error instanceof ApiError ? (
        <AppText variant="caption" color={colors.brandWarning} style={{ marginBottom: spacing.md }}>
          Không lưu được mục tiêu, nhưng bạn vẫn có thể tiếp tục.
        </AppText>
      ) : null}

      <Button
        title="Tiếp tục"
        onPress={() => mutation.mutate()}
        loading={mutation.isPending}
        disabled={!canContinue}
        testID="onboarding-continue"
      />
    </Screen>
  );
}
