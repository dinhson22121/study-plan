# Phase 1 — Security Closure (Workstream G)

> Mục tiêu: không còn known Critical/High issue trước launch. Bổ sung cho
> `server/SECURITY.md` (chi tiết từng fix backend).

- **Trạng thái:** PASS có điều kiện — xem §4 (việc cần làm khi deploy AWS).
- **Ngày scan:** chạy lại ngay trước mỗi release.

---

## 1. Checklist bắt buộc (PHASE1 §11)

| # | Mục | Trạng thái | Ghi chú |
|---|---|---|---|
| 1 | JWT secret thật | ✅ | Prod config từ chối secret default/<32 ký tự (`config.validateProduction`) |
| 2 | Secret management chuẩn | ✅ | `EDU_*` env / `.env` (gitignored) / SSM; `config.example.yaml` + `.env.example` |
| 3 | Rate limiting auth endpoints | ✅ | Redis fixed-window (Lua atomic) cho login/register/refresh, 429 |
| 4 | HTTPS everywhere | ✅* | Nginx TLS 1.2/1.3 + HTTP→HTTPS redirect (`deploy/nginx`); *cần cert thật khi deploy |
| 5 | Secure headers | ✅ | HSTS, nosniff, X-Frame-Options DENY, Referrer/Permissions-Policy |
| 6 | Private S3 bucket | ⏳ | Bucket policy private — áp khi provision (xem `deploy/README.md`) |
| 7 | Redis/Kafka không public | ✅* | prod compose bỏ host ports; *managed services cần SG/VPC private |
| 8 | Dependency audit pass | ✅/⚠️ | Backend govulncheck: 0; admin npm: 0; client: 33 (build-tooling, §3) |
| 9 | Admin action audit log | ⏳ | Hành động admin có structured log + correlation-id; audit-trail bảng riêng = Phase 2 |
| 10 | Firebase credential quản lý đúng | ✅ | `firebase-service-account.json` gitignored, mount qua volume; `google-services.json` gitignored ở client |
| 11 | Play Store privacy/data disclosure | ⏳ | Cần điền Data Safety form + privacy policy URL khi nộp Play |

## 2. Đã đóng trong hardening (tóm tắt — chi tiết ở server/SECURITY.md)

- Password policy (10–72 byte, chữ+số), bcrypt, HS256 pinned, refresh rotation+revocation, **access-token revocation** (jti + Redis blocklist).
- Rate limiting, prod config validation (weak secret + sslmode), readiness probe.
- Observability: metrics `/metrics`, Sentry (gated), correlation-id.
- IDOR/ownership checks, parameterized SQL, answer-key ẩn với student, broadcast idempotency.

## 3. Dependency audit — kết quả thật

| Thành phần | Công cụ | Kết quả |
|---|---|---|
| **server/** (Go) | `govulncheck ./...` | **No vulnerabilities found** ✅ |
| **admin/** (npm) | `npm audit` | **0 vulnerabilities** ✅ |
| **client/** (npm) | `npm audit` | **33 (27 moderate, 6 high)** ⚠️ |
| **worker/** (pip) | `pip-audit` | ⏳ chạy trong CI/Docker (máy build không có Python) |

**Client — phân tích 33 cảnh báo:** toàn bộ nằm trong **chuỗi build-tooling của Expo**
(`@expo/prebuild-config` → `@expo/config` / `@expo/config-plugins`, và `expo-splash-screen`).
Đây là **dev/build-time tooling**, KHÔNG phải code runtime ship cho người dùng.
- **Rủi ro thực tế:** thấp (không nằm trong app bundle chạy trên thiết bị).
- **Remediation:** nâng `expo-splash-screen` + chạy `npx expo install --fix` theo bản vá Expo SDK; chạy lại `npm audit` trước release. Theo dõi trong §4.

## 4. Việc còn lại trước public launch (infra-bound)

1. **Cert TLS thật** (ACM/Let's Encrypt) gắn vào Nginx/ALB; bật HSTS preload sau khi xác nhận.
2. **S3 bucket policy private** + chặn public access; IAM least-privilege cho app/worker.
3. **Redis/Kafka private** trong VPC (SG chỉ cho app/worker); MSK dùng TLS/SASL.
4. **Client deps:** `expo install --fix`, audit lại về 0 high.
5. **worker pip-audit** chạy trong Docker/CI.
6. **Play Data Safety** + privacy policy URL.
7. **Audit-trail** cho hành động admin (nếu yêu cầu compliance) — cân nhắc Phase 2.

## 5. Verdict

- **0 known Critical**, **0 known High ở backend/admin runtime.**
- Client high-severity là build-tooling (không runtime) → **không chặn internal/closed beta**; cần đưa về 0 trước **public production**.
- **Điều kiện PASS đầy đủ:** hoàn tất §4 mục 1–4 trên môi trường AWS thật + re-scan.

## 6. Sign-off

| Vai trò | Tên | Ngày | Trạng thái |
|---|---|---|---|
| Security reviewer | | | ☐ |
| Tech Lead | | | ☐ |
