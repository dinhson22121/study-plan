# Phase 1 — QA Report (Workstream H)

> Đảm bảo launch không vỡ vì blocker bug/crash. Cập nhật mỗi vòng QA/beta.

- **Ngày chạy:** (cập nhật mỗi lần)
- **Môi trường máy build hiện tại:** Windows, **không có Docker / Python interpreter / Android emulator** → các test cần infra/thiết bị được đánh dấu ⏳ và phải chạy trong CI/AWS/Play.

---

## 1. Tổng quan kết quả

| Hạng mục | Công cụ | Kết quả |
|---|---|---|
| Backend unit tests | `go test ./...` | ✅ **216 test func, all pass** |
| Backend coverage (business logic) | `go test -cover` | ✅ domain 77–97%, application 70–93%, shared 68–100% |
| Backend integration/e2e | `go test -tags=integration` | ⏳ **14 file**, cần Postgres/Redis/Kafka (chạy CI/infra) |
| Backend security scan | `govulncheck` | ✅ No vulnerabilities |
| Admin build | `npm run build` (tsc+vite) | ✅ pass |
| Admin unit (Vitest) | `npx vitest run` | ✅ **19 test, 5 file** |
| Admin e2e smoke (Playwright) | `npx playwright test` | ✅ **7 test** (backend mock qua route interception) |
| Admin dependency audit | `npm audit` | ✅ 0 vuln |
| Mobile typecheck | `tsc --noEmit` | ✅ clean (toàn bộ 13 màn) |
| Mobile device QA | thủ công / Firebase Test Lab | ⏳ cần .aab + thiết bị |
| Worker compile/lint | `py_compile` / `ruff` | ⏳ cần Python (chạy trong Docker/CI) |

## 2. Backend QA (H1)

- **Unit:** 216 test func pass; bao phủ tốt domain + application (logic nghiệp vụ).
- **Coverage chú thích:** các package báo 0% là `interfaces/http`, `infrastructure`,
  module-wiring — chúng được phủ bởi **14 file integration/e2e** (build tag `integration`)
  cần DB/Redis/Kafka. Trên CI/infra: chạy `make test-integration`.
- **Migration rollback:** mỗi migration có `.down.sql` (15 cặp). Test rollback:
  `make migrate-down` rồi `make migrate-up` trên DB staging.
- **Smoke sau deploy:** `deploy/scripts/smoke-test.sh` (`/health`, `/health/ready`,
  `/metrics`, optional register+login).
- **Còn thiếu để đạt 80% tổng:** integration tests cho interfaces/http + infrastructure
  phải chạy (không tính trong coverage unit vì cần infra).

## 3. Admin QA (H2)

- **Auth/session:** Vitest phủ token store, JWT helper, api client (envelope + **single-flight 401 refresh**), AuthContext.
- **Flow upload/parse/review/publish:** Playwright smoke (login → upload init→complete → draft review → publish) với backend mock.
- **Browser sanity:** chạy Playwright trên Chromium; cần bổ sung Firefox/WebKit cho cross-browser đầy đủ.

## 4. Mobile QA (H3) — ⏳ cần build + thiết bị

Checklist phải chạy khi có `.aab` (EAS) + thiết bị:
- [ ] fresh install · login · onboarding · placement · sinh study plan · quiz · progress
- [ ] mở push notification · logout/login lại
- [ ] Android 10/11/12/13/14 · ≥1 thiết bị RAM thấp
- [ ] mạng yếu / mất mạng / reconnect
- [ ] background/foreground; session restore (SecureStore)

## 5. Crash gate (H4)

| Chỉ tiêu | Ngưỡng | Trạng thái |
|---|---|---|
| Crash-free sessions (beta) | ≥ 99.5% | ⏳ đo qua Sentry/Play sau beta |
| ANR blocker | 0 | ⏳ |
| Startup crash | 0 | ⏳ |

## 6. Blocker list

| ID | Mô tả | Mức | Trạng thái |
|---|---|---|---|
| — | (chưa có P0/P1 đã biết ở phần chạy được) | — | — |

## 7. Exit criteria

- P0 = 0, P1 = 0 (xem `docs/PHASE1_SCOPE_FREEZE.md` §4)
- crash-free beta đạt ngưỡng
- integration/e2e + mobile device QA pass trên môi trường thật

## 8. Việc chạy được ngay (không cần hạ tầng) — ĐÃ PASS

`go test ./...` · `go vet ./...` · `govulncheck` · admin `npm run build` + Vitest + Playwright + `npm audit` · client `tsc --noEmit`.
