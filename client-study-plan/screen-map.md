# AI Study Coach — Screen Map

## 1. Danh sách màn hình

| ID | Tên màn hình | Mục tiêu | CTA chính | Đi tới |
|---|---|---|---|---|
| S01 | Splash | Giới thiệu app | Bắt đầu | S02 |
| S02 | Login | Đăng nhập | Đăng nhập | S03 |
| S03 | Onboarding | Chọn mục tiêu học | Tiếp tục | S04 |
| S04 | Chẩn đoán đầu vào | Bắt đầu lộ trình cá nhân hóa | Bắt đầu chẩn đoán | S05 |
| S05 | Kết quả chẩn đoán | Tạo niềm tin và gợi ý lộ trình | Xem dashboard | S06 |
| S06 | Dashboard | Xem việc hôm nay và trạng thái học tập | Tiếp tục bài học | S08 |
| S07 | Lộ trình học | Nhìn milestone và bước tiếp theo | Mở bài học | S08 |
| S08 | Bài học | Học nhanh một chủ đề | Làm quiz kiểm tra | S09 |
| S09 | Quiz | Kiểm tra nhanh sau bài học | Nộp bài | S10 |
| S10 | Kết quả quiz | Xem điểm và lỗi sai | Về dashboard | S06 |
| S11 | Tiến độ | Xem số liệu học tập | Xem lộ trình | S07 |
| S12 | Cá nhân | Quản lý cài đặt và mục tiêu | Lưu cài đặt | S12 |
| S13 | Thông báo | Xem nhắc học và cập nhật mới | Mở dashboard | S06 |

---

## 2. Bottom navigation

| Tab | Màn gốc | Mục đích |
|---|---|---|
| Trang chủ | Dashboard | việc cần làm hôm nay |
| Lộ trình | Lộ trình học | xem milestone |
| Quiz | Quiz | luyện nhanh |
| Tiến độ | Tiến độ | xem tiến bộ |
| Cá nhân | Cá nhân | cài đặt |

---

## 3. Luồng quan trọng

### Flow 1 — Khởi động
Splash → Login → Onboarding → Chẩn đoán đầu vào → Kết quả chẩn đoán → Dashboard

### Flow 2 — Học và kiểm tra
Dashboard → Bài học → Quiz → Kết quả quiz → Dashboard

### Flow 3 — Xem chiến lược học
Dashboard → Lộ trình học → Bài học

### Flow 4 — Theo dõi tiến bộ
Dashboard → Tiến độ → Lộ trình học

### Flow 5 — Nhắc học
Dashboard → Thông báo → Dashboard

---

## 4. Mapping màn hình → tính năng

| Màn hình | Tính năng chính | Dữ liệu cần có | Trạng thái đặc biệt |
|---|---|---|---|
| Splash | branding + intro | tên app, tagline | animation |
| Login | auth giả lập | email, password | error |
| Onboarding | chọn mục tiêu học | mục tiêu, thời lượng học | selected state |
| Chẩn đoán | giới thiệu chẩn đoán | môn học, mô tả | loading |
| Kết quả chẩn đoán | insight cá nhân hóa | mức độ từng môn | completed |
| Dashboard | task/streak/progress | user, streak, tasks | loading, empty |
| Lộ trình | milestone cards | topics, completion | locked |
| Bài học | lesson summary | title, bullets, examples | progress state |
| Quiz | câu hỏi + đáp án | 3 câu, 4 đáp án | selected state |
| Kết quả quiz | điểm + giải thích | số câu đúng, feedback | success / warning |
| Tiến độ | chart + stats | streak, giờ học, completion | empty |
| Cá nhân | settings | reminder, goal | toggle |
| Thông báo | alerts | reminder, milestone, new quiz | unread badge |

