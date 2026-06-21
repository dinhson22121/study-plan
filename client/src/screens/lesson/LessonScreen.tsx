// LessonScreen — show a lesson (by lessonId or first lesson of a topic) + quiz CTA.
import { View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import {
  AppText,
  Button,
  Card,
  Screen,
  ScreenHeader,
  StatePlaceholder,
} from '@/components/ui';
import { contentApi } from '@/api/content';
import { colors, spacing } from '@/theme/tokens';
import type { Lesson } from '@/api/types';
import type { AppStackParamList } from '@/navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Lesson'>;

async function loadLesson(
  lessonId?: string,
  topicId?: string,
): Promise<Lesson | null> {
  if (lessonId) {
    return contentApi.get(lessonId);
  }
  if (topicId) {
    const lessons = await contentApi.listByTopic(topicId);
    return lessons[0] ?? null;
  }
  return null;
}

export default function LessonScreen({ route, navigation }: Props) {
  const { lessonId, topicId } = route.params ?? {};

  const lessonQuery = useQuery<Lesson | null>({
    queryKey: ['lesson', lessonId ?? null, topicId ?? null],
    queryFn: () => loadLesson(lessonId, topicId),
    enabled: Boolean(lessonId || topicId),
  });

  const goBack = () => navigation.goBack();

  if (!lessonId && !topicId) {
    return (
      <Screen>
        <ScreenHeader title="Bài học" onBack={goBack} />
        <StatePlaceholder
          kind="error"
          message="Không xác định được bài học cần mở."
        />
      </Screen>
    );
  }

  if (lessonQuery.isLoading) {
    return (
      <Screen>
        <ScreenHeader title="Bài học" onBack={goBack} />
        <StatePlaceholder kind="loading" />
      </Screen>
    );
  }

  if (lessonQuery.isError) {
    return (
      <Screen>
        <ScreenHeader title="Bài học" onBack={goBack} />
        <StatePlaceholder
          kind="error"
          message="Không tải được nội dung bài học."
          onRetry={() => lessonQuery.refetch()}
        />
      </Screen>
    );
  }

  const lesson = lessonQuery.data;

  if (!lesson) {
    return (
      <Screen>
        <ScreenHeader title="Bài học" onBack={goBack} />
        <StatePlaceholder
          kind="empty"
          message="Chủ đề này chưa có bài học nào."
        />
      </Screen>
    );
  }

  const summaryPoints = lesson.summary_points ?? [];
  const quizTopicId = lesson.topic_id ?? topicId;

  return (
    <Screen scroll>
      <ScreenHeader title="Bài học" subtitle={lesson.title} onBack={goBack} />

      <AppText variant="hero" weight="bold" display>
        {lesson.title}
      </AppText>

      {summaryPoints.length > 0 ? (
        <>
          <View style={{ height: spacing.lg }} />
          <Card soft>
            <AppText variant="bodyLarge" weight="semibold">
              Ý chính
            </AppText>
            <View style={{ height: spacing.md }} />
            <View style={{ gap: spacing.sm }}>
              {summaryPoints.map((point, index) => (
                <View
                  key={`${index}-${point}`}
                  style={{ flexDirection: 'row', gap: spacing.sm }}
                >
                  <AppText weight="bold" color={colors.brandPrimary}>
                    {`${index + 1}.`}
                  </AppText>
                  <AppText style={{ flex: 1 }} color={colors.textSecondary}>
                    {point}
                  </AppText>
                </View>
              ))}
            </View>
          </Card>
        </>
      ) : null}

      {lesson.body ? (
        <>
          <View style={{ height: spacing.lg }} />
          <Card>
            <AppText variant="body" color={colors.textPrimary} style={{ lineHeight: 22 }}>
              {lesson.body}
            </AppText>
          </Card>
        </>
      ) : null}

      <View style={{ height: spacing.xxl }} />

      <Button
        title="Làm quiz kiểm tra"
        onPress={() => navigation.navigate('Quiz', { topicId: quizTopicId })}
      />
    </Screen>
  );
}
