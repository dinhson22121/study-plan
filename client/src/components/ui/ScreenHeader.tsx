import { Pressable, View } from 'react-native';

import { AppText } from '@/components/ui/AppText';
import { colors, hitTargetMin, spacing } from '@/theme/tokens';

type Props = {
  title: string;
  subtitle?: string;
  onBack?: () => void;
  rightIcon?: string;
  onRightPress?: () => void;
  rightBadge?: boolean;
};

export function ScreenHeader({
  title,
  subtitle,
  onBack,
  rightIcon,
  onRightPress,
  rightBadge = false,
}: Props) {
  return (
    <View
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: spacing.lg,
      }}
    >
      <View style={{ flexDirection: 'row', alignItems: 'center', flex: 1, gap: spacing.sm }}>
        {onBack ? (
          <Pressable
            onPress={onBack}
            accessibilityRole="button"
            accessibilityLabel="Quay lại"
            style={{ width: hitTargetMin, height: hitTargetMin, justifyContent: 'center' }}
          >
            <AppText variant="title">‹</AppText>
          </Pressable>
        ) : null}
        <View style={{ flex: 1 }}>
          <AppText variant="title" weight="bold" display>
            {title}
          </AppText>
          {subtitle ? (
            <AppText variant="caption" color={colors.textMuted}>
              {subtitle}
            </AppText>
          ) : null}
        </View>
      </View>
      {rightIcon ? (
        <Pressable
          onPress={onRightPress}
          accessibilityRole="button"
          style={{ width: hitTargetMin, height: hitTargetMin, alignItems: 'center', justifyContent: 'center' }}
        >
          <AppText variant="title">{rightIcon}</AppText>
          {rightBadge ? (
            <View
              style={{
                position: 'absolute',
                top: 8,
                right: 8,
                width: 10,
                height: 10,
                borderRadius: 5,
                backgroundColor: colors.brandDanger,
              }}
            />
          ) : null}
        </Pressable>
      ) : null}
    </View>
  );
}
