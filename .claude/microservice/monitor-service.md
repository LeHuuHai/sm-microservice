# Monitor Service

**monitor-service** là trái tim của hệ thống giám sát. Dịch vụ này xử lý trạng thái trực tuyến (online/offline) của các server, ghi nhận nhật ký sự kiện, kiểm tra timeout chủ động, và tính toán/sinh báo cáo uptime.

## 1. Công nghệ & Triển khai

- **Ngôn ngữ & Framework:** Go (Golang) + Gin (REST) + gRPC.
- **Databases & Caching:**
  - **PostgreSQL:** Lưu bản sao danh sách server (`monitored_servers`) đồng bộ từ `server-service`, và trạng thái trực tuyến hiện tại (`live_status`).
  - **Elasticsearch:** Lưu nhật ký sự kiện dạng chuỗi thời gian (timeseries `status_logs`) phục vụ tính toán uptime bằng tính năng Aggregation.
  - **Redis:** Làm lớp Cache (sử dụng MapReduce pattern kết hợp Elasticsearch) để lưu kết quả báo cáo của các ngày đã qua, tăng tốc độ xử lý lên đến 95%.
- **Background Processing:**
  - **Batch Writers:** Sử dụng Go channels (`pgChan`, `esChan`) để gom các sự kiện heartbeat/ping nhỏ lẻ thành các batch lớn (ví dụ: batch 2000 records hoặc 1 giây timeout) trước khi ghi vào PostgreSQL và Elasticsearch nhằm tối ưu I/O.
  - **Active Checker:** Một loop chạy ngầm quét trực tiếp `monitored_servers` và `live_status` để phát hiện server mất heartbeat, sau đó đẩy yêu cầu lên Kafka để `ping-worker` kiểm tra.
  - **Daily Report Cron:** Cronjob tự động sinh báo cáo mỗi đêm lúc 00:00.

## 2. Giao tiếp (APIs & RPCs)

Dịch vụ này vận hành đa giao thức (REST, gRPC, và Kafka) và đảm nhận nhiều vai trò xử lý bất đồng bộ.

### HTTP REST APIs
- `POST /report`: Gửi yêu cầu sinh báo cáo khẩn cấp ngoài chu kỳ. Nhận HTTP 202 Accepted ngay lập tức (xử lý bất đồng bộ qua Kafka). Áp dụng `RoleCheckMiddleware`.

### Internal gRPC (Cung cấp cho `mail-worker`)
- `InternalFileTransferService.DownloadReport`: Mở luồng stream (Server-streaming) để gửi tệp báo cáo nhị phân dạng chunk cho `mail-worker` tải về, thay vì nhét file lớn vào payload Kafka. Sử dụng `APIKeyCheckStreamGRPCInterceptor` để bảo mật nội bộ.

### Kafka Event Streaming
- **Consumers (Đăng ký nhận):**
  - `server_lifecycle`: Nhận thông báo tạo/sửa/xóa từ `server-service` để cập nhật bảng `monitored_servers` local. Áp dụng phiên bản (Version) để khử trùng và chống Lost Update.
  - `heartbeats`: Nhận sự kiện có tín hiệu từ `heartbeat-gateway`.
  - `ping_res`: Nhận kết quả ping ICMP từ `ping-worker`.
  - *Cơ chế Consumer:* Tắt auto-commit, chỉ commit offset khi luồng Batch Writer đã ghi thành công vào DB (At-Least-Once).
- **Publishers (Phát hành):**
  - `ping`: Đẩy yêu cầu kiểm tra ICMP khi phát hiện mất heartbeat.
  - `mail`: Đẩy thông báo chứa `filename` để `mail-worker` thực hiện tải file và gửi email.

## 3. Cấu trúc thư mục nội bộ
- `api/`: OpenAPI generated code cho HTTP REST.
- `cmd/main.go`: Khởi tạo toàn bộ dependency (DBs, Caches, Kafka Brokers), Worker ngầm, Batch Services, REST Server và gRPC Server.
- `internal/`:
  - `domain/`: Định nghĩa các Service interfaces (`MonitorServiceInterface`, `ReportServiceInterface`), Repository interfaces.
  - `handler/`: `report_rest_handler.go` nhận yêu cầu báo cáo.
  - `infra/`: `elasticsearch` (Aggregator/Writer), `kafka` (Consumer/Publisher logic), `postgres` (Repositories), `redis` (Cache), `worker` (Consumers loop, Active checker).
  - `model/`: Các entity cho LiveStatus, StatusLog.
  - `rpc/`: Implement gRPC server cho việc truyền tệp (`TransferHandler`).
  - `service/`: Chứa core business logic (ví dụ `batch_service.go`, `monitor_service.go`, `report_service.go`).
