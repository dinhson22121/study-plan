# Hướng dẫn cài đặt — GitHub Copilot Mobile App Builder

Bộ file này giúp GitHub Copilot tạo prototype mobile app theo đúng phong cách
như prototype "AI Study Coach".

## Cấu trúc

```
.github/
├── copilot-instructions.md      ← luật chung, Copilot đọc tự động
└── prompts/
    └── mobile-app.prompt.md     ← prompt gọi nhanh bằng /mobile-app
```

## Cách cài

1. **Chép thư mục `.github/`** (trong gói này) vào **gốc repo** của bạn.
   Repo của bạn sẽ có `your-project/.github/copilot-instructions.md`.
2. Mở repo trong **VS Code** (có cài extension GitHub Copilot + Copilot Chat).
3. Bật **Agent mode**: mở Copilot Chat → ở ô chọn chế độ, chọn **Agent**
   (không phải "Ask"). Agent mode mới tự tạo/sửa nhiều file được.

## Cách dùng

**Cách 1 — dùng prompt có sẵn:**
Trong Copilot Chat gõ:
```
/mobile-app
```
rồi dán spec (hoặc tên file spec) khi được hỏi.

**Cách 2 — chat tự do:**
Bỏ file spec (vd `app_design.md`) vào repo, rồi gõ:
```
Đọc app_design.md và tạo prototype mobile app theo copilot-instructions.md.
Hỏi tôi làm rõ trước nếu thiếu thông tin.
```

## Mẹo để ra kết quả tốt

- **Spec càng chi tiết càng tốt:** liệt kê từng màn hình + luồng đi + nội dung mẫu.
- **Trả lời dứt khoát** các câu Copilot hỏi (đừng để "tùy bạn").
- **Lặp từng phần:** xem prototype rồi yêu cầu chỉnh nhỏ ("đổi màu chuỗi học",
  "thêm màn Thông báo"), thay vì làm lại từ đầu.
- Nếu có **design system / brand**, đưa file CSS tokens vào repo và bảo Copilot
  dùng nó thay cho bộ màu mặc định.

## Lưu ý

Copilot mạnh khi code trong IDE, nhưng phần *hỏi làm rõ trước* và *xuất một file
prototype hoàn chỉnh* cần bạn nhắc kỹ và lặp vài lần. Đó là điều bình thường.
