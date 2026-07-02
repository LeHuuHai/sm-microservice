# Chi tiết các Microservices (Microservice Profiles)

Dưới đây là tóm tắt nhanh về các Microservice trong hệ thống. Để xem thông tin kỹ thuật chuyên sâu về từng dịch vụ, vui lòng nhấp vào các liên kết chi tiết tương ứng.

---

## 1. `auth-service`
*Quản lý tài khoản, phân quyền và cấp phát/thu hồi JWT.*
- **Giao tiếp:** HTTP REST
- **Database:** PostgreSQL, Redis
- **Chi tiết:** [Đọc thêm tại auth-service.md](auth-service.md)

---

## 2. `server-service`
*Quản lý danh mục máy chủ (Inventory Catalog). Nguồn lưu trữ cấu hình tĩnh duy nhất.*
- **Giao tiếp:** HTTP REST, phát hành Kafka `server_lifecycle`
- **Database:** PostgreSQL (Có dùng Transactional Outbox Pattern)
- **Chi tiết:** [Đọc thêm tại server-service.md](server-service.md)

---

## 3. `heartbeat-gateway`
*Cửa ngõ tiếp nhận tín hiệu sống (heartbeat) cực nhanh từ các máy chủ.*
- **Giao tiếp:** HTTP REST (nhận payload), phát hành Kafka `heartbeats`
- **Database:** Stateless (Không DB)
- **Chi tiết:** [Đọc thêm tại heartbeat-gateway.md](heartbeat-gateway.md)

---

## 4. `ping-worker`
*Daemon ICMP ping chủ động kiểm tra trạng thái.*
- **Giao tiếp:** Kafka (lắng nghe `ping`, phát hành `ping_res`)
- **Database:** Stateless
- **Chi tiết:** [Đọc thêm tại ping-worker.md](ping-worker.md)

---

## 5. `monitor-service`
*Trái tim của hệ thống giám sát. Xử lý timeout, logs, và báo cáo.*
- **Giao tiếp:** HTTP REST, gRPC (phục vụ mail-worker), Kafka (đa dạng topics)
- **Database:** PostgreSQL, Elasticsearch, Redis
- **Chi tiết:** [Đọc thêm tại monitor-service.md](monitor-service.md)

---

## 6. `mail-worker`
*Daemon tải báo cáo và gửi email SMTP.*
- **Giao tiếp:** Kafka (nhận `mail`), gRPC Client (tải file)
- **Database:** Stateless
- **Chi tiết:** [Đọc thêm tại mail-worker.md](mail-worker.md)

---

## 7. `agent`
*Chương trình daemon chạy trên host server.*
- **Giao tiếp:** HTTP POST đến heartbeat-gateway
- **Chi tiết:** [Đọc thêm tại agent.md](agent.md)
