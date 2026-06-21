import {
  ActivityIndicator,
  Pressable,
  StyleProp,
  ViewStyle,
} from 'react-native';

import { AppText } from '@/components/ui/AppText';
import { colors, hitTargetMin, radius, spacing } from '@/theme/tokens';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';

type Props = {
  title: string;
  onPress: () => void;
  variant?: Variant;
  loading?: boolean;
  disabled?: boolean;
  style?: StyleProp<ViewStyle>;
  testID?: string;
};

const bg: Record<Variant, string> = {
  primary: colors.brandPrimary,
  secondary: colors.surfaceSoft,
  ghost: 'transparent',
  danger: colors.brandDanger,
};

const fg: Record<Variant, string> = {
  primary: colors.textOnBrand,
  secondary: colors.brandPrimary,
  ghost: colors.brandPrimary,
  danger: colors.textOnBrand,
};

export function Button({
  title,
  onPress,
  variant = 'primary',
  loading = false,
  disabled = false,
  style,
  testID,
}: Props) {
  const isDisabled = disabled || loading;
  return (
    <Pressable
      testID={testID}
      onPress={onPress}
      disabled={isDisabled}
      accessibilityRole="button"
      accessibilityState={{ disabled: isDisabled, busy: loading }}
      style={({ pressed }) => [
        {
          minHeight: hitTargetMin,
          borderRadius: radius.button,
          paddingVertical: spacing.md,
          paddingHorizontal: spacing.xl,
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: bg[variant],
          opacity: isDisabled ? 0.5 : pressed ? 0.85 : 1,
        },
        style,
      ]}
    >
      {loading ? (
        <ActivityIndicator color={fg[variant]} />
      ) : (
        <AppText weight="semibold" color={fg[variant]}>
          {title}
        </AppText>
      )}
    </Pressable>
  );
}
