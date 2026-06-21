import * as SecureStore from 'expo-secure-store';

// Auth tokens live in the OS keystore/keychain (encrypted at rest) rather than
// MMKV/AsyncStorage, satisfying the "secure token storage" requirement.

const ACCESS_KEY = 'auth.access_token';
const REFRESH_KEY = 'auth.refresh_token';

export type TokenPair = {
  accessToken: string;
  refreshToken: string;
};

export async function saveTokens(pair: TokenPair): Promise<void> {
  await Promise.all([
    SecureStore.setItemAsync(ACCESS_KEY, pair.accessToken),
    SecureStore.setItemAsync(REFRESH_KEY, pair.refreshToken),
  ]);
}

export async function getAccessToken(): Promise<string | null> {
  return SecureStore.getItemAsync(ACCESS_KEY);
}

export async function getRefreshToken(): Promise<string | null> {
  return SecureStore.getItemAsync(REFRESH_KEY);
}

export async function clearTokens(): Promise<void> {
  await Promise.all([
    SecureStore.deleteItemAsync(ACCESS_KEY),
    SecureStore.deleteItemAsync(REFRESH_KEY),
  ]);
}
