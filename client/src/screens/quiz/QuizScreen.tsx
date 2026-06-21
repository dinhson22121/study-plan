// QuizScreen — used as both the 'QuizTab' tab and the stack 'Quiz' screen.
// Reads an optional topicId via useRoute; navigates to QuizResult on submit.
import { useState } from 'react';
import { View } from 'react-native';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigation, useRoute } from '@react-navigation/native';
import type { RouteProp } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';

import {
  AppText,
  Button,
  Card,
  Chip,
  Screen,
  ScreenHeader,
  StatePlaceholder,
} from '@/components/ui';
import { quizApi } from '@/api/quiz';
import { colors, spacing } from '@/theme/tokens';
import type { Quiz } from '@/api/types';
import type { AppStackParamList } from '@/navigation/types';

type QuizRoute = RouteProp<AppStackParamList, 'Quiz'>;

type Answers = Record<string, string>;

function QuizQuestions({ quiz }: { quiz: Quiz }) {
  const navigation =
    useNavigation<NativeStackNavigationProp<AppStackParamList>>();
  const queryClient = useQueryClient();
  const [answers, setAnswers] = useState<Answers>({});

  const submitMutation = useMutation({
    mutationFn: () => {
      const payload = quiz.questions.map((question) => ({
        question_id: question.id,
        selected_option_id: answers[question.id] ?? '',
      }));
      return quizApi.submit(quiz.id, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['progress'] });
      queryClient.invalidateQueries({ queryKey: ['quizzes'] });
      navigation.replace('QuizResult', { quizId: quiz.id });
    },
  });

  const select = (questionId: string, optionId: string) =>
    setAnswers((prev) => ({ ...prev, [questionId]: optionId }));

  const answeredCount = quiz.questions.filter((q) => answers[q.id]).length;
  const allAnswered = answeredCount === quiz.questions.length;

  return (
    <Screen scroll>
      <ScreenHeader
        title="Quiz kiểm tra"
        subtitle={`Đã trả lời ${answeredCount}/${quiz.questions.length} câu`}
      />

      <View style={{ gap: spacing.lg }}>
        {quiz.questions.map((question, index) => (
          <Card key={question.id}>
            <AppText variant="bodyLarge" weight="semibold">
              {`Câu ${index + 1}. ${question.stem}`}
            </AppText>
            <View style={{ height: spacing.md }} />
            <View style={{ gap: spacing.sm }}>
              {question.options.map((option) => (
                <Chip
                  key={option.id}
                  label={`${option.label}. ${option.text}`}
                  selected={answers[question.id] === option.id}
                  onPress={() => select(question.id, option.id)}
                />
              ))}
            </View>
          </Card>
        ))}
      </View>

      <View style={{ height: spacing.xxl }} />

      <Button
        title="Nộp bài"
        disabled={!allAnswered}
        loading={submitMutation.isPending}
        onPress={() => submitMutation.mutate()}
      />
      {submitMutation.isError ? (
        <>
          <View style={{ height: spacing.sm }} />
          <AppText variant="caption" color={colors.brandDanger} center>
            Nộp bài thất bại. Vui lòng thử lại.
          </AppText>
        </>
      ) : null}
    </Screen>
  );
}

export default function QuizScreen() {
  const route = useRoute<QuizRoute>();
  const topicId = route.params?.topicId;
  const [quiz, setQuiz] = useState<Quiz | null>(null);

  const startMutation = useMutation({
    mutationFn: () => quizApi.start(topicId),
    onSuccess: (data) => setQuiz(data),
  });

  if (startMutation.isPending) {
    return (
      <Screen>
        <ScreenHeader title="Quiz kiểm tra" />
        <StatePlaceholder kind="loading" message="Đang tạo bộ câu hỏi…" />
      </Screen>
    );
  }

  if (startMutation.isError) {
    return (
      <Screen>
        <ScreenHeader title="Quiz kiểm tra" />
        <StatePlaceholder
          kind="error"
          message="Không tạo được quiz. Vui lòng thử lại."
          onRetry={() => startMutation.mutate()}
        />
      </Screen>
    );
  }

  if (quiz) {
    if (quiz.questions.length === 0) {
      return (
        <Screen>
          <ScreenHeader title="Quiz kiểm tra" />
          <StatePlaceholder
            kind="empty"
            message="Quiz này chưa có câu hỏi nào."
          />
        </Screen>
      );
    }
    return <QuizQuestions quiz={quiz} />;
  }

  return (
    <Screen>
      <ScreenHeader
        title="Quiz kiểm tra"
        subtitle="Kiểm tra nhanh kiến thức của bạn"
      />
      <View style={{ flex: 1, justifyContent: 'center', gap: spacing.lg }}>
        <Card>
          <AppText variant="bodyLarge" weight="semibold">
            Sẵn sàng làm quiz?
          </AppText>
          <View style={{ height: spacing.sm }} />
          <AppText variant="body" color={colors.textSecondary}>
            Một vài câu trắc nghiệm ngắn giúp bạn củng cố kiến thức ngay.
          </AppText>
          <View style={{ height: spacing.lg }} />
          <Button title="Bắt đầu" onPress={() => startMutation.mutate()} />
        </Card>
      </View>
    </Screen>
  );
}
