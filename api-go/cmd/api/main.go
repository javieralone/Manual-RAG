package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	authadapters "api-go/internal/adapters/auth"
	"api-go/internal/adapters/clients"
	configadapter "api-go/internal/adapters/config"
	"api-go/internal/adapters/decorators"
	adaptersHTTP "api-go/internal/adapters/http"
	"api-go/internal/adapters/http/handlers"
	"api-go/internal/adapters/http/middlewares"
	"api-go/internal/adapters/observability"
	"api-go/internal/core/domain"
	"api-go/internal/core/services"
)

func main() {
	config, err := configadapter.Load()
	if err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}
	shutdownTracer, err := observability.InitTracer(context.Background(), config.OTLPEndpoint)
	if err != nil {
		log.Printf("tracing OTLP deshabilitado: %v", err)
	} else {
		defer shutdownTracer(context.Background())
	}

	httpClient := &http.Client{Timeout: config.HTTPClientTimeout}
	metrics := observability.NewMetrics(prometheus.DefaultRegisterer)
	logger := observability.NewLogger()

	// 1. Adaptadores Secundarios
	ragAdapter := clients.NewPythonRAGClient(config.PythonEngineURL, httpClient, metrics)
	ollamaAdapter := clients.NewOllamaClient(config.OllamaURL, config.OllamaModel, httpClient, metrics)
	tokenService, err := authadapters.NewJWTService(
		config.Auth.JWTSecret,
		config.Auth.RefreshSecret,
		config.Auth.Issuer,
		config.Auth.Audience,
		config.Auth.AccessTTL,
		config.Auth.RefreshTTL,
	)
	if err != nil {
		log.Fatalf("configuración JWT inválida: %v", err)
	}
	userRepository := authadapters.NewInMemoryUserRepository(domain.User{
		Username:     config.Auth.AdminUsername,
		PasswordHash: config.Auth.AdminPasswordHash,
		Roles:        config.Auth.AdminRoles,
	})
	authService := services.NewAuthService(userRepository, authadapters.NewBcryptPasswordHasher(), tokenService)

	// 2. Caso de Uso Core
	rawUseCase := services.NewQueryOrchestrator(ragAdapter, ollamaAdapter)

	// 3. Decorador Worker Pool (Máximo 2 peticiones concurrentes a la vez)
	useCaseWithWorkerPool := decorators.NewWorkerPoolUseCaseDecorator(rawUseCase, config.WorkerLimit, metrics.WorkerInFlight, metrics.WorkerRejections)

	// 4. Handler
	queryHandler := handlers.NewQueryHandler(useCaseWithWorkerPool)
	authHandler := handlers.NewAuthHandler(authService, logger, metrics)
	healthHandler := handlers.NewHealthHandler(metrics,
		clients.NewURLHealthChecker(strings.TrimRight(config.PythonEngineURL, "/")+"/health", httpClient),
		clients.NewURLHealthChecker(strings.TrimRight(config.OllamaURL, "/")+"/api/tags", httpClient),
	)
	readinessContext, cancelReadiness := context.WithCancel(context.Background())
	defer cancelReadiness()
	healthHandler.StartReadinessMonitor(readinessContext, config.ReadinessInterval)

	// 5. Router con Middleware de Timeout
	authenticate := middlewares.AuthenticationMiddleware(tokenService)
	authorize := middlewares.RequireAnyRole(domain.RoleAdmin, domain.RoleOperator, domain.RoleUser)
	rateLimit := func(next http.Handler) http.Handler { return next }
	if config.RateLimitEnabled {
		rateLimit = middlewares.RateLimitMiddleware(config.RateLimitRequests, config.RateLimitWindow, metrics, logger)
	}
	router := adaptersHTTP.NewRouter(queryHandler, authHandler, healthHandler, authenticate, authorize, rateLimit)
	handlerWithMiddleware := middlewares.TraceMiddleware(middlewares.MetricsMiddleware(metrics)(middlewares.TimeoutMiddleware(config.RequestTimeout)(router)))

	logger.Info("api_gateway_started", "port", config.HTTPPort, "worker_limit", config.WorkerLimit, "readiness_interval", config.ReadinessInterval.String())
	if err := http.ListenAndServe(config.HTTPPort, handlerWithMiddleware); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
