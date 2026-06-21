import { ActivityIndicator, View } from 'react-native';

import { AppText } from '@/components/ui/AppText';
import { Button } from '@/components/ui/Button';
import { colors, spacing } from '@/theme/tokens';

type Props = {
  kind: 'loading' | 'empty' | 'error';
  title?: string;
  message?: string;
  onRetry?: () => void;
};

const defaults = {
  loading: { title: 'Đang tải…', message: '' },
  empty: { title: 'Chưa có dữ liệu', message: 'Nội dung sẽ xuất hiện ở đây.' },
  error: { title: 'Đã có lỗi xảy ra', message: 'Vui lòng thử lại.' },
} as const;

export function StatePlaceholder({ kind, title, message, onRetry }: Props) {
  const d = defaults[kind];
  return (
    <View
      style={{
        flex: 1,
        alignItems: 'center',
        justifyContent: 'center',
        padding: spacing.xl,
        gap: spacing.md,
      }}
    >
      {kind === 'loading' ? (
        <ActivityIndicator size="large" color={colors.brandPrimary} />
      ) : null}
      <AppText variant="bodyLarge" weight="semibold" center>
        {title ?? d.title}
      </AppText>
      {(message ?? d.message) ? (
        <AppText variant="body" color={colors.textSecondary} center>
          {message ?? d.message}
        </AppText>
      ) : null}
      {kind === 'error' && onRetry ? (
        <Button title="Thử lại" variant="secondary" onPress={onRetry} />
      ) : null}
    </View>
  );
}
