package config

import (
	"os"
	"strconv"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type Config struct {
	AuthDBConfig    *pkgconfig.PostgresConfig
	ServerDBConfig  *pkgconfig.PostgresConfig
	MonitorDBConfig *pkgconfig.PostgresConfig
	ESConfig        *pkgconfig.ElasticsearchConfig
	KafkaBroker     string
}

func Load() (*Config, error) {
	// auth db
	authPgPort, _ := strconv.Atoi(os.Getenv("AUTH_DB_PORT"))
	authDBConfig := &pkgconfig.PostgresConfig{
		Host:     os.Getenv("AUTH_DB_HOST"),
		Username: os.Getenv("AUTH_DB_USER"),
		Password: pkgconfig.ReadSecret("auth_db_password"),
		Port:     authPgPort,
		Database: os.Getenv("AUTH_DB_DBNAME"),
	}

	// server db
	serverPgPort, _ := strconv.Atoi(os.Getenv("SERVER_DB_PORT"))
	serverDBConfig := &pkgconfig.PostgresConfig{
		Host:     os.Getenv("SERVER_DB_HOST"),
		Username: os.Getenv("SERVER_DB_USER"),
		Password: pkgconfig.ReadSecret("server_db_password"),
		Port:     serverPgPort,
		Database: os.Getenv("SERVER_DB_DBNAME"),
	}

	// monitor db
	monitorPgPort, _ := strconv.Atoi(os.Getenv("MONITOR_DB_PORT"))
	monitorDBConfig := &pkgconfig.PostgresConfig{
		Host:     os.Getenv("MONITOR_DB_HOST"),
		Username: os.Getenv("MONITOR_DB_USER"),
		Password: pkgconfig.ReadSecret("monitor_db_password"),
		Port:     monitorPgPort,
		Database: os.Getenv("MONITOR_DB_DBNAME"),
	}

	// es
	esConfig := &pkgconfig.ElasticsearchConfig{
		URL:   os.Getenv("ES_URL"),
		Index: os.Getenv("ES_INDEX"),
	}

	cfg := Config{
		AuthDBConfig:    authDBConfig,
		ServerDBConfig:  serverDBConfig,
		MonitorDBConfig: monitorDBConfig,
		ESConfig:        esConfig,
		KafkaBroker:     os.Getenv("KAFKA_BROKER"),
	}

	return &cfg, nil
}
