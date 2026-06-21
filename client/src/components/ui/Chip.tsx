import { Pressable } from 'react-native';

import { AppText } from '@/components/ui/AppText';
import { colors, hitTargetMin, radius, spacing } from '@/theme/tokens';

type Props = {
  label: string;
  selected?: boolean;
  onPress?: () => void;
  testID?: string;
};

export function Chip({ label, selected = false, onPress, testID }: Props) {
  return (
    <Pressable
      testID={testID}
      onPress={onPress}
      accessibilityRole="button"
      accessibilityState={{ selected }}
      style={{
        minHeight: hitTargetMin,
        justifyContent: 'center',
        borderRadius: radius.chip,
        paddingVertical: spacing.sm,
        paddingHorizontal: spacing.lg,
        backgroundColor: selected ? colors.brandPrimary : colors.surfaceSoft,
        borderWidth: 1,
        borderColor: selected ? colors.brandPrimary : colors.borderSoft,
      }}
    >
      <AppText
        weight={selected ? 'semibold' : 'medium'}
        color={selected ? colors.textOnBrand : colors.textSecondary}
      >
        {label}
      </AppText>
    </Pressable>
  );
}
