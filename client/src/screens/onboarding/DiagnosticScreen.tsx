// DiagnosticScreen — intro to the placement test, then starts it.
import { View } from 'react-native';
import { useMutation, useQuery } from '@tanstack/react-query';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { AppText, Button, Card, Screen } from '@/components/ui';
import { ApiError } from '@/api/client';
import { curriculumApi } from '@/api/curriculum';
import { placementApi } from '@/api/placement';
import { colors, spacing } from '@/theme/tokens';
import type { OnboardingStackParamList } from '@/navigation/types';

type Props = NativeStackScreenProps<OnboardingStackParamList, 'Diagnostic'>;

// Shown when the curriculum endpoint is empty or unavailable.
const FALLBACK_SUBJECTS = ['Toán', 'Tiếng Anh', 'Vật Lý'];

export default function DiagnosticScreen({ navigation }: Props) {
  const subjectsQuery = useQuery({
    queryKey: ['curriculum', 'subjects'],
    queryFn: () => curriculumApi.listSubjects(),
  });

  const fetched = subjectsQuery.data?.map((s) => s.name) ?? [];
  const subjectNames = fetched.length > 0 ? fetched : FALLBACK_SUBJECTS;

  const startMutation = useMutation({
    mutationFn: () => placementApi.startTest(),
    onSuccess: () => navigation.navigate('DiagnosticResult'),
  });

  return (
    <Screen scroll>
      <View style={{ gap: spacing.xs, marginBottom: spacing.xxl }}>
        <AppText variant="hero" weight="bold" display>
          Chẩn đoán đầu vào
        </AppText>
        <AppText variant="body" color={colors.textSecondary}>
          Một bài kiểm tra ngắn giúp chúng tôi hiểu trình độ hiện tại của bạn ở từng môn.
        </AppText>
      </View>

      <Card style={{ gap: spacing.md, marginBottom: spacing.xxl }}>
        <AppText variant="bodyLarge" weight="semibold">
          Các môn sẽ được đánh giá
        </AppText>
        {subjectsQuery.isLoading ? (
          <AppText variant="body" color={colors.textMuted}>
            Đang tải danh sách môn…
          </AppText>
        ) : (
          <View style={{ gap: spacing.sm }}>
            {subjectNames.map((name) => (
              <View
                key={name}
                style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}
              >
                <View
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: 4,
                    backgroundColor: colors.brandPrimary,
                  }}
                />
                <AppText variant="bodyLarge">{name}</AppText>
              </View>
            ))}
          </View>
        )}
      </Card>

      {startMutation.isError ? (
        <AppText variant="body" weight="medium" color={colors.brandDanger} style={{ marginBottom: spacing.md }}>
          {startMutation.error instanceof ApiError
            ? startMutation.error.message
            : 'Không bắt đầu được bài chẩn đoán. Vui lòng thử lại.'}
        </AppText>
      ) : null}

      <Button
        title="Bắt đầu bài chẩn đoán"
        onPress={() => startMutation.mutate()}
        loading={startMutation.isPending}
        disabled={startMutation.isPending}
        testID="diagnostic-start"
      />
    </Screen>
  );
}
