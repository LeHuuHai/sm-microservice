# Mail Worker

**mail-worker** là daemon chịu trách nhiệm tải tệp tin báo cáo và gửi email tự động cho quản trị viên.

## 1. Công nghệ & Triển khai

- **Ngôn ngữ:** Go (Golang) dạng Background Worker (không mở port HTTP).
- **Thư viện Mail:** Sử dụng `gomail` (SMTP) để đóng gói và gửi thư.
- **Bảo mật:** Giao tiếp gRPC nội bộ sử dụng `APIKeyBindStreamGRPCInterceptor` để xác thực quyền kéo tệp tin.

## 2. Giao tiếp (Kafka & gRPC)

- **Kafka Consumer:** Lắng nghe Kafka topic `mail`. Payload nhận được rất nhỏ (thường chỉ chứa `filename` và danh sách `receivers`). Mục đích là không chứa file nhị phân lớn trong Kafka nhằm tối ưu memory cho broker.
- **gRPC Client:** Khi nhận được sự kiện, `mail-worker` mở kết nối gRPC Client tới `InternalFileTransferService` của `monitor-service` (cấu hình qua `REPORT_REPO_ADDR`). Sử dụng stream để kéo (pull) file nhị phân báo cáo về bộ nhớ local một cách an toàn.
- **SMTP Outbound:** Đính kèm file vừa kéo về vào thư và đẩy đi qua giao thức SMTP (Gmail, SendGrid...).

## 3. Cấu trúc thư mục nội bộ
- `cmd/main.go`: Khởi tạo kết nối gRPC (Client), Kafka Consumer, GomailSender, và kích hoạt `MailWorker`.
- `internal/`:
  - `infra/`: `kafka` (MailConsumer), `mail` (GomailSender), `runtime`, `worker` (Luồng chính nhận event và gọi service).
  - `service/`: `download_service.go` (Thực thi logic gọi gRPC Stream để lấy file bytes).
