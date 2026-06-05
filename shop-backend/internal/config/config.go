package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Log      LogConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

// DatabaseConfig holds PostgreSQL settings.
// DB_TYPE selects the active storage backend: "postgres" (default) or "redis".
type DatabaseConfig struct {
	Type       string // postgres | redis
	Host       string
	Port       string
	Name       string
	User       string
	Password   string
	SSLMode    string
	LogQueries bool
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type LogConfig struct {
	Level string
}

type JWTConfig struct {
	Secret string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Type:       getEnv("DB_TYPE", "postgres"),
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnv("DB_PORT", "5432"),
			Name:       getEnv("DB_NAME", "shop"),
			User:       getEnv("DB_USER", "shop"),
			Password:   getEnv("DB_PASSWORD", "shop"),
			SSLMode:    getEnv("DB_SSLMODE", "disable"),
			LogQueries: getEnv("DB_LOG_QUERIES", "false") == "true",
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
