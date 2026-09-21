package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"api-go/internal/core/domain"
)

type Config struct {
	PythonEngineURL   string
	OllamaURL         string
	OllamaModel       string
	HTTPPort          string
	HTTPClientTimeout time.Duration
	RequestTimeout    time.Duration
	ReadinessInterval time.Duration
	WorkerLimit       int
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   time.Duration
	OTLPEndpoint      string
	Auth              AuthConfig
}

type AuthConfig struct {
	JWTSecret         string
	RefreshSecret     string
	Issuer            string
	Audience          string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	AdminUsername     string
	AdminPasswordHash string
	AdminRoles        []domain.Role
}

func Load() (Config, error) {
	accessSecret, err := requiredEnv("AUTH_JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	refreshSecret, err := requiredEnv("AUTH_REFRESH_SECRET")
	if err != nil {
		return Config{}, err
	}
	adminUsername, err := requiredEnv("AUTH_ADMIN_USERNAME")
	if err != nil {
		return Config{}, err
	}
	adminPasswordHash, err := requiredEnv("AUTH_ADMIN_PASSWORD_HASH")
	if err != nil {
		return Config{}, err
	}

	accessTTL, err := durationEnv("AUTH_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := durationEnv("AUTH_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	workerLimit, err := intEnv("WORKER_LIMIT", 2)
	if err != nil || workerLimit < 1 {
		return Config{}, errors.New("WORKER_LIMIT debe ser un entero positivo")
	}
	httpClientTimeout, err := durationEnv("HTTP_CLIENT_TIMEOUT", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	requestTimeout, err := durationEnv("REQUEST_TIMEOUT", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}
	readinessInterval, err := durationEnv("READINESS_INTERVAL", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	rateLimitRequests, err := intEnv("RATE_LIMIT_REQUESTS", 60)
	if err != nil || rateLimitRequests < 1 {
		return Config{}, errors.New("RATE_LIMIT_REQUESTS debe ser un entero positivo")
	}
	rateLimitWindow, err := durationEnv("RATE_LIMIT_WINDOW", time.Minute)
	if err != nil {
		return Config{}, err
	}

	roles, err := parseRoles(os.Getenv("AUTH_ADMIN_ROLES"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		PythonEngineURL:   envOrDefault("PYTHON_ENGINE_URL", "http://rag-engine:8000"),
		OllamaURL:         envOrDefault("OLLAMA_URL", "http://host.docker.internal:11434"),
		OllamaModel:       envOrDefault("OLLAMA_MODEL", "qwen2.5:1.5b"),
		HTTPPort:          envOrDefault("HTTP_PORT", ":8080"),
		HTTPClientTimeout: httpClientTimeout,
		RequestTimeout:    requestTimeout,
		ReadinessInterval: readinessInterval,
		WorkerLimit:       workerLimit,
		RateLimitEnabled:  boolEnv("RATE_LIMIT_ENABLED", true),
		RateLimitRequests: rateLimitRequests,
		RateLimitWindow:   rateLimitWindow,
		OTLPEndpoint:      os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Auth: AuthConfig{
			JWTSecret:         accessSecret,
			RefreshSecret:     refreshSecret,
			Issuer:            envOrDefault("AUTH_ISSUER", "manual-rag-api"),
			Audience:          envOrDefault("AUTH_AUDIENCE", "manual-rag-client"),
			AccessTTL:         accessTTL,
			RefreshTTL:        refreshTTL,
			AdminUsername:     adminUsername,
			AdminPasswordHash: adminPasswordHash,
			AdminRoles:        roles,
		},
	}, nil
}

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("falta la variable de entorno requerida %s", name)
	}
	return value, nil
}

func envOrDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s debe ser una duración positiva, por ejemplo 15m", name)
	}
	return parsed, nil
}

func intEnv(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func boolEnv(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseRoles(raw string) ([]domain.Role, error) {
	if strings.TrimSpace(raw) == "" {
		return []domain.Role{domain.RoleAdmin}, nil
	}
	var roles []domain.Role
	for _, value := range strings.Split(raw, ",") {
		role := domain.Role(strings.TrimSpace(value))
		if role != domain.RoleAdmin && role != domain.RoleOperator && role != domain.RoleUser {
			return nil, fmt.Errorf("rol no soportado: %s", role)
		}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return nil, errors.New("AUTH_ADMIN_ROLES no puede estar vacío")
	}
	return roles, nil
}
