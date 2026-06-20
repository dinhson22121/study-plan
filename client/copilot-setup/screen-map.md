# Screen Map Template

> Dùng file này để chốt nhanh: **màn nào có gì**, **đi từ đâu sang đâu**, và
> **mỗi màn phục vụ mục tiêu gì**.

---

## 1. Danh sách màn hình

| ID | Tên màn hình | Mục tiêu | CTA chính | Đi tới |
|---|---|---|---|---|
| S01 | Splash | Giới thiệu app ngắn gọn | Bắt đầu | S02 |
| S02 | Login | Đăng nhập / vào app | Đăng nhập | S05 |
| S03 | Onboarding 1 | Giải thích giá trị | Tiếp tục | S04 |
| S04 | Onboarding 2 | Chọn mục tiêu học | Hoàn tất | S05 |
| S05 | Dashboard | Xem việc hôm nay | Tiếp tục học | S07 |

> Thay bảng trên bằng flow thật của dự án.

---

## 2. Bottom navigation

| Tab | Màn gốc | Mục đích |
|---|---|---|
| Trang chủ | Dashboard | tổng quan |
| Lộ trình | Roadmap | xem tiến trình học |
| Quiz | Quiz hub | luyện tập |
| Tiến độ | Progress | xem số liệu |
| Cá nhân | Profile | cài đặt / hồ sơ |

Nếu app không dùng bottom nav, ghi rõ navigation pattern khác.

---

## 3. Luồng quan trọng

### Flow 1 — Onboarding

Splash → Login → Onboarding → Dashboard

### Flow 2 — Học bài

Dashboard → Lesson → Quiz → Quiz Result → Dashboard

### Flow 3 — Xem tiến độ

Dashboard → Progress → Roadmap → Lesson

> Viết lại theo app thật của bạn.

---

## 4. Mapping màn hình → tính năng

| Màn hình | Tính năng chính | Dữ liệu cần có | Trạng thái đặc biệt |
|---|---|---|---|
| Dashboard | nhiệm vụ hôm nay, streak, tiến độ tuần | tên user, task, % tiến độ | loading, empty |
| Quiz | làm bài trắc nghiệm | câu hỏi, đáp án | đúng/sai, timeout |
| Progress | biểu đồ tiến độ | điểm số, streak, completion | no data |

---

## 5. Quy tắc review

Trước khi Copilot code:

1. Đã đủ danh sách màn hình chưa?
2. Luồng điều hướng đã rõ chưa?
3. Mỗi màn đã có CTA chính chưa?
4. Có màn nào đang thừa hoặc thiếu không?
5. Có màn nào cần empty/error/loading state không?

