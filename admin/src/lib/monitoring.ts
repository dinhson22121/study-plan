import { env } from "./env";

interface SentryLike {
  init: (options: Record<string, unknown>) => void;
}

function isSentryLike(mod: unknown): mod is SentryLike {
  return typeof mod === "object" && mod !== null && typeof (mod as SentryLike).init === "function";
}

/**
 * Optional, lightweight error monitoring.
 *
 * Wiring is fully guarded: when `VITE_SENTRY_DSN` is unset this is a no-op and
 * pulls in zero extra code. When a DSN is present, `@sentry/browser` is loaded
 * dynamically so the dependency stays optional — if it is not installed the
 * failure is swallowed and the app continues to run unmonitored.
 */
export async function initMonitoring(): Promise<void> {
  const dsn = env.sentryDsn;
  if (!dsn) return;

  try {
    // Indirect specifier so TypeScript/Vite do not require the optional package
    // to be present at build time; resolved at runtime only when a DSN is set.
    const specifier = "@sentry/browser";
    const mod: unknown = await import(/* @vite-ignore */ specifier).catch(() => null);
    if (!isSentryLike(mod)) return;
    mod.init({
      dsn,
      environment: env.appEnv,
      tracesSampleRate: 0,
    });
  } catch {
    // Never let monitoring setup break app startup.
  }
}
