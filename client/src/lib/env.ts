import Constants from 'expo-constants';

// Resolution order: EXPO_PUBLIC_* env (EAS build profiles) → app.json `extra` → dev default.
const extra = (Constants.expoConfig?.extra ?? {}) as {
  apiBaseUrl?: string;
  sentryDsn?: string;
};

export const env = {
  // 10.0.2.2 is the Android emulator alias for the host machine's localhost.
  apiBaseUrl:
    process.env.EXPO_PUBLIC_API_BASE_URL ??
    extra.apiBaseUrl ??
    'http://10.0.2.2:8080/api/v1',
  sentryDsn: process.env.EXPO_PUBLIC_SENTRY_DSN ?? extra.sentryDsn ?? '',
} as const;
