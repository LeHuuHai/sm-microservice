# Server Management System (Microservices)

Dự án này là phiên bản **Kiến trúc Microservice** của Server Management System, một hệ thống giám sát hạ tầng phân tán, hướng sự kiện (event-driven) để theo dõi tính khả dụng và metrics của các server fleet.

---

## 1. Sơ đồ Kiến trúc Tổng thể (Architecture Diagram)

Hệ thống hoạt động theo mô hình **Hybrid Event-Driven, REST HTTP & gRPC**:
- **Luồng dữ liệu thu thập / xử lý (Uptime Checking & Write-path):** Đi qua Kafka để đảm bảo khả năng scale lớn, chịu tải tốt, và cách ly lỗi.
- **Luồng điều khiển / giao tiếp Client (Read-path & Admin):** Đi qua REST HTTP thông qua Traefik API Gateway.
- **Luồng truyền tải file nội bộ:** Đi qua gRPC Stream để tối ưu tốc độ.

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

## 2. Cây Thư Mục (Project Structure)

Dự án được cấu trúc theo dạng **Go Workspace**, quản lý nhiều microservices độc lập chia sẻ chung module `pkg`.

```text
microservices/
├── go.work                       # File khai báo workspace của Go
├── pkg/                          # THƯ VIỆN DÙNG CHUNG (Shared Libraries)
│   ├── apperr/, auth/, db/, es/, mq/, pb/ ...
│
├── auth-service/                 # MICROSERVICE: QUẢN LÝ TÀI KHOẢN & XÁC THỰC
├── server-service/               # MICROSERVICE: QUẢN LÝ DANH MỤC SERVER
├── heartbeat-gateway/            # MICROSERVICE: THU THẬP HTTP HEARTBEATS
├── monitor-service/              # MICROSERVICE: PHÂN TÍCH, GHI NHẬT KÝ & BÁO CÁO
├── ping-worker/                  # WORKER DAEMON: THỰC HIỆN ICMP PING
├── mail-worker/                  # WORKER DAEMON: GỬI BÁO CÁO QUA SMTP
├── agent/                        # HOST-LEVEL DAEMON: GỬI HEARTBEAT
└── deploy/                       # DOCKER SWARM STACKS
```

---

## 3. Design Decomposition (Cơ sở phân chia Microservices)

Tuân thủ phương pháp **Domain-Driven Design (DDD)**, hệ thống được phân rã qua 4 bước ánh xạ (mapping):

**Bước 1: Xác định Năng lực nghiệp vụ (Business Capabilities)**
1. Quản lý truy cập tài khoản.
2. Quản lý danh mục máy chủ.
3. Tiếp nhận tín hiệu sống tần suất cao.
4. Kiểm tra ICMP ping chủ động.
5. Phân tích trạng thái, ghi log và báo cáo.
6. Gửi thông báo (Email).

**Bước 2: Phân loại Subdomain (Problem Space)**
- **Core Subdomain:** Monitoring & Analytics (logic tính toán uptime, MapReduce báo cáo).
- **Supporting Subdomains:** Server Inventory, Heartbeat Ingestion, Active Probing, Notification.
- **Generic Subdomain:** Account & Access (chuẩn công nghiệp).

**Bước 3 & Bước 4: Ánh xạ thành Bounded Contexts và Microservices (Solution Space)**

| Phân loại | Bounded Context (Ngữ cảnh) | Năng lực nghiệp vụ được giao | Thuật ngữ Domain | Microservice |
| --- | --- | --- | --- | --- |
| **Generic** | **Auth Context** | Đăng nhập, phân quyền, cấp phát/thu hồi JWT | `Account`, `Token`, `Role` | `auth-service` |
| **Supporting**| **Server Context** | Quản lý bản ghi server tĩnh | `ServerProfile`, `UpdatedAt` | `server-service` |
| **Supporting**| **Gateway Context** | Thu thập heartbeat tần suất cao | `RawHeartbeat` | `heartbeat-gateway` |
| **Supporting**| **Ping Context** | Thực thi lệnh ICMP ping chủ động | `PingTask`, `ICMPResult`| `ping-worker` |
| **Core** | **Monitor Context** | Cập nhật trạng thái trực tuyến, ghi log | `LiveStatus`, `StatusLog` | `monitor-service` |
| **Supporting**| **Mail Context** | Đóng gói báo cáo và gửi email SMTP | `EmailPayload` | `mail-worker` |

*(Agent là thành phần Host-level chạy ở Client).*

### Định nghĩa thuật ngữ ranh giới (Core Domain Terms)
Sự phân tách thể hiện rõ nhất giữa Server Context và Monitor Context:
- **`ServerProfile` (Server Context)**: Chỉ chứa thông tin tĩnh của server (`ServerID`, `ServerName`, `IPv4`). Hoàn toàn không chứa trạng thái hoạt động để đảm bảo chia cắt domain.
- **`LiveStatus` (Monitor Context)**: Theo dõi các metric uptime hiện tại (`Status`, `LastPingAt`, `LastHeartbeatAt`), đồng bộ `ServerProfile` nội bộ từ các Kafka events.

---

## 4. Danh Sách Các Services & Cấu hình (Environment Variables)

Mỗi service sử dụng `.env` riêng biệt, được load tại runtime.

### Traefik API Gateway
- Đóng vai trò Reverse Proxy tổng, định tuyến HTTP REST (Port `8080` khi chạy local).

### `auth-service` (Port: 8080)
- **Vai trò**: Quản lý accounts, issue JWT, Traefik ForwardAuth.
- **ENV**: `APP_PORT`, `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `REDIS_URL`, `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `JWT_ACCESS_EXPIRED`.

### `server-service` (Port: 8080)
- **Vai trò**: Quản lý Server Catalog tĩnh, xuất Kafka event `server_lifecycle`.
- **ENV**: `APP_PORT`, `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `KAFKA_BROKER`, `KAFKA_LIFECYCLE_TOPIC`.

### `heartbeat-gateway` (Port: 8080)
- **Vai trò**: Ingestion layer tĩnh nhận POST /heartbeat.
- **ENV**: `APP_PORT`, `APP_HEARTBEAT_KEY`, `KAFKA_BROKER`, `KAFKA_HEARTBEAT_TOPIC`.

### `monitor-service` (Port: 8080 - HTTP, 8180 - gRPC)
- **Vai trò**: Consumer heartbeat/ping, tính toán báo cáo MapReduce, gRPC file stream.
- **ENV**: `APP_PORT`, `APP_GRPC_PORT`, `DB_HOST`, `ES_URL`, `REDIS_URL`, `KAFKA_BROKER`, `KAFKA_HEARTBEAT_TOPIC`, `KAFKA_LIFECYCLE_TOPIC`, `KAFKA_PING_RES_TOPIC`, `APP_REPORT_KEY`.

### `ping-worker` (Daemon)
- **Vai trò**: Tiêu thụ event, thực hiện ICMP ping (yêu cầu `CAP_NET_RAW` trên Linux).
- **ENV**: `APP_NUM_THREAD`, `KAFKA_BROKER`, `KAFKA_PING_TOPIC`, `KAFKA_PING_RES_TOPIC`.

### `mail-worker` (Daemon)
- **Vai trò**: Tiêu thụ event mail, pull gRPC stream, gửi SMTP.
- **ENV**: `KAFKA_BROKER`, `KAFKA_MAIL_TOPIC`, `REPORT_REPO_ADDR` (trỏ về `monitor-service:8180`), `APP_REPORT_KEY`, `GOMAIL_ADDR`, `GOMAIL_PORT`, `GOMAIL_FROM`, `GOMAIL_PASSWORD`.

### `agent` (Daemon)
- **Vai trò**: Chạy trên host để gửi hearbeat.
- **ENV**: `APP_SERVER_ID`, `APP_HEARTBEAT_URL`, `APP_HEARTBEAT_KEY`, `APP_CYCLE_HEARTBEAT`.

---

## 5. API Specification (Chi tiết Endpoints)

Tất cả các endpoint đi qua Traefik (trừ heartbeat/report stream) đều yêu cầu `Authorization: Bearer <access_token>`. Traefik sẽ dùng ForwardAuth để decode token thành `X-User-Role`.

### 1. List Servers (`server-service`)
*   **Method / Path**: `GET /servers`
*   **Query Parameters**: `from`, `to`, `sort_field`, `desc`
*   **Response (200 OK)**: Trả về danh sách ServerProfile tĩnh (không chứa trạng thái).

### 2. Create Server (`server-service`)
*   **Method / Path**: `POST /servers`
*   **Request Body**:
    ```json
    { "server_id": "sv1", "server_name": "Web Proxy 1", "ipv4": "10.10.10.101" }
    ```
*   **Response (201 Created)**: Tạo mới và publish event `ServerCreated`.

### 3. Import Servers (`server-service`)
*   **Method / Path**: `POST /servers/import`
*   **Content-Type**: `multipart/form-data` (file: `.xlsx`)

### 4. Heartbeat (`heartbeat-gateway`)
*   **Method / Path**: `POST /heartbeat`
*   **Header**: `X-API-Key: <key>`
*   **Request Body**: `{"server_id": "sv1", "timestamp": 1690000000}`
*   **Response (200 OK)**: Xử lý siêu tốc, đẩy thẳng vào Kafka.

### 5. Generate Report (`monitor-service`)
*   **Method / Path**: `POST /report`
*   **Request Body**:
    ```json
    {
      "from": "2026-06-07T00:00:00Z",
      "to": "2026-06-08T00:00:00Z",
      "receivers": ["admin@mycompany.com"]
    }
    ```
*   **Response (202 Accepted)**: Xử lý báo cáo bất đồng bộ và đẩy xuống `mail-worker` qua Kafka.

---

## 6. Giao thức Giao tiếp Bất Đồng Bộ (Kafka Topics)

Kafka là xương sống giao tiếp của dự án:
- `server_lifecycle`: Phân phối các sự kiện CRUD máy chủ (`server-service` $\rightarrow$ `monitor-service`).
- `heartbeats`: Chuyển tiếp tín hiệu sống dạng fire-and-forget (`heartbeat-gateway` $\rightarrow$ `monitor-service`).
- `ping`: Yêu cầu kiểm tra chủ động (`monitor-service` $\rightarrow$ `ping-worker`).
- `ping_res`: Trả về kết quả ICMP Ping (`ping-worker` $\rightarrow$ `monitor-service`).
- `mail`: Yêu cầu gửi email có đính kèm tên file (`monitor-service` $\rightarrow$ `mail-worker`).

---

## 7. Deployment

Hệ thống được thiết kế để triển khai native trên **Docker Swarm** với cấu trúc mạng overlay và quản lý bảo mật qua **Docker Secrets**. Toàn bộ file cấu hình nằm trong thư mục `microservices/deploy/`.

### Bước 1: Khởi tạo Docker Swarm & Network
Đảm bảo máy chủ của bạn đã bật chế độ Swarm và tạo mạng overlay dùng chung:
```bash
docker swarm init
docker network create -d overlay --attachable sm_network
```

### Bước 2: Khởi tạo Secrets
Hệ thống sử dụng Docker Secrets thay cho biến môi trường để bảo vệ thông tin nhạy cảm. Chạy script để tự động nạp các secret mặc định (bạn có thể sửa lại nội dung file script trước khi chạy để thay đổi password):
```bash
cd microservices/deploy
bash init-secrets.sh
```

### Bước 3: Triển khai Infrastructure Stack
Khởi chạy các database (PostgreSQL, Redis, Elasticsearch) và message broker (Kafka). Chúng được gộp chung vào một stack riêng để tối ưu I/O và đảm bảo dữ liệu:
```bash
docker stack deploy -c docker-stack-infra.yml sm-infra
```
*Đợi khoảng 1-2 phút để Kafka broker và Elasticsearch hoàn tất quá trình khởi động.*

### Bước 4: (Tùy chọn) Build & Push Images
Nếu bạn có thay đổi mã nguồn, hãy sử dụng powershell script để tự động build và push toàn bộ 7 images lên Docker Registry:
```powershell
.\build-push.ps1
```
*(Cần set biến môi trường `$env:DOCKER_REGISTRY` trong script trước khi chạy).*

### Bước 5: Triển khai Application Stack
Khởi chạy tầng ứng dụng gồm Traefik API Gateway và 6 microservices:
```bash
docker stack deploy -c docker-stack-app.yml sm-app
```
Traefik sẽ tự động định tuyến các request từ Host port `80` (hoặc cấu hình tùy chỉnh) vào mạng Swarm nội bộ (`8080` của từng service).

### Bước 6: Khởi chạy Mock Agents
Để kiểm thử hệ thống nhận Heartbeat thực tế, dự án đi kèm một kịch bản giả lập 10 agent liên tục gửi tín hiệu:
```bash
bash run-mock-agents.sh
```
Script này sẽ spin up 10 container độc lập đại diện cho các servers (với mã `server_00001` đến `server_00010`). Bạn cần dùng postman gọi API `POST /servers` để khai báo danh sách server tĩnh khớp với IDs này. Để dừng agent, chạy `bash stop-mock-agents.sh`.

---

## 8. Tài liệu kỹ thuật chi tiết (Developer Documentation)

Để giữ cho `README.md` ngắn gọn, các thiết kế kiến trúc chuyên sâu và cấu trúc mã nguồn được lưu trữ trong thư mục `.claude/microservice/`. Các developer hoặc maintainer nên đọc thêm các tài liệu sau:

- 📖 **[Developer Documentation Hub](.claude/microservice/index.md)**: Trang mục lục tổng hợp toàn bộ tài liệu về thiết kế Microservice.
- 📐 **[Kiến trúc và Luồng giao tiếp](.claude/microservice/architecture.md)**: Giải thích chi tiết về Decentralized Auth, các luồng Event-Driven (Heartbeat, Ping) qua Kafka, và gRPC file streaming.
- 📂 **[Quy tắc tổ chức Workspace & Thư mục](.claude/microservice/directory-structure.md)**: Diễn giải cấu trúc `go.work`, các quy chuẩn Clean Architecture áp dụng trong từng service, và cách dùng chung package `pkg`.
- 🧩 **[Chi tiết từng Microservices](.claude/microservice/services.md)**: Tài liệu bóc tách vai trò của 7 microservices (`auth`, `server`, `monitor`, `heartbeat-gw`, `mail-worker`, `ping-worker`, `agent`) và cấu trúc `.env` chuyên sâu của chúng.
- 📦 **[Thư viện dùng chung (Shared Packages)](.claude/microservice/shared-packages.md)**: Diễn giải về các thư viện xử lý Kafka, DB, Redis, gRPC protobuf... nằm trong thư mục `pkg`.
