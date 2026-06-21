import { ReactNode } from 'react';
import { ScrollView, StyleProp, View, ViewStyle } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { colors, spacing } from '@/theme/tokens';

type Props = {
  children: ReactNode;
  scroll?: boolean;
  padded?: boolean;
  background?: string;
  contentStyle?: StyleProp<ViewStyle>;
};

export function Screen({
  children,
  scroll = false,
  padded = true,
  background = colors.surfaceApp,
  contentStyle,
}: Props) {
  const inner: StyleProp<ViewStyle> = [
    padded ? { padding: spacing.xl } : null,
    contentStyle,
  ];

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: background }} edges={['top']}>
      {scroll ? (
        <ScrollView
          contentContainerStyle={[{ flexGrow: 1 }, inner]}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          {children}
        </ScrollView>
      ) : (
        <View style={[{ flex: 1 }, inner]}>{children}</View>
      )}
    </SafeAreaView>
  );
}
