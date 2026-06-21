# AI Study Coach — Product Spec cho Prototype

## 1. Tên sản phẩm

- **Tên app:** AI Study Coach
- **Mô tả ngắn 1 câu:** App học tập cá nhân hóa giúp học sinh THPT có lộ trình rõ ràng, làm quiz ngắn mỗi ngày và theo dõi tiến độ thật sự.
- **Mục tiêu chính của prototype:** Demo một trải nghiệm học tập trọn vẹn từ đăng nhập → onboarding → chẩn đoán đầu vào → dashboard → lộ trình → bài học → quiz → kết quả → tiến độ.

---

## 2. Đối tượng người dùng

- **Nhóm người dùng chính:** Học sinh THPT
- **Độ tuổi / trình độ:** 15–18 tuổi
- **Bối cảnh sử dụng:** Học sau giờ học chính khóa, ôn thi giữa kỳ/cuối kỳ và ôn thi đại học
- **Nỗi đau chính:** Không biết nên học gì trước, thiếu động lực duy trì đều đặn, khó nhìn thấy tiến bộ
- **Động lực / lợi ích mong muốn:** Có kế hoạch học rõ, cảm thấy mình đang tiến bộ từng ngày, có bài quiz ngắn để kiểm tra ngay

---

## 3. Brand / phong cách

- **Ngôn ngữ UI:** Tiếng Việt
- **Phong cách visual:** Sạch, sáng, hiện đại, trẻ trung, premium vừa phải
- **Mức gamification:** Vừa
- **Brand riêng:** Chưa có, dùng palette mặc định của AI Study Coach

---

## 4. Danh sách màn hình cần có

1. Splash
2. Login
3. Onboarding — Chọn mục tiêu học
4. Chẩn đoán đầu vào
5. Kết quả chẩn đoán
6. Dashboard
7. Lộ trình học
8. Bài học
9. Quiz
10. Kết quả quiz
11. Tiến độ
12. Cá nhân
13. Thông báo

---

## 5. Luồng điều hướng

- Splash → Login
- Login thành công → Onboarding
- Onboarding → Chẩn đoán đầu vào
- Chẩn đoán đầu vào → Kết quả chẩn đoán
- Kết quả chẩn đoán → Dashboard
- Dashboard:
  - “Tiếp tục bài học” → Bài học
  - “Xem lộ trình” → Lộ trình học
  - “Làm quiz nhanh” → Quiz
  - icon chuông → Thông báo
- Bài học → Quiz
- Quiz nộp bài → Kết quả quiz
- Kết quả quiz → Dashboard hoặc Lộ trình
- Bottom nav:
  - Trang chủ → Dashboard
  - Lộ trình → Lộ trình học
  - Quiz → Quiz
  - Tiến độ → Tiến độ
  - Cá nhân → Cá nhân

---

## 6. Nội dung từng màn hình

### Splash
- Hero logo AI Study Coach
- Tagline: “Học đúng trọng tâm, tiến bộ mỗi ngày”

### Login
- Email, mật khẩu
- CTA: Đăng nhập

### Onboarding
- Chọn mục tiêu: tăng điểm Toán, giữ streak học đều, ôn thi tốt nghiệp
- Chọn thời lượng học mỗi ngày: 20 / 40 / 60 phút

### Chẩn đoán đầu vào
- 3 môn hiển thị: Toán, Tiếng Anh, Vật Lý
- CTA: Bắt đầu bài chẩn đoán

### Kết quả chẩn đoán
- Toán: Cần củng cố Logarit
- Tiếng Anh: Tốt phần đọc hiểu, yếu mệnh đề quan hệ
- Vật Lý: Cần luyện điện xoay chiều

### Dashboard
- Chào “Minh Anh”
- Chuỗi học: 12 ngày
- Tiến độ tuần: 68%
- Nhiệm vụ hôm nay:
  - Logarit cơ bản — 15 phút
  - Quiz tiếng Anh — 5 câu
  - Ôn lại điện xoay chiều — 10 phút

### Lộ trình học
- 3 milestone:
  - Củng cố nền tảng
  - Luyện đề theo chuyên đề
  - Tăng tốc trước kỳ thi

### Bài học
- Chủ đề: Logarit cơ bản
- 3 ý chính
- ví dụ minh họa
- CTA: Làm quiz kiểm tra

### Quiz
- 3 câu trắc nghiệm thật, ngữ cảnh THPT

### Kết quả quiz
- Điểm: 2/3
- Câu sai có giải thích ngắn
- CTA: Học lại bài / Xem lộ trình

### Tiến độ
- Streak
- Tổng giờ học tuần
- Tỷ lệ hoàn thành nhiệm vụ
- Môn tiến bộ nhanh nhất

### Cá nhân
- Mục tiêu học
- Thời gian nhắc học
- Bật/tắt thông báo

### Thông báo
- Nhắc học tối nay 19:30
- Hoàn thành milestone Logarit
- Quiz tuần mới đã sẵn sàng

---

## 7. Dữ liệu mẫu / nội dung thật

- **Tên người dùng:** Minh Anh
- **Môn học:** Toán, Tiếng Anh, Vật Lý
- **Chủ đề:** Logarit, mệnh đề quan hệ, điện xoay chiều
- **Tên bài học:** Logarit cơ bản — đổi cơ số và ứng dụng
- **Điểm số:** 2/3, 78%, 8.1/10
- **Tiến độ:** 68% tuần này
- **Thông báo:** Nhắc học, mở quiz mới, hoàn thành milestone

### Câu hỏi quiz mẫu

1. Nếu `log2(8) = ?`
   - A. 2
   - B. 3
   - C. 4
   - D. 8
   - Đáp án: B

2. Chọn câu đúng về mệnh đề quan hệ:
   - A. who dùng cho vật
   - B. which dùng cho người
   - C. who dùng cho người
   - D. where dùng cho đại từ nhân xưng
   - Đáp án: C

3. Đại lượng nào thay đổi tuần hoàn theo thời gian trong dòng điện xoay chiều?
   - A. Chỉ cường độ dòng điện
   - B. Chỉ điện áp
   - C. Cả điện áp và cường độ dòng điện
   - D. Không đại lượng nào
   - Đáp án: C

---

## 8. Trạng thái cần thể hiện

- [x] loading
- [x] empty state
- [x] error state
- [x] completed state
- [x] locked / disabled state
- [x] notification badge

---

## 9. Component / pattern đặc biệt

- progress ring
- streak card
- weekly progress chart
- roadmap milestone card
- filter chips
- quick action buttons

---

## 10. Constraint kỹ thuật cho prototype

- **Xuất 1 file HTML duy nhất:** có
- **Tương tác thật bằng JS:** có
- **Có animation:** nhẹ
- **Khung điện thoại:** 390×844
- **Có bottom nav:** có

---

## 11. Acceptance criteria

1. Prototype có tối thiểu 10 màn.
2. Các CTA chính đều chuyển màn thật.
3. Có luồng hoàn chỉnh từ login đến quiz result.
4. Dashboard có chuỗi học, nhiệm vụ hôm nay và tiến độ tuần.
5. Quiz có ít nhất 3 câu thật và có chấm điểm.
6. Không dùng lorem ipsum hay placeholder chung chung.
7. Có ít nhất loading/error/empty/completed state ở các màn phù hợp.

---

## 12. Ghi chú thêm

- Giữ đúng palette AI Study Coach.
- Không thêm tính năng ngoài scope trên.
- Sau khi code xong phải có component breakdown và screen-to-feature mapping.
