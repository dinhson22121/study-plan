# Enhancement Plan cho `edu-app-prd-tdd-v3.html`

## Mục tiêu

Nâng tài liệu hiện tại từ một bản **technical design thiên về Notification Module** thành bộ tài liệu đủ để:

- Product chốt scope và KPI.
- Backend triển khai đúng domain boundary và contract.
- Mobile tích hợp API, push notification, deeplink.
- QA có acceptance criteria và test matrix rõ ràng.
- DevOps có yêu cầu observability, scaling và runbook vận hành.

---

## 1. Các phần sẽ sửa trong tài liệu hiện tại

| Section hiện tại | Vấn đề hiện tại | Nội dung sẽ sửa |
|---|---|---|
| Tổng quan kiến trúc | Mô tả tốt về backend nhưng chưa nói rõ scope MVP và ranh giới giữa monolith với service tách riêng | Bổ sung architectural decision cho MVP, điều kiện tách Notification Service, nguyên tắc sync vs async communication |
| Project Structure (DDD) | Mới mô tả cấu trúc thư mục, chưa thể hiện ownership và dependency rule giữa module | Thêm nguyên tắc dependency, shared package policy, placement của tests, migrations, configs |
| 12 Domain Modules | Mới chỉ liệt kê tên module | Bổ sung aggregate/entity chính, command/query, published events, consumed events, dependency matrix cho từng module |
| Content / Upload capability | Chưa có thiết kế rõ cho luồng admin upload đề bài và lưu file | Bổ sung Upload Service hoặc Upload submodule, luồng upload lên S3, metadata management, RBAC và lifecycle file |
| Notification Strategy | Tập trung vào loại notification nhưng chưa có policy đầy đủ | Bổ sung quiet hours, locale, deeplink rules, segmentation, campaign rules, retry classification, fallback content |
| Database Schema | Chưa đủ cho multi-device, campaign/audit, analytics và replay | Điều chỉnh schema và index để hỗ trợ multi-device, broadcast campaign, delivery tracking, retention |
| Kafka Topics | Có schema cơ bản nhưng chưa có partition key, retention, versioning | Bổ sung message key strategy, schema version, retention policy, replay/DLQ handling, producer/consumer ownership |
| Notification Flow | Mới có happy path và retry path chính | Bổ sung flow cho broadcast, disabled preference, invalid token, duplicate message, DLQ replay |
| API Endpoints | Chưa có request/response schema, validation, pagination, RBAC, error envelope | Bổ sung contract đầy đủ cho client và QA dùng chung |
| Error Handling | Chưa map hết domain error, chưa chuẩn hóa response body | Thêm error code matrix, trace ID, retryable vs non-retryable classification |
| Scaling Plan | Mới ở mức định tính | Bổ sung threshold, SLO, bottleneck, chi phí ước tính, điều kiện scale theo metric |
| Init Project | Tốt cho bootstrap nhưng chưa đủ cho team onboarding | Bổ sung `.env.example`, Makefile targets, migration command, local run flow, seed data, CI entrypoints |
| Future Enhancements | Có ý tưởng nhưng chưa có mức độ ưu tiên | Phân loại theo P1/P2/P3 và điều kiện kích hoạt |

---

## 2. Các section mới cần thêm để hoàn chỉnh PRD

### 2.1 Product Vision & Problem Statement

Mục này cần trả lời:

- Edu App giải quyết vấn đề gì.
- Người dùng mục tiêu là ai.
- Tại sao notification là năng lực quan trọng thay vì chỉ là tính năng phụ.
- Giá trị cốt lõi của phiên bản đầu tiên.

### 2.2 Personas & User Journeys

Cần thêm ít nhất các persona:

- Student
- Admin
- Parent/Guardian *(nếu nằm trong roadmap gần)*
- Teacher/Counselor *(nếu thực sự có trong scope sản phẩm)*

Với mỗi persona, cần mô tả:

- Mục tiêu
- Pain points
- Tương tác chính với app
- Điểm cần notification hỗ trợ

### 2.3 Goals, Success Metrics & KPI

Cần bổ sung KPI ở 3 lớp:

- **Product KPI:** DAU/WAU, D7 retention, study streak rate, weekly completion rate
- **Notification KPI:** send success rate, open rate, CTR, unsubscribe/preference disable rate
- **Platform KPI:** queue lag, retry rate, DLQ rate, token invalidation rate, p95 API latency

### 2.4 MVP Scope / Out of Scope

Phải chốt rõ:

- Tính năng nào thuộc MVP
- Tính năng nào deferred
- Tính năng nào chỉ là future enhancement

Ví dụ cần làm rõ:

- Broadcast có segment ngay từ MVP hay chỉ all users
- Parent notification có thuộc v1 hay chỉ roadmap
- Analytics ở mức dashboard hay chỉ raw log

### 2.5 User Stories & Acceptance Criteria

Cần thêm user stories cho các luồng chính:

1. Đăng ký tài khoản
2. Đăng nhập / refresh / logout
3. Đăng ký device token
4. Bật/tắt từng loại notification
5. Daily study reminder
6. Weekly quiz reminder
7. Study plan reminder
8. Achievement notification
9. Re-engagement notification
10. Admin broadcast
11. Xem notification history

Mỗi story cần có:

- Trigger
- Preconditions
- Main flow
- Edge cases
- Acceptance criteria dạng kiểm thử được

### 2.6 Non-functional Requirements

Cần có một section riêng cho:

- Availability target
- Latency target
- Throughput target
- Durability của message
- Data retention
- Security requirements
- Auditability
- Localization/timezone rules

### 2.7 Security & Privacy

Bổ sung:

- JWT claim design
- RBAC cho `ADMIN`
- Secret management cho FCM/Kafka/Postgres
- PII được phép xuất hiện ở đâu
- Chính sách log masking
- Audit trail cho admin broadcast và preference changes

### 2.8 Testing Strategy

Đây là phần còn thiếu nhiều nhất nếu tài liệu thật sự muốn mang tên `TDD`.

Cần thêm:

- Unit test matrix theo module
- Integration test matrix cho DB, Redis, Kafka, FCM adapter
- Contract test cho API và Kafka events
- End-to-end test cho các luồng chính
- Failure injection test cho retry, invalid token, DLQ
- Timezone test cases
- Idempotency test cases

### 2.9 Observability & Runbook

Cần thêm:

- Metrics cần expose
- Structured log fields
- Tracing fields (`correlationId`, `userId`, `notificationType`, `campaignId`)
- Alert rules
- Dashboard suggestions
- Runbook cho DLQ replay, token cleanup, Kafka lag, FCM outage

### 2.10 Release Roadmap

Nên chia thành:

- **Phase 1:** MVP monolith + auth + device token + preference + reminder cơ bản
- **Phase 2:** segmentation + analytics + campaign management
- **Phase 3:** tách notification service + advanced personalization

### 2.11 Risks, Assumptions & Open Questions

Phải có bảng riêng cho:

- Rủi ro kỹ thuật
- Giả định sản phẩm
- Phần cần stakeholder xác nhận

### 2.12 Admin Upload Service for Exam Content

Cần thêm một section riêng mô tả năng lực upload file để admin đưa đề bài/tài liệu lên hệ thống.

Phần này nên trả lời:

- Admin được upload những loại file nào: PDF, DOCX, PNG/JPG, ZIP hay chỉ PDF/image
- File được lưu trực tiếp qua backend hay dùng presigned URL để upload thẳng lên S3
- Sau khi upload xong thì file gắn với entity nào: `question`, `exam`, `content`, `attachment`
- Hệ thống có parse/OCR/import câu hỏi tự động hay chỉ lưu file thô ở giai đoạn đầu
- Có versioning, soft delete, replace file, audit log hay không

Đề xuất mặc định cho MVP:

- Backend tạo **upload session** và trả về **presigned PUT URL**
- Admin client upload file trực tiếp lên S3 để tránh app server giữ file lớn
- Backend nhận callback/complete request để lưu metadata vào database
- Sau khi upload hoàn tất, backend tạo **parse job** cho **Python worker**
- Python worker pull file PDF từ S3, extract text, parse thành **draft questions** và lưu xuống DB
- Chỉ `ADMIN` mới có quyền tạo/xóa asset
- File đề bài được gắn metadata để dùng lại cho Question/Content modules
- Admin review lại kết quả parse trước khi publish chính thức vào question bank

---

## 3. Các điểm chưa nhất quán cần sửa ngay trước khi triển khai

### 3.1 Trạng thái `SKIPPED` chưa đồng bộ

Hiện flow và schema có dùng `SKIPPED`, nhưng enum domain chưa có.  
Việc cần sửa:

- Thêm `StatusSkipped` vào domain model
- Quy định rõ khi nào log `SKIPPED`
- Map rõ `ErrPreferenceDisabled` trong error policy

### 3.2 `device_token` đang thiên về single-device trong code, multi-device trong schema

Schema hiện cho phép nhiều token trên một user, nhưng repository interface chỉ trả về một token.  
Việc cần sửa:

- Đổi contract từ `FindDeviceToken(userID)` sang `FindActiveDeviceTokens(userID)`
- Làm rõ behavior khi user có nhiều thiết bị
- Quy định logic logout một thiết bị hay toàn bộ thiết bị

### 3.3 Tên Kafka topic chưa thống nhất

Cần thống nhất một naming duy nhất cho toàn tài liệu:

- `notification.schedule`
- `notification.send`
- `notification.result`
- `notification.dlq`

### 3.4 Luồng `Achievement Notification` chưa chốt

Hiện tài liệu có chỗ ghi direct FCM, có chỗ ghi flow chung qua Kafka.  
Phải quyết định một trong hai:

- Tất cả notification đều qua Kafka để đồng nhất observability
- Hoặc achievement đi direct FCM với lý do rõ ràng về latency

### 3.5 Broadcast chưa rõ scope

Cần làm rõ:

- Broadcast cho toàn bộ users
- Broadcast theo segment
- Broadcast theo grade/level/timezone
- Có cần schedule hay chỉ manual send

### 3.6 Error contract chưa đầy đủ

Cần thêm:

- JSON error envelope chuẩn
- Error code catalog
- Mapping domain error → HTTP status → client action

### 3.7 Chưa có thiết kế chuẩn cho upload file của admin

Hiện tài liệu chưa mô tả upload service, nên nếu triển khai ngay rất dễ phát sinh thiết kế rời rạc giữa backend, mobile/web admin và storage.  
Việc cần bổ sung:

- Chốt dùng **S3 object storage** cho file đề bài
- Chốt upload flow: `init upload` → `presigned URL` → `complete upload`
- Chốt parse flow: `complete upload` → `create parse_job` → `Python worker` → `draft questions`
- Chốt metadata schema và quan hệ giữa file với exam/question/content
- Chốt validation: MIME type, size limit, checksum, duplicate detection
- Chốt quyền hạn: chỉ `ADMIN`, có audit log cho upload/delete/replace

---

## 4. Các nội dung chi tiết nên bổ sung cho Notification Module

### 4.1 Database đề xuất mở rộng

Nên xem xét thêm hoặc làm rõ:

- `device_id`
- `last_seen_at`
- `app_version`
- `locale`
- `timezone`
- `campaign_id`
- `provider_message_id`
- `delivered_at`
- `opened_at` *(nếu tracking được từ client)*

### 4.2 Bảng/cấu trúc mới có thể cần

- `notification_campaign`
- `notification_delivery`
- `notification_segment`
- `notification_event_audit`

### 4.3 API cần chi tiết thêm

Cho từng endpoint cần có:

- request schema
- response schema
- validation rules
- auth scope
- sample payload
- sample errors

### 4.4 Event contract cần version hóa

Mỗi event nên có thêm:

- `eventVersion`
- `occurredAt`
- `source`
- `campaignId` *(nếu applicable)*
- `traceId` hoặc dùng lại `correlationId`

---

## 4B. Các nội dung chi tiết nên bổ sung cho Admin Upload Service

### 4B.1 Mục tiêu nghiệp vụ

Upload service này dùng để hỗ trợ admin:

- upload đề thi, đề bài, tài liệu học tập, hình minh họa
- gắn file vào question bank hoặc exam package
- quản lý metadata file thay vì lưu blob trong database
- tái sử dụng một asset cho nhiều nội dung học

### 4B.2 Kiến trúc đề xuất

Đề xuất cho MVP:

1. `Admin API` nhận request tạo upload session
2. Backend validate quyền `ADMIN`, sinh object key và presigned URL
3. Client upload file trực tiếp lên S3
4. Client gọi API complete upload
5. Backend verify object tồn tại, lưu metadata vào DB
6. Backend tạo `parse_job` với status `QUEUED`
7. Python worker tải PDF từ S3, extract text và parse ra bộ câu hỏi nháp
8. Lưu draft questions xuống DB để admin review/chỉnh sửa
9. Sau khi admin xác nhận, asset và question draft được attach vào `question`, `exam`, hoặc `content` entity

Lợi ích:

- tránh upload file lớn qua app server
- giảm memory/CPU pressure cho backend
- dễ scale độc lập storage với business API
- phù hợp cho web admin sau này
- tách riêng parsing workload khỏi request/response path của admin API
- dễ nâng cấp parser/OCR mà không ảnh hưởng upload flow

### 4B.3 Domain boundary đề xuất

Có thể triển khai theo một trong hai cách:

- **Cách 1:** tạo module riêng `upload`
- **Cách 2:** đặt trong `content` module dưới dạng subdomain `asset management`

Khuyến nghị:

- Nếu upload chỉ phục vụ đề bài/tài liệu: đặt trong `content`
- Nếu upload là năng lực dùng chung cho exam, question, avatar, attachment: tách module `upload`

### 4B.4 Database schema đề xuất

Nên có ít nhất bảng `uploaded_asset`:

| Column | Type | Ghi chú |
|---|---|---|
| `id` | UUID | PK |
| `object_key` | VARCHAR | S3 key duy nhất |
| `bucket_name` | VARCHAR | bucket lưu trữ |
| `original_filename` | VARCHAR | tên file gốc |
| `content_type` | VARCHAR | MIME type |
| `file_size` | BIGINT | số byte |
| `checksum_sha256` | VARCHAR | chống duplicate / verify integrity |
| `status` | VARCHAR | `PENDING`, `UPLOADED`, `VERIFIED`, `DELETED`, `FAILED` |
| `uploaded_by` | UUID | admin user id |
| `entity_type` | VARCHAR | `QUESTION`, `EXAM`, `CONTENT`, `ATTACHMENT` |
| `entity_id` | UUID | id entity liên kết |
| `storage_provider` | VARCHAR | `S3` |
| `created_at` | TIMESTAMPTZ | thời điểm tạo |
| `verified_at` | TIMESTAMPTZ | thời điểm verify object |
| `deleted_at` | TIMESTAMPTZ | soft delete |

Có thể cần thêm:

- `upload_session`
- `asset_version`
- `asset_access_log`
- `exam_attachment`
- `parse_job`
- `parsed_question_draft`
- `parsed_question_option_draft`

### 4B.5 S3 key strategy

Nên thống nhất format key từ đầu, ví dụ:

- `exam-assets/{year}/{month}/{assetId}-{sanitizedFilename}`
- `question-assets/{questionId}/{assetId}`
- `draft-uploads/{adminId}/{sessionId}/{filename}`

Yêu cầu:

- key không phụ thuộc hoàn toàn vào tên file gốc
- dễ truy vết theo module/ngữ cảnh
- tránh collision

### 4B.6 API contract nên bổ sung

Ít nhất cần các endpoint:

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| `POST` | `/api/v1/admin/uploads/init` | `ADMIN` | Tạo upload session, trả presigned URL |
| `POST` | `/api/v1/admin/uploads/complete` | `ADMIN` | Xác nhận upload hoàn tất và lưu metadata |
| `POST` | `/api/v1/admin/uploads/:id/parse` | `ADMIN` | Tạo parse job thủ công nếu cần retry/reparse |
| `GET` | `/api/v1/admin/uploads` | `ADMIN` | Danh sách asset, filter theo loại/nội dung |
| `GET` | `/api/v1/admin/uploads/:id` | `ADMIN` | Chi tiết asset |
| `GET` | `/api/v1/admin/uploads/:id/parse-jobs` | `ADMIN` | Lịch sử parse job của asset |
| `GET` | `/api/v1/admin/uploads/:id/draft-questions` | `ADMIN` | Kết quả parse nháp theo asset |
| `DELETE` | `/api/v1/admin/uploads/:id` | `ADMIN` | Soft delete asset |
| `POST` | `/api/v1/admin/uploads/:id/attach` | `ADMIN` | Gắn asset vào question/exam/content |

Nếu có admin web UI, response `init upload` nên gồm:

- `uploadUrl`
- `method`
- `headers` cần gửi
- `objectKey`
- `assetId`
- `expiresAt`

### 4B.7 Validation rules

Cần chốt ngay trong tài liệu:

- max file size
- allowed MIME types / extensions
- có chặn executable/unsafe file hay không
- số file tối đa cho mỗi exam/question
- có bắt buộc checksum hay không
- có scan virus/malware hay không

Khuyến nghị MVP:

- chỉ cho phép `pdf`, `png`, `jpg`, `jpeg`
- giới hạn 20MB hoặc 50MB tùy use case
- reject file type không khớp MIME và extension
- lưu checksum SHA-256

### 4B.8 Security & access control

Upload service phải có:

- RBAC chỉ cho `ADMIN`
- presigned URL có TTL ngắn
- bucket policy chặn public write
- object không public mặc định
- download nên đi qua signed URL hoặc backend authorize
- audit log cho create/delete/replace

### 4B.9 Lifecycle & retention

Cần định nghĩa:

- upload session hết hạn sau bao lâu
- file `PENDING` không complete thì cleanup khi nào
- soft delete có xóa object vật lý ngay không
- object không còn gắn với entity nào thì xử lý ra sao
- có versioning khi admin thay file hay không

### 4B.10 Tích hợp với các module khác

Upload service nên liên kết rõ với:

- `question` module: hình minh họa, file đề bài theo câu hỏi
- `content` module: tài liệu học tập
- `curriculum` module: file theo chapter/topic
- `admin` workflow: import đề thi, duyệt nội dung, cập nhật asset

Nếu sau này có OCR/import parser, upload service sẽ là điểm vào chuẩn cho pipeline đó.

### 4B.11 Python Parse Worker đề xuất

Worker này chịu trách nhiệm:

- poll hoặc consume `parse_job`
- tải file PDF từ S3
- extract text từ PDF text-based
- parse thành draft questions, options, answer hints
- lưu kết quả xuống DB
- cập nhật trạng thái job và lỗi parse

Khuyến nghị công nghệ:

- `Python 3.x`
- `boto3` để tải file từ S3
- `PyMuPDF` hoặc `pdfplumber` để extract text từ PDF có text
- parser rule-based bằng regex + heuristic theo format đề
- `psycopg` hoặc `SQLAlchemy` để ghi DB

Phạm vi MVP:

- ưu tiên **PDF có text sẵn**
- chưa bắt buộc OCR cho scan/image ở phase đầu
- output là **draft questions**, không publish thẳng

Các trạng thái nên có cho `parse_job`:

- `QUEUED`
- `PROCESSING`
- `PARSED`
- `REVIEW_REQUIRED`
- `FAILED`

### 4B.12 Dữ liệu đầu ra từ parser

Tối thiểu parser nên lưu được:

- `question_number`
- `question_text`
- `question_type`
- `option_label`
- `option_text`
- `explanation_raw` *(nếu có)*
- `answer_key_raw` *(nếu có)*
- `source_asset_id`
- `parse_confidence` *(nếu có scoring)*

Khuyến nghị:

- lưu dưới dạng **draft tables** trước
- admin chỉnh sửa và xác nhận
- chỉ sau khi confirm mới map sang bảng question chính thức

### 4B.13 Parse failure policy

Tài liệu nên quy định rõ:

- parse lỗi toàn bộ thì job `FAILED`
- parse được một phần thì job `REVIEW_REQUIRED`
- không tạo duplicate draft nếu cùng asset bị retry
- luôn lưu `raw_extracted_text` hoặc link đến artifact debug nếu cần phân tích lỗi
- cho phép admin re-run parse sau khi đổi parser version hoặc rule set

---

## 5. Testing checklist cần bổ sung vào tài liệu

### 5.1 Unit tests

- NotificationType validation
- Template rendering
- Preference gate
- Retry policy
- Domain error mapping

### 5.2 Integration tests

- Postgres repository
- Redis idempotency
- Kafka producer/consumer
- FCM adapter with fake provider

### 5.3 End-to-end tests

- Register token → send reminder → log success
- Disabled preference → skip + log `SKIPPED`
- Invalid token → deactivate token + log `FAILED`
- HTTP 429 → retry → success/fail
- Duplicate message → no duplicate delivery
- Broadcast → fan-out đúng scope

### 5.4 Edge cases

- User đổi timezone
- User có nhiều thiết bị
- Token bị rotate
- Kafka consumer restart
- FCM downtime
- Redis unavailable

### 5.5 Upload service tests

- Tạo presigned URL thành công với role `ADMIN`
- User không phải admin bị từ chối
- Upload file quá size bị reject
- Upload MIME không hợp lệ bị reject
- Complete upload khi object không tồn tại bị fail
- Duplicate checksum được detect đúng theo policy
- Delete asset chỉ soft delete metadata hoặc cả object theo lifecycle rule
- Attach asset vào wrong entity type bị reject
- Presigned URL hết hạn không upload được
- S3 outage được surface lỗi đúng cho admin client

### 5.6 Python parse worker tests

- Worker tạo kết quả đúng với PDF có text sẵn
- PDF corrupt hoặc empty làm job `FAILED`
- Retry parse job không tạo duplicate draft questions
- Parser tách đúng câu hỏi trắc nghiệm A/B/C/D theo format chuẩn
- Parser vẫn lưu partial result khi đề parse được một phần
- Worker cập nhật trạng thái `QUEUED` → `PROCESSING` → `PARSED/FAILED`
- Download từ S3 lỗi thì worker log lỗi và mark job đúng trạng thái
- Admin chỉ publish được question sau khi review draft

---

## 6. Observability checklist cần có

- Metric `notifications_scheduled_total`
- Metric `notifications_sent_total`
- Metric `notifications_failed_total`
- Metric `notifications_skipped_total`
- Metric `notifications_retry_total`
- Metric `notifications_dlq_total`
- Metric `fcm_invalid_token_total`
- Metric `kafka_consumer_lag`
- Metric `api_request_duration_seconds`
- Metric `upload_init_total`
- Metric `upload_complete_total`
- Metric `upload_failed_total`
- Metric `upload_bytes_total`
- Metric `s3_presign_error_total`
- Metric `parse_job_created_total`
- Metric `parse_job_completed_total`
- Metric `parse_job_failed_total`
- Metric `parse_question_draft_total`
- Metric `parse_duration_seconds`

Structured log tối thiểu cần có:

- `correlationId`
- `notificationType`
- `userId`
- `campaignId`
- `deviceTokenHash`
- `status`
- `retryCount`
- `errorCode`
- `assetId`
- `objectKey`
- `contentType`
- `fileSize`
- `parseJobId`
- `parserVersion`
- `assetStatus`
- `parseStatus`

---

## 7. Tài liệu/phụ lục nên tạo thêm sau bước enhancement này

- `PRD section` hoàn chỉnh cho product scope
- `API contract` chi tiết hoặc OpenAPI spec
- `Test plan` cho QA + backend
- `Operations runbook` cho notification pipeline
- `Upload service contract` + object lifecycle policy
- `Parse worker spec` + draft question schema
- `Domain event catalog`
- `Migration plan` cho schema ban đầu

---

## 8. Definition of Done cho bản tài liệu nâng cấp

Tài liệu được xem là đủ tốt để triển khai khi:

1. Có scope MVP rõ ràng và measurable KPI.
2. Mỗi module có boundary và contract rõ.
3. Notification flow không còn mâu thuẫn giữa domain, DB, Kafka và API.
4. API có request/response/error schema đầy đủ.
5. Có acceptance criteria cho các luồng người dùng chính.
6. Có test strategy đủ để backend, mobile, QA cùng bám vào.
7. Có observability và runbook ở mức đủ vận hành production sớm.
8. Có thiết kế upload service đủ để admin upload đề bài lên S3 an toàn và truy vết được.
9. Có parse pipeline rõ ràng để PDF được chuyển thành draft questions trong DB.

---

## 9. Thông tin cần xác nhận thêm từ stakeholder

Đây là các câu hỏi cần được xác nhận để chốt bản PRD/TDD hoàn chỉnh:

1. Edu App đang phục vụ nhóm học sinh nào: phổ thông, luyện thi, hay đại học?
2. MVP có những role nào ngoài `student` và `admin`?
3. `Parent notification` có thuộc roadmap gần hay chỉ là future idea?
4. Broadcast có cần segmentation ngay từ v1 không?
5. Có cần tracking open/click notification từ mobile app trong MVP không?
6. Có quiet hours hoặc chính sách chống spam notification không?
7. KPI quan trọng nhất của v1 là retention, completion rate, hay activation?
8. Có cần đa ngôn ngữ ngay từ phiên bản đầu tiên không?
9. Admin sẽ upload những loại file nào cho đề bài?
10. MVP chỉ hỗ trợ **PDF có text sẵn** hay cần cả scan/OCR ngay từ đầu?
11. File upload có cần cho download công khai trong app hay chỉ dùng nội bộ admin?
12. Giới hạn dung lượng và số lượng file trên mỗi đề là bao nhiêu?
13. Kết quả parse sẽ map sang dạng câu hỏi nào trước: trắc nghiệm A/B/C/D, đúng-sai, hay cả tự luận?
14. Có bắt buộc admin review draft questions trước khi publish không?

---

## 10. Hướng cập nhật đề xuất

Thứ tự cập nhật nên là:

1. Sửa các inconsistency hiện tại trong domain/schema/API.
2. Bổ sung phần Product Vision, Personas, Goals, MVP Scope.
3. Bổ sung User Stories + Acceptance Criteria.
4. Bổ sung Testing Strategy và Observability.
5. Bổ sung thiết kế Upload Service + S3 lifecycle + RBAC.
6. Bổ sung Python Parse Worker + draft question pipeline.
7. Chốt Roadmap, Risks, Assumptions.
