package decorators

import (
	"context"
	"errors"

	"api-go/internal/core/domain"
	"api-go/internal/core/ports"
	"github.com/prometheus/client_golang/prometheus"
)

var ErrServerBusy = errors.New("el servidor alcanzó el límite de peticiones simultáneas, intenta en unos momentos")

// QueryService agrupa los casos de uso normal y en streaming para compartir un único pool de workers.
type QueryService interface {
	ports.QueryUseCase
	ports.QueryStreamUseCase
}

type WorkerPoolUseCaseDecorator struct {
	wrapped    QueryService
	workers    chan struct{}
	inFlight   prometheus.Gauge
	rejections prometheus.Counter
}

func NewWorkerPoolUseCaseDecorator(useCase QueryService, maxWorkers int, inFlight prometheus.Gauge, rejections prometheus.Counter) *WorkerPoolUseCaseDecorator {
	return &WorkerPoolUseCaseDecorator{
		wrapped:    useCase,
		workers:    make(chan struct{}, maxWorkers),
		inFlight:   inFlight,
		rejections: rejections,
	}
}

func (d *WorkerPoolUseCaseDecorator) ExecuteQuery(ctx context.Context, question string) (*domain.QueryResponse, error) {
	return d.executeQuery(ctx, question, domain.QueryFilters{})
}

func (d *WorkerPoolUseCaseDecorator) ExecuteQueryWithFilters(ctx context.Context, question string, filters domain.QueryFilters) (*domain.QueryResponse, error) {
	return d.executeQuery(ctx, question, filters)
}

func (d *WorkerPoolUseCaseDecorator) executeQuery(ctx context.Context, question string, filters domain.QueryFilters) (*domain.QueryResponse, error) {
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

	if filtered, ok := d.wrapped.(ports.FilteredQueryUseCase); ok {
		return filtered.ExecuteQueryWithFilters(ctx, question, filters)
	}
	return d.wrapped.ExecuteQuery(ctx, question)
}

// ExecuteQueryStream reutiliza el mismo pool de workers; el slot se mantiene ocupado durante todo
// el streaming porque la llamada es síncrona hasta que el evento "complete"/"error" se emite.
func (d *WorkerPoolUseCaseDecorator) ExecuteQueryStream(ctx context.Context, question string, sink ports.StreamSink) error {
	return d.executeQueryStream(ctx, question, domain.QueryFilters{}, sink)
}

func (d *WorkerPoolUseCaseDecorator) ExecuteQueryStreamWithFilters(ctx context.Context, question string, filters domain.QueryFilters, sink ports.StreamSink) error {
	return d.executeQueryStream(ctx, question, filters, sink)
}

func (d *WorkerPoolUseCaseDecorator) executeQueryStream(ctx context.Context, question string, filters domain.QueryFilters, sink ports.StreamSink) error {
	select {
	case d.workers <- struct{}{}:
		defer func() { <-d.workers }()
		if d.inFlight != nil {
			d.inFlight.Inc()
			defer d.inFlight.Dec()
		}
	case <-ctx.Done():
		return ctx.Err()
	default:
		if d.rejections != nil {
			d.rejections.Inc()
		}
		return sink.SendError(ErrServerBusy)
	}

	if filtered, ok := d.wrapped.(ports.FilteredQueryStreamUseCase); ok {
		return filtered.ExecuteQueryStreamWithFilters(ctx, question, filters, sink)
	}
	return d.wrapped.ExecuteQueryStream(ctx, question, sink)
}
