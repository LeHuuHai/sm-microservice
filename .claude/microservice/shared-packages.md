# Thư viện Dùng chung (Shared Packages - `/pkg`)

Thư mục `microservices/pkg/` chứa các thư viện dùng chung cho toàn bộ các microservices. Việc tập trung hóa các thư viện này giúp duy trì sự nhất quán về mặt công nghệ, giảm thiểu trùng lặp code và chuẩn hóa các giao thức kết nối.

---

## 1. Danh sách các Shared Packages

| Thư mục | Chức năng | Các thành phần chính |
|---|---|---|
| [`apperr`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/apperr) | Định nghĩa lỗi nghiệp vụ | Danh sách mã lỗi chuẩn hóa, chuyển đổi từ Domain Error sang HTTP status codes / gRPC status codes. |
| [`auth`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/auth) | Xác thực & Phân quyền | `RoleCheckMiddleware` (cho Gin HTTP REST), các Interceptors cho gRPC như `APIKeyCheckStreamGRPCInterceptor` (Stream auth server), `APIKeyBindStreamGRPCInterceptor` (Stream auth client). Định nghĩa `Role` & `Scope`. |
| [`cache`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/cache) | Kết nối Redis (cơ bản) | Hàm khởi tạo Redis client không bọc lỗi chuẩn hóa. |
| [`config`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/config) | Định nghĩa Struct Cấu hình | Các struct config cho DB, Redis, Kafka, ES. Hỗ trợ log formatters qua `slog.LogValue`. |
| [`db`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/db) | Database Driver & Transactions | Khởi tạo Postgres (GORM), định nghĩa `TxManager` hỗ trợ Transactional Outbox Pattern cho HTTP handler. |
| [`es`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/es) | Elasticsearch Connection | Wrapper khởi tạo ES client và cấu hình bulk operations. |
| [`jwt`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/jwt) | Cấp phát & Verify JWT | Sinh Access Token, Refresh Token và phân tích claims. |
| [`model`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/model) | Shared Domain Entities | Model `Server` (versioned), `Account` và `OutboxEvent` dùng chung. |
| [`mq`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/mq) | Message Queue wrappers | Kafka Reader & Writer wrappers sử dụng `segmentio/kafka-go` hỗ trợ manual offset commits. |
| [`pb`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/pb) | Generated Protobuf Code | Các struct và service clients biên dịch từ gRPC `.proto` contracts (sử dụng chủ yếu cho `InternalFileTransferService`). |
| [`rdb`](file:///c:/Users/hailh22/WorkSpace/checkpoint1/microservice/server-management/microservices/pkg/rdb) | Kết nối Redis (chuẩn hóa lỗi) | Khởi tạo Redis client và map lỗi kết nối thông qua `apperr`. |

---

## 2. Các Pattern Thiết kế Quan trọng áp dụng tại `/pkg`

### Pattern 1: Functional Options Pattern cho Initialization
Để đảm bảo code khởi tạo dùng chung linh hoạt và tương thích ngược (không làm crash các service khác khi một service cần thêm tuỳ chỉnh cấu hình), tất cả các hàm constructor trong `pkg/` đều sử dụng Functional Options.

Ví dụ:
```go
// Thay vì truyền nhiều tham số cố định
func NewKafkaWriter(broker string, topic string, retries int) *Writer { ... }

// Sử dụng Functional Option
func NewKafkaWriter(opts ...WriterOption) *Writer { ... }
```

### Pattern 2: Chuẩn hóa Lỗi Tập trung (Centralized Error Handling)
Gói `apperr` là "từ điển" lỗi duy nhất cho toàn hệ thống. Mọi service phải map lỗi từ infra (ví dụ `gorm.ErrRecordNotFound`) sang `apperr` (như `apperr.ErrRecordNotFound`) trước khi trả về tầng HTTP Handler hoặc gRPC để đảm bảo client nhận được thông điệp lỗi nhất quán.
