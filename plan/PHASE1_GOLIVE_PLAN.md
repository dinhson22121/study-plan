# Phase 1 Go-Live Plan

## Mục tiêu

Đưa sản phẩm từ trạng thái **MVP/dev + prototype** lên mức **production-ready** cho
Phase 1, với các điều kiện:

- **Backend thật** chạy trên **EC2**
- **Database thật** dùng **PostgreSQL production**
- **Admin web** thao tác ổn định ở mức prod
- **Mobile app Android** đủ chất lượng để đẩy lên **Google Play**
- **PDF upload/parse worker** vận hành được ngoài môi trường local
- **Không còn known Critical/High security issue**
- **Không còn known blocker bug**
- Có **monitoring, backup, rollback, smoke test** và **release checklist**

---

## 1. Trạng thái hiện tại

| Thành phần | Trạng thái hiện tại | Ghi chú |
|---|---|---|
| `server/` | Khá hoàn chỉnh | `go test ./...` pass, nhiều module đã xong |
| `admin/` | Khá hoàn chỉnh | FE build pass, test pass ở mức cơ bản |
| `worker/` | MVP | Đủ cho parse PDF text-based, cần hardening prod |
| `client/` | Chưa production-ready | Hiện mới là spec + prototype HTML |
| Infra | Dev/local ready | Có Docker Compose local, chưa phải prod infra |
| Security | Chưa đạt prod | `server/SECURITY.md` còn deferred items bắt buộc phải đóng |

---

## 2. Định nghĩa Go-Live Phase 1

Sản phẩm chỉ được coi là **đủ điều kiện go-live Phase 1** khi đạt đủ:

1. Backend production deploy pass trên AWS
2. Admin web production deploy pass
3. Worker production deploy pass
4. Mobile Android release app pass QA + beta
5. Không còn known Critical/High security issue
6. Blocker bug = 0
7. Crash-free beta đạt ngưỡng
8. Có backup/restore, monitoring, rollback

---

## 3. Kiến trúc production target

### 3.1 Kiến trúc khuyến nghị

| Thành phần | Chọn cho Phase 1 |
|---|---|
| API backend | **Go app trên EC2** |
| Reverse proxy / TLS | **Nginx + HTTPS** |
| Database | **Amazon RDS PostgreSQL** |
| Object storage | **Amazon S3** |
| Redis | **Amazon ElastiCache Redis** *(khuyến nghị)* |
| Kafka | **Amazon MSK Serverless** *(khuyến nghị)* |
| Parse worker | **Python worker trên EC2** |
| Admin web | build static từ `admin/`, serve qua **Nginx** hoặc **S3 + CloudFront** |
| Mobile app | React Native TypeScript build Android release |
| Monitoring | **CloudWatch + Sentry** |
| Secrets | **AWS SSM Parameter Store / Secrets Manager** |

### 3.2 Kiến trúc đơn giản nhất để launch

Nếu cần đơn giản hóa trong giai đoạn đầu:

- 1 EC2 cho API + worker + Nginx
- 1 RDS PostgreSQL
- 1 S3 bucket
- Redis/Kafka có thể self-host tạm trên EC2

> Tuy nhiên đây **không phải** cấu hình tôi khuyên dùng lâu dài cho prod.

---

## 4. Workstream tổng thể

| Workstream | Kết quả cần đạt |
|---|---|
| A. Scope freeze | Chốt phạm vi Phase 1 và quality bar |
| B. Backend hardening | Backend đạt chuẩn production |
| C. AWS infra | Môi trường prod thật sẵn sàng |
| D. Admin web prod | Admin thao tác ổn định ở mức prod |
| E. Mobile app | App Android thật đủ để lên Google Play |
| F. Worker hardening | Parse pipeline chạy ổn trên prod |
| G. Security closure | Không còn critical/high issue |
| H. QA / Beta | Không crash, không blocker |
| I. Release ops | Có deploy rehearsal, rollback, runbook |

---

## 5. Workstream A — Scope freeze & release criteria

### Mục tiêu

Khóa đúng những gì thuộc Phase 1 để không kéo dài dự án vô hạn.

### Việc phải làm

1. Chốt **in-scope**:
   - auth
   - onboarding
   - placement
   - study plan
   - lesson/topic
   - quiz
   - progress
   - notification
   - admin upload/parse/review/publish
2. Chốt **out-of-scope**:
   - OCR scan nâng cao
   - iOS release
   - advanced analytics
   - multi-region
3. Chốt **go/no-go gates**
4. Chốt **severity policy**
5. Chốt **pilot user group**

### Deliverable

- scope document
- acceptance criteria
- severity matrix
- go/no-go checklist

### Exit criteria

- Scope được signoff bởi owner sản phẩm
- Không còn ambiguity lớn về feature list Phase 1

---

## 6. Workstream B — Backend production hardening (`server/`)

### Mục tiêu

Đưa `server/` từ mức MVP/dev lên mức production-safe.

### B1. Đóng các mục deferred trong `server/SECURITY.md`

Phải xử lý dứt điểm:

1. **Rate limiting**
   - login/register/auth endpoints
   - brute-force protection
2. **Dev secrets**
   - bỏ default secrets khỏi config commit
   - dùng env/secret manager
3. **Transport security**
   - PostgreSQL `sslmode=require`
   - Redis/Kafka có TLS hoặc service managed
4. **Password policy**
   - tăng yêu cầu tối thiểu
   - cân nhắc password quality rule
5. **Access token revocation**
   - define strategy nếu cần revoke sớm

### B2. Config & environment cleanup

Việc làm:

- loại bỏ hoặc vô hiệu hóa các dev defaults khỏi production path
- tạo `.env.example` / deployment config template
- validate runtime config đầy đủ ở startup
- tách env dev/staging/prod

### B3. Observability

Tối thiểu cần có:

- structured logging
- correlation id propagation
- Sentry backend
- metrics:
  - request latency
  - 4xx / 5xx rate
  - Kafka lag
  - parse job failure rate
  - upload failure rate

### B4. Reliability

- verify graceful shutdown
- verify retry/backoff behavior
- migration rollback plan
- readiness/liveness checks rõ ràng
- smoke test sau deploy

### B5. API freeze

- chốt contract cho mobile/admin
- publish API spec hoặc tài liệu contract cố định

### Deliverable

- backend release branch
- security closure report
- prod config template
- monitoring dashboard
- backend runbook

### Exit criteria

- `go test ./...` pass
- integration/e2e pass
- deferred security items closed
- prod config validated

---

## 7. Workstream C — AWS production infra

### Mục tiêu

Có môi trường thật để deploy backend/admin/worker.

### C1. Provision resources

Phải có:

- VPC
- subnets
- security groups
- IAM roles
- EC2
- RDS PostgreSQL
- S3 bucket
- Redis
- Kafka
- Route53 / domain

### C2. Server provisioning

Trên EC2:

- Docker / Docker Compose hoặc systemd
- Nginx
- HTTPS
- log rotation
- backup script / monitoring agent

### C3. Secrets management

- SSM Parameter Store hoặc Secrets Manager
- không hardcode secret vào repo
- rotation policy cho JWT secret / DB creds nếu có

### C4. Data layer readiness

- RDS migration pass
- backup snapshot policy
- restore test
- S3 lifecycle / bucket policy

### C5. Deploy pipeline

Khuyến nghị:

- GitHub Actions
- build artifacts/images
- deploy to EC2
- smoke test sau deploy

### Deliverable

- AWS prod environment
- deploy pipeline
- infra doc
- restore doc

### Exit criteria

- Domain + HTTPS live
- API healthcheck pass trên domain thật
- Backup + restore test pass ít nhất 1 lần

---

## 8. Workstream D — Admin web production readiness (`admin/`)

### Mục tiêu

Admin thao tác được trong môi trường thật, không chỉ local demo.

### D1. Auth/session hardening

- refresh token flow ổn định
- unauthorized redirect đúng
- logout sạch
- error state rõ

### D2. UX production

- loading/error/empty đầy đủ
- pagination usable
- search/filter usable
- delete/publish confirm dialog
- state sync sau mutation

### D3. Testing

- Vitest cho auth / token / API helpers
- thêm test cho flows quan trọng
- Playwright smoke:
  - login
  - upload
  - complete upload
  - review draft
  - publish

### D4. Deployment

- build production pass
- serve static qua Nginx hoặc S3+CloudFront
- env prod rõ ràng
- source map / Sentry release nếu dùng

### Deliverable

- admin prod build
- FE test suite
- admin deployment doc
- UAT signoff

### Exit criteria

- `npm run build` pass
- FE smoke test pass
- Admin UAT pass

---

## 9. Workstream E — Mobile app production implementation (`client/`)

### Mục tiêu

Biến `client/` từ prototype HTML thành **React Native app thật** có thể lên Google Play.

### E1. Chốt mobile tech stack

Theo tài liệu hiện có:

- React Native
- TypeScript
- React Navigation
- Zustand
- Axios
- MMKV
- Firebase Cloud Messaging
- Sentry React Native

### E2. Module Phase 1 bắt buộc

1. Splash
2. Login
3. Onboarding
4. Placement Test
5. Generate Study Plan
6. Dashboard
7. Study Topic / Lesson
8. Quiz
9. Quiz Result
10. Progress
11. Notifications
12. Profile

### E3. Chất lượng app

Phải có:

- secure token storage
- global error handling
- offline/error state
- retry logic hợp lý
- crash reporting
- loading states nhất quán

### E4. Android release readiness

- package name
- app icon / splash
- signing config
- privacy policy
- permission disclosure
- Play Console listing
- screenshots / description
- signed `.aab`

### E5. QA for mobile

Phải test:

- fresh install
- login
- onboarding
- placement
- study plan generation
- quiz
- progress
- push notification open
- logout/login lại

Test trên:

- Android 10/11/12/13/14
- ít nhất 1 thiết bị RAM thấp
- mạng yếu / mất mạng / reconnect

### Deliverable

- React Native app thật
- signed AAB
- Play Console assets
- beta feedback log

### Exit criteria

- app release build pass
- pilot/beta pass
- crash-free beta đạt ngưỡng

---

## 10. Workstream F — Worker hardening (`worker/`)

### Mục tiêu

Luồng parse PDF chạy ổn trên prod.

### Việc làm

1. Docker/systemd runtime rõ ràng
2. restart policy
3. logging + Sentry nếu cần
4. stuck job detection
5. retry / failure handling
6. supported PDF format policy
7. unsupported file UX rõ cho admin

### Deliverable

- worker production config
- worker runbook
- parse support matrix

### Exit criteria

- worker chạy ổn nhiều vòng
- retry/reparse pass
- failure alerts hoạt động

---

## 11. Workstream G — Security closure

### Mục tiêu

Không còn known Critical/High issue trước launch.

### Checklist bắt buộc

- [ ] JWT secret thật
- [ ] secret management chuẩn
- [ ] rate limiting auth endpoints
- [ ] HTTPS everywhere
- [ ] secure headers
- [ ] private S3 bucket
- [ ] Redis/Kafka không public lung tung
- [ ] dependency audit pass
- [ ] admin action audit log
- [ ] Firebase credential quản lý đúng
- [ ] Play Store privacy/data disclosure đúng

### Deliverable

- security signoff
- dependency audit report
- remediation log

### Exit criteria

- 0 known Critical
- 0 known High

---

## 12. Workstream H — QA, beta, non-crash gate

### Mục tiêu

Đảm bảo launch không vỡ vì blocker bug hoặc crash loop.

### H1. Backend QA

- unit tests
- integration tests
- e2e tests
- deploy smoke tests
- migration rollback test

### H2. Admin QA

- auth/session tests
- upload/parse/review/publish flow
- browser sanity check

### H3. Mobile QA

- real device test matrix
- poor network tests
- notification tests
- session restore tests
- background/foreground behavior

### H4. Crash gate

Ngưỡng đề xuất:

- **Crash-free beta >= 99.5%**
- **ANR = 0 blocker**
- **Startup crash = 0**

### Deliverable

- QA report
- beta report
- blocker list = 0

### Exit criteria

- P0 bug = 0
- P1 bug = 0
- crash-free beta đạt ngưỡng

---

## 13. Workstream I — Release operations

### Mục tiêu

Go-live có kiểm soát, rollback được nếu có vấn đề.

### Việc làm

1. Release checklist
2. Cutover plan
3. Rollback plan
4. Post-launch monitoring window 24h / 7 ngày
5. Incident owner + escalation matrix

### Deliverable

- release runbook
- rollback runbook
- support ownership list

### Exit criteria

- dry-run release pass
- rollback drill pass

---

## 14. Thứ tự triển khai khuyến nghị

1. Scope freeze
2. Backend hardening
3. AWS infra
4. Admin prod readiness
5. Mobile app implementation
6. Worker hardening
7. QA + beta
8. Release rehearsal
9. Go-live

---

## 15. Go/No-Go gates

### Gate 1 — Backend prod-ready

- backend tests pass
- security deferred closed
- deploy rehearsal pass

### Gate 2 — Admin prod-ready

- FE build/test pass
- admin UAT pass

### Gate 3 — Mobile beta-ready

- Android release build pass
- real-device QA pass
- Play internal testing pass

### Gate 4 — Launch-ready

- infra stable
- monitoring live
- rollback ready
- no high/critical issue

---

## 16. Deliverable cuối cùng của Phase 1

Khi hoàn tất, phải có:

- backend production trên EC2
- PostgreSQL production trên RDS
- S3 production cho asset
- worker production chạy ổn
- admin web production
- Android app release AAB sẵn sàng cho Google Play
- push notification chạy thật
- monitoring + alerting + backup + rollback
- 0 blocker bug / 0 known high security issue

---

## 17. Timeline ước lượng

> Ước lượng thực tế nếu làm nghiêm túc và có thể chạy song song.

| Nhóm việc | Ước lượng |
|---|---|
| Backend + infra hardening | 2–4 tuần |
| Admin prod readiness | 1–2 tuần |
| Mobile app production implementation | 4–8 tuần |
| Worker hardening | 1–2 tuần |
| QA / beta / release rehearsal | 2–3 tuần |

**Tổng:** khoảng **8–15 tuần**

---

## 18. Recommendation

### Cách launch ít rủi ro nhất

- **Wave 1:** backend + worker + admin + infra production → internal UAT
- **Wave 2:** Android closed beta
- **Wave 3:** Google Play production release

### Verdict hiện tại

- **Internal demo / UAT:** có thể
- **Public go-live:** chưa nên cho đến khi hoàn tất toàn bộ gates phía trên
