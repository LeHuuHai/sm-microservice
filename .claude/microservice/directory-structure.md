# Cấu trúc Thư mục Workspace Microservice

Thư mục `/microservices` được tổ chức theo cơ chế **Go Workspaces (`go.work`)**, cho phép quản lý nhiều microservices độc lập trong một repository duy nhất mà vẫn chia sẻ được code dùng chung.

---

## 1. Sơ đồ Tổng quan Workspace

```text
microservices/
├── go.work                       # File khai báo workspace của Go
├── go.work.sum                   # File sum chung của workspace
│
├── pkg/                          # THƯ VIỆN DÙNG CHUNG (Shared Libraries)
│   ├── apperr/                   # Định nghĩa lỗi hệ thống
│   ├── auth/                     # Middleware RoleCheck và Interceptors
│   ├── db/                       # Khởi tạo Postgres & transaction helper
│   ├── es/                       # Khởi tạo Elasticsearch client
│   ├── mq/                       # Wrappers cho Kafka reader & writer
│   ├── pb/                       # Generated Go code từ Protobuf (.proto)
│   └── ...
│
├── auth-service/                 # MICROSERVICE: QUẢN LÝ TÀI KHOẢN & XÁC THỰC
│   ├── cmd/                      # Entrypoint (main.go)
│   ├── api/                      # OpenAPI generated code
│   └── internal/
│       ├── handler/              # HTTP REST Handler (`auth_rest_handler.go`)
│       ├── domain/               # Core business interfaces
│       ├── infra/                # Triển khai PostgreSQL, Redis blocklist
│       └── service/              # Triển khai business logic
│
├── server-service/               # MICROSERVICE: QUẢN LÝ DANH MỤC SERVER
│   ├── cmd/                      # Entrypoint (main.go)
│   ├── api/                      # OpenAPI generated code
│   └── internal/
│       ├── handler/              # HTTP REST Handler (`server_rest_handler.go`)
│       ├── domain/               # Core interfaces, Import/Export logic
│       ├── infra/                # Postgres, Kafka Publisher, File Export/Import
│       └── service/              # Triển khai business logic
│
├── heartbeat-gateway/            # MICROSERVICE: THU THẬP HTTP HEARTBEATS
│   ├── cmd/                      # Entrypoint (main.go)
│   ├── api/                      # REST API Contract
│   └── internal/
│       ├── handler/              # HTTP Gin handler nhận heartbeat
│       ├── infra/                # Triển khai Kafka publisher (Acks=0)
│       ├── middleware/           # API Key checker middleware
│       └── service/              # Gateway service push event
│
├── monitor-service/              # MICROSERVICE: PHÂN TÍCH, GHI NHẬT KÝ & BÁO CÁO
│   ├── cmd/                      # Entrypoint (main.go)
│   ├── api/                      # REST API Contract cho HTTP
│   └── internal/
│       ├── handler/              # HTTP REST Handler cho report
│       ├── rpc/                  # gRPC Handler cho InternalFileTransferService
│       ├── domain/               # Interfaces nghiệp vụ, models
│       ├── infra/                # Postgres, Redis, Elasticsearch, Kafka, Worker loops
│       └── service/              # Core logic: monitor, report, batch processing
│
├── ping-worker/                  # WORKER DAEMON: THỰC HIỆN ICMP PING
│   ├── cmd/                      # Entrypoint (main.go)
│   └── internal/
│       ├── infra/                # Kafka reader/writer, runtime config
│       └── worker/               # Worker pool quản lý ping concurrent
│
├── mail-worker/                  # WORKER DAEMON: GỬI BÁO CÁO QUA SMTP
│   ├── cmd/                      # Entrypoint (main.go)
│   └── internal/
│       ├── infra/                # Kafka Consumer, Gomail Sender
│       ├── service/              # Download service (gRPC client pull)
│       └── worker/               # Worker nhận Kafka message
│
├── agent/                        # HOST-LEVEL DAEMON: GỬI HEARTBEAT
│   ├── cmd/                      # Entrypoint chứa loop Ticker & HTTP Client
│   ├── config/                   # Đọc cấu hình ENV
│   └── internal/
│       └── model/                # Model heartbeat
└── deploy/                       # CHỨA CÁC FILE TRIỂN KHAI DOCKER SWARM/COMPOSE
```

## 2. Quy tắc Tổ chức Code trong từng Microservice

Mỗi microservice trong danh sách trên đều tuân theo nguyên lý thiết kế **Dependency Inversion** của Clean Architecture:

- **Tách biệt Interface và Implementation:** Interface nghiệp vụ được định nghĩa trong `internal/domain/`. Các thư viện bên ngoài hoặc driver DB (Infra layer) được định nghĩa trong `internal/infra/` và bắt buộc phải implement các interface của domain.
- **Tách biệt Config và Runtime:** File cấu hình (`internal/config/`) chỉ đọc môi trường. Việc khởi tạo các kết nối và inject dependencies được thực hiện trong `internal/infra/runtime/` (thường xuất ra một struct `App` dùng chung).
- **Lớp REST/RPC:** `internal/handler` chịu trách nhiệm xử lý các request HTTP (REST) và map sang domain model, trong khi `internal/rpc` (nếu có) xử lý việc map dữ liệu từ protobuf struct sang domain model, thực thi service, và bắt lỗi chuyển thành gRPC codes tương ứng.

## 3. Thông tin điều hướng chi tiết

- **Tài liệu cấu trúc chi tiết từng Service:** Tham khảo [services.md](services.md) hoặc các file md riêng biệt của từng service.
