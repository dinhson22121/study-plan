import * as Sentry from '@sentry/react-native';

import { env } from '@/lib/env';

// Guarded init: a no-op when no DSN is configured (dev / local builds).
export function initMonitoring() {
  if (!env.sentryDsn) return;
  Sentry.init({
    dsn: env.sentryDsn,
    tracesSampleRate: 0.1,
    enableAutoSessionTracking: true,
  });
}

export function captureError(error: unknown) {
  if (!env.sentryDsn) return;
  Sentry.captureException(error);
}
