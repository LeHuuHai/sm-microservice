# Heartbeat Gateway

**heartbeat-gateway** là service đóng vai trò cửa ngõ tiếp nhận tín hiệu sống (heartbeat) từ tất cả các máy chủ trong hệ thống với tần suất cao (high-frequency ingestion).

## 1. Công nghệ & Triển khai

- **Ngôn ngữ & Framework:** Go (Golang) + Gin Web Framework.
- **Databases:** Không có (Stateless). Thiết kế này giúp gateway không bị nghẽn bởi connection pool của database khi chịu tải lớn.
- **Bảo mật:** Sử dụng API Key (truyền qua header `X-API-Key`) thay vì JWT để giảm chi phí tính toán khi xác thực hàng triệu request/giây. API key được cấu hình cố định hoặc cấp theo môi trường (`APP_HEARTBEAT_KEY`).

## 2. Giao tiếp (APIs)

- **Port hoạt động:** Cấu hình qua `APP_PORT` (thường là 8082).

### REST API
- `POST /heartbeat`: Nhận payload JSON chứa `server_id` và `timestamp` từ agent.

### Kafka Event Publishing
- Ghi trực tiếp các sự kiện heartbeat nhận được vào Kafka topic: `heartbeats`.
- *Tối ưu:* Service này sử dụng Kafka Publisher kết nối thẳng đến broker, cho phép đạt thông lượng (throughput) cực cao (thường <5ms response time) do không có bước nào block chờ DB I/O.

## 3. Cấu trúc thư mục nội bộ
- `api/`: OpenAPI generated code cho HTTP REST.
- `cmd/main.go`: Khởi tạo HTTP Server và Kafka Publisher.
- `internal/`:
  - `handler/`: Chứa `heartbeat_handler.go` nhận payload HTTP.
  - `infra/`: `kafka` (triển khai KafkaPublisher), `runtime` (App struct).
  - `middleware/`: Triển khai `APIKeyMiddleware`.
  - `service/`: `gw_service.go` đóng vai trò đẩy event sang Kafka publisher.
