# Ping Worker

**ping-worker** là một background daemon có nhiệm vụ thực hiện ping ICMP chủ động đến các server khi nhận được yêu cầu (thường là khi server đó bị miss heartbeat).

## 1. Công nghệ & Triển khai

- **Ngôn ngữ:** Go (Golang) dạng Background Worker (không mở port HTTP/gRPC).
- **Network Protocol:** Sử dụng ICMP (Raw socket) để ping. *Lưu ý: Quá trình này yêu cầu quyền root hoặc capability `CAP_NET_RAW` trên Linux environment.*
- **Concurrency:** Áp dụng mô hình Worker Pool (`workerPool`). Có thể cấu hình số lượng luồng thực thi song song qua biến môi trường `NUM_THREAD`. Cho phép scale horizontally một cách tự nhiên.

## 2. Giao tiếp (Kafka)

Dịch vụ hoàn toàn hướng sự kiện (Event-driven):
- **Đăng ký nhận (Consume):** Lắng nghe Kafka topic `ping`. Khi `monitor-service` phát hiện một server bị timeout, nó sẽ gửi một sự kiện `PingRequested` vào đây.
- **Phát hành (Publish):** Sau khi ping xong, kết quả (thành công hay thất bại, độ trễ) sẽ được đóng gói thành sự kiện `PingResult` và đẩy vào Kafka topic `ping_res`. `monitor-service` sẽ nghe topic này để cập nhật trạng thái.

## 3. Cấu trúc thư mục nội bộ
- `cmd/main.go`: Khởi tạo Kafka Consumer, Publisher, và kích hoạt `PingWorkerPool`.
- `internal/`:
  - `infra/`: `kafka` (Consumer/Publisher), `runtime` (App config), `service` (Implement logic ping thực tế), `worker` (Quản lý pool các goroutines).
