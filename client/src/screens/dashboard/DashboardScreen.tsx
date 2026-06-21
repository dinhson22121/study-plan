// DashboardScreen — greeting, streak, weekly progress, today's tasks, quick actions.
import { useMemo } from 'react';
import { View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import type { CompositeScreenProps } from '@react-navigation/native';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import {
  AppText,
  Button,
  Card,
  ProgressRing,
  Screen,
  ScreenHeader,
  StatePlaceholder,
} from '@/components/ui';
import { progressApi } from '@/api/progress';
import { studyPlanApi } from '@/api/studyplan';
import { useAuthStore } from '@/store/authStore';
import { colors, spacing } from '@/theme/tokens';
import type { Milestone, StudyPlan } from '@/api/types';
import type { AppStackParamList, TabsParamList } from '@/navigation/types';

type Props = CompositeScreenProps<
  BottomTabScreenProps<TabsParamList, 'Dashboard'>,
  NativeStackScreenProps<AppStackParamList>
>;

type DailyTask = {
  id: string;
  title: string;
  topicId?: string;
};

function deriveTasks(plan: StudyPlan | undefined): DailyTask[] {
  if (!plan) {
    return [];
  }
  const active = plan.milestones.find(
    (m: Milestone) => m.status === 'IN_PROGRESS' || m.status === 'ACTIVE',
  );
  const milestone = active ?? plan.milestones[0];
  if (!milestone) {
    return [];
  }
  const topicIds = milestone.topic_ids ?? [];
  if (topicIds.length === 0) {
    return [{ id: milestone.id, title: milestone.title }];
  }
  return topicIds.slice(0, 3).map((topicId, index) => ({
    id: `${milestone.id}-${topicId}`,
    title: `${milestone.title} — phần ${index + 1}`,
    topicId,
  }));
}

export default function DashboardScreen({ navigation }: Props) {
  const user = useAuthStore((s) => s.user);
  const greetingName = user?.display_name?.trim() || 'bạn';

  const overviewQuery = useQuery({
    queryKey: ['progress', 'overview'],
    queryFn: () => progressApi.overview(),
  });

  const plansQuery = useQuery({
    queryKey: ['studyplans', 'latest'],
    queryFn: async () => {
      const plans = await studyPlanApi.list();
      const latest = plans[plans.length - 1];
      if (!latest) {
        return null;
      }
      return studyPlanApi.get(latest.id);
    },
  });

  const tasks = useMemo(
    () => deriveTasks(plansQuery.data ?? undefined),
    [plansQuery.data],
  );
  const firstTopicId = tasks.find((t) => t.topicId)?.topicId;

  const goToNotifications = () => navigation.navigate('Notifications');

  if (overviewQuery.isLoading) {
    return (
      <Screen>
        <StatePlaceholder kind="loading" />
      </Screen>
    );
  }

  if (overviewQuery.isError) {
    return (
      <Screen>
        <StatePlaceholder
          kind="error"
          message="Không tải được dữ liệu trang chủ."
          onRetry={() => overviewQuery.refetch()}
        />
      </Screen>
    );
  }

  const overview = overviewQuery.data;
  const weeklyPct = overview?.weekly_progress_pct ?? 0;
  const streak = overview?.streak_days ?? 0;

  return (
    <Screen scroll>
      <ScreenHeader
        title={`Chào ${greetingName}`}
        subtitle="Cùng học một chút hôm nay nhé"
        rightIcon="🔔"
        onRightPress={goToNotifications}
        rightBadge
      />

      <Card style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.xl }}>
        <ProgressRing progress={weeklyPct} label="Tuần này" />
        <View style={{ flex: 1, gap: spacing.xs }}>
          <AppText variant="caption" color={colors.textMuted}>
            Chuỗi học
          </AppText>
          <AppText variant="hero" weight="bold" display>
            {`${streak} ngày`}
          </AppText>
          <AppText variant="caption" color={colors.textSecondary}>
            Giữ vững phong độ mỗi ngày!
          </AppText>
        </View>
      </Card>

      <View style={{ height: spacing.xl }} />

      <AppText variant="title" weight="bold" display>
        Nhiệm vụ hôm nay
      </AppText>
      <View style={{ height: spacing.md }} />

      {plansQuery.isLoading ? (
        <Card soft>
          <AppText color={colors.textSecondary}>Đang tải nhiệm vụ…</AppText>
        </Card>
      ) : plansQuery.isError ? (
        <Card soft>
          <AppText color={colors.textSecondary}>
            Không tải được nhiệm vụ.
          </AppText>
          <View style={{ height: spacing.sm }} />
          <Button
            title="Thử lại"
            variant="secondary"
            onPress={() => plansQuery.refetch()}
          />
        </Card>
      ) : tasks.length === 0 ? (
        <Card soft>
          <AppText weight="semibold">Chưa có nhiệm vụ nào</AppText>
          <View style={{ height: spacing.xs }} />
          <AppText variant="body" color={colors.textSecondary}>
            Hãy tạo lộ trình học để nhận nhiệm vụ mỗi ngày.
          </AppText>
          <View style={{ height: spacing.md }} />
          <Button
            title="Xem lộ trình"
            variant="secondary"
            onPress={() => navigation.navigate('Main', { screen: 'Roadmap' })}
          />
        </Card>
      ) : (
        <View style={{ gap: spacing.md }}>
          {tasks.map((task) => (
            <Card key={task.id}>
              <AppText weight="semibold" variant="bodyLarge">
                {task.title}
              </AppText>
              <View style={{ height: spacing.md }} />
              <Button
                title="Bắt đầu"
                variant={task.topicId ? 'primary' : 'secondary'}
                disabled={!task.topicId}
                onPress={() =>
                  task.topicId &&
                  navigation.navigate('Lesson', { topicId: task.topicId })
                }
              />
            </Card>
          ))}
        </View>
      )}

      <View style={{ height: spacing.xl }} />

      <AppText variant="title" weight="bold" display>
        Hành động nhanh
      </AppText>
      <View style={{ height: spacing.md }} />

      <View style={{ gap: spacing.md }}>
        <Button
          title="Tiếp tục bài học"
          variant="primary"
          disabled={!firstTopicId}
          onPress={() =>
            firstTopicId &&
            navigation.navigate('Lesson', { topicId: firstTopicId })
          }
        />
        <Button
          title="Xem lộ trình"
          variant="secondary"
          onPress={() => navigation.navigate('Main', { screen: 'Roadmap' })}
        />
        <Button
          title="Làm quiz nhanh"
          variant="secondary"
          onPress={() => navigation.navigate('Quiz')}
        />
      </View>
    </Screen>
  );
}
