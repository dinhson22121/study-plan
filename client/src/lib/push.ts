import { Platform } from 'react-native';
import messaging from '@react-native-firebase/messaging';

import { notificationApi } from '@/api/notification';
import { captureError } from '@/lib/monitoring';

// Requests push permission, obtains the FCM token, and registers it with the
// backend. Safe to call after login; failures are non-fatal (logged to Sentry).
export async function registerForPush(): Promise<void> {
  try {
    const authStatus = await messaging().requestPermission();
    const granted =
      authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
      authStatus === messaging.AuthorizationStatus.PROVISIONAL;
    if (!granted) return;

    const token = await messaging().getToken();
    if (token) {
      await notificationApi.registerDevice(token, Platform.OS);
    }
  } catch (err) {
    captureError(err);
  }
}

export async function unregisterPush(): Promise<void> {
  try {
    const token = await messaging().getToken();
    if (token) await notificationApi.deleteDevice(token);
  } catch {
    // best-effort
  }
}

// Refresh the registered token whenever FCM rotates it.
export function listenForTokenRefresh(): () => void {
  return messaging().onTokenRefresh((token) => {
    notificationApi.registerDevice(token, Platform.OS).catch(captureError);
  });
}
