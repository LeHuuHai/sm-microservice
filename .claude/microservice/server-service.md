# Server Service

**server-service** là microservice đóng vai trò nguồn chân lý (Source of Truth) quản lý danh mục máy chủ (Inventory Catalog). Bất kỳ thay đổi cấu hình tĩnh nào của server đều được quản lý tại đây.

## 1. Công nghệ & Triển khai

- **Ngôn ngữ & Framework:** Go (Golang) + Gin Web Framework.
- **API Router:** Code OpenAPI được sinh tự động thông qua `oapi-codegen`.
- **Database:** PostgreSQL lưu trữ bảng `servers` (id, tên, IP, metadata) và bảng `outbox` để phục vụ Outbox Pattern.
- **Background Worker:** `OutboxPoller` chạy ngầm để quét bảng `outbox` và đẩy sự kiện lên Kafka.
- **Bảo mật:** Sử dụng middleware `RoleCheckMiddleware` (nằm trong thư mục `pkg/auth`) để kiểm tra quyền truy cập dựa trên Role được header `X-User-Role` của Traefik truyền xuống.

## 2. Giao tiếp

Dịch vụ này sử dụng giao tiếp **HTTP REST** kết hợp với **Kafka** cho luồng event-driven (không dùng gRPC nội bộ).

### REST APIs (HTTP)

- **Port hoạt động:** Được cấu hình qua biến môi trường (thường là 8080).
- Các Endpoints:
  - `GET /servers`: Lấy danh sách server (có phân trang và filter).
  - `POST /servers`: Đăng ký máy chủ mới.
  - `PATCH /servers/{id}`: Cập nhật thông tin server.
  - `DELETE /servers/{id}`: Xóa server.
  - `POST /servers/import`: Nhập danh sách máy chủ hàng loạt thông qua file XLSX (tích hợp internal `xlsximport`).
  - `GET /servers/export`: Trích xuất file báo cáo danh sách XLSX (tích hợp internal `xlsxexport`).

### Kafka Event Publishing

Áp dụng **Transactional Outbox Pattern**:
1. Thêm/Sửa/Xóa dữ liệu trong bảng `servers`.
2. Lưu một Record vào bảng `outbox` mang sự kiện `ServerCreated`, `ServerUpdated`, hoặc `ServerDeleted`.
3. Cả 2 thao tác trên thực hiện trong **cùng 1 Database Transaction**.
4. Worker ngầm đọc bảng `outbox` và publish lên **Kafka topic:** `server_lifecycle`.
- *Điều này đảm bảo tính At-Least-Once delivery cho các thay đổi dữ liệu, để các microservice khác (như `monitor-service`) đồng bộ một cách an toàn.*

## 3. Cấu trúc thư mục nội bộ
- `api/`: Nơi chứa interface và các struct generated từ OpenAPI.
- `cmd/main.go`: Khởi tạo dependencies (TxManager, OutboxRepo, ServerRepo, EventPublisher) và bắt đầu REST Server cùng `OutboxPoller`.
- `internal/`:
  - `domain/`: Các định nghĩa interface `service`, logic parsing file `export`/`deserialize`.
  - `handler/`: `server_rest_handler.go` nhận HTTP request.
  - `infra/`: `file` (logic read/write XLSX), `kafka` (publish event), `postgres` (SQL query/tx), `worker` (outbox poller).
