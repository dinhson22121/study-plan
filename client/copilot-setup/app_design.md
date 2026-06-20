# App Design Template

> Dùng file này làm **spec gốc** khi yêu cầu Copilot tạo prototype mobile app.
> Càng điền cụ thể, prototype càng sát nhu cầu thật.

---

## 1. Tên sản phẩm

- **Tên app:**
- **Mô tả ngắn 1 câu:**
- **Mục tiêu chính của prototype:**

Ví dụ:

- Tên app: AI Study Coach
- Mô tả ngắn: App học tập cá nhân hóa cho học sinh THPT
- Mục tiêu chính: Demo onboarding + dashboard + lộ trình + quiz để pitch nội bộ

---

## 2. Đối tượng người dùng

- **Nhóm người dùng chính:**
- **Độ tuổi / trình độ:**
- **Bối cảnh sử dụng:**
- **Nỗi đau chính:**
- **Động lực / lợi ích mong muốn:**

Ví dụ:

- Học sinh lớp 10–12
- Dùng app sau giờ học hoặc trước kỳ thi
- Muốn có lộ trình học rõ ràng, nhắc học, quiz ngắn và xem tiến độ

---

## 3. Brand / phong cách

- **Ngôn ngữ UI:** (vd: Tiếng Việt)
- **Phong cách visual:** (vd: trẻ trung, sạch, premium, nghiêm túc, enterprise)
- **Mức gamification:** thấp / vừa / cao
- **Có dùng brand riêng không:** có / không
- Nếu có, ghi rõ:
  - màu chủ đạo:
  - màu phụ:
  - font:
  - logo / icon style:

Nếu có file `brand-tokens.css`, nói rõ Copilot phải ưu tiên token trong file đó.

---

## 4. Danh sách màn hình cần có

> Liệt kê rõ từng màn hình, không ghi chung chung.

Ví dụ format:

1. Splash
2. Login
3. Onboarding bước 1
4. Onboarding bước 2
5. Dashboard
6. Lộ trình học
7. Bài học
8. Quiz
9. Kết quả quiz
10. Tiến độ
11. Cá nhân
12. Thông báo

---

## 5. Luồng điều hướng

> Viết theo kiểu “màn nào bấm gì sẽ đi đâu”.

Ví dụ:

- Splash → Login
- Login thành công → Dashboard
- Dashboard bấm “Tiếp tục học” → Bài học hiện tại
- Dashboard bấm “Làm quiz” → Quiz
- Quiz nộp bài → Kết quả quiz
- Bottom nav:
  - Trang chủ → Dashboard
  - Lộ trình → Lộ trình học
  - Quiz → Quiz hub
  - Tiến độ → Progress
  - Cá nhân → Profile

---

## 6. Nội dung từng màn hình

> Với mỗi màn, ghi rõ:
> - mục tiêu màn
> - thành phần chính
> - CTA
> - dữ liệu mẫu

### Ví dụ format

### Màn: Dashboard

- **Mục tiêu:** cho học sinh thấy việc cần làm hôm nay
- **Thành phần chính:**
  - hero chào người dùng
  - chuỗi học
  - nhiệm vụ hôm nay
  - tiến độ tuần
  - quick actions
- **CTA chính:** “Tiếp tục bài học”
- **Dữ liệu mẫu:**
  - tên: Minh Anh
  - chuỗi học: 12 ngày
  - môn ưu tiên: Toán / Logarit

Lặp lại format này cho từng màn.

---

## 7. Dữ liệu mẫu / nội dung thật

> Không dùng placeholder. Hãy ghi dữ liệu thật để Copilot dùng.

- tên người dùng:
- môn học:
- chủ đề:
- tên bài học:
- điểm số:
- tiến độ:
- thông báo:
- câu hỏi quiz mẫu:
- đáp án mẫu:

---

## 8. Trạng thái cần thể hiện

> Copilot thường bỏ qua phần này nếu bạn không ghi rõ.

Đánh dấu các trạng thái nào phải có:

- [ ] loading
- [ ] empty state
- [ ] error state
- [ ] completed state
- [ ] locked / disabled state
- [ ] notification badge
- [ ] offline / no connection

Nếu có, mô tả cụ thể từng trạng thái.

---

## 9. Component / pattern đặc biệt

Ghi rõ nếu app cần:

- progress ring
- streak card
- chart
- tree roadmap
- tab switcher
- bottom sheet
- filter chips
- flashcard
- leaderboard
- AI chat widget

---

## 10. Constraint kỹ thuật cho prototype

- **Xuất 1 file HTML duy nhất:** có / không
- **Tương tác thật bằng JS:** có / không
- **Có animation:** nhẹ / vừa / không
- **Khung điện thoại:** 390x844 / custom
- **Có bottom nav:** có / không

---

## 11. Acceptance criteria

> Đây là phần cực quan trọng để review output.

Ví dụ:

1. Prototype phải có tối thiểu 8 màn.
2. Bấm CTA phải chuyển màn thật.
3. Dashboard phải có chuỗi học, nhiệm vụ hôm nay và tiến độ tuần.
4. Quiz phải có ít nhất 3 câu mẫu và màn kết quả.
5. Không dùng lorem ipsum hoặc text giả.

---

## 12. Ghi chú thêm cho Copilot

Ví dụ:

- Hỏi tôi làm rõ trước nếu thiếu màn hoặc thiếu luồng.
- Không thêm màn mới nếu tôi chưa duyệt.
- Sau khi code xong, liệt kê component breakdown và screen-to-feature mapping.

