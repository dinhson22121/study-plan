# GCP Deployment Architecture — Edu App

## Mục tiêu

Thiết kế kiến trúc deploy **toàn bộ project trên Google Cloud** theo stack hiện tại:

- **server/** → Go/Gin modular monolith
- **admin/** → React admin web
- **worker/** → Python PDF parse worker
- **client/** → React Native mobile app dùng Firebase/FCM

Tài liệu này ưu tiên:

1. **chi phí hợp lý** cho giai đoạn Phase 1
2. **vận hành đơn giản**
3. **ít thay đổi code nhất có thể**
4. vẫn giữ đường nâng cấp khi scale tăng

---

## 1. Kết luận kiến trúc khuyến nghị

### Nếu đi GCP cho project này, tôi khuyên:

| Thành phần | GCP service khuyến nghị |
|---|---|
| Admin web | **Firebase Hosting** |
| Backend API | **Cloud Run** |
| Worker parse PDF | **Compute Engine VM nhỏ** *(Phase 1)* |
| PostgreSQL | **Cloud SQL for PostgreSQL** |
| Object storage | **Cloud Storage** |
| Cache / token / idempotency | **Memorystore Redis** *(hoặc self-host Redis nếu cần tiết kiệm hơn)* |
| Queue / event bus | **Pub/Sub** *(khuyến nghị)* |
| Secrets | **Secret Manager** |
| Logs / metrics | **Cloud Logging + Cloud Monitoring** |
| Error reporting | **Sentry** |
| CI/CD | **GitHub Actions + Artifact Registry** |
| Mobile push | **Firebase / FCM** |

---

## 2. Quan điểm thiết kế quan trọng

## 2.1 Không cố map 1:1 mọi thứ từ AWS/Kafka sang GCP nếu không cần

GCP **không có lựa chọn first-party rẻ và đơn giản cho Kafka** như cách AWS có SQS/EventBridge/MSK.

Với scale **100–200 user**, tôi **không khuyên** giữ Kafka như production requirement bắt buộc.

### Khuyến nghị

- **Notification/event pipeline** nên đi theo **Pub/Sub**
- Nếu cần giữ logic gần hiện tại nhất:
  - dùng Go publisher/subscriber abstraction
  - giữ topic naming tương đương logic cũ

## 2.2 Worker không nên cố nhét vào Cloud Functions

Worker parse PDF:

- tải file từ object storage
- parse bằng PyMuPDF
- ghi DB
- có retry / stuck-job recovery

Đây **không phải use case đẹp cho Cloud Functions**.

### Khuyến nghị

**Phase 1**:
- chạy worker dạng **container trên Compute Engine**

**Phase 2**:
- refactor để dùng **Pub/Sub + Cloud Run Jobs**

## 2.3 Cloud Run hợp backend hơn Cloud Functions

Với Go API hiện tại:

- nhiều route
- auth
- DB
- Redis
- metrics
- middleware

=> **Cloud Run** phù hợp hơn **Cloud Functions** rất nhiều.

---

## 3. Kiến trúc tổng thể

```text
Mobile App (React Native)
        |
        | HTTPS
        v
Firebase Hosting / CDN  ---->  Admin Web (React static build)
        |
        v
Cloud Load Balancer / Domain
        |
        v
Cloud Run (Go API)
   |         |           \
   |         |            \
   |         |             ---> Cloud Storage (uploads, assets)
   |         |
   |         ---> Cloud SQL PostgreSQL
   |
   ---> Memorystore Redis
   |
   ---> Pub/Sub (recommended for events / async tasks)

Compute Engine VM
   └── Python parse worker
       ├── poll/consume parse jobs
       ├── read PDF from Cloud Storage
       └── write drafts to Cloud SQL

Firebase / FCM
   └── push notifications to mobile app
```

---

## 4. Mapping từ stack hiện tại sang GCP

| Hiện tại | GCP target | Ghi chú |
|---|---|---|
| `server/` Go API | **Cloud Run** | deploy container |
| `worker/` Python | **Compute Engine** | giữ gần logic hiện tại nhất |
| Postgres local/RDS plan | **Cloud SQL PostgreSQL** | managed DB |
| MinIO/S3 | **Cloud Storage** | đổi endpoint + credentials |
| Redis | **Memorystore** | refresh token, rate limit, idempotency |
| Kafka | **Pub/Sub** | khuyến nghị thay ở Phase 1 |
| Admin React | **Firebase Hosting** | static site rất hợp |
| Mobile RN + FCM | **Firebase** | cùng hệ GCP |
| Secrets Manager/SSM | **Secret Manager** | chuẩn GCP |
| CloudWatch | **Cloud Logging / Monitoring** | chuẩn GCP |

---

## 5. Thiết kế chi tiết từng thành phần

## 5.1 Admin web (`admin/`)

### Service

- **Firebase Hosting**

### Lý do

- static site
- HTTPS/CDN dễ
- deploy nhanh
- hợp với app admin React build ra `dist/`
- miễn phí / rất rẻ ở scale đầu

### Deploy flow

1. `npm ci`
2. `npm run build`
3. deploy `dist/` lên Firebase Hosting

### Env cần có

- `VITE_API_BASE_URL=https://api.<domain>/api/v1`
- `VITE_APP_NAME=Edu Admin`
- `VITE_POLL_INTERVAL_MS=5000`

### Domain

- `admin.<domain>` hoặc `console.<domain>`

---

## 5.2 Backend API (`server/`)

### Service

- **Cloud Run**

### Lý do

- hợp containerized Go service
- scale-to-zero / scale-out tốt
- không cần tự quản VM cho API
- deploy dễ qua image

### Runtime flow

1. Build Docker image từ `server/`
2. Push lên **Artifact Registry**
3. Deploy image lên Cloud Run
4. Gắn custom domain:
   - `api.<domain>`

### Config gợi ý Phase 1

- min instances: `0` hoặc `1` nếu muốn giảm cold start
- max instances: `2–5`
- CPU: `1 vCPU`
- RAM: `512MB–1GB`

### Secret injection

Từ **Secret Manager** vào Cloud Run env:

- `EDU_JWT_SECRET`
- `EDU_POSTGRES_URL`
- `EDU_REDIS_URL`
- `EDU_KAFKA_BROKERS` *(nếu còn dùng)*
- `EDU_S3_*` *(nếu tạm giữ abstraction s3-like)*
- `EDU_SENTRY_DSN`

### Mạng

Cloud Run cần:

- **Serverless VPC Access** nếu truy cập Cloud SQL/Memorystore private IP

---

## 5.3 Database

### Service

- **Cloud SQL for PostgreSQL**

### Lý do

- managed
- backup
- monitoring
- dễ kết nối từ Cloud Run và VM

### Gợi ý Phase 1

- PostgreSQL 16
- single zone/single instance để tiết kiệm
- auto backup bật
- private IP

### Cần làm

- migrate schema
- backup retention
- restore test

---

## 5.4 Object storage

### Service

- **Cloud Storage**

### Cách dùng với code hiện tại

Vì backend hiện đang đi theo abstraction kiểu S3-compatible, bạn có 2 hướng:

### Hướng A — ngắn hạn

- giữ local MinIO cho dev
- viết adapter GCS riêng cho prod

### Hướng B — trung hạn

- trừu tượng hóa storage theo interface chung:
  - presign upload
  - head object
  - read object

Tôi khuyên dùng **interface storage chung**, rồi:

- local → MinIO
- prod GCP → Cloud Storage

### Bucket gợi ý

- `edu-assets-prod`
- bật:
  - versioning nếu cần
  - uniform bucket-level access
  - private bucket

---

## 5.5 Redis

### Service

- **Memorystore Redis**

### Dùng cho

- refresh token blocklist/store
- rate limiting
- idempotency

### Lưu ý

Memorystore trên GCP **khá đắt** nếu dùng instance managed ngay từ đầu.

### Khuyến nghị

**Phase 1**:

- nếu cần tối ưu chi phí, có thể **self-host Redis cùng VM worker**
- nếu ưu tiên managed và ổn định hơn, dùng Memorystore

---

## 5.6 Queue / event system

### Khuyến nghị production Phase 1

- **Pub/Sub**

### Dùng cho

- upload parse trigger
- notification async event
- future background tasks

### Vì sao không giữ Kafka?

- GCP không có lựa chọn first-party rẻ/đơn giản như bạn cần
- Confluent / self-host Kafka sẽ tăng ops và cost

### Chiến lược

**Phase 1**:
- migrate async event sang Pub/Sub

**Nếu chưa refactor kịp**:
- tạm self-host Kafka trên Compute Engine

Nhưng tôi **không khuyên** cách này nếu muốn gọn và rẻ.

---

## 5.7 Worker parse PDF (`worker/`)

### Service khuyến nghị cho Phase 1

- **Compute Engine VM nhỏ**

### Tại sao không Cloud Run ngay?

Worker hiện:

- long-running
- poll job queue
- xử lý file
- reconnect/retry DB

=> hợp với **VM/container luôn chạy** hơn là serverless request model.

### Cách chạy

- Docker container hoặc systemd service
- VM riêng hoặc chung với Redis/self-host Kafka nếu cần tiết kiệm

### Config gợi ý

- `e2-small` hoặc `e2-medium`
- Ubuntu/Debian
- restart policy
- Sentry + logs

### Hướng nâng cấp Phase 2

- đổi sang **Pub/Sub trigger + Cloud Run Jobs**

---

## 5.8 Mobile app (`client/`)

### Dịch vụ liên quan trên GCP

- **Firebase**
- **FCM**
- **Google Play Console**

### Vai trò

- push notification
- crash reporting (qua Sentry riêng)
- hosting static deep-link association nếu cần

### Build/release

- EAS build ra `.aab`
- upload lên Play Console
- backend API trỏ tới `https://api.<domain>/api/v1`

---

## 6. Domain, TLS và routing

## Domain gợi ý

- `api.<domain>` → Cloud Run
- `admin.<domain>` → Firebase Hosting

## TLS

- Cloud Run custom domain + managed cert
- Firebase Hosting cũng có managed SSL

=> đơn giản hơn tự cài Nginx/TLS trên VM

---

## 7. CI/CD đề xuất

## 7.1 Backend

GitHub Actions:

1. test `server/`
2. build Docker image
3. push Artifact Registry
4. deploy Cloud Run

## 7.2 Admin

1. test `admin/`
2. build
3. deploy Firebase Hosting

## 7.3 Worker

1. test/lint `worker/`
2. build image
3. deploy container lên Compute Engine
   - pull image
   - restart service

## 7.4 Mobile

1. typecheck/test
2. EAS build
3. upload Play internal track

---

## 8. Security design

## 8.1 Secrets

Dùng **Secret Manager** cho:

- DB URL
- JWT secret
- Redis URL
- storage credentials nếu có
- Sentry DSN

## 8.2 Network

- Cloud SQL private IP
- Redis private
- worker VM private nếu có thể
- Cloud Run qua VPC connector

## 8.3 Storage

- bucket private
- signed upload URL
- least privilege IAM

## 8.4 Audit / monitoring

- Cloud Logging
- Sentry
- metrics dashboard

---

## 9. Cost-optimized GCP architecture (khuyên dùng)

## Mục tiêu

Giữ chi phí hợp lý cho **100–200 user**.

### Đề xuất

- Admin → Firebase Hosting
- API → Cloud Run
- DB → Cloud SQL nhỏ
- Object storage → Cloud Storage
- Queue → Pub/Sub
- Worker → 1 Compute Engine nhỏ
- Redis → cân nhắc self-host trước, Memorystore sau

### Ước tính

- Cloud Run: `~$0–10`
- Cloud SQL: `~$15–30`
- Cloud Storage: `<$1`
- Pub/Sub: `~$0–1`
- VM worker: `~$5–15`
- Firebase Hosting/admin: gần như `~$0–3`

### Tổng

**~$25–50/tháng**

---

## 10. Full managed GCP architecture

Nếu muốn managed nhiều hơn:

- Admin → Firebase Hosting
- API → Cloud Run
- DB → Cloud SQL
- Storage → Cloud Storage
- Queue → Pub/Sub
- Redis → Memorystore
- Worker → Cloud Run Jobs *(sau khi refactor)*

### Ước tính

**~$40–80+/tháng**

---

## 11. Kiến trúc “giữ stack hiện tại nhất”

Nếu bạn muốn ít đổi code nhất:

- Admin → Firebase Hosting
- API → Cloud Run
- DB → Cloud SQL
- Storage → MinIO local/dev, Cloud Storage prod qua adapter
- Redis → self-host trên worker VM hoặc Memorystore
- Kafka → self-host trên VM
- Worker → VM

### Nhược điểm

- ops nhiều hơn
- không đẹp theo style GCP-native
- cost/complexity không tối ưu

### Kết luận

Chỉ dùng nếu bạn muốn **ít refactor ngay lập tức**.

---

## 12. Khuyến nghị cuối

### Nếu là tôi, tôi sẽ chọn:

**Phase 1 trên GCP**

1. **Admin** → Firebase Hosting
2. **API** → Cloud Run
3. **DB** → Cloud SQL
4. **Object storage** → Cloud Storage
5. **Queue** → Pub/Sub
6. **Worker** → Compute Engine
7. **Secrets** → Secret Manager
8. **Push** → Firebase / FCM

### Vì sao?

- hợp với mobile app dùng Firebase
- managed vừa đủ
- không quá đắt
- không phải ôm Kafka managed
- vẫn giữ worker riêng dễ kiểm soát

---

## 13. Phase migration roadmap

### Phase 1

- Cloud Run API
- Cloud SQL
- Cloud Storage
- Firebase Hosting
- Worker trên VM
- Pub/Sub cho event mới

### Phase 2

- refactor worker sang Cloud Run Jobs
- bỏ polling nếu cần
- giảm VM footprint
- scale queue/event sạch hơn

---

## 14. Go-live checklist cho kiến trúc GCP này

- [ ] Admin deploy lên Firebase Hosting
- [ ] API deploy lên Cloud Run
- [ ] Cloud SQL migrate pass
- [ ] Cloud Storage private bucket + signed upload pass
- [ ] Worker VM parse flow pass
- [ ] Pub/Sub event flow pass
- [ ] Secret Manager wiring pass
- [ ] Sentry + Logging + Monitoring pass
- [ ] Mobile app trỏ đúng `api.<domain>`
- [ ] closed beta pass

---

## 15. Recommendation summary

### Nếu bạn muốn:

- **rẻ nhất gần managed** → **GCP cost-optimized architecture**
- **ít đổi code nhất** → **worker VM + Cloud Run API**
- **đẹp, cloud-native hơn về sau** → **Pub/Sub + Cloud Run Jobs**

**Chốt lại:** với project của bạn, **GCP là lựa chọn rất hợp**, nhưng tôi khuyên
không cố giữ Kafka nguyên xi. Dùng **Cloud Run + Cloud SQL + Cloud Storage + Pub/Sub + Firebase Hosting + worker VM** là phương án tốt nhất cho Phase 1.
