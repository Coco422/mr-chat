package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	CORS     CORSConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
}

type HTTPConfig struct {
	Host            string
	Port            string
	ReadTimeout     Duration
	WriteTimeout    Duration
	ShutdownTimeout Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

type PostgresConfig struct {
	DSNOverride     string
	Enabled         bool
	AutoMigrate     bool
	MigrationsDir   string
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime Duration
}

type RedisConfig struct {
	Enabled      bool
	Host         string
	Port         int
	Password     string
	DB           int
	DialTimeout  Duration
	ReadTimeout  Duration
	WriteTimeout Duration
}

type AuthConfig struct {
	JWTSecret           string
	JWTIssuer           string
	AccessTTL           Duration
	RefreshTTL          Duration
	RefreshCookieName   string
	RefreshCookieDomain string
	RefreshCookieSecure bool
}

type Duration struct {
	value time.Duration
}

func (d Duration) Duration() time.Duration {
	return d.value
}

func (h HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%s", h.Host, h.Port)
}

func (c PostgresConfig) DSN() string {
	if c.DSNOverride != "" {
		return c.DSNOverride
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
		c.SSLMode,
	)
}

func (c RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func Load() (Config, error) {
	cfg := Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "mrchat-api"),
			Version:     getEnv("APP_VERSION", "dev"),
			Environment: getEnv("APP_ENV", "local"),
		},
		HTTP: HTTPConfig{
			Host:            getEnv("HTTP_HOST", "0.0.0.0"),
			Port:            getEnv("HTTP_PORT", "8080"),
			ReadTimeout:     mustDuration("HTTP_READ_TIMEOUT", "10s"),
			WriteTimeout:    mustDuration("HTTP_WRITE_TIMEOUT", "30s"),
			ShutdownTimeout: mustDuration("HTTP_SHUTDOWN_TIMEOUT", "15s"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173")),
			AllowCredentials: mustBool("CORS_ALLOW_CREDENTIALS", true),
		},
		Postgres: PostgresConfig{
			DSNOverride:     getEnv("POSTGRES_DSN", ""),
			Enabled:         mustBool("POSTGRES_ENABLED", true),
			AutoMigrate:     mustBool("POSTGRES_AUTO_MIGRATE", true),
			MigrationsDir:   getEnv("POSTGRES_MIGRATIONS_DIR", "db/migrations"),
			Host:            getEnv("POSTGRES_HOST", "127.0.0.1"),
			Port:            mustInt("POSTGRES_PORT", 5432),
			User:            getEnv("POSTGRES_USER", "mrchat"),
			Password:        getEnv("POSTGRES_PASSWORD", "mrchat"),
			Database:        getEnv("POSTGRES_DB", "mrchat"),
			SSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
			MaxOpenConns:    mustInt("POSTGRES_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    mustInt("POSTGRES_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: mustDuration("POSTGRES_CONN_MAX_LIFETIME", "30m"),
		},
		Redis: RedisConfig{
			Enabled:      mustBool("REDIS_ENABLED", true),
			Host:         getEnv("REDIS_HOST", "127.0.0.1"),
			Port:         mustInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           mustInt("REDIS_DB", 0),
			DialTimeout:  mustDuration("REDIS_DIAL_TIMEOUT", "3s"),
			ReadTimeout:  mustDuration("REDIS_READ_TIMEOUT", "2s"),
			WriteTimeout: mustDuration("REDIS_WRITE_TIMEOUT", "2s"),
		},
		Auth: AuthConfig{
			JWTSecret:           getEnv("JWT_SECRET", "change-me"),
			JWTIssuer:           getEnv("JWT_ISSUER", getEnv("APP_NAME", "mrchat-api")),
			AccessTTL:           mustDuration("JWT_ACCESS_TTL", "1h"),
			RefreshTTL:          mustDuration("JWT_REFRESH_TTL", "168h"),
			RefreshCookieName:   getEnv("AUTH_REFRESH_COOKIE_NAME", "mrchat_refresh_token"),
			RefreshCookieDomain: getEnv("AUTH_REFRESH_COOKIE_DOMAIN", ""),
			RefreshCookieSecure: mustBool("AUTH_REFRESH_COOKIE_SECURE", false),
		},
	}

	if cfg.Auth.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET cannot be empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func mustBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func mustInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func mustDuration(key, fallback string) Duration {
	value := getEnv(key, fallback)

	parsed, err := time.ParseDuration(value)
	if err != nil {
		parsed, _ = time.ParseDuration(fallback)
	}

	return Duration{value: parsed}
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	parts := make([]string, 0)
	current := ""
	for _, r := range value {
		if r == ',' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			continue
		}
		current += string(r)
	}
	if current != "" {
		parts = append(parts, current)
	}

	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := trimSpaces(part); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return cleaned
}

func trimSpaces(value string) string {
	start := 0
	end := len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t' || value[start] == '\n' || value[start] == '\r') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t' || value[end-1] == '\n' || value[end-1] == '\r') {
		end--
	}
	return value[start:end]
}
