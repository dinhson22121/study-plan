export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1",
  appName: import.meta.env.VITE_APP_NAME ?? "Edu Admin",
  pollIntervalMs: Number(import.meta.env.VITE_POLL_INTERVAL_MS ?? 5000),
  uploadMaxFileSizeBytes: Number(import.meta.env.VITE_UPLOAD_MAX_FILE_SIZE_BYTES ?? 20 * 1024 * 1024),
};
