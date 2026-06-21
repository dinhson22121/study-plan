// NotificationsScreen — history of reminders, milestones and new-quiz alerts.
import { View } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import {
  AppText,
  Card,
  Screen,
  ScreenHeader,
  StatePlaceholder,
} from '@/components/ui';
import { notificationApi } from '@/api/notification';
import { colors, spacing } from '@/theme/tokens';
import type { AppStackParamList } from '@/navigation/types';
import type { NotificationItem } from '@/api/types';

type Props = NativeStackScreenProps<AppStackParamList, 'Notifications'>;

const MINUTE_MS = 60 * 1000;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;

function relativeTime(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return '';
  const diff = Date.now() - then;
  if (diff < MINUTE_MS) return 'Vừa xong';
  if (diff < HOUR_MS) return `${Math.floor(diff / MINUTE_MS)} phút trước`;
  if (diff < DAY_MS) return `${Math.floor(diff / HOUR_MS)} giờ trước`;
  return `${Math.floor(diff / DAY_MS)} ngày trước`;
}

function NotificationRow({ item }: { item: NotificationItem }) {
  return (
    <Card soft style={{ flexDirection: 'row', gap: spacing.md, alignItems: 'flex-start' }}>
      <View
        style={{
          width: 10,
          height: 10,
          borderRadius: 5,
          marginTop: spacing.xs,
          backgroundColor: item.read ? 'transparent' : colors.brandPrimary,
        }}
        accessibilityLabel={item.read ? undefined : 'Chưa đọc'}
      />
      <View style={{ flex: 1, gap: spacing.xs }}>
        <AppText variant="body" weight={item.read ? 'medium' : 'semibold'}>
          {item.title}
        </AppText>
        <AppText variant="body" color={colors.textSecondary}>
          {item.body}
        </AppText>
        <AppText variant="caption" color={colors.textMuted}>
          {relativeTime(item.created_at)}
        </AppText>
      </View>
    </Card>
  );
}

export default function NotificationsScreen({ navigation }: Props) {
  const historyQuery = useQuery<NotificationItem[]>({
    queryKey: ['notifications', 'history'],
    queryFn: () => notificationApi.history(),
  });

  const renderBody = () => {
    if (historyQuery.isLoading) {
      return <StatePlaceholder kind="loading" />;
    }
    if (historyQuery.isError) {
      return (
        <StatePlaceholder
          kind="error"
          message="Không tải được thông báo."
          onRetry={() => historyQuery.refetch()}
        />
      );
    }
    const items = historyQuery.data ?? [];
    if (items.length === 0) {
      return (
        <StatePlaceholder
          kind="empty"
          title="Chưa có thông báo"
          message="Các nhắc học và cập nhật mới sẽ xuất hiện ở đây."
        />
      );
    }
    return (
      <View style={{ gap: spacing.md }}>
        {items.map((item) => (
          <NotificationRow key={item.id} item={item} />
        ))}
      </View>
    );
  };

  return (
    <Screen scroll>
      <ScreenHeader title="Thông báo" onBack={() => navigation.goBack()} />
      {renderBody()}
    </Screen>
  );
}
