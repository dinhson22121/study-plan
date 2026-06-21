// RoadmapScreen — latest study plan with milestone cards; generate plan if none.
import { View } from 'react-native';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { CompositeScreenProps } from '@react-navigation/native';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
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
import { studyPlanApi } from '@/api/studyplan';
import { colors, spacing } from '@/theme/tokens';
import type { Milestone, StudyPlan } from '@/api/types';
import type { AppStackParamList, TabsParamList } from '@/navigation/types';

type Props = CompositeScreenProps<
  BottomTabScreenProps<TabsParamList, 'Roadmap'>,
  NativeStackScreenProps<AppStackParamList>
>;

const LATEST_PLAN_KEY = ['studyplans', 'latest'] as const;

type MilestoneState = 'completed' | 'active' | 'locked';

function milestoneState(milestone: Milestone): MilestoneState {
  const status = milestone.status.toUpperCase();
  if (status === 'COMPLETED' || status === 'DONE') {
    return 'completed';
  }
  if (status === 'IN_PROGRESS' || status === 'ACTIVE' || status === 'CURRENT') {
    return 'active';
  }
  return 'locked';
}

const stateLabel: Record<MilestoneState, string> = {
  completed: 'Đã hoàn thành',
  active: 'Đang học',
  locked: 'Chưa mở khóa',
};

function MilestoneCard({
  milestone,
  index,
  onOpen,
}: {
  milestone: Milestone;
  index: number;
  onOpen: (topicId: string) => void;
}) {
  const state = milestoneState(milestone);
  const topicId = milestone.topic_ids?.[0];
  const canOpen = state === 'active' && Boolean(topicId);

  return (
    <Card soft={state === 'locked'}>
      <View
        style={{
          flexDirection: 'row',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: spacing.md,
        }}
      >
        <AppText
          variant="bodyLarge"
          weight="bold"
          color={state === 'locked' ? colors.textMuted : colors.textPrimary}
          style={{ flex: 1 }}
        >
          {`${index + 1}. ${milestone.title}`}
        </AppText>
        <Chip label={state === 'locked' ? '🔒' : stateLabel[state]} selected={state === 'active'} />
      </View>

      {canOpen ? (
        <>
          <View style={{ height: spacing.md }} />
          <Button title="Mở bài học" onPress={() => topicId && onOpen(topicId)} />
        </>
      ) : null}
    </Card>
  );
}

export default function RoadmapScreen({ navigation }: Props) {
  const queryClient = useQueryClient();

  const planQuery = useQuery<StudyPlan | null>({
    queryKey: LATEST_PLAN_KEY,
    queryFn: async () => {
      const plans = await studyPlanApi.list();
      const latest = plans[plans.length - 1];
      if (!latest) {
        return null;
      }
      return studyPlanApi.get(latest.id);
    },
  });

  const generateMutation = useMutation({
    mutationFn: () => studyPlanApi.generate(),
    onSuccess: (plan) => {
      queryClient.setQueryData(LATEST_PLAN_KEY, plan);
      queryClient.invalidateQueries({ queryKey: ['studyplans'] });
    },
  });

  const openLesson = (topicId: string) =>
    navigation.navigate('Lesson', { topicId });

  if (planQuery.isLoading) {
    return (
      <Screen>
        <ScreenHeader title="Lộ trình học" />
        <StatePlaceholder kind="loading" />
      </Screen>
    );
  }

  if (planQuery.isError) {
    return (
      <Screen>
        <ScreenHeader title="Lộ trình học" />
        <StatePlaceholder
          kind="error"
          message="Không tải được lộ trình của bạn."
          onRetry={() => planQuery.refetch()}
        />
      </Screen>
    );
  }

  const plan = planQuery.data;
  const milestones = plan?.milestones ?? [];

  if (!plan || milestones.length === 0) {
    return (
      <Screen>
        <ScreenHeader
          title="Lộ trình học"
          subtitle="Bắt đầu hành trình học tập cá nhân hóa"
        />
        <View style={{ flex: 1, justifyContent: 'center', gap: spacing.lg }}>
          <Card>
            <AppText variant="bodyLarge" weight="semibold">
              Chưa có lộ trình học
            </AppText>
            <View style={{ height: spacing.sm }} />
            <AppText variant="body" color={colors.textSecondary}>
              Tạo lộ trình để nhận các milestone phù hợp với mục tiêu của bạn.
            </AppText>
            <View style={{ height: spacing.lg }} />
            <Button
              title="Tạo lộ trình"
              loading={generateMutation.isPending}
              onPress={() => generateMutation.mutate()}
            />
            {generateMutation.isError ? (
              <>
                <View style={{ height: spacing.sm }} />
                <AppText variant="caption" color={colors.brandDanger}>
                  Tạo lộ trình thất bại. Vui lòng thử lại.
                </AppText>
              </>
            ) : null}
          </Card>
        </View>
      </Screen>
    );
  }

  return (
    <Screen scroll>
      <ScreenHeader
        title="Lộ trình học"
        subtitle="Theo dõi từng cột mốc của bạn"
      />
      <View style={{ gap: spacing.md }}>
        {milestones.map((milestone, index) => (
          <MilestoneCard
            key={milestone.id}
            milestone={milestone}
            index={index}
            onOpen={openLesson}
          />
        ))}
      </View>
    </Screen>
  );
}
