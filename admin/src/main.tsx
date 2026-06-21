import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Providers } from "@/app/providers";
import { App } from "@/app/App";
import { initMonitoring } from "@/lib/monitoring";
import "@/styles/globals.css";

// Fire-and-forget; guarded so it is a no-op when no Sentry DSN is configured.
void initMonitoring();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <Providers>
      <App />
    </Providers>
  </StrictMode>,
);
