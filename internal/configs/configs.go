package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/subosito/gotenv"
)

type Config struct {
	AppConfig struct {
		Port      int
		JWTSecret string
	}
	PostgresConfig struct {
		Host     string
		Port     int
		DBName   string
		Username string
		Password string
	}
}

func InitConfig() (*Config, error) {
	_ = gotenv.Load()
	var cfg Config
	var err error

	if cfg.AppConfig.Port, err = getEnvInt("APP_PORT"); err != nil {
		return nil, err
	}

	if cfg.AppConfig.JWTSecret, err = getEnv("JWT_SECRET"); err != nil {
		return nil, err
	}

	if cfg.PostgresConfig.Host, err = getEnv("POSTGRES_HOST"); err != nil {
		return nil, err
	}

	if cfg.PostgresConfig.Port, err = getEnvInt("POSTGRES_PORT"); err != nil {
		return nil, err
	}

	if cfg.PostgresConfig.DBName, err = getEnv("POSTGRES_DBNAME"); err != nil {
		return nil, err
	}

	if cfg.PostgresConfig.Username, err = getEnv("POSTGRES_USERNAME"); err != nil {
		return nil, err
	}

	if cfg.PostgresConfig.Password, err = getEnv("POSTGRES_PASSWORD"); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getEnv(key string) (string, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" && key != "MYSQL_PASSWORD" {
		return "", fmt.Errorf("missing required env %s", key)
	}
	return val, nil
}

func getEnvInt(key string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, fmt.Errorf("missing required env %s", key)
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid int for %s: %v", key, err)
	}
	return v, nil
}
