# Phase 2 Microservices Plan

## Mục tiêu

Sau khi **Phase 1 đã go-live ổn định**, triển khai kế hoạch **microservice hóa có kiểm soát**
cho Edu App, với các mục tiêu:

- tách dần các bounded context có nhu cầu scale/vận hành riêng
- giảm coupling trong modular monolith hiện tại
- đưa **Kafka** (hoặc event bus tương đương) về đúng vai trò **async backbone**
- tránh “big bang rewrite”
- không phá vỡ sản phẩm đang chạy production

> **Nguyên tắc cốt lõi:** Phase 2 là **evolutionary extraction**, không phải viết lại toàn bộ hệ thống.

---

## 1. Khi nào mới được bắt đầu Phase 2

Chỉ bắt đầu microservice hóa khi Phase 1 đã đạt đủ:

1. production ổn định ít nhất **4–8 tuần**
2. không còn P0/P1 kéo dài
3. monitoring + tracing + runbook đã đủ dùng
4. team đã có dữ liệu thật về:
   - traffic
   - bottleneck
   - chi phí
   - error rate
   - latency

### Không nên bắt đầu Phase 2 nếu:

- vẫn còn loay hoay release mobile
- backend chưa ổn định
- chưa có dữ liệu tải thực tế
- chưa có owner rõ cho từng domain

---

## 2. Mục tiêu thực tế của microservices

Microservices ở đây không nhằm “cho hiện đại”, mà phải giải quyết ít nhất một trong các bài toán:

- **scale khác nhau** giữa các domain
- **tài nguyên khác nhau** (API vs worker vs parsing vs notification)
- **team ownership rõ ràng**
- **release độc lập**
- **failure isolation**
- **event-driven side effects**

---

## 3. Quan điểm tách service

## 3.1 Không tách 12 module thành 12 service ngay

Hiện tại trong monolith có nhiều module, nhưng nếu đi microservices quá mịn sẽ tạo:

- quá nhiều network hop
- quá nhiều DB
- quá nhiều deployment unit
- quá nhiều operational overhead

### Gợi ý số service hợp lý cho Phase 2

**6–7 service** là đủ hợp lý:

1. **Identity Service**
2. **Learning Catalog Service**
3. **Assessment Service**
4. **Planning Service**
5. **Progress & Analytics Service**
6. **Notification Service**
7. **Asset / Import Service**

---

## 4. Service decomposition đề xuất

## 4.1 Identity Service

### Chứa

- auth
- user credential
- refresh token
- role / permission
- profile cơ bản

### Trách nhiệm

- register
- login
- refresh/logout
- token validation
- user basic identity

### Lý do giữ riêng

- security-sensitive
- contract rõ
- ít phụ thuộc vào domain học tập

---

## 4.2 Learning Catalog Service

### Chứa

- curriculum
- subject
- chapter
- topic
- lesson
- content item

### Trách nhiệm

- quản lý nội dung tĩnh / bán tĩnh
- phục vụ lesson browsing
- phục vụ question/topic metadata

### Ghi chú

Question bank có thể:

- để chung với Assessment ở giai đoạn đầu
- hoặc tách sau khi đủ lớn

---

## 4.3 Assessment Service

### Chứa

- placement
- question bank
- quiz
- grading

### Trách nhiệm

- sinh quiz
- submit quiz
- placement test
- review kết quả

### Lý do nên là 1 service riêng

- đây là domain có nhiều logic sync
- latency quan trọng
- load có thể tăng mạnh theo mùa thi

> Đây là service **không nên phụ thuộc Kafka cho request chính**. Kafka chỉ dùng cho event phát sinh sau khi submit/grade xong.

---

## 4.4 Planning Service

### Chứa

- goal
- studyplan
- milestone generation

### Trách nhiệm

- lưu goal
- dùng placement + curriculum để sinh study plan
- cập nhật trạng thái milestone

### Đặc điểm

- sync vừa phải
- phụ thuộc dữ liệu từ Assessment + Catalog

---

## 4.5 Progress & Analytics Service

### Chứa

- progress
- streak
- mastery
- achievements
- activity events
- weak-topic summary

### Trách nhiệm

- consume event từ quiz/lesson/plan
- update progress read model
- phục vụ dashboard analytics

### Lý do rất hợp microservice

- thiên về **event-driven**
- có thể scale độc lập
- dễ tách read models

---

## 4.6 Notification Service

### Chứa

- device token
- notification preference
- notification template
- notification log
- push sender
- reminder scheduler
- broadcast
- re-engagement

### Trách nhiệm

- nhận event cần notify
- gửi push
- retry / DLQ
- lưu lịch sử gửi

### Vì sao đây là service đầu tiên nên tách

- đã async sẵn
- đã có pipeline Kafka rõ nhất
- side-effect rõ ràng
- phụ thuộc FCM, retry, delivery logic riêng

---

## 4.7 Asset / Import Service

### Chứa

- uploaded_asset
- parse_job
- draft question
- PDF parse worker

### Trách nhiệm

- upload file
- object storage
- parse PDF
- draft review input
- import sang question bank

### Vì sao nên tách sớm

- workload khác hẳn API chính
- có file/object storage
- có worker dài hạn
- tiêu tốn CPU/RAM khác biệt

---

## 5. Kafka sẽ đóng vai trò gì?

## 5.1 Kafka không phải request/response bus

Kafka **không** dùng để:

- login
- load dashboard trực tiếp
- tạo quiz rồi trả câu hỏi
- submit quiz rồi chờ trả kết quả
- đọc lesson/topic

Những cái này phải giữ ở:

- **HTTP/gRPC synchronous**

## 5.2 Kafka là async integration backbone

Kafka phù hợp cho:

- event fan-out
- eventual consistency
- side-effect processing
- replay
- backpressure
- audit stream

## 5.3 Kafka nên xử lý các event nào?

### Identity

- `identity.user.registered`

### Assessment

- `assessment.placement.completed`
- `assessment.quiz.started`
- `assessment.quiz.completed`

### Planning

- `planning.studyplan.generated`
- `planning.milestone.completed`

### Progress & Analytics

- `progress.updated`
- `analytics.activity.recorded`
- `achievement.awarded`

### Notification

- `notification.requested`
- `notification.scheduled`
- `notification.sent`
- `notification.failed`
- `notification.dlq`

### Asset / Import

- `asset.uploaded`
- `parse.requested`
- `parse.completed`
- `parse.failed`

---

## 6. Event flow mẫu

## 6.1 Quiz completed

1. Client submit quiz → **Assessment Service**
2. Assessment Service grade xong → trả kết quả ngay cho client
3. Assessment publish event:
   - `assessment.quiz.completed`
4. Progress & Analytics Service consume:
   - update mastery
   - update streak
   - update weak topics
5. Notification Service consume:
   - nếu đạt milestone hoặc achievement → push notification

## 6.2 Study plan generated

1. Client tạo goal / generate plan → **Planning Service**
2. Planning generate xong → trả studyplan
3. Planning publish event:
   - `planning.studyplan.generated`
4. Notification Service consume:
   - schedule reminder

## 6.3 Asset uploaded

1. Admin upload file → **Asset Service**
2. Asset Service verify object xong → tạo parse job
3. Asset Service publish event:
   - `parse.requested`
4. Parse worker consume → parse PDF → ghi draft
5. Publish event:
   - `parse.completed` hoặc `parse.failed`

---

## 7. Database strategy

## 7.1 Nguyên tắc

- mỗi service **own data** của nó
- không query DB service khác trực tiếp
- không join cross-service DB

## 7.2 Cách triển khai thực tế ở Phase 2

### Giai đoạn đầu

Để tiết kiệm chi phí, bạn **không cần** mở 7 database managed riêng ngay.

### Phương án thực tế

- dùng **1 PostgreSQL cluster**
- chia thành:
  - nhiều **schema**
  - hoặc nhiều database logical

Ví dụ:

- `identity.*`
- `catalog.*`
- `assessment.*`
- `planning.*`
- `progress.*`
- `notification.*`
- `asset.*`

### Giai đoạn sau

Khi scale lớn hơn:

- tách hẳn DB vật lý theo service cần scale độc lập

## 7.3 Data migration strategy

### Mục tiêu

Di chuyển dữ liệu của từng bounded context về đúng **service-owned schema/database**
mà **không làm gián đoạn production**.

### Nguyên tắc

1. **Không migrate kiểu big bang**
2. **Không đổi vừa code vừa data vừa topology trong một lần**
3. Mỗi lần migrate phải có:
   - backfill plan
   - verification plan
   - rollback plan
4. Ưu tiên:
   - **schema split trước**
   - **service extraction sau**
5. Đầu tiên tách ở mức:
   - **schema riêng trong cùng 1 PostgreSQL cluster**
   - chưa cần tách DB vật lý ngay

## 7.4 Service ownership map (data)

### Identity Service

Schema đề xuất:

- `identity.user_credential`
- `identity.user_profile`
- `identity.refresh_token_store` *(nếu sau này persistence hóa khỏi Redis hoặc cần audit)*

### Learning Catalog Service

Schema đề xuất:

- `catalog.subject`
- `catalog.chapter`
- `catalog.topic`
- `catalog.lesson`
- `catalog.content_item`

### Assessment Service

Schema đề xuất:

- `assessment.question`
- `assessment.answer_option`
- `assessment.placement_test`
- `assessment.placement_result`
- `assessment.quiz`
- `assessment.quiz_question`
- `assessment.quiz_answer`
- `assessment.quiz_result`

### Planning Service

Schema đề xuất:

- `planning.goal`
- `planning.study_plan`
- `planning.study_plan_milestone`

### Progress & Analytics Service

Schema đề xuất:

- `progress.topic_progress`
- `progress.mastery`
- `progress.streak`
- `progress.achievement`
- `analytics.activity_event`
- `analytics.weak_topic_snapshot`

### Notification Service

Schema đề xuất:

- `notification.device_token`
- `notification.preference`
- `notification.template`
- `notification.log`
- `notification.audit` *(nếu thêm sau)*

### Asset / Import Service

Schema đề xuất:

- `asset.uploaded_asset`
- `asset.parse_job`
- `asset.question_draft`
- `asset.question_draft_option`

## 7.5 Migration pattern đề xuất cho mỗi service

### Step 1 — Ownership freeze

Trước khi di chuyển:

- xác định bảng nào thuộc service nào
- freeze thay đổi schema liên quan trong monolith
- ghi rõ source of truth hiện tại

### Step 2 — Tạo schema mới

Ví dụ:

- `notification.*`
- `asset.*`
- `progress.*`

Tạo:

- schema mới
- bảng mới
- index mới
- FK nội bộ trong cùng schema nếu cần

### Step 3 — Backfill dữ liệu

Backfill từ bảng cũ → bảng mới:

- chạy theo batch
- có idempotency
- có checkpoint
- có row count verification

### Step 4 — Dual write / change capture

Trong thời gian chuyển đổi:

- monolith ghi cả **old tables** và **new tables**
- hoặc dùng **outbox/event** để sync chênh lệch

> Với Phase 2, tôi khuyên **dual write có thời hạn ngắn** thay vì CDC phức tạp nếu traffic chưa quá lớn.

### Step 5 — Read switch

Chuyển reader/service sang đọc từ:

- schema mới

nhưng vẫn giữ:

- old path để rollback nhanh

### Step 6 — Write switch

Khi đã verify read ổn:

- dừng ghi vào bảng cũ
- chỉ ghi vào bảng mới

### Step 7 — Decommission

Sau 1 khoảng quan sát an toàn:

- archive snapshot
- drop write path cũ
- xóa bảng cũ ở phase sau

## 7.6 Verification checklist cho mỗi migration

Mỗi lần migrate bảng/service phải verify:

1. row count source = row count target
2. checksum/sample data đúng
3. read API không thay đổi contract
4. write path mới pass integration test
5. dashboard/metrics không lệch
6. rollback path test được

## 7.7 Rollback strategy

Mỗi migration phải có rollback riêng:

### Nếu lỗi ở backfill

- dừng job backfill
- truncate data target nếu cần
- chạy lại từ checkpoint

### Nếu lỗi sau read switch

- chuyển read path về bảng cũ

### Nếu lỗi sau write switch

- bật lại dual write hoặc trả write về bảng cũ

### Luôn cần

- DB snapshot trước migration lớn
- migration log có timestamp + batch id + operator

## 7.8 Service-by-service migration order

### 1. Notification Service

Bảng migrate trước:

- `device_token`
- `notification_preference`
- `notification_template`
- `notification_log`

Lý do:

- domain rõ
- ít ảnh hưởng business sync
- Kafka/event flow sẵn có

### 2. Asset / Import Service

Bảng migrate:

- `uploaded_asset`
- `parse_job`
- `question_draft`
- `question_draft_option`

Lý do:

- workflow riêng
- admin-only
- worker-heavy

### 3. Progress & Analytics Service

Bảng migrate:

- progress/mastery/streak
- achievement
- activity_event

Lý do:

- event-driven
- read model tự nhiên

### 4. Assessment Service

Bảng migrate:

- `question*`
- `placement*`
- `quiz*`

Lý do:

- logic sync mạnh
- ảnh hưởng trực tiếp client
- nên làm sau khi event backbone đã ổn

### 5. Planning Service

Bảng migrate:

- `goal`
- `study_plan`
- `study_plan_milestone`

### 6. Identity Service

Identity có thể migrate sớm hoặc muộn, nhưng nếu auth đang ổn định và ít bottleneck
thì có thể **để sau** để giảm rủi ro login/auth regression.

## 7.9 Deliverable bắt buộc cho phần migration

Trước khi tách mỗi service phải có:

- ownership matrix
- bảng mapping source → target schema
- SQL migrations
- backfill script/job
- dual write window plan
- rollback playbook
- verification report

## 7.10 Điều kiện để được tách DB vật lý riêng

Chỉ tách khỏi 1 cluster Postgres chung khi có ít nhất một lý do mạnh:

- volume quá lớn
- workload quá khác nhau
- cần HA/backup riêng
- team ownership đủ mạnh
- compliance/security boundary yêu cầu

Nếu chưa có các điều kiện này, giữ:

- **1 cluster**
- **nhiều schema**

vẫn là phương án hợp lý nhất về cost và ops.

---

## 8. Cách tách theo phase

## Phase 2A — Chuẩn bị trước khi tách

### Mục tiêu

Làm monolith sẵn sàng cho extraction.

### Việc cần làm

1. Chuẩn hóa service boundaries trong code
2. Tạo interface rõ cho outbound dependency
3. Chuẩn hóa domain events
4. Thêm correlation-id / trace-id end-to-end
5. Chuẩn hóa API contract và event contract
6. Tạo service ownership matrix

### Deliverable

- service boundary doc
- event catalog v1
- dependency matrix

---

## Phase 2B — Extract Notification Service

### Vì sao làm đầu tiên

- async nhất
- side effect rõ
- tách ít ảnh hưởng nhất

### Việc cần làm

1. tách repo/DB/schema notification
2. tách consumer/producer riêng
3. để monolith publish event thay vì gọi notification internals
4. expose notification management API riêng nếu cần

### Deliverable

- Notification Service deploy riêng
- Kafka topics riêng hóa
- no direct package import từ monolith cũ

### Exit criteria

- push flow end-to-end pass
- reminder / broadcast / re-engagement pass
- DLQ + retry pass

---

## Phase 2C — Extract Asset / Import Service

### Vì sao làm sớm

- worker + file storage là workload riêng
- admin flow phụ thuộc mạnh

### Việc cần làm

1. tách upload API
2. tách parse worker
3. tách draft question store
4. giữ publish sang Assessment/Question bank qua API hoặc event

### Deliverable

- Asset Service riêng
- worker riêng
- object storage ownership riêng

---

## Phase 2D — Extract Progress & Analytics Service

### Vì sao

- event-driven tự nhiên
- phù hợp read-model service

### Việc cần làm

1. consume `assessment.quiz.completed`
2. build read model cho dashboard
3. expose analytics API riêng

### Deliverable

- progress read model riêng
- analytics API riêng

---

## Phase 2E — Extract Assessment Service

### Việc cần làm

1. tách question bank
2. tách quiz engine
3. tách placement engine
4. publish events ra Kafka

### Lưu ý

Đây là bước khó hơn vì ảnh hưởng trực tiếp tới UX sync.

---

## Phase 2F — Extract Planning Service

### Việc cần làm

1. tách goal
2. tách studyplan generation
3. consume placement/catalog data qua API/events

---

## 9. Sync vs Async rule

## Dùng HTTP/gRPC khi:

- cần response ngay cho client
- cần validate business rule đồng bộ
- request duration phải predictable

## Dùng Kafka khi:

- không cần trả kết quả ngay
- có nhiều consumer
- cần retry/fan-out/replay
- side effect nằm ngoài request chính

---

## 10. Service communication rule

### Synchronous

- Identity → token validation
- Assessment ← Catalog (question/topic metadata)
- Planning ← Assessment/Catalog

### Asynchronous

- Assessment → Progress
- Assessment → Analytics
- Assessment → Notification
- Planning → Notification
- Asset → Parse worker

---

## 11. Anti-pattern cần tránh

1. tách quá nhiều service nhỏ
2. dùng Kafka cho mọi thứ
3. để service A query DB của service B
4. thiếu event versioning
5. big bang rewrite
6. chưa có observability đã tách service

---

## 12. Điều kiện để mỗi extraction được phép làm

Mỗi service chỉ được tách khi:

- có owner rõ
- có API/event contract rõ
- có test contract
- có rollback plan
- có metrics
- có dashboard/logs/alerts

---

## 13. Timeline đề xuất

### Sau go-live Phase 1

**0–1 tháng đầu**
- chỉ monitor prod
- chưa tách service

**Tháng 2–3**
- Phase 2A
- Phase 2B (Notification)

**Tháng 3–4**
- Phase 2C (Asset / Import)

**Tháng 4–6**
- Phase 2D (Progress & Analytics)

**Sau đó**
- mới xét tiếp Assessment / Planning nếu cần

---

## 14. Recommendation cuối

Nếu làm Phase 2 đúng cách, tôi khuyên:

1. **Không tách tất cả cùng lúc**
2. **Bắt đầu với Notification**
3. **Tách Asset/Import ngay sau đó**
4. **Kafka chỉ dùng cho event backbone**
5. **Sync request vẫn để HTTP/gRPC**
6. **1 Postgres cluster + nhiều schema** là đủ cho giai đoạn đầu của microservices

---

## 15. Kết luận

### Microservice plan hợp lý cho Edu App là:

- **7 coarse-grained services**
- **Kafka làm async backbone**
- **extraction theo từng bước**
- **không rewrite toàn bộ**

### Chốt một câu:

**Phase 2 nên là quá trình tách dần từ modular monolith sang coarse-grained microservices, trong đó Notification và Asset/Import là hai service nên được extract đầu tiên, còn Kafka giữ vai trò event backbone cho side-effect, fan-out và eventual consistency.**
