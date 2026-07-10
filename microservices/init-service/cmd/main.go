package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/LeHuuHai/server-management/microservices/init-service/config"
	internaldb "github.com/LeHuuHai/server-management/microservices/init-service/internal/db"
	internales "github.com/LeHuuHai/server-management/microservices/init-service/internal/es"
	internalmq "github.com/LeHuuHai/server-management/microservices/init-service/internal/mq"
	"github.com/LeHuuHai/server-management/microservices/pkg/es"
)

//go:embed migrations/*/*.sql
var migrationsFS embed.FS

func main() {
	slog.Info("Starting init-service...")
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 1. Init Kafka
	if err := internalmq.InitKafkaTopics(cfg.KafkaBroker); err != nil {
		slog.Error("Failed to init kafka topics", "err", err)
	} else {
		slog.Info("Kafka topics initialized successfully")
	}

	// 2. Init Elasticsearch
	esClient, err := es.Connect(cfg.ESConfig)
	if err != nil {
		slog.Error("Failed to connect to elasticsearch", "err", err)
	} else {
		if err := internales.InitHeartbeatIndex(esClient); err != nil {
			slog.Error("Failed to init elasticsearch heartbeat index", "err", err)
		} else {
			slog.Info("Elasticsearch index initialized successfully")
		}
	}

	// 3. Init Auth DB
	if err := internaldb.RunMigrations(cfg.AuthDBConfig, migrationsFS, "migrations/auth"); err != nil {
		slog.Error("Failed to run Auth DB migrations", "error", err)
	} else {
		slog.Info("Auth DB migrations ran successfully")
	}

	// 4. Init Server DB
	if err := internaldb.RunMigrations(cfg.ServerDBConfig, migrationsFS, "migrations/server"); err != nil {
		slog.Error("Failed to run Server DB migrations", "error", err)
	} else {
		slog.Info("Server DB migrations ran successfully")
	}

	// 5. Init Monitor DB
	if err := internaldb.RunMigrations(cfg.MonitorDBConfig, migrationsFS, "migrations/monitor"); err != nil {
		slog.Error("Failed to run Monitor DB migrations", "error", err)
	} else {
		slog.Info("Monitor DB migrations ran successfully")
	}

	slog.Info("init-service completed successfully.")
}
