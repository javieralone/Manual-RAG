package decorators

import (
	"context"
	"errors"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrServerBusy = errors.New("el servidor alcanzó el límite de peticiones simultáneas, intenta en unos momentos")

type WorkerPoolUseCaseDecorator struct {
	wrapped    ports.QueryUseCase
	workers    chan struct{}
	inFlight   prometheus.Gauge
	rejections prometheus.Counter
}

func NewWorkerPoolUseCaseDecorator(useCase ports.QueryUseCase, maxWorkers int, inFlight prometheus.Gauge, rejections prometheus.Counter) *WorkerPoolUseCaseDecorator {
	return &WorkerPoolUseCaseDecorator{
		wrapped:    useCase,
		workers:    make(chan struct{}, maxWorkers),
		inFlight:   inFlight,
		rejections: rejections,
	}
}

func (d *WorkerPoolUseCaseDecorator) ExecuteQuery(ctx context.Context, question string) (*domain.QueryResponse, error) {
	select {
	case d.workers <- struct{}{}:
		defer func() { <-d.workers }()
		if d.inFlight != nil {
			d.inFlight.Inc()
			defer d.inFlight.Dec()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if d.rejections != nil {
			d.rejections.Inc()
		}
		// Devuelve un error rápido de saturación sin tocar el UseCase ni la RAM
		return nil, ErrServerBusy
	}

	return d.wrapped.ExecuteQuery(ctx, question)
}
