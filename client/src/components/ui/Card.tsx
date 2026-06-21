import { ReactNode } from 'react';
import { StyleProp, View, ViewStyle } from 'react-native';

import { colors, radius, shadowCard, spacing } from '@/theme/tokens';

type Props = {
  children: ReactNode;
  soft?: boolean;
  style?: StyleProp<ViewStyle>;
};

export function Card({ children, soft = false, style }: Props) {
  return (
    <View
      style={[
        {
          backgroundColor: soft ? colors.surfaceSoft : colors.surfaceCard,
          borderRadius: radius.card,
          padding: spacing.xl,
          borderWidth: 1,
          borderColor: colors.borderSoft,
        },
        soft ? null : shadowCard,
        style,
      ]}
    >
      {children}
    </View>
  );
}
