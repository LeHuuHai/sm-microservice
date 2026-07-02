# Auth Service

**auth-service** là microservice chịu trách nhiệm quản lý tài khoản, xác thực người dùng, phân quyền và cấp phát/thu hồi JWT (JSON Web Tokens).

## 1. Công nghệ & Triển khai

- **Ngôn ngữ & Framework:** Go (Golang) + Gin Web Framework.
- **API Router:** Code OpenAPI được sinh tự động thông qua `oapi-codegen` (xem thư mục `api`).
- **Database:** PostgreSQL (Lưu trữ `accounts` với hashed password) thông qua GORM.
- **Cache / In-memory DB:** Redis (Lưu trữ token blocklist cho các Refresh Token đã bị thu hồi hoặc hết hạn).
- **Cơ chế đặc trưng:** 
  - Triển khai **Token Blocklist**: Khi người dùng logout, Refresh Token được đưa vào Redis blocklist với TTL tương ứng thời gian sống còn lại của token để tối ưu dung lượng RAM và ngăn ngừa replay attack.
  - Sử dụng Strict Handler Middleware: Trích xuất `Authorization: Bearer <token>` và truyền vào `context` để sử dụng ở endpoint `Verify`.

## 2. Giao tiếp (APIs)

Dịch vụ này phơi bày các API thông qua giao thức **HTTP REST** (không dùng gRPC như mô tả ở thiết kế cũ).

- **Port hoạt động:** Được cấu hình qua biến môi trường (`APP_PORT`, thường là 8080 trong Docker container hoặc config local).

Các REST Endpoint chính:
- `POST /auth/login`: Xác thực thông tin người dùng và trả về Access Token + Refresh Token.
- `POST /auth/refresh`: Dùng Refresh Token hợp lệ (chưa bị block, chưa hết hạn) để cấp mới Access Token.
- `POST /auth/logout`: Thu hồi Refresh Token, đưa vào Redis blocklist.
- `GET /auth/verify`: Trích xuất token từ header, xác thực hợp lệ và trả về 200 OK kèm các headers `X-User-ID` và `X-User-Role`. *Endpoint này đóng vai trò quan trọng làm **ForwardAuth middleware** cho Traefik Gateway để xác thực mọi request.*

## 3. Cấu trúc thư mục nội bộ
- `api/`: Nơi chứa code generated từ OpenAPI.
- `cmd/main.go`: Entry point khởi tạo các dependencies và HTTP Server.
- `internal/`:
  - `domain/`: Các interface nghiệp vụ (AuthServiceInterface, AccountRepository...).
  - `handler/`: `auth_rest_handler.go` xử lý HTTP request / response.
  - `infra/`: `postgres` (implement repo), `redis` (implement blocklist), `runtime` (khởi tạo app).
  - `service/`: Triển khai logic nghiệp vụ cấp phát JWT.
