package config

import (
	"os"
	"strconv"
)

// Config drži sve konfiguracione vrednosti aplikacije.
// U Spring-u bi ovo bio @ConfigurationProperties klasa.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host       string
	Port       string
	Name       string
	User       string
	Password   string
	SSLMode    string
	LogQueries bool
}

type LogConfig struct {
	Level string
}

// Load čita konfiguraciju iz environment varijabli (.env fajl ili K8s ConfigMap/Secret).
// U Spring-u bi ovo bio application.properties / application.yml.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnv("DB_PORT", "5432"),
			Name:       getEnv("DB_NAME", "shop"),
			User:       getEnv("DB_USER", "shop"),
			Password:   getEnv("DB_PASSWORD", "shop"),
			SSLMode:    getEnv("DB_SSLMODE", "disable"),
			LogQueries: getEnv("DB_LOG_QUERIES", "false") == "true",
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
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
