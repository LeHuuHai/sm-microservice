module github.com/LeHuuHai/server-management/microservices/init-service

go 1.25.3

replace github.com/LeHuuHai/server-management/microservices/pkg => ../pkg

require (
	github.com/LeHuuHai/server-management/microservices/pkg v0.0.0-00010101000000-000000000000
	github.com/elastic/go-elasticsearch/v8 v8.19.6
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/segmentio/kafka-go v0.4.51
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.16 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
)
