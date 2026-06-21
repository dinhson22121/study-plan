# Phase 1 — Scope Freeze & Release Criteria (Workstream A)

> Chốt phạm vi Phase 1 và quality bar. Mọi thay đổi sau khi freeze phải qua
> change-control (xem §6). Tài liệu này mở rộng `plan/PHASE1_GOLIVE_PLAN.md`.

- **Trạng thái:** ĐỀ XUẤT — chờ product owner sign-off
- **Ngày freeze (dự kiến):** điền khi sign-off
- **Phiên bản:** 1.0

---

## 1. In-scope (Phase 1)

| Năng lực | Thành phần | Trạng thái build |
|---|---|---|
| Auth (đăng ký/đăng nhập/refresh/logout, revoke) | server `auth`, admin, client | ✅ |
| Onboarding (chọn mục tiêu + thời lượng) | client, server `goal` | ✅ |
| Placement / chẩn đoán đầu vào | server `placement`, client | ✅ (luồng submit câu hỏi: planned) |
| Study plan (sinh lộ trình + milestone) | server `studyplan`, client | ✅ |
| Lesson / topic | server `content`/`curriculum`, client | ✅ |
| Quiz (làm + chấm + review) | server `quiz`, client | ✅ |
| Progress & analytics (streak, weak topics) | server `progress`/`analytics`, client | ✅ |
| Notification (FCM push, preferences, history) | server `notification`, worker, client | ✅ |
| Admin: upload → parse PDF → review draft → publish | server `content`/`question`, worker, admin | ✅ |

## 2. Out-of-scope (Phase 1)

- OCR / parse PDF scan ảnh (chỉ hỗ trợ PDF có text)
- iOS release (Android-first; iOS để Phase 2)
- Advanced analytics / báo cáo chuyên sâu
- Multi-region / HA nhiều vùng
- Realtime collaboration, social/leaderboard
- Thanh toán / gói trả phí

## 3. Acceptance criteria (theo workstream exit-criteria)

- **Backend:** `go test ./...` pass; integration/e2e pass trên infra; deferred security đã đóng; prod config validate. ✅ (đơn vị) / ⏳ (integration cần infra)
- **Admin:** `npm run build` pass; Vitest + Playwright smoke pass; UAT pass. ✅ build/test / ⏳ UAT
- **Mobile:** release build (.aab) pass; QA thiết bị thật pass; crash-free beta ≥ 99.5%. ⏳ (cần EAS build + thiết bị)
- **Worker:** chạy ổn nhiều vòng; retry/reparse pass; alert hoạt động. ⏳ (cần infra)
- **Infra:** domain + HTTPS live; healthcheck pass trên domain thật; backup+restore test ≥ 1 lần. ⏳ (cần AWS)

## 4. Severity matrix

| Mức | Định nghĩa | SLA fix | Chặn release? |
|---|---|---|---|
| **P0 / Critical** | Mất dữ liệu, lỗ hổng bảo mật, app không khởi động, crash loop, không đăng nhập được | Ngay lập tức | CÓ |
| **P1 / High** | Luồng chính hỏng (quiz/plan/upload), không workaround | < 24h | CÓ |
| **P2 / Medium** | Lỗi có workaround, ảnh hưởng phụ | < 1 tuần | KHÔNG |
| **P3 / Low** | Cosmetic, nội dung, tối ưu nhỏ | Backlog | KHÔNG |

**Điều kiện go-live:** P0 = 0 và P1 = 0.

## 5. Go/No-Go gates

| Gate | Điều kiện |
|---|---|
| **Gate 1 — Backend prod-ready** | tests pass · deferred security đóng · deploy rehearsal pass |
| **Gate 2 — Admin prod-ready** | build/test pass · admin UAT pass |
| **Gate 3 — Mobile beta-ready** | Android release build pass · QA thiết bị thật pass · Play internal testing pass |
| **Gate 4 — Launch-ready** | infra ổn định · monitoring live · rollback sẵn sàng · 0 P0/P1 |

## 6. Change control sau freeze

- Mọi đề xuất thêm/đổi scope → ghi vào `docs/SCOPE_CHANGE_LOG.md`, đánh giá tác động, owner duyệt.
- Bug P0/P1 mới phát hiện vẫn được fix trong scope freeze (không tính là thay đổi scope).
- Tính năng mới → đẩy sang Phase 2 trừ khi owner duyệt khẩn.

## 7. Pilot / beta group

- **Wave 1 (internal UAT):** team nội bộ + 3–5 giáo viên reviewer.
- **Wave 2 (closed beta):** ~20–50 học sinh THPT qua Google Play internal/closed testing.
- **Wave 3 (production):** mở public sau khi đạt crash-free ≥ 99.5% và 0 P0/P1.

## 8. Sign-off

| Vai trò | Tên | Ngày | Trạng thái |
|---|---|---|---|
| Product Owner | | | ☐ |
| Tech Lead | | | ☐ |
| QA Lead | | | ☐ |
