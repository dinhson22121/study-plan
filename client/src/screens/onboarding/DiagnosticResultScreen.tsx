// DiagnosticResultScreen — per-subject insight; enters the main app on continue.
import { View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { AppText, Button, Card, Screen, StatePlaceholder } from '@/components/ui';
import { placementApi } from '@/api/placement';
import { useAuthStore } from '@/store/authStore';
import { colors, spacing } from '@/theme/tokens';
import type { PlacementResult } from '@/api/types';
import type { OnboardingStackParamList } from '@/navigation/types';

type Props = NativeStackScreenProps<OnboardingStackParamList, 'DiagnosticResult'>;

function levelColor(level?: string): string {
  switch ((level ?? '').toUpperCase()) {
    case 'STRONG':
    case 'GOOD':
      return colors.brandAccent;
    case 'WEAK':
    case 'NEEDS_WORK':
      return colors.brandWarning;
    default:
      return colors.brandPrimary;
  }
}

function ResultCard({ result }: { result: PlacementResult }) {
  const title = result.subject_name ?? 'Môn học';
  return (
    <Card style={{ gap: spacing.sm }}>
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
        <AppText variant="bodyLarge" weight="semibold">
          {title}
        </AppText>
        {result.level ? (
          <AppText variant="caption" weight="semibold" color={levelColor(result.level)}>
            {result.level}
          </AppText>
        ) : null}
      </View>
      <AppText variant="body" color={colors.textSecondary}>
        {result.summary ?? 'Chưa có nhận xét chi tiết cho môn này.'}
      </AppText>
    </Card>
  );
}

export default function DiagnosticResultScreen(_props: Props) {
  const setOnboarded = useAuthStore((s) => s.setOnboarded);

  const resultsQuery = useQuery({
    queryKey: ['placement', 'results'],
    queryFn: () => placementApi.listResults(),
  });

  const goToDashboard = () => {
    // Flip the onboarded flag so RootNavigator swaps to the main app stack.
    setOnboarded(true);
  };

  if (resultsQuery.isLoading) {
    return (
      <Screen>
        <StatePlaceholder kind="loading" message="Đang phân tích kết quả của bạn…" />
      </Screen>
    );
  }

  if (resultsQuery.isError) {
    return (
      <Screen>
        <StatePlaceholder
          kind="error"
          title="Không tải được kết quả"
          message="Vui lòng thử lại để xem nhận xét theo từng môn."
          onRetry={() => void resultsQuery.refetch()}
        />
      </Screen>
    );
  }

  const results = resultsQuery.data ?? [];

  return (
    <Screen scroll>
      <View style={{ gap: spacing.xs, marginBottom: spacing.xxl }}>
        <AppText variant="hero" weight="bold" display>
          Kết quả chẩn đoán
        </AppText>
        <AppText variant="body" color={colors.textSecondary}>
          Đây là điểm mạnh và phần cần củng cố của bạn.
        </AppText>
      </View>

      {results.length === 0 ? (
        <View style={{ minHeight: 220 }}>
          <StatePlaceholder
            kind="empty"
            title="Chưa có nhận xét"
            message="Chúng tôi sẽ cập nhật nhận xét sau khi bạn hoàn thành bài chẩn đoán."
          />
        </View>
      ) : (
        <View style={{ gap: spacing.md, marginBottom: spacing.xxl }}>
          {results.map((result) => (
            <ResultCard key={result.id} result={result} />
          ))}
        </View>
      )}

      <Button title="Xem dashboard" onPress={goToDashboard} testID="diagnostic-result-continue" />
    </Screen>
  );
}
