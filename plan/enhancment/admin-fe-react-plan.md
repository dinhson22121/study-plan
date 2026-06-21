# Admin FE Plan — React Latest Stable

## Mục tiêu

Xây dựng một **admin web app** riêng để vận hành luồng:

- đăng nhập admin
- upload PDF đề bài lên S3/MinIO qua presigned URL
- theo dõi parse job
- review/sửa draft questions
- publish từng draft hoặc publish theo asset
- link asset vào `QUESTION` / `EXAM` / `CONTENT`

Plan này bám theo backend hiện có trong repo. Hiện tại backend đã có API cho luồng admin, nhưng **chưa có frontend admin**.

---

## 1. Quyết định công nghệ

### 1.1 Stack chính

- **React latest stable**: **React 19**
- **TypeScript**
- **Vite** để scaffold/build
- **React Router** cho routing SPA
- **TanStack Query** cho data fetching, cache, polling
- **React Hook Form + Zod** cho form và validation
- **Tailwind CSS 4** cho styling
- **shadcn/ui** cho component library admin
- **Axios** cho HTTP client và auth interceptor

### 1.2 Vì sao chọn stack này

- React + Vite nhanh để dựng admin dashboard độc lập
- TypeScript giúp bắt lỗi contract sớm khi tích hợp với backend Go
- TanStack Query hợp với flow polling parse job và refresh data
- React Hook Form + Zod giảm lỗi form upload/review
- Tailwind + shadcn/ui giúp ra giao diện admin nhanh, ít phụ thuộc design system sớm

---

## 2. Vị trí trong repo

Khuyến nghị tạo app FE riêng tại:

```text
apps/
  admin-web/
```

Không embed vào Gin ở phase đầu.

### Lý do

- tách lifecycle FE và BE
- frontend build/deploy độc lập
- dễ dùng local dev song song với backend hiện tại
- tránh làm backend Go phức tạp vì static asset pipeline

---

## 3. Backend readiness hiện tại

Backend hiện đã có các API đủ để FE admin bắt đầu:

### Auth

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`

### Upload / parse

- `POST /api/v1/admin/uploads/init`
- `POST /api/v1/admin/uploads/complete`
- `GET /api/v1/admin/uploads`
- `GET /api/v1/admin/uploads/:id`
- `POST /api/v1/admin/uploads/:id/parse`
- `POST /api/v1/admin/uploads/:id/link`
- `GET /api/v1/admin/uploads/:id/parse-jobs`
- `GET /api/v1/admin/uploads/:id/draft-questions`
- `DELETE /api/v1/admin/uploads/:id`

### Draft review / publish

- `PUT /api/v1/admin/question-drafts/:id`
- `PUT /api/v1/admin/question-drafts/:id/options/:optionId`
- `POST /api/v1/admin/question-drafts/:id/publish`
- `POST /api/v1/admin/uploads/:id/publish`

### Curriculum / topic picker

- `GET /api/v1/curriculum/subjects`
- `GET /api/v1/curriculum/subjects/:id/chapters`
- `GET /api/v1/curriculum/chapters/:id/topics`
- `GET /api/v1/curriculum/topics/:id`

### Question bank

- `GET /api/v1/questions`
- `GET /api/v1/questions/:id`
- `POST /api/v1/questions`

---

## 4. Scope của admin FE MVP

## 4.1 In scope

1. Admin login
2. Asset upload bằng file PDF
3. Theo dõi danh sách asset và parse status
4. Xem chi tiết asset
5. Xem draft questions theo asset
6. Sửa stem / explanation / options / correct answer
7. Publish từng draft
8. Publish toàn bộ draft của asset
9. Link asset vào entity
10. Chọn `topic_id` và `difficulty` khi publish

## 4.2 Out of scope ở phase đầu

- OCR preview
- drag-drop reorder phức tạp
- real-time collaboration
- advanced analytics dashboard
- batch import nhiều file cùng lúc
- exam builder hoàn chỉnh

---

## 5. Luồng người dùng chính

## 5.1 Admin login

1. Admin nhập email/password
2. FE gọi `POST /auth/login`
3. Nhận `access_token`, `refresh_token`, `expires_at`
4. FE decode JWT để đọc `role`
5. Nếu role != `ADMIN` thì chặn truy cập
6. Vào dashboard

## 5.2 Upload PDF

1. Admin chọn file PDF
2. FE tính `checksum_sha256` trên browser
3. FE gọi `POST /admin/uploads/init`
4. Nhận `upload_url`, `asset_id`, `headers`
5. FE `PUT` file trực tiếp lên S3/MinIO
6. FE gọi `POST /admin/uploads/complete`
7. Asset chuyển `UPLOADED`, backend tạo `parse_job`
8. FE điều hướng sang asset detail

## 5.3 Theo dõi parse

1. FE poll `GET /admin/uploads/:id/parse-jobs`
2. Khi job `PARSED` hoặc `REVIEW_REQUIRED`, FE enable nút review
3. Admin mở danh sách draft questions

## 5.4 Review draft

1. FE gọi `GET /admin/uploads/:id/draft-questions`
2. Admin sửa stem/explanation/option/correct answer
3. FE gọi:
   - `PUT /admin/question-drafts/:id`
   - `PUT /admin/question-drafts/:id/options/:optionId`

## 5.5 Publish

1. Admin chọn `topic` + `difficulty`
2. Publish từng draft:
   - `POST /admin/question-drafts/:id/publish`
3. Hoặc publish cả asset:
   - `POST /admin/uploads/:id/publish`

## 5.6 Link asset

1. Admin chọn entity type
2. Nhập/chọn entity id
3. FE gọi `POST /admin/uploads/:id/link`

---

## 6. Kiến trúc frontend

## 6.1 App modules

### Auth

- login form
- token persistence
- refresh token flow
- route guard

### Dashboard

- quick stats cơ bản
- recent uploads
- parse jobs cần review

### Assets

- upload page
- assets list
- asset detail
- parse history
- link entity modal

### Drafts

- draft list by asset
- draft editor
- single publish
- batch publish

### Curriculum

- subject/chapter/topic picker để chọn `topic_id`

---

## 6.2 Cấu trúc thư mục đề xuất

```text
apps/admin-web/
  public/
  src/
    app/
      providers/
      router.tsx
      store.ts
    api/
      client.ts
      auth.ts
      uploads.ts
      drafts.ts
      curriculum.ts
      questions.ts
      types.ts
    components/
      layout/
      forms/
      tables/
      feedback/
      ui/
    features/
      auth/
      dashboard/
      uploads/
      drafts/
      curriculum/
    hooks/
      useAuth.ts
      useTokenRefresh.ts
      useUploadFlow.ts
      useDraftPublish.ts
    lib/
      jwt.ts
      checksum.ts
      env.ts
      utils.ts
    pages/
      LoginPage.tsx
      DashboardPage.tsx
      UploadsPage.tsx
      UploadDetailPage.tsx
      DraftReviewPage.tsx
      NotFoundPage.tsx
    styles/
      globals.css
    main.tsx
  package.json
  tsconfig.json
  vite.config.ts
```

---

## 7. Auth strategy

## 7.1 Token handling

Vì backend hiện trả token qua JSON, FE MVP sẽ dùng:

- `access_token` lưu trong **memory + sessionStorage**
- `refresh_token` lưu trong **sessionStorage**

### Quy tắc

- access token được gắn vào `Authorization: Bearer ...`
- khi 401, FE tự gọi `POST /auth/refresh`
- refresh thành công thì retry request
- logout gọi `POST /auth/logout` rồi clear session

## 7.2 Role check

JWT access token hiện có `role` claim, nên FE có thể:

- decode access token
- nếu role != `ADMIN` thì chặn route admin

### Lưu ý

Về sau có thể thêm `/auth/me`, nhưng **MVP chưa bắt buộc** vì token đã có claim đủ dùng.

---

## 8. Routing plan

## 8.1 Public routes

- `/login`

## 8.2 Protected admin routes

- `/`
- `/dashboard`
- `/uploads`
- `/uploads/new`
- `/uploads/:assetId`
- `/uploads/:assetId/drafts`
- `/drafts/:draftId`

### Hành vi

- nếu chưa login → redirect `/login`
- nếu token hết hạn và refresh fail → redirect `/login`
- nếu không phải admin → hiện `403`

---

## 9. UI screens cần làm

## 9.1 Login page

### Thành phần

- email input
- password input
- submit button
- error banner

### Thành công

- redirect dashboard

## 9.2 Dashboard

### Mục tiêu

Chỉ cần dashboard nhẹ cho MVP:

- số asset gần đây
- asset cần review
- parse job fail gần đây
- shortcut tới upload mới

## 9.3 Uploads page

### Nội dung

- bảng danh sách assets
- filter theo status
- search theo filename *(nếu chưa có API search thì client-side trên page hiện tại)*
- button “Upload PDF”

### Cột gợi ý

- filename
- status
- uploaded by
- file size
- created at
- latest parse status
- actions

## 9.4 Upload modal / page

### Step 1

- chọn file PDF
- validate extension / MIME / size
- tính checksum SHA-256

### Step 2

- gọi init upload
- upload trực tiếp lên presigned URL

### Step 3

- gọi complete upload
- điều hướng sang asset detail

## 9.5 Upload detail page

### Nội dung

- metadata asset
- parse jobs timeline
- button retry parse
- button link entity
- button review drafts
- button batch publish

## 9.6 Draft review page

### Nội dung

- bảng danh sách draft questions theo asset
- click mở editor bên phải hoặc modal
- sửa:
  - stem
  - explanation
  - options
  - correct answer
- chọn `difficulty`
- chọn `topic`

### Action

- save draft changes
- publish one
- publish all

---

## 10. Component plan

### Layout

- `AdminShell`
- `Sidebar`
- `Topbar`
- `ProtectedRoute`

### Upload

- `UploadPdfForm`
- `UploadProgressCard`
- `AssetsTable`
- `AssetStatusBadge`
- `ParseJobsPanel`
- `LinkEntityDialog`

### Drafts

- `DraftQuestionTable`
- `DraftEditorPanel`
- `OptionEditor`
- `PublishDialog`
- `TopicSelector`

### Shared

- `AppTable`
- `EmptyState`
- `ErrorState`
- `ConfirmDialog`
- `ToastProvider`

---

## 11. API integration contract

## 11.1 Response envelope

Backend đang trả response kiểu:

```json
{
  "success": true,
  "data": {},
  "meta": {}
}
```

Error:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "..."
  }
}
```

FE cần chuẩn hóa parser cho envelope này từ đầu.

## 11.2 Typed API client

Nên tạo:

- `api/client.ts`
- `api/types.ts`
- `api/auth.ts`
- `api/uploads.ts`
- `api/drafts.ts`
- `api/curriculum.ts`

Không gọi `fetch` trực tiếp trong component.

---

## 12. State management plan

Không cần global state phức tạp.

### Dùng:

- **TanStack Query** cho server state
- **React Context** cho auth/session
- local component state cho form/editor

### Không cần ở MVP:

- Redux
- Zustand
- xstate

---

## 13. UX details quan trọng

### Upload UX

- hiển thị progress upload thật
- disable nút submit khi đang upload
- show asset id sau khi thành công

### Parse UX

- poll mỗi 5 giây ở asset detail
- badge rõ cho `QUEUED / PROCESSING / PARSED / REVIEW_REQUIRED / FAILED`
- nếu `FAILED`, show `error_message`

### Draft review UX

- hiển thị confidence nếu backend có field
- highlight draft thiếu correct answer
- xác nhận trước khi publish batch

### Error UX

- toast cho action thành công
- error banner cho lỗi auth/validation
- route-level fallback cho lỗi tải dữ liệu

---

## 14. Testing plan cho FE

## 14.1 Unit / component

- login form validation
- upload form validation
- draft editor state transitions
- topic selector behavior

## 14.2 API / integration

- auth refresh interceptor
- upload init → PUT → complete flow
- poll parse jobs
- publish one / publish all

## 14.3 E2E

Dùng **Playwright**:

1. login admin
2. upload PDF
3. complete upload
4. xem parse job
5. review draft
6. publish question

---

## 15. Phase implementation

## Phase 1 — Scaffold & foundation

### Làm

1. tạo `apps/admin-web` bằng Vite React TS
2. setup Tailwind + shadcn/ui
3. setup router
4. setup Axios client
5. setup auth context
6. setup env config

### Done khi

- app chạy local
- gọi được `/auth/login`
- có protected route

## Phase 2 — Upload module

### Làm

1. uploads list page
2. upload PDF form
3. init upload
4. direct PUT to presigned URL
5. complete upload
6. asset detail
7. parse jobs panel

### Done khi

- admin upload được 1 PDF và thấy asset trong list

## Phase 3 — Draft review module

### Làm

1. draft list by asset
2. draft editor
3. update draft
4. update option
5. curriculum topic picker
6. publish single
7. publish by asset

### Done khi

- admin review + publish được question thật

## Phase 4 — Polish & hardening

### Làm

1. better loading/error states
2. toast + confirm dialogs
3. responsive admin layout
4. Playwright E2E
5. basic dashboard

### Done khi

- admin FE đủ dùng cho demo/UAT

---

## 16. Env vars cho FE

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_NAME=Edu Admin
VITE_POLL_INTERVAL_MS=5000
```

---

## 17. Lệnh khởi tạo đề xuất

```bash
npm create vite@latest apps/admin-web -- --template react-ts
cd apps/admin-web
npm install
npm install axios @tanstack/react-query react-router-dom react-hook-form zod @hookform/resolvers
npm install -D tailwindcss @tailwindcss/vite
```

Sau đó thêm:

- shadcn/ui
- Playwright

---

## 18. Definition of Done

Admin FE được xem là đạt MVP khi:

1. Admin login được bằng backend hiện tại.
2. Có protected admin routes.
3. Upload được PDF qua presigned URL.
4. Xem được asset list + asset detail + parse jobs.
5. Xem và sửa được draft questions.
6. Publish được một draft và publish được toàn bộ draft của asset.
7. Link được asset vào entity.
8. Có Playwright happy-path test cho flow upload → review → publish.

---

## 19. Khuyến nghị cuối

Nếu bắt đầu implement ngay, tôi khuyên đi theo thứ tự:

1. **Phase 1**
2. **Phase 2**
3. **Phase 3**
4. rồi mới làm **dashboard/polish**

Với backend hiện tại, **admin FE hoàn toàn có thể bắt đầu ngay** mà chưa cần thay đổi lớn ở API.

