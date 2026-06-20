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

Structured log tối thiểu cần có:

- `correlationId`
- `notificationType`
- `userId`
- `campaignId`
- `deviceTokenHash`
- `status`
- `retryCount`
- `errorCode`

---

## 7. Tài liệu/phụ lục nên tạo thêm sau bước enhancement này

- `PRD section` hoàn chỉnh cho product scope
- `API contract` chi tiết hoặc OpenAPI spec
- `Test plan` cho QA + backend
- `Operations runbook` cho notification pipeline
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

---

## 10. Hướng cập nhật đề xuất

Thứ tự cập nhật nên là:

1. Sửa các inconsistency hiện tại trong domain/schema/API.
2. Bổ sung phần Product Vision, Personas, Goals, MVP Scope.
3. Bổ sung User Stories + Acceptance Criteria.
4. Bổ sung Testing Strategy và Observability.
5. Chốt Roadmap, Risks, Assumptions.

