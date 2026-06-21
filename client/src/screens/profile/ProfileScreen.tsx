// ProfileScreen — user info, study goal editing, reminder + notification toggles, logout.
import { useEffect, useState } from 'react';
import { Switch, View } from 'react-native';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import {
  AppText,
  Button,
  Card,
  Chip,
  Screen,
  ScreenHeader,
  StatePlaceholder,
} from '@/components/ui';
import { goalApi } from '@/api/goal';
import { notificationApi } from '@/api/notification';
import { useAuthStore } from '@/store/authStore';
import { unregisterPush } from '@/lib/push';
import { colors, spacing } from '@/theme/tokens';
import type { TabsParamList, AppStackParamList } from '@/navigation/types';
import type { Goal, NotificationPreference } from '@/api/types';

type Props = CompositeScreenProps<
  BottomTabScreenProps<TabsParamList, 'Profile'>,
  NativeStackScreenProps<AppStackParamList>
>;

type GoalOption = { value: string; label: string };

const GOAL_OPTIONS: readonly GoalOption[] = [
  { value: 'boost_math', label: 'Tăng điểm Toán' },
  { value: 'keep_streak', label: 'Giữ streak học đều' },
  { value: 'exam_prep', label: 'Ôn thi tốt nghiệp' },
];

const MINUTE_OPTIONS: readonly number[] = [20, 40, 60];

const NOTIFICATION_LABELS: Record<string, string> = {
  reminder: 'Nhắc học hằng ngày',
  milestone: 'Hoàn thành cột mốc',
  new_quiz: 'Quiz mới sẵn sàng',
};

function notificationLabel(type: string): string {
  return NOTIFICATION_LABELS[type] ?? type;
}

export default function ProfileScreen() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const queryClient = useQueryClient();

  const goalQuery = useQuery<Goal>({
    queryKey: ['goal'],
    queryFn: () => goalApi.get(),
  });
  const prefsQuery = useQuery<NotificationPreference[]>({
    queryKey: ['notifications', 'preferences'],
    queryFn: () => notificationApi.listPreferences(),
  });

  const [goalType, setGoalType] = useState<string | undefined>(undefined);
  const [dailyMinutes, setDailyMinutes] = useState<number | undefined>(undefined);

  useEffect(() => {
    if (goalQuery.data) {
      setGoalType(goalQuery.data.goal_type);
      setDailyMinutes(goalQuery.data.daily_minutes);
    }
  }, [goalQuery.data]);

  const saveGoal = useMutation({
    mutationFn: (goal: Goal) => goalApi.set(goal),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['goal'] });
    },
  });

  const togglePref = useMutation({
    mutationFn: ({ type, enabled }: { type: string; enabled: boolean }) =>
      notificationApi.setPreference(type, enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications', 'preferences'] });
    },
  });

  const handleLogout = async () => {
    await unregisterPush();
    await logout();
  };

  if (goalQuery.isLoading || prefsQuery.isLoading) {
    return (
      <Screen>
        <StatePlaceholder kind="loading" />
      </Screen>
    );
  }

  if (goalQuery.isError || prefsQuery.isError) {
    return (
      <Screen>
        <StatePlaceholder
          kind="error"
          message="Không tải được cài đặt cá nhân."
          onRetry={() => {
            goalQuery.refetch();
            prefsQuery.refetch();
          }}
        />
      </Screen>
    );
  }

  const prefs = prefsQuery.data ?? [];
  const reminderTime = goalQuery.data?.target_date ? undefined : '19:30';

  return (
    <Screen scroll>
      <ScreenHeader title="Cá nhân" subtitle="Quản lý mục tiêu và cài đặt" />

      <Card style={{ marginBottom: spacing.lg }}>
        <AppText variant="title" weight="bold" display>
          {user?.display_name ?? 'Học sinh'}
        </AppText>
        <AppText variant="body" color={colors.textSecondary}>
          {user?.email ?? ''}
        </AppText>
      </Card>

      <Card style={{ gap: spacing.md, marginBottom: spacing.lg }}>
        <AppText variant="bodyLarge" weight="semibold">
          Mục tiêu học
        </AppText>
        <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm }}>
          {GOAL_OPTIONS.map((opt) => (
            <Chip
              key={opt.value}
              label={opt.label}
              selected={goalType === opt.value}
              onPress={() => setGoalType(opt.value)}
              testID={`goal-${opt.value}`}
            />
          ))}
        </View>

        <AppText variant="caption" color={colors.textSecondary}>
          Thời lượng học mỗi ngày
        </AppText>
        <View style={{ flexDirection: 'row', gap: spacing.sm }}>
          {MINUTE_OPTIONS.map((min) => (
            <Chip
              key={min}
              label={`${min} phút`}
              selected={dailyMinutes === min}
              onPress={() => setDailyMinutes(min)}
              testID={`minutes-${min}`}
            />
          ))}
        </View>

        <Button
          title="Lưu cài đặt"
          variant="secondary"
          loading={saveGoal.isPending}
          disabled={!goalType}
          onPress={() => {
            if (!goalType) return;
            saveGoal.mutate({ goal_type: goalType, daily_minutes: dailyMinutes });
          }}
          testID="profile-save-goal"
        />
        {saveGoal.isError ? (
          <AppText variant="caption" color={colors.brandDanger}>
            Lưu thất bại. Vui lòng thử lại.
          </AppText>
        ) : null}
      </Card>

      <Card style={{ marginBottom: spacing.lg }}>
        <AppText variant="caption" color={colors.textSecondary}>
          Thời gian nhắc học
        </AppText>
        <AppText variant="title" weight="bold" display color={colors.brandPrimary}>
          {reminderTime ?? '19:30'}
        </AppText>
      </Card>

      <Card style={{ gap: spacing.lg, marginBottom: spacing.lg }}>
        <AppText variant="bodyLarge" weight="semibold">
          Thông báo
        </AppText>
        {prefs.length === 0 ? (
          <AppText variant="body" color={colors.textSecondary}>
            Chưa có tùy chọn thông báo.
          </AppText>
        ) : (
          prefs.map((pref) => (
            <View
              key={pref.type}
              style={{
                flexDirection: 'row',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <AppText variant="body" weight="medium" style={{ flex: 1 }}>
                {notificationLabel(pref.type)}
              </AppText>
              <Switch
                value={pref.enabled}
                disabled={togglePref.isPending}
                onValueChange={(enabled) =>
                  togglePref.mutate({ type: pref.type, enabled })
                }
                accessibilityRole="switch"
                accessibilityLabel={notificationLabel(pref.type)}
                accessibilityState={{ checked: pref.enabled }}
                trackColor={{ false: colors.surfaceSoft, true: colors.brandPrimary }}
              />
            </View>
          ))
        )}
      </Card>

      <Button
        title="Đăng xuất"
        variant="danger"
        onPress={handleLogout}
        testID="profile-logout"
      />
    </Screen>
  );
}
