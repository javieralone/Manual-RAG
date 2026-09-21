package main

import (
	"log"
	"net/http"

	authadapters "api-go/internal/adapters/auth"
	"api-go/internal/adapters/clients"
	configadapter "api-go/internal/adapters/config"
	"api-go/internal/adapters/decorators"
	adaptersHTTP "api-go/internal/adapters/http"
	"api-go/internal/adapters/http/handlers"
	"api-go/internal/adapters/http/middlewares"
	"api-go/internal/core/domain"
	"api-go/internal/core/services"
)

func main() {
	config, err := configadapter.Load()
	if err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}

	httpClient := &http.Client{Timeout: config.HTTPClientTimeout}

	// 1. Adaptadores Secundarios
	ragAdapter := clients.NewPythonRAGClient(config.PythonEngineURL, httpClient)
	ollamaAdapter := clients.NewOllamaClient(config.OllamaURL, config.OllamaModel, httpClient)
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
	useCaseWithWorkerPool := decorators.NewWorkerPoolUseCaseDecorator(rawUseCase, config.WorkerLimit)

	// 4. Handler
	queryHandler := handlers.NewQueryHandler(useCaseWithWorkerPool)
	authHandler := handlers.NewAuthHandler(authService, log.Default())

	// 5. Router con Middleware de Timeout (Límite global de 60s por Request)
	authenticate := middlewares.AuthenticationMiddleware(tokenService)
	authorize := middlewares.RequireAnyRole(domain.RoleAdmin, domain.RoleOperator, domain.RoleUser)
	router := adaptersHTTP.NewRouter(queryHandler, authHandler, authenticate, authorize)
	handlerWithMiddleware := middlewares.TimeoutMiddleware(config.RequestTimeout)(router)

	log.Printf("API Gateway corriendo en %s con autenticación JWT y WorkerPool...", config.HTTPPort)
	if err := http.ListenAndServe(config.HTTPPort, handlerWithMiddleware); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
