/* eslint-disable @typescript-eslint/no-empty-function */
// Mocks for native modules so unit tests run under jest-expo (node) without a device.

jest.mock('expo-secure-store', () => {
  const store: Record<string, string> = {};
  return {
    setItemAsync: jest.fn(async (k: string, v: string) => {
      store[k] = v;
    }),
    getItemAsync: jest.fn(async (k: string) => store[k] ?? null),
    deleteItemAsync: jest.fn(async (k: string) => {
      delete store[k];
    }),
  };
});

jest.mock('expo-constants', () => ({
  expoConfig: { extra: { apiBaseUrl: 'http://localhost/api/v1', sentryDsn: '' } },
}));

jest.mock('@react-native-firebase/messaging', () => () => ({
  requestPermission: jest.fn().mockResolvedValue(1),
  getToken: jest.fn().mockResolvedValue('fcm-test-token'),
  onTokenRefresh: jest.fn(() => () => {}),
}));

jest.mock('@sentry/react-native', () => ({
  init: jest.fn(),
  captureException: jest.fn(),
}));
