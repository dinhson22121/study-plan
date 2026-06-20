# Implementation Plan từ `enhance.md`

## Mục tiêu

Biến các đề xuất trong `enhance.md` thành backlog triển khai có thể code theo phase, bám sát repo hiện tại:

- Go/Gin modular monolith
- `internal\content` và `internal\question`
- PostgreSQL + Redis + Kafka
- thêm S3-compatible storage
- thêm Python worker để parse PDF thành draft questions

---

## 1. Các giả định đã chốt để implement

### 1.1 MVP decisions

- Admin upload file đề bài bằng **presigned PUT URL**
- File gốc lưu trên **S3**; local/dev dùng **MinIO**
- Backend Go chỉ xử lý:
  - auth/rbac
  - tạo upload session
  - validate metadata
  - lưu DB
  - tạo `parse_job`
- **Python worker** sẽ:
  - poll `parse_job`
  - tải PDF từ S3
  - extract text
  - parse ra **draft questions**
  - lưu draft xuống PostgreSQL
- **Admin phải review draft** trước khi publish thành câu hỏi chính thức
- MVP chỉ ưu tiên **PDF có text sẵn**
- MVP parse trước cho **MCQ A/B/C/D**; free-text và OCR scan để phase sau

### 1.2 Ownership giữa module

- `content` **own**:
  - uploaded asset
  - upload session
  - parse job
  - asset lifecycle
- `question` **own**:
  - final question aggregate
  - publish draft → question chính thức
- Kết quả parse ban đầu có thể lưu ở `content` hoặc `question_draft`, nhưng khuyến nghị:
  - metadata + parse job ở `content`
  - draft question records ở `question`-adjacent schema/tables

---

## 2. Deliverables cần có

| Deliverable | Mô tả |
|---|---|
| `enhance.md` được chốt | Tài liệu enhancement ở mức product/technical |
| Upload service trong Go | Admin có thể init upload, complete upload, list asset, retry parse |
| S3/MinIO integration | Có client/presigner và config runtime |
| DB migrations mới | Thêm bảng upload, parse job, draft question |
| Python parse worker | Poll job, tải PDF, parse, lưu draft |
| Review/publish flow | Admin xem draft và publish vào question bank |
| Observability | Metrics, logs, trace fields cho upload + parse |
| Test coverage tối thiểu | Unit + integration + worker tests cho happy path và lỗi chính |

---

## 3. Repo mapping đề xuất

### 3.1 Các path sẽ thêm hoặc sửa

| Path | Hành động | Mục đích |
|---|---|---|
| `config\config.go` | Sửa | Thêm `S3Config`, `UploadConfig`, `ParserConfig` |
| `config\config.yaml` | Sửa | Thêm config mẫu cho S3/MinIO và parse worker |
| `internal\app\deps.go` | Sửa | Thêm dependency cho S3/presigner nếu dùng chung trong monolith |
| `pkg\s3\` | Thêm | Wrapper cho S3 client + presigned URL |
| `internal\content\domain\` | Sửa/Thêm | Model/ports cho asset, upload session, parse job |
| `internal\content\application\` | Sửa/Thêm | Use cases init/complete/list/delete/trigger parse |
| `internal\content\infrastructure\` | Sửa/Thêm | PG repo + S3 adapter |
| `internal\content\interfaces\http\` | Sửa/Thêm | Admin upload handlers |
| `internal\question\` | Sửa/Thêm | Draft review/publish flow nếu đặt publish ở question module |
| `migrations\` | Thêm | Migration cho upload + parse tables |
| `docker-compose.yml` | Sửa | Thêm MinIO cho local dev |
| `workers\pdf_parser\` | Thêm | Python worker riêng |
| `plan\enhancment\` | Sửa | Bổ sung implementation notes nếu cần |

### 3.2 Cấu trúc thư mục khuyến nghị

```text
pkg/
  s3/
    client.go
    presign.go

internal/
  content/
    domain/
      asset.go
      upload_session.go
      parse_job.go
      ports.go
    application/
      service.go
      admin_upload_service.go
      parse_job_service.go
    infrastructure/
      pg_asset_repository.go
      pg_parse_job_repository.go
      s3_presigner.go
    interfaces/
      http/
        admin_upload_handler.go

  question/
    application/
      draft_service.go
      publish_service.go
    infrastructure/
      pg_draft_repository.go
    interfaces/
      http/
        admin_question_draft_handler.go

workers/
  pdf_parser/
    requirements.txt
    main.py
    config.py
    s3_client.py
    job_repository.py
    pdf_extract.py
    parser.py
    persistence.py
```

---

## 4. Phase triển khai

## Phase 0 — Chốt contract và update docs gốc

### Mục tiêu

Đồng bộ các quyết định hiện đã chốt để code không bị lệch.

### Việc làm

1. Update `plan\edu-app-prd-tdd-v3.html`:
   - thêm upload service
   - thêm parse worker
   - thêm decision “PDF text-based only for MVP”
   - thêm review-before-publish flow
2. Đồng bộ các inconsistency cũ:
   - `SKIPPED` status
   - topic naming
   - multi-device token
   - broadcast scope
3. Thêm section “Open Decisions” chỉ còn lại các thông tin chưa cần để bắt đầu code.

### Done khi

- Tài liệu gốc phản ánh đúng architecture sẽ implement
- Không còn mâu thuẫn giữa flow upload, parse, question publish

---

## Phase 1 — Storage, config và local infra

### Mục tiêu

Chuẩn bị S3-compatible storage và runtime config.

### Việc làm

1. Thêm vào `config.Config`:
   - `S3.Endpoint`
   - `S3.Region`
   - `S3.AccessKey`
   - `S3.SecretKey`
   - `S3.Bucket`
   - `S3.UsePathStyle`
   - `Upload.MaxFileSizeBytes`
   - `Upload.PresignTTL`
   - `Parser.PollInterval`
   - `Parser.BatchSize`
2. Bind env tương ứng:
   - `EDU_S3_ENDPOINT`
   - `EDU_S3_REGION`
   - `EDU_S3_ACCESS_KEY`
   - `EDU_S3_SECRET_KEY`
   - `EDU_S3_BUCKET`
   - `EDU_S3_USE_PATH_STYLE`
   - `EDU_UPLOAD_MAX_FILE_SIZE_BYTES`
   - `EDU_UPLOAD_PRESIGN_TTL`
   - `EDU_PARSER_POLL_INTERVAL`
   - `EDU_PARSER_BATCH_SIZE`
3. Tạo `pkg\s3`:
   - shared S3 client factory
   - presigned PUT URL helper
   - object existence/head helper
4. Sửa `docker-compose.yml`:
   - thêm `minio`
   - expose console cho local dev
   - add persistent volume
5. Nếu cần, tạo script/init bucket cho local dev.

### Done khi

- Server boot được với MinIO/S3 config
- Có thể tạo presigned URL local
- Có thể `HEAD` object để verify upload complete

---

## Phase 2 — Database schema và domain model

### Mục tiêu

Tạo nền dữ liệu cho upload và parse pipeline.

### Migration đề xuất

1. `013_init_upload_asset.up.sql`
2. `013_init_upload_asset.down.sql`
3. `014_init_parse_job.up.sql`
4. `014_init_parse_job.down.sql`
5. `015_init_question_draft.up.sql`
6. `015_init_question_draft.down.sql`

### Bảng cần thêm

#### 2.1 `uploaded_asset`

Các cột tối thiểu:

- `id`
- `object_key`
- `bucket_name`
- `original_filename`
- `content_type`
- `file_size`
- `checksum_sha256`
- `status`
- `uploaded_by`
- `entity_type`
- `entity_id`
- `storage_provider`
- `created_at`
- `verified_at`
- `deleted_at`

### 2.2 `parse_job`

Các cột tối thiểu:

- `id`
- `asset_id`
- `status`
- `parser_version`
- `attempt_count`
- `error_message`
- `claimed_by`
- `claimed_at`
- `started_at`
- `finished_at`
- `created_by`
- `created_at`
- `updated_at`

Khuyến nghị enum:

- `QUEUED`
- `PROCESSING`
- `PARSED`
- `REVIEW_REQUIRED`
- `FAILED`

### 2.3 `question_draft`

Các cột tối thiểu:

- `id`
- `asset_id`
- `parse_job_id`
- `question_number`
- `question_type`
- `stem`
- `explanation_raw`
- `answer_key_raw`
- `parse_confidence`
- `status`
- `reviewed_by`
- `reviewed_at`
- `published_question_id`
- `created_at`
- `updated_at`

### 2.4 `question_draft_option`

Các cột tối thiểu:

- `id`
- `question_draft_id`
- `option_label`
- `option_text`
- `is_correct_inferred`
- `order_index`

### Domain modelling trong Go

Thêm trong `internal\content\domain`:

- `AssetStatus`
- `UploadedAsset`
- `ParseJobStatus`
- `ParseJob`
- repository ports cho asset và parse job

Nếu review/publish nằm ở `question`, thêm trong `internal\question`:

- `QuestionDraft`
- `QuestionDraftOption`
- service publish draft → `Question`

### Done khi

- Migrate lên/down sạch
- Domain model đủ để code API và worker
- Query path cho list asset / list jobs / list drafts rõ ràng

---

## Phase 3 — Go upload service

### Mục tiêu

Cho admin upload đề bài an toàn, có metadata và tạo parse job.

### API sẽ implement

| Method | Path | Ghi chú |
|---|---|---|
| `POST` | `/api/v1/admin/uploads/init` | Tạo upload session + presigned URL |
| `POST` | `/api/v1/admin/uploads/complete` | Verify object và lưu metadata |
| `GET` | `/api/v1/admin/uploads` | List/filter asset |
| `GET` | `/api/v1/admin/uploads/:id` | Chi tiết asset |
| `POST` | `/api/v1/admin/uploads/:id/parse` | Retry/reparse thủ công |
| `GET` | `/api/v1/admin/uploads/:id/parse-jobs` | Lịch sử parse |
| `DELETE` | `/api/v1/admin/uploads/:id` | Soft delete |

### Flow `init upload`

1. Validate JWT
2. Validate role `ADMIN`
3. Validate filename/content_type/file_size
4. Sinh `asset_id`
5. Sinh `object_key`
6. Tạo record `uploaded_asset` với status `PENDING`
7. Tạo presigned PUT URL
8. Trả về:
   - `assetId`
   - `objectKey`
   - `uploadUrl`
   - `expiresAt`
   - `requiredHeaders`

### Flow `complete upload`

1. Validate role `ADMIN`
2. Lấy asset theo `asset_id`
3. Verify object tồn tại trên S3 qua `HEAD`
4. So khớp `content_type`, `file_size` nếu có
5. Update `uploaded_asset.status = UPLOADED`
6. Tạo `parse_job.status = QUEUED`
7. Trả response chi tiết asset + parse job id

### Validation rules MVP

- chỉ nhận `application/pdf`
- file size giới hạn theo config
- reject nếu asset đã complete/deleted
- reject nếu user không có role `ADMIN`
- chưa cho publish trực tiếp khi chưa parse/review

### Go files gợi ý

- `internal\content\application\admin_upload_service.go`
- `internal\content\interfaces\http\admin_upload_handler.go`
- `internal\content\infrastructure\pg_asset_repository.go`
- `internal\content\infrastructure\pg_parse_job_repository.go`
- `internal\content\infrastructure\s3_presigner.go`

### Done khi

- Admin lấy được presigned URL
- Upload xong gọi complete và tạo được `parse_job`
- Asset list/detail trả đúng metadata

---

## Phase 4 — Python parse worker

### Mục tiêu

Parse PDF thành draft questions lưu DB mà không block admin API.

### Cách claim job MVP

Khuyến nghị dùng **polling Postgres**, chưa cần Kafka:

1. Worker query `parse_job` status `QUEUED`
2. Claim job bằng transaction + `FOR UPDATE SKIP LOCKED`
3. Set `PROCESSING`
4. Parse xong thì update `PARSED` hoặc `REVIEW_REQUIRED`
5. Lỗi thì set `FAILED`

Lý do:

- đơn giản hơn Kafka cho worker đầu tiên
- retry dễ
- phù hợp khi traffic admin upload còn thấp

### Worker flow

1. Load config
2. Poll jobs theo interval
3. Claim 1 job
4. Lấy metadata asset từ DB
5. Download PDF từ S3
6. Extract text bằng `PyMuPDF`
7. Parse text theo rule:
   - nhận diện `Câu 1`, `Question 1`, hoặc pattern tương tự
   - tách A/B/C/D
   - cố gắng detect answer key nếu có
8. Ghi `question_draft` + `question_draft_option`
9. Update `parse_job`

### Package Python gợi ý

- `boto3`
- `PyMuPDF`
- `psycopg` hoặc `SQLAlchemy`
- `pydantic` cho config/schema nội bộ nếu cần

### File layout gợi ý

| File | Mục đích |
|---|---|
| `main.py` | worker loop |
| `config.py` | env/config loader |
| `job_repository.py` | claim/update parse job |
| `s3_client.py` | download PDF |
| `pdf_extract.py` | read text from PDF |
| `parser.py` | split question/options |
| `persistence.py` | insert draft records |

### Parse strategy MVP

- ưu tiên đề trắc nghiệm định dạng khá chuẩn
- chưa xử lý OCR image/scan
- chưa parse công thức phức tạp
- nếu chỉ parse được một phần thì:
  - vẫn lưu partial draft
  - set `REVIEW_REQUIRED`

### Done khi

- Worker xử lý được 1 file PDF text-based end-to-end
- Draft records được lưu đúng
- Job retry không sinh duplicate draft

---

## Phase 5 — Review và publish draft questions

### Mục tiêu

Không publish tự động; admin phải duyệt và sửa nếu cần.

### API nên thêm

| Method | Path | Mô tả |
|---|---|---|
| `GET` | `/api/v1/admin/uploads/:id/draft-questions` | Xem draft questions theo asset |
| `PUT` | `/api/v1/admin/question-drafts/:id` | Sửa draft question |
| `PUT` | `/api/v1/admin/question-drafts/:id/options/:optionId` | Sửa option |
| `POST` | `/api/v1/admin/question-drafts/:id/publish` | Publish 1 draft thành question |
| `POST` | `/api/v1/admin/uploads/:id/publish` | Publish batch các draft đã review |

### Publish flow

1. Validate admin role
2. Lấy `question_draft` + options
3. Map sang `question.Question`
4. Dùng `question` repository/service để create question chính thức
5. Update `question_draft.published_question_id`
6. Mark draft là `PUBLISHED` hoặc `REVIEWED_PUBLISHED`

### Liên kết với model hiện tại

Repo hiện tại đã có:

- `question.Question`
- `question.AnswerOption`
- `question.TypeMCQ`
- `question.TypeFreeText`

MVP nên publish trước cho:

- `MCQ`

Free text có thể để phase tiếp theo nếu parser chưa đủ tin cậy.

### Done khi

- Admin xem được draft
- Admin sửa được stem/options
- Publish xong tạo record trong question module

---

## Phase 6 — Security, observability và hardening

### Security

1. RBAC:
   - mọi endpoint upload/parse/review/publish phải `ADMIN`
2. Storage:
   - object private mặc định
   - presigned PUT TTL ngắn
3. Audit:
   - log init upload
   - log complete upload
   - log delete
   - log parse retry
   - log publish draft

### Observability

Metrics tối thiểu:

- `upload_init_total`
- `upload_complete_total`
- `upload_failed_total`
- `parse_job_created_total`
- `parse_job_completed_total`
- `parse_job_failed_total`
- `parse_duration_seconds`
- `parse_question_draft_total`

Structured log fields:

- `assetId`
- `parseJobId`
- `objectKey`
- `parserVersion`
- `status`
- `errorCode`
- `uploadedBy`

### Hardening

- idempotent complete upload
- idempotent parse retry
- cleanup asset `PENDING` quá hạn
- soft delete trước, hard delete sau
- constraint tránh duplicate option order

### Done khi

- Có metric/log cơ bản để debug production
- Lỗi S3/PDF/DB được surface rõ, không silent fail

---

## Phase 7 — Testing

## 7.1 Go tests

### Unit

- object key generator
- mime/file size validator
- complete upload state transition
- parse job creation rules

### Integration

- PG repository cho asset
- PG repository cho parse job
- S3 presign + head object flow
- admin upload handler auth/rbac

## 7.2 Python worker tests

- extract text từ PDF có text
- parse MCQ 4 đáp án
- partial parse tạo `REVIEW_REQUIRED`
- corrupted PDF tạo `FAILED`
- retry không duplicate draft

## 7.3 End-to-end

1. Init upload
2. Upload PDF lên MinIO
3. Complete upload
4. Worker parse
5. Admin xem draft
6. Admin publish
7. Question xuất hiện trong `question` module

### Done khi

- Có ít nhất 1 E2E chạy qua full happy path local

---

## 8. Backlog theo ưu tiên

## P0 — phải làm trước

1. Thêm config S3/upload/parser
2. Thêm MinIO local
3. Thêm migrations `uploaded_asset`, `parse_job`, `question_draft`
4. Implement upload init/complete
5. Implement Python worker polling + parse PDF text-based
6. Implement draft review + publish cho MCQ

## P1 — nên làm ngay sau P0

1. Retry parse job thủ công
2. Asset soft delete
3. Audit log rõ hơn
4. Parser confidence scoring
5. Batch publish

## P2 — phase sau

1. OCR cho PDF scan/image
2. Hỗ trợ DOCX/image ZIP import
3. Hỗ trợ FREE_TEXT parse
4. Malware scanning
5. Rule engine cho nhiều format đề khác nhau
6. Chuyển job dispatch sang Kafka nếu tải tăng

---

## 9. Rủi ro chính khi implement

| Rủi ro | Ảnh hưởng | Giảm thiểu |
|---|---|---|
| PDF format quá đa dạng | parse sai | giới hạn MVP cho PDF text-based chuẩn |
| Admin upload file scan | worker parse kém | reject hoặc mark unsupported trong MVP |
| Duplicate parse khi retry | draft trùng | unique constraint + job claim đúng cách |
| Publish draft sai | question bank bẩn | review bắt buộc trước publish |
| S3 config lệch local/prod | upload fail | MinIO + config mẫu + smoke test |

---

## 10. Các quyết định còn mở nhưng không chặn P0

1. Giới hạn file size chốt là `20MB` hay `50MB`
2. Có lưu `raw_extracted_text` trong DB hay object storage riêng
3. Có cần attach 1 asset vào nhiều entity ngay MVP không
4. Có cần batch publish trong MVP không

---

## 11. Definition of Done cho nhánh implement này

Hoàn thành khi:

1. Admin upload được PDF qua presigned URL.
2. Metadata asset được lưu DB và verify được object trên S3.
3. `parse_job` được tạo tự động sau complete upload.
4. Python worker parse được PDF text-based thành draft questions.
5. Admin review và publish được draft MCQ vào question bank.
6. Có test happy path local với MinIO + Postgres.
7. Tài liệu gốc PRD/TDD được update theo implementation thực tế.

