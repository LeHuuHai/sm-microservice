# Microservices Architecture - Developer Documentation

> **Chào mừng bạn đến với tài liệu hướng dẫn phiên bản Microservice.** Đây là entrypoint chi tiết giải thích kiến trúc, cấu trúc thư mục, các service thành phần và luồng dữ liệu của phiên bản Microservice đã được di chuyển từ Monolith.

---

## 1. Bản đồ Tài liệu (Documentation Map)

Để hiểu rõ hơn về các khía cạnh của phiên bản Microservice, vui lòng xem các tài liệu sau:

| Tài liệu cốt lõi | Nội dung |
|---|---|
| [`architecture.md`](architecture.md) | Kiến trúc tổng quan, mô hình giao tiếp (gRPC, Kafka, REST HTTP) và cơ chế Bảo mật/Xác thực. |
| [`directory-structure.md`](directory-structure.md) | Cấu trúc thư mục của workspace `microservices/` và quy tắc điều hướng mã nguồn. |
| [`services.md`](services.md) | Tóm tắt nhanh danh sách các microservices và host-level agent. |
| [`shared-packages.md`](shared-packages.md) | Hướng dẫn sử dụng các thư viện dùng chung trong `microservices/pkg/`. |

### Tài liệu chi tiết từng Service
- [`auth-service`](auth-service.md)
- [`server-service`](server-service.md)
- [`monitor-service`](monitor-service.md)
- [`heartbeat-gateway`](heartbeat-gateway.md)
- [`ping-worker`](ping-worker.md)
- [`mail-worker`](mail-worker.md)
- [`agent`](agent.md)

---

## 2. So sánh Nhanh: Monolith vs. Microservice

| Đặc tả | Phiên bản Monolith (Cũ) | Phiên bản Microservice (Mới) |
|---|---|---|
| **Cấu trúc Code** | `/cmd/` và `/internal/` ở root | `/microservices/` độc lập (Go Workspaces) |
| **Giao tiếp nội bộ** | Interface in-memory | HTTP REST (Gin) + gRPC + Kafka (Asynchronous) |
| **Bảo mật** | JWT/API Key xử lý tập trung trong Gin | Traefik ForwardAuth + RoleCheck Middleware API/gRPC |
| **Database Ghi nhận** | `pgwriter`/`eswriter` | `monitor-service` tích hợp Batch Writers ghi từ Kafka |
| **Báo cáo Uptime** | Tạo file XLSX đồng bộ qua REST HTTP | Asynchronous 202: Gửi link tải qua gRPC stream + Cache Redis |
| **Triển khai** | Docker Compose thuần | Docker Swarm (Rolling updates, Secrets, Traefik Gateway) |

---

## 3. Các Nguyên Tắc Vận Hành & Phát Triển (Core Guidelines)

1. **Clean Architecture tại từng Microservice:**
   Mỗi service tự quản lý cấu trúc Clean Architecture của riêng nó: `domain` $\rightarrow$ `infra` $\rightarrow$ `handler`/`rpc`.
2. **Không phá vỡ Interface dùng chung (Functional Options Pattern):**
   Mọi helper hoặc wrap-connection trong `pkg/` phải áp dụng Functional Option Pattern để tránh breaking changes.
3. **Event-Driven & Event Ordering:**
   Các thay đổi trạng thái được truyền phát qua Kafka. Sự kiện được đánh `Version` để khử trùng.
4. **Không sử dụng `.env` ở Production:**
   Docker Swarm sẽ inject trực tiếp biến môi trường. Docker Secrets phân phối passwords qua `/run/secrets/`.
