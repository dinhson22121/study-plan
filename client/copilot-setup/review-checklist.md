# Prototype Review Checklist

> Dùng file này sau khi Copilot xuất `prototype.html`.

---

## 1. Strategy / flow

- [ ] Prototype bám đúng mục tiêu sản phẩm
- [ ] Không tự thêm màn hình ngoài scope
- [ ] Luồng điều hướng hợp lý từ đầu đến cuối
- [ ] CTA chính của từng màn đủ rõ
- [ ] Có thể demo được các flow quan trọng

---

## 2. Visual / brand

- [ ] Màu sắc đúng brand hoặc đúng palette mặc định đã chốt
- [ ] Typography nhất quán
- [ ] Card, button, spacing đồng nhất
- [ ] Không dùng gradient lòe loẹt / shadow quá nặng
- [ ] Nội dung không bị dồn dập hoặc thiếu khoảng thở

---

## 3. Content

- [ ] Dùng dữ liệu thật, không lorem ipsum
- [ ] Text đúng ngôn ngữ đã yêu cầu
- [ ] Tên màn, CTA, label rõ nghĩa
- [ ] Nội dung phù hợp đúng đối tượng người dùng

---

## 4. Interaction

- [ ] Các nút chính đều bấm được
- [ ] Chuyển màn hoạt động
- [ ] Animation nhẹ, không phô
- [ ] Hit target đủ lớn
- [ ] Không có dead-end screen

---

## 5. States

- [ ] Có loading state ở màn cần dữ liệu
- [ ] Có empty state ở nơi có thể rỗng
- [ ] Có error/cảnh báo ở nơi cần thiết
- [ ] Có completed/success state nếu flow có hoàn tất
- [ ] Có disabled/locked state nếu có gating

---

## 6. Handoff quality

- [ ] Copilot có liệt kê danh sách màn hình
- [ ] Copilot có mô tả screen-to-feature mapping
- [ ] Copilot có component breakdown
- [ ] Có ghi rõ giả định / phần còn mở
- [ ] Có thể dùng prototype để review với stakeholder

---

## 7. Quyết định sau review

Đánh dấu một trong các lựa chọn:

- [ ] Duyệt, không cần sửa
- [ ] Duyệt với chỉnh sửa nhỏ
- [ ] Cần sửa lớn nhưng giữ concept
- [ ] Làm lại flow từ đầu

### Ghi chú chỉnh sửa

- Màn nào cần đổi?
- Text nào cần sửa?
- Thành phần nào thừa / thiếu?
- Màu / spacing / CTA nào cần đổi?

