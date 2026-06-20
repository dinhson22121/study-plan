# Edu Admin (apps/admin-web)

React 19 + Vite + TypeScript admin web app for the edu-app backend: login, upload
exam PDFs to S3/MinIO via presigned URLs, track parse jobs, review/edit parsed
question drafts, and publish them into the question bank.

## Stack
React 19 · Vite 6 · TypeScript · React Router 7 · TanStack Query 5 ·
React Hook Form + Zod · Axios · Tailwind CSS 4 (hand-rolled UI primitives).

## Setup
```bash
cp .env.example .env   # adjust VITE_API_BASE_URL if needed
npm install
npm run dev            # http://localhost:5173
```
Requires the backend running (default `http://localhost:8080/api/v1`) and an
ADMIN account. The app blocks non-ADMIN users.

## Scripts
- `npm run dev` — dev server
- `npm run build` — typecheck (`tsc --noEmit`) + production build to `dist/`
- `npm run preview` — serve the production build
- `npm run typecheck` — types only
- `npm run test` — watch mode unit tests (Vitest)
- `npm run test:run` — run unit tests once

## Structure
```
src/
  api/         typed client (axios + envelope unwrap + 401 refresh) + endpoints
  auth/        token store, AuthContext, ProtectedRoute (role=ADMIN)
  components/  ui/ (Tailwind primitives), layout/, status badges, TopicSelector
  hooks/       TanStack Query hooks (assets, parse jobs polling, drafts, curriculum)
  pages/       Login, Dashboard, Uploads, UploadNew, UploadDetail, DraftReview
  lib/         env, jwt decode, sha256 checksum, utils
```

## Flow
login → upload PDF (browser SHA-256 → init → PUT to storage → complete) →
asset detail (parse jobs poll every 5s, retry, link entity) →
draft review (edit stem/options/answer, pick topic + difficulty) →
publish single or publish-all.
