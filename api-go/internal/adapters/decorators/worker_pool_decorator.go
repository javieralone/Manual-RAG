package decorators

import (
	"context"
	"errors"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
)

var ErrServerBusy = errors.New("el servidor alcanzó el límite de peticiones simultáneas, intenta en unos momentos")

type WorkerPoolUseCaseDecorator struct {
	wrapped ports.QueryUseCase
	workers chan struct{}
}

func NewWorkerPoolUseCaseDecorator(useCase ports.QueryUseCase, maxWorkers int) *WorkerPoolUseCaseDecorator {
	return &WorkerPoolUseCaseDecorator{
		wrapped: useCase,
		workers: make(chan struct{}, maxWorkers),
	}
}

func (d *WorkerPoolUseCaseDecorator) ExecuteQuery(ctx context.Context, question string) (*domain.QueryResponse, error) {
	select {
	case d.workers <- struct{}{}:
		defer func() { <-d.workers }()
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Devuelve un error rápido de saturación sin tocar el UseCase ni la RAM
		return nil, ErrServerBusy
	}

	return d.wrapped.ExecuteQuery(ctx, question)
}