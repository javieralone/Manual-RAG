package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	authadapters "api-go/internal/adapters/auth"
	"api-go/internal/adapters/clients"
	configadapter "api-go/internal/adapters/config"
	"api-go/internal/adapters/decorators"
	adaptersHTTP "api-go/internal/adapters/http"
	"api-go/internal/adapters/http/handlers"
	"api-go/internal/adapters/http/middlewares"
	"api-go/internal/adapters/observability"
	"api-go/internal/application/use_cases"
	"api-go/internal/core/domain"
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
	resilience := clients.ResilienceConfig{Retries: config.RetryAttempts, Backoff: config.RetryBackoff, MaxFailures: config.CircuitFailures, ResetAfter: config.CircuitReset}
	ragAdapter := clients.NewPythonRAGClient(config.PythonEngineURL, httpClient, metrics, resilience)
	ollamaAdapter := clients.NewOllamaClient(config.OllamaURL, config.OllamaModel, httpClient, metrics, resilience)
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
	redisOptions, err := redis.ParseURL(config.SessionRedisURL)
	if err != nil {
		log.Fatalf("configuración Redis inválida: %v", err)
	}
	redisClient := redis.NewClient(redisOptions)
	refreshSessions := authadapters.NewRedisRefreshSessionStore(redisClient)
	defer refreshSessions.Close()
	userRepository := authadapters.NewInMemoryUserRepository(domain.User{
		Username:     config.Auth.AdminUsername,
		PasswordHash: config.Auth.AdminPasswordHash,
		Roles:        config.Auth.AdminRoles,
	})
	authService := use_cases.NewAuthService(userRepository, authadapters.NewBcryptPasswordHasher(), tokenService, refreshSessions)

	// 2. Caso de Uso Core
	rawUseCase := use_cases.NewQueryOrchestrator(ragAdapter, ollamaAdapter)

	// 3. Decorador Worker Pool (Máximo 2 peticiones concurrentes a la vez)
	useCaseWithWorkerPool := decorators.NewWorkerPoolUseCaseDecorator(rawUseCase, config.WorkerLimit, metrics.WorkerInFlight, metrics.WorkerRejections)

	// 4. Handler
	queryHandler := handlers.NewQueryHandler(useCaseWithWorkerPool)
	authHandler := handlers.NewAuthHandler(authService, logger, metrics, handlers.RefreshCookieConfig{Secure: config.Auth.CookieSecure, MaxAge: int(config.Auth.RefreshTTL.Seconds())})
	healthHandler := handlers.NewHealthHandler(metrics,
		clients.NewURLHealthChecker(strings.TrimRight(config.PythonEngineURL, "/")+"/ready", httpClient),
		clients.NewURLHealthChecker(strings.TrimRight(config.OllamaURL, "/")+"/api/tags", httpClient),
		refreshSessions,
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
	handlerWithMiddleware := middlewares.CORSMiddleware()(middlewares.TraceMiddleware(middlewares.MetricsMiddleware(metrics)(middlewares.TimeoutMiddleware(config.RequestTimeout)(middlewares.BodyLimitMiddleware(config.MaxRequestBodyBytes)(router)))))

	server := &http.Server{Addr: config.HTTPPort, Handler: handlerWithMiddleware, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: config.RequestTimeout, WriteTimeout: config.RequestTimeout, IdleTimeout: 60 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	logger.Info("api_gateway_started", "port", config.HTTPPort, "worker_limit", config.WorkerLimit, "readiness_interval", config.ReadinessInterval.String())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error iniciando servidor: %v", err)
		}
	case <-signals:
		shutdownContext, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("shutdown incompleto: %v", err)
		}
		_ = shutdownTracer(shutdownContext)
	}
}
