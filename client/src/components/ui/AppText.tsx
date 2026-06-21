import { Text, TextProps, TextStyle } from 'react-native';

import { colors, fontFamily, fontSize } from '@/theme/tokens';

type Variant = 'hero' | 'title' | 'body' | 'bodyLarge' | 'caption';
type Weight = 'regular' | 'medium' | 'semibold' | 'bold';

const variantStyle: Record<Variant, TextStyle> = {
  hero: { fontSize: fontSize.hero, color: colors.textPrimary },
  title: { fontSize: fontSize.title, color: colors.textPrimary },
  bodyLarge: { fontSize: fontSize.bodyLarge, color: colors.textPrimary },
  body: { fontSize: fontSize.body, color: colors.textPrimary },
  caption: { fontSize: fontSize.caption, color: colors.textSecondary },
};

const weightFamily: Record<Weight, string> = {
  regular: fontFamily.ui,
  medium: fontFamily.uiMedium,
  semibold: fontFamily.uiSemiBold,
  bold: fontFamily.uiBold,
};

type Props = TextProps & {
  variant?: Variant;
  weight?: Weight;
  color?: string;
  display?: boolean;
  center?: boolean;
};

export function AppText({
  variant = 'body',
  weight = 'regular',
  color,
  display = false,
  center = false,
  style,
  ...rest
}: Props) {
  return (
    <Text
      {...rest}
      style={[
        variantStyle[variant],
        {
          fontFamily: display
            ? weight === 'bold'
              ? fontFamily.displayBold
              : fontFamily.display
            : weightFamily[weight],
        },
        color ? { color } : null,
        center ? { textAlign: 'center' } : null,
        style,
      ]}
    />
  );
}
