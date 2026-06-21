# AI Study Coach — Mobile App

React Native (Expo, TypeScript) client for the Edu platform. Implements the 13
Phase‑1 screens wired to the backend contract in `../server/docs/API.md`.

The product spec, screen map and HTML prototype live alongside this app
(`app_design.md`, `screen-map.md`, `prototype.html`, `copilot-setup/`).

## Stack

- Expo SDK 52 + React Native 0.76 + TypeScript (strict)
- React Navigation (native-stack + bottom-tabs)
- TanStack Query (server state) + Zustand (auth/session)
- Axios (envelope unwrap + single‑flight 401 refresh)
- expo-secure-store (encrypted token storage)
- @react-native-firebase/messaging (FCM push) + @sentry/react-native (crash reporting)
- EAS Build → Android App Bundle for Google Play

## Setup

```bash
cd client
npm install
cp .env.example .env   # set EXPO_PUBLIC_API_BASE_URL (Android emulator: http://10.0.2.2:8080/api/v1)
```

FCM requires native config not committed here: add `google-services.json`
(Android) / `GoogleService-Info.plist` (iOS) and set the Sentry org/EAS project
id in `app.json` before building a dev client.

## Run (dev client — MMKV/FCM need a custom client, not Expo Go)

```bash
npm run typecheck      # tsc --noEmit (CI gate)
npx expo prebuild      # generate native projects
npm run android        # or: npm run ios
```

## Build for Google Play

```bash
npx eas build --profile production --platform android   # → signed .aab
```

Profiles live in `eas.json` (development / preview / production), each with its
own `EXPO_PUBLIC_API_BASE_URL`.

## Structure

```
src/
  api/          axios client + per-domain endpoint modules + DTOs (types.ts)
  components/ui/ design-system primitives (Button, Card, ProgressRing, …)
  lib/          env, tokenStore (SecureStore), queryClient, monitoring (Sentry), push (FCM)
  navigation/   RootNavigator (auth → onboarding → app), stacks + bottom tabs
  screens/      13 screens grouped by flow (auth, onboarding, dashboard, …)
  store/        Zustand auth store
  theme/        design tokens (from copilot-setup/brand-tokens.css)
```

## Known follow-ups (confirm against backend)

- **Goal enum:** screens use goal-type constants; align with the server's goal schema values.
- **Diagnostic flow:** the diagnostic screen starts a placement test then reads
  results; presenting/submitting placement questions in-app is a planned extension.
- **Weak-topic accuracy scale:** rendered as a 0–1 fraction; confirm the API scale.
- **Reminder time / "today's tasks":** no dedicated API fields yet — derived/placeholder
  until the backend exposes them.
