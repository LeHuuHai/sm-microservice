# Kiến trúc và Cơ chế Giao tiếp (System Architecture & Communication)

Tài liệu này mô tả chi tiết thiết kế hệ thống, các luồng tương tác và cơ chế bảo mật phi tập trung (Decentralized Auth) của phiên bản Microservice.

---

## 1. Sơ đồ Kiến trúc Microservice (Architecture Diagram)

Hệ thống hoạt động theo mô hình **Hybrid Event-Driven, REST HTTP & gRPC**:
- **Luồng dữ liệu thu thập / xử lý (Uptime Checking & Write-path):** Đi qua Kafka để đảm bảo khả năng scale lớn, chịu tải tốt, và cách ly lỗi (fault isolation).
- **Luồng điều khiển / giao tiếp Client (Read-path & Admin):** Đi qua REST HTTP thông qua Traefik API Gateway để dễ dàng tương tác với Frontend.
- **Luồng truyền tải file nội bộ:** Đi qua gRPC Stream để tối ưu tốc độ và bộ nhớ.

```text
                  ┌──────────────────────┐
                  │    Web/Mobile App    │
                  └──────────┬───────────┘
                             │ REST HTTP (Authorization: Bearer JWT)
                             ▼
                           ┌───────────┐         ForwardAuth
                           │  Traefik  │ ──────────────────────────────┐
                           │  Gateway  │                               │
                           └─────┬─────┘                               │
                                 │                                     ▼
┌─────────────┐                  │ HTTP REST                 ┌───────────┐
│ host-agent  │ ───► POST /hb ───┤ (X-User-ID & X-User-Role) │   Auth    │
│ (monitored) │ (X-API-Key)      ├─────────────────────────► │  Service  │
└─────────────┘                  │                           └───────────┘
                                 │
                                 ├──► HTTP REST ───────────► ┌───────────┐
                                 │                           │  Server   │
                                 │                           │  Service  │
                                 │                           └─────┬─────┘
                                 │                                 │ Server Lifecycle
                                 │                                 │ Events (Versioned)
                                 ├──► HTTP REST (X-API-Key)  ┌───────────┐
                                 │                           │ Heartbeat │
                                 │                           │  Gateway  │
                                 │                           └─────┬─────┘
                                 │                                 │ Heartbeats
                                 │                                 │ (Acks=0)
                                 ▼                                 ▼
┌────────────────────────────────┼─────────────────────────────────┼─┐
│                                │          KAFKA BUS              │ │
│  Topics:                       ▼                                 ▼ │
│  - heartbeats ◄─────────────────────────────────────────────────┐  │
│  - server_lifecycle ─────────────────────────────────────────┐  │  │
│  - ping ──────────────────────────────────────────────────┐  │  │  │
│  - ping_res ◄──────────────────────────────────────────┐  │  │  │  │
│  - mail ───────────────────────────┐                   │  │  │  │  │
└────────────────────────────────────┼───────────────────┼──┼──┼──┼──┘
                                     │                   │  │  │  │
    ┌───────────┐                    │                   │  │  │  │
    │   Ping    │ ◄── (Consume) ─────┼───────────────────┼──┼──┘  │
    │  Worker   │ ─── (Publish) ─────┼───────────────────┘  │     │
    └───────────┘                    │                      │     │
                                     │                      ▼     ▼
                                     │               ┌───────────┐
                                     │               │  Monitor  │ ◄── (Từ Web: HTTP POST /report)
                                     │               │  Service  │
                                     │               └─────┬─────┘
                                     │                     │
                                     │ Kafka (mail)        │ gRPC Stream
                                     │                     │ (X-API-Key)
                                     ▼                     ▼
                               ┌───────────┐
                               │   Mail    │ ◄─────────────┘
                               │  Worker   │
                               └───────────┘
```

---

## 2. Các Luồng Nghiệp Vụ Chính (Process Flows)

### Flow A: Thu thập Heartbeat thụ động (Passive Checking)
1. **Agent** gửi HTTP `POST /heartbeat` đến `heartbeat-gateway` thông qua Traefik kèm theo header `X-API-Key`.
2. **Heartbeat Gateway** xác thực API Key, đóng gói event và publish nhanh vào Kafka topic `heartbeats` (sử dụng cấu hình `RequiredAcks = 0` / fire-and-forget để tối ưu hiệu năng).
3. **Monitor Service** (chứa consumer) đọc message từ Kafka:
   - **Ghi trạng thái hiện tại:** Cập nhật `LiveStatus` vào PostgreSQL bằng cơ chế gom lô (Buffered Micro-Batch).
   - **Ghi lịch sử:** Cập nhật `StatusLog` vào Elasticsearch phục vụ báo cáo.

### Luồng phụ: Đồng bộ hóa danh mục Server (Server Synchronization)
1. Khi có bất kỳ thay đổi nào về server (thêm, sửa, xóa) tại `server-service`, một event tương ứng được gửi vào Kafka topic `server_lifecycle` thông qua Transactional Outbox Pattern.
2. **Monitor Service** đọc event này để đồng bộ và duy trì một **bản sao danh sách server** cục bộ (`MonitoredServer` table) trong PostgreSQL.
3. Bản sao cục bộ này dùng làm căn cứ danh mục máy chủ để background checker thực hiện quét và lấy thông tin địa chỉ IP.

### Flow B: Ping chủ động kiểm tra (Active Checking)
1. **Monitor Service** chạy background checker định kỳ truy vấn danh sách server được giám sát (`monitoredServerRepo`) và các trạng thái trực tuyến (`liveStatusRepo`) trực tiếp từ PostgreSQL local database. Nếu phát hiện server có thời gian nhận heartbeat gần nhất (`LastHeartbeatAt`) vượt quá timeout quy định, service sẽ gửi một event `PingRequested` vào Kafka.
2. **Ping Worker** nhận event từ topic `ping`, thực hiện ICMP ping trực tiếp tới IP của server mục tiêu.
3. Ping Worker publish kết quả `PingResult` vào Kafka topic `ping_res`.
4. **Monitor Service** đọc kết quả, cập nhật `LiveStatus` trong Postgres và ghi log Elasticsearch qua batcher.

### Flow C: Báo cáo và Gửi Mail (Report Generation & Delivery)
1. Client gọi REST API `POST /report` tạo báo cáo thông qua Traefik Gateway.
2. **Monitor Service** tiến hành truy vấn Redis cache (hoặc MapReduce ES logs) để sinh báo cáo Excel và lưu file cục bộ.
3. Sau khi sinh file thành công, **Monitor Service** đẩy tin nhắn chứa `filename` vào Kafka topic `mail`, sau đó trả về HTTP `202 Accepted` cho Client (báo hiệu quá trình gửi mail sẽ được xử lý ngầm).
4. **Mail Worker** nhận event từ Kafka, thực hiện gọi gRPC Client kết nối đến `InternalFileTransferService` của `monitor-service`.
5. Dữ liệu file XLSX được truyền tải an toàn bằng **gRPC Stream** qua mạng nội bộ.
6. **Mail Worker** nhận đủ stream file, tiến hành gửi email qua SMTP.

---

## 3. Kiến trúc Xác thực & Phân quyền Phi tập trung (Decentralized Security)

Để tránh nút thắt cổ chai và đảm bảo nguyên lý bảo mật Zero-Trust, hệ thống phân chia trách nhiệm xác thực theo mô hình ForwardAuth:

```text
[REST Client] 
     │ Authorization: Bearer JWT
     ▼
┌───────────────┐
│  Traefik      │ ─── (ForwardAuth Request) ───► ┌───────────────┐
│  Gateway      │ ◄── (X-User-ID, X-User-Role) ──┤ auth-service  │ (Verify JWT)
└───────┬───────┘                                └───────────────┘
        │
        │ HTTP Headers:
        │   X-User-ID: "usr_123"
        │   X-User-Role: "admin"
        ▼
┌───────────────┐
│ Backend REST  │  (Chạy RoleCheckMiddleware so khớp role với required scope)
│ (server/mon)  │
└───────────────┘
```

1. **Kiểm tra JWT tại Traefik Gateway:** Traefik sử dụng middleware ForwardAuth để gửi request tới endpoint `GET /auth/verify` của `auth-service`. Nếu token hợp lệ, `auth-service` sẽ trả về mã 200 OK kèm theo các HTTP Headers `X-User-ID` và `X-User-Role`. Traefik sau đó tự động inject các header này vào request gốc và chuyển tiếp (forward) xuống các service nội bộ (`server-service`, `monitor-service`).
2. **Ủy quyền tại Service (Gin Middleware):** `server-service` và `monitor-service` sử dụng `RoleCheckMiddleware` của Gin để chặn các API requests. Nó đọc giá trị header `X-User-Role` và gọi hàm kiểm tra xem role đó có quyền truy cập endpoint hiện tại hay không, độc lập hoàn toàn với `auth-service`.
3. **Mã API Key nội bộ (Streaming Interceptor):** Các luồng streaming (như `DownloadReport`) không đi qua Traefik mà được kết nối trực tiếp bởi `mail-worker` tới `monitor-service`. Luồng gRPC này được bảo vệ bằng API Key dùng chung, tự động inject ở Client qua `APIKeyBindStreamGRPCInterceptor` và được xác thực ở Server qua `APIKeyCheckStreamGRPCInterceptor` bằng metadata.
4. **Agent Verification:** `heartbeat-gateway` không dùng JWT hay role, nó xác thực request gửi lên từ Agent trực tiếp bằng Gin middleware (`APIKeyMiddleware`) dựa vào header HTTP `X-API-Key` với cấu hình cứng trong secret.
