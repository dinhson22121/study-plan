# Pipeline Prompt — AI Study Coach (Claude Design)

Mobile app học tập cá nhân hóa cho học sinh THPT lớp 12 ôn thi THPT Quốc Gia.

**Cách dùng:** Chạy theo thứ tự. Chạy **Bước 0** trước (thiết lập design system chung), sau đó paste tuần tự từng màn hình vào Claude Design.

---

## BƯỚC 0 — Khởi tạo Design System (chạy đầu tiên)

```
Thiết lập design system cho mobile app "AI Study Coach" — app học tập cá nhân hóa cho học sinh THPT lớp 12 ôn thi THPT Quốc Gia. Áp dụng cho TẤT CẢ màn hình sau.

Khung: mobile 390x844. Phong cách: clean, minimal, student-friendly, trẻ trung, gamification nhẹ. KHÔNG dark academic.
Màu: Primary = Blue, Success = Green, Warning = Orange, Error = Red. Nền sáng, bo góc mềm, khoảng trắng thoáng, typography dễ đọc trên màn nhỏ.
Bottom Tabs: Home / Study Plan / Quiz / Progress / Profile.
Nội dung UI bằng tiếng Việt.

Hãy tạo style guide: bảng màu, typography scale, component cơ bản (button, card, input, progress ring/bar, badge, topic status chip). Giữ nguyên hệ này cho các màn tiếp theo.
```

---

## MÀN 1 — Splash

```
Theo design system AI Study Coach, thiết kế màn Splash.
- Logo app ở giữa
- Trạng thái "Loading session..."
- Tối giản, animation-ready, brand màu Blue.
```

---

## MÀN 2 — Login

```
Theo design system AI Study Coach, thiết kế màn Login.
- Nút "Đăng nhập với Google" (primary)
- Dòng "Điều khoản & Điều kiện" (Terms & Conditions) ở dưới
- Tối giản, 1 hành động chính rõ ràng.
```

---

## MÀN 3 — Onboarding (5 bước)

```
Theo design system AI Study Coach, thiết kế Onboarding dạng 5 bước có progress indicator ở trên (step 1/5...).
Bước 1 - Thông tin học sinh: Họ tên, Trường, Tỉnh/Thành, Lớp
Bước 2 - Mục tiêu: Trường ĐH mục tiêu, Ngành mục tiêu
Bước 3 - Môn học: chọn nhiều (Toán, Tiếng Anh, Lý)
Bước 4 - Điểm số: Điểm hiện tại, Điểm mục tiêu (mỗi môn)
Bước 5 - Thời gian học: Giờ/ngày, Ngày/tuần
Mỗi bước có nút "Tiếp tục". Form gọn, dễ điền trên mobile.
```

---

## MÀN 4 — Placement Test

```
Theo design system AI Study Coach, thiết kế luồng Placement Test:
A) Màn chọn môn để làm bài
B) Màn làm bài: hiển thị câu hỏi + đáp án trắc nghiệm, thanh tiến độ câu (vd 5/20), timer (15-20 phút), nút Tiếp/Submit
C) Màn kết quả: "Tính toán trình độ" → hiển thị level được xếp.
20 câu/môn. Giao diện tập trung, ít sao nhãng.
```

---

## MÀN 5 — Dashboard (điểm nhấn chính)

```
Theo design system AI Study Coach, thiết kế Dashboard — màn quan trọng nhất, tối ưu retention và one-click vào bài học hôm nay.
Sections (theo thứ tự ưu tiên):
1. "Hôm nay cần học" — checklist nhiệm vụ (vd: ✓ Logarit, ✓ 10 bài tập, ✓ Quiz nhanh) + nút lớn "Học ngay"
2. Current Streak — số ngày liên tục, icon lửa
3. "Tiến độ tuần" — progress ring/bar hiển thị 65%
4. Upcoming Quiz — card quiz sắp tới
5. Motivation Card — câu động viên
Header chào học sinh. Layout cực kỳ rõ ràng, dồn sự chú ý vào "Học ngay".
```

---

## MÀN 6 — Study Plan

```
Theo design system AI Study Coach, thiết kế màn Study Plan dạng cây lộ trình theo tuần.
Cấu trúc: Week 1 → Môn (Toán, Anh) → các Topic con (vd Toán: Khái niệm Log, Công thức Log; Anh: Vocabulary, Reading).
Mỗi Topic có status phân biệt rõ bằng màu + icon:
- Locked (xám/khóa)
- Available (xanh primary)
- In Progress (cam)
- Completed (xanh success/tick)
Trực quan, dễ hiểu, cảm giác tiến trình rõ ràng.
```

---

## MÀN 7 — Topic Learning

```
Theo design system AI Study Coach, thiết kế màn Topic Learning.
Layout:
- Topic Header (tên topic, môn, trạng thái)
- Learning Materials: tab/danh sách PDF, Slide, Notes
- Practice Questions
- Nút "Làm Quiz" nổi bật ở cuối
Bố cục cuộn dọc, ưu tiên đọc tài liệu rồi chuyển sang Quiz.
```

---

## MÀN 8 — Quiz

```
Theo design system AI Study Coach, thiết kế luồng Quiz:
A) Màn câu hỏi: Question → Options (trắc nghiệm) → nút Next → Submit, có thanh tiến độ câu
B) Màn Result: Score lớn, số câu Đúng / Sai, danh sách Explanation cho từng câu.
Phản hồi đúng/sai dùng màu Success/Error rõ ràng.
```

---

## MÀN 9 — Progress

```
Theo design system AI Study Coach, thiết kế màn Progress với metrics dạng chart/visual:
- Topic Completion
- Subject Completion
- Quiz Average
- Study Hours
Dùng progress ring, bar chart, số liệu lớn dễ đọc. Cảm giác thành tựu, khích lệ tiếp tục.
```

---

## MÀN 10 — Profile

```
Theo design system AI Study Coach, thiết kế màn Profile.
- Avatar + tên học sinh
- Study Goal (trường/ngành mục tiêu)
- Notification Settings (toggle)
- Quiet Hours (giờ im lặng)
- Nút Logout (màu Error, dưới cùng)
Gọn gàng, dạng danh sách settings.
```

---

## (Tùy chọn) BƯỚC CUỐI — Notification UX

```
Theo design system AI Study Coach, thiết kế 3 mẫu push notification / in-app banner:
1. Daily Reminder: "Bạn còn 2 nhiệm vụ hôm nay."
2. Weekly Reminder: "Bạn chưa hoàn thành Quiz tuần này."
3. Achievement: "Chúc mừng! Bạn đã hoàn thành Logarit."
Phong cách thân thiện, dùng màu theo loại (reminder = primary/warning, achievement = success).
```
