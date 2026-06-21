// QuizResultScreen — score + per-wrong-answer explanations + navigation CTAs.
import { View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
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
import { quizApi } from '@/api/quiz';
import { colors, spacing } from '@/theme/tokens';
import type { QuizResult } from '@/api/types';
import type { AppStackParamList } from '@/navigation/types';

type Props = NativeStackScreenProps<AppStackParamList, 'QuizResult'>;

const PASS_THRESHOLD = 0.6;

function scoreColor(ratio: number): string {
  return ratio >= PASS_THRESHOLD ? colors.brandAccent : colors.brandWarning;
}

function headlineFor(ratio: number): string {
  return ratio >= PASS_THRESHOLD
    ? 'Làm tốt lắm! Bạn đã nắm chắc bài học.'
    : 'Cần ôn lại một chút. Đừng nản nhé!';
}

export default function QuizResultScreen({ route, navigation }: Props) {
  const { quizId } = route.params;

  const resultQuery = useQuery<QuizResult>({
    queryKey: ['quizzes', quizId, 'result'],
    queryFn: () => quizApi.getResult(quizId),
  });

  const goHome = () => navigation.navigate('Main', { screen: 'Dashboard' });
  const goRoadmap = () => navigation.navigate('Main', { screen: 'Roadmap' });

  if (resultQuery.isLoading) {
    return (
      <Screen>
        <ScreenHeader title="Kết quả quiz" />
        <StatePlaceholder kind="loading" />
      </Screen>
    );
  }

  if (resultQuery.isError || !resultQuery.data) {
    return (
      <Screen>
        <ScreenHeader title="Kết quả quiz" />
        <StatePlaceholder
          kind="error"
          message="Không tải được kết quả quiz."
          onRetry={() => resultQuery.refetch()}
        />
      </Screen>
    );
  }

  const result = resultQuery.data;
  const total = result.total || 1;
  const ratio = result.score / total;
  const pct = Math.round(ratio * 100);
  const accent = scoreColor(ratio);
  const wrongAnswers = result.answers.filter((a) => !a.is_correct);

  return (
    <Screen scroll>
      <ScreenHeader title="Kết quả quiz" />

      <Card style={{ alignItems: 'center', gap: spacing.lg }}>
        <ProgressRing progress={pct} color={accent} label="Chính xác" />
        <AppText variant="hero" weight="bold" display color={accent}>
          {`${result.score}/${result.total}`}
        </AppText>
        <AppText variant="body" color={colors.textSecondary} center>
          {headlineFor(ratio)}
        </AppText>
      </Card>

      {wrongAnswers.length > 0 ? (
        <>
          <View style={{ height: spacing.xl }} />
          <AppText variant="title" weight="bold" display>
            Câu cần xem lại
          </AppText>
          <View style={{ height: spacing.md }} />
          <View style={{ gap: spacing.md }}>
            {wrongAnswers.map((answer, index) => (
              <Card key={answer.question_id} soft>
                <AppText weight="semibold" color={colors.brandDanger}>
                  {`Câu sai ${index + 1}`}
                </AppText>
                {answer.explanation ? (
                  <>
                    <View style={{ height: spacing.sm }} />
                    <AppText variant="body" color={colors.textSecondary}>
                      {answer.explanation}
                    </AppText>
                  </>
                ) : (
                  <>
                    <View style={{ height: spacing.sm }} />
                    <AppText variant="body" color={colors.textMuted}>
                      Hãy ôn lại phần lý thuyết liên quan.
                    </AppText>
                  </>
                )}
              </Card>
            ))}
          </View>
        </>
      ) : (
        <>
          <View style={{ height: spacing.xl }} />
          <Card soft>
            <AppText weight="semibold" color={colors.brandAccent}>
              Tuyệt vời! Bạn trả lời đúng tất cả các câu.
            </AppText>
          </Card>
        </>
      )}

      <View style={{ height: spacing.xxl }} />

      <View style={{ gap: spacing.md }}>
        <Button title="Học lại bài" variant="secondary" onPress={goRoadmap} />
        <Button title="Về trang chủ" onPress={goHome} />
      </View>
    </Screen>
  );
}
