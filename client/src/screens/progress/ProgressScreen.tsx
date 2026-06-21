// ProgressScreen — streak, weekly stats, per-topic completion and weak topics.
import { View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';
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
import { analyticsApi, progressApi } from '@/api/progress';
import { colors, radius, spacing } from '@/theme/tokens';
import type { TabsParamList, AppStackParamList } from '@/navigation/types';
import type {
  ProgressOverview,
  TopicProgress,
  WeakTopic,
} from '@/api/types';

type Props = CompositeScreenProps<
  BottomTabScreenProps<TabsParamList, 'Progress'>,
  NativeStackScreenProps<AppStackParamList>
>;

const WEAK_TOPICS_LIMIT = 5;

function StatTile({ value, label }: { value: string; label: string }) {
  return (
    <Card soft style={{ flex: 1, padding: spacing.lg }}>
      <AppText variant="title" weight="bold" display>
        {value}
      </AppText>
      <AppText variant="caption" color={colors.textSecondary}>
        {label}
      </AppText>
    </Card>
  );
}

function CompletionBar({ pct, color }: { pct: number; color: string }) {
  const clamped = Math.max(0, Math.min(100, pct));
  return (
    <View
      style={{
        height: 8,
        borderRadius: radius.chip,
        backgroundColor: colors.surfaceSoft,
        overflow: 'hidden',
      }}
    >
      <View
        style={{
          width: `${clamped}%`,
          height: '100%',
          borderRadius: radius.chip,
          backgroundColor: color,
        }}
      />
    </View>
  );
}

function TopicRow({ topic }: { topic: TopicProgress }) {
  return (
    <View style={{ gap: spacing.sm }}>
      <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
        <AppText variant="body" weight="medium" style={{ flex: 1 }} numberOfLines={1}>
          {topic.topic_name}
        </AppText>
        <AppText variant="body" weight="semibold" color={colors.brandPrimary}>
          {`${Math.round(topic.completion_pct)}%`}
        </AppText>
      </View>
      <CompletionBar pct={topic.completion_pct} color={colors.brandPrimary} />
    </View>
  );
}

function WeakTopicRow({ topic }: { topic: WeakTopic }) {
  const pct = Math.round(topic.accuracy * 100);
  return (
    <View style={{ gap: spacing.sm }}>
      <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
        <AppText variant="body" weight="medium" style={{ flex: 1 }} numberOfLines={1}>
          {topic.topic_name}
        </AppText>
        <AppText variant="body" weight="semibold" color={colors.brandWarning}>
          {`${pct}%`}
        </AppText>
      </View>
      <CompletionBar pct={pct} color={colors.brandWarning} />
    </View>
  );
}

export default function ProgressScreen({ navigation }: Props) {
  const overviewQuery = useQuery<ProgressOverview>({
    queryKey: ['progress', 'overview'],
    queryFn: () => progressApi.overview(),
  });
  const topicsQuery = useQuery<TopicProgress[]>({
    queryKey: ['progress', 'topics'],
    queryFn: () => progressApi.topics(),
  });
  const weakQuery = useQuery<WeakTopic[]>({
    queryKey: ['analytics', 'weak-topics', WEAK_TOPICS_LIMIT],
    queryFn: () => analyticsApi.weakTopics(WEAK_TOPICS_LIMIT),
  });

  if (overviewQuery.isLoading || topicsQuery.isLoading || weakQuery.isLoading) {
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
          message="Không tải được tiến độ học tập."
          onRetry={() => {
            overviewQuery.refetch();
            topicsQuery.refetch();
            weakQuery.refetch();
          }}
        />
      </Screen>
    );
  }

  const overview = overviewQuery.data;
  const topics = topicsQuery.data ?? [];
  const weakTopics = weakQuery.data ?? [];

  const hasData =
    !!overview &&
    (overview.streak_days > 0 ||
      overview.weekly_progress_pct > 0 ||
      topics.length > 0);

  if (!hasData) {
    return (
      <Screen>
        <ScreenHeader title="Tiến độ" subtitle="Theo dõi tiến bộ mỗi ngày" />
        <StatePlaceholder
          kind="empty"
          title="Chưa có dữ liệu tiến độ"
          message="Hoàn thành bài học và quiz đầu tiên để bắt đầu theo dõi."
        />
      </Screen>
    );
  }

  const weeklyMinutes = overview?.weekly_minutes ?? 0;
  const taskCompletion = Math.round(overview?.task_completion_pct ?? 0);
  const fastest = overview?.fastest_improving_subject;

  return (
    <Screen scroll>
      <ScreenHeader title="Tiến độ" subtitle="Theo dõi tiến bộ mỗi ngày" />

      <Card style={{ alignItems: 'center', gap: spacing.md, marginBottom: spacing.lg }}>
        <ProgressRing
          progress={overview?.weekly_progress_pct ?? 0}
          size={120}
          label="tuần này"
        />
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
          <AppText variant="hero" weight="bold" display color={colors.brandAccent}>
            {`${overview?.streak_days ?? 0}`}
          </AppText>
          <AppText variant="body" color={colors.textSecondary}>
            ngày liên tiếp
          </AppText>
        </View>
      </Card>

      <View style={{ flexDirection: 'row', gap: spacing.md, marginBottom: spacing.lg }}>
        <StatTile value={`${weeklyMinutes} phút`} label="Giờ học tuần" />
        <StatTile value={`${taskCompletion}%`} label="Hoàn thành nhiệm vụ" />
      </View>

      {fastest ? (
        <Card soft style={{ marginBottom: spacing.lg }}>
          <AppText variant="caption" color={colors.textSecondary}>
            Môn tiến bộ nhanh nhất
          </AppText>
          <AppText variant="title" weight="bold" display color={colors.brandPrimary}>
            {fastest}
          </AppText>
        </Card>
      ) : null}

      {topics.length > 0 ? (
        <Card style={{ gap: spacing.lg, marginBottom: spacing.lg }}>
          <AppText variant="bodyLarge" weight="semibold">
            Hoàn thành theo chủ đề
          </AppText>
          {topics.map((topic) => (
            <TopicRow key={topic.topic_id} topic={topic} />
          ))}
        </Card>
      ) : null}

      {weakTopics.length > 0 ? (
        <Card style={{ gap: spacing.lg, marginBottom: spacing.lg }}>
          <AppText variant="bodyLarge" weight="semibold">
            Chủ đề cần củng cố
          </AppText>
          {weakTopics.map((topic) => (
            <WeakTopicRow key={topic.topic_id} topic={topic} />
          ))}
        </Card>
      ) : null}

      <Button
        title="Xem lộ trình"
        onPress={() => navigation.navigate('Main', { screen: 'Roadmap' })}
        testID="progress-view-roadmap"
      />
    </Screen>
  );
}
