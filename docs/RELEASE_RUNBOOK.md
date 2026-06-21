# Phase 1 — Release Runbook (Workstream I)

> Go-live có kiểm soát, rollback được. Dùng kèm `deploy/README.md` (provision)
> và `deploy/scripts/` (deploy + smoke).

---

## 1. Release checklist (pre-cutover)

- [ ] Tất cả go/no-go gates pass (`docs/PHASE1_SCOPE_FREEZE.md` §5)
- [ ] CI xanh trên commit release (`.github/workflows/ci.yml`)
- [ ] `docs/SECURITY_CLOSURE.md` PASS; cert TLS thật đã gắn
- [ ] DB migration review + chạy thử trên staging (`make migrate-up` / rollback test)
- [ ] Backup snapshot RDS ngay trước cutover
- [ ] Secrets đã nạp (SSM/Secrets Manager): `EDU_JWT_SECRET`, DB creds, `EDU_S3_*`, `EDU_SENTRY_DSN`
- [ ] Image tag release đã build & push; ghi lại tag **đang chạy hiện tại** (để rollback)
- [ ] Monitoring/alerting live (Sentry, CloudWatch, `/metrics` scrape)
- [ ] Thông báo maintenance window cho stakeholders

## 2. Cutover (backend + worker + admin + infra)

```bash
# Trên EC2 (hoặc qua .github/workflows/deploy.yml)
cd /opt/edu-app
git fetch && git checkout <release-tag>
export IMAGE_TAG=<release-tag>          # và các EDU_* secrets từ SSM
bash deploy/scripts/deploy.sh           # build → migrate (one-shot) → up -d (prod override) → smoke
```

`deploy.sh` tự gọi `smoke-test.sh`. Nếu smoke fail → **không tiếp tục**, chuyển §4.

- **Admin web:** build (`npm ci && npm run build`) → serve `admin/dist` qua Nginx/S3+CloudFront.
- **Mobile:** `eas build --profile production --platform android` → `.aab` → Play Console
  (internal → closed → production). Mobile cutover **độc lập** với backend.

## 3. Post-launch monitoring window

| Cửa sổ | Theo dõi |
|---|---|
| **0–2h** | error rate (Sentry), 5xx (`edu_http_requests_total`), latency, healthcheck, DB conns |
| **2–24h** | parse job failure, push delivery, Kafka lag, đăng nhập/đăng ký thành công |
| **1–7 ngày** | crash-free %, retention, P1/P2 mới, dung lượng/scale |

Ngưỡng cảnh báo gợi ý: 5xx > 1% (5m) · p95 latency > 1s · `/health/ready` fail · crash-free < 99.5%.

## 4. Rollback runbook

**Khi nào:** P0/P1 mới sau deploy, 5xx tăng vọt, healthcheck fail, migration hỏng.

**App/worker (nhanh, ưu tiên):**
```bash
cd /opt/edu-app
export IMAGE_TAG=<previous-good-tag>
docker compose -f docker-compose.yml -f deploy/docker-compose.prod.yml up -d --no-deps app worker
bash deploy/scripts/smoke-test.sh
```

**Database:**
- Migration mới chỉ **thêm** (backward-compatible) → thường không cần rollback DB; chỉ rollback app.
- Nếu migration phá vỡ tương thích → `make migrate-down` (đã có `.down.sql` cho 15 migration) hoặc khôi phục từ snapshot RDS chụp ở §1.
- **Quy tắc:** ưu tiên expand-then-contract để tránh rollback DB.

**Mobile:** Play Console — halt rollout / rollback sang bản trước (staged rollout giúp giảm bán kính ảnh hưởng).

## 5. Incident ownership & escalation

| Vai trò | Trách nhiệm | Liên hệ |
|---|---|---|
| Incident owner (on-call) | Điều phối, quyết định rollback | (điền) |
| Backend/infra | server, DB, AWS | (điền) |
| Mobile | client, Play Console | (điền) |
| Product owner | quyết định go/no-go, comms | (điền) |

**Escalation:** on-call → tech lead (15p không khắc phục) → product owner (ảnh hưởng người dùng > 30p).

## 6. Dry-run trước launch (exit criteria của I)

- [ ] **Deploy rehearsal** trên staging: `deploy.sh` chạy hết, smoke pass
- [ ] **Rollback drill** trên staging: deploy tag mới rồi rollback về tag cũ, smoke pass
- [ ] **Restore test:** khôi phục RDS snapshot sang instance tạm, verify dữ liệu
- [ ] Toàn bộ owner xác nhận đã đọc runbook

## 7. Wave plan (giảm rủi ro)

1. **Wave 1:** backend + worker + admin + infra → internal UAT
2. **Wave 2:** Android closed beta (Play internal/closed testing)
3. **Wave 3:** Google Play production (staged rollout 10% → 50% → 100%)
