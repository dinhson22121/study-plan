// Design tokens for AI Study Coach, derived from client/copilot-setup/brand-tokens.css.
// Keep this the single source of truth for colors, spacing, radius and type.

export const colors = {
  brandPrimary: '#3b5bff',
  brandPrimaryStrong: '#2745e8',
  brandSecondary: '#7c5cff',
  brandAccent: '#15b981',
  brandWarning: '#f5871f',
  brandDanger: '#f0474b',

  surfaceApp: '#f5f7fc',
  surfaceCard: '#ffffff',
  surfaceSoft: '#eef2ff',

  textPrimary: '#181b34',
  textSecondary: '#6b7186',
  textMuted: '#9ca2b8',
  textOnBrand: '#ffffff',

  borderSoft: 'rgba(24, 27, 52, 0.08)',
} as const;

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32,
} as const;

export const radius = {
  card: 20,
  button: 16,
  chip: 999,
} as const;

export const fontFamily = {
  // Loaded in App.tsx via @expo-google-fonts equivalents / expo-font.
  ui: 'PlusJakartaSans_400Regular',
  uiMedium: 'PlusJakartaSans_500Medium',
  uiSemiBold: 'PlusJakartaSans_600SemiBold',
  uiBold: 'PlusJakartaSans_700Bold',
  display: 'SpaceGrotesk_600SemiBold',
  displayBold: 'SpaceGrotesk_700Bold',
} as const;

export const fontSize = {
  caption: 12,
  body: 14,
  bodyLarge: 16,
  title: 20,
  hero: 28,
} as const;

export const hitTargetMin = 44;

export const shadowCard = {
  shadowColor: '#1c264c',
  shadowOffset: { width: 0, height: 10 },
  shadowOpacity: 0.12,
  shadowRadius: 16,
  elevation: 4,
} as const;

export const heroGradient = [colors.brandPrimary, colors.brandSecondary] as const;
