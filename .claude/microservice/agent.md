# Agent

**agent** là một chương trình siêu nhẹ (daemon) được cài đặt và chạy trực tiếp trên các máy chủ cần giám sát (Host-Level Daemon).

## 1. Công nghệ & Triển khai

- **Ngôn ngữ:** Go (Golang). Được compile thành một binary tĩnh (static binary) nhỏ gọn để dễ dàng phân phối và chạy trên nhiều hệ điều hành khác nhau.
- **Dependencies:** Chỉ sử dụng standard library (như `net/http`, `time`), không phụ thuộc thư viện thứ 3 phức tạp.
- **Cấu hình:** Thông qua biến môi trường (Ví dụ: `APP_SERVER_ID`, `APP_HEARTBEAT_URL`, `APP_HEARTBEAT_KEY`, `APP_CYCLE_HEARTBEAT`).

## 2. Cơ chế hoạt động

1. **Khởi tạo:** Đọc file cấu hình hoặc biến môi trường.
2. **First Heartbeat:** Ngay khi khởi động, lập tức gửi một tín hiệu heartbeat đầu tiên để báo cáo trạng thái `ONLINE` sớm nhất có thể.
3. **Vòng lặp (Ticker):** Thiết lập một `time.Ticker` chạy ngầm. Cứ mỗi chu kỳ (ví dụ: 3000ms), nó đóng gói JSON chứa `ServerID` và `Timestamp` hiện tại.
4. **HTTP POST:** Sử dụng HTTP Client (có timeout 5 giây để tránh treo thread) bắn request POST đến `heartbeat-gateway`. Đính kèm header `X-API-Key` để xác thực.

## 3. Cấu trúc thư mục nội bộ
- `cmd/main.go`: Chứa toàn bộ logic vòng lặp Ticker và hàm `sendHeartbeat` gọi HTTP POST.
- `config/`: Load cấu hình (AppConfig).
- `internal/model/`: Định nghĩa struct `Heartbeat` map với JSON payload.
