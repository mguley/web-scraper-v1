package crawler

import (
	"context"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"github.com/mguley/web-scraper-v1/internal/tor"
	"log"
)

// DispatcherConfig holds configuration settings for the Dispatcher.
//
// Fields:
// - MaxWorkers int: Maximum number of concurrent workers.
// - BatchLimit int: Maximum number of units to process in one batch
type DispatcherConfig struct {
	MaxWorkers int
	BatchLimit int
}

// Dispatcher orchestrates the distribution of work units to workers.
// It manages the lifecycle and operations of workers through the WorkerManager and uses configuration settings
// provided via DispatcherConfig.
//
// Type parameter:
//   - T any: The type of data that the Dispatcher processes. This ensures type safety throughout the processing pipeline.
//
// Fields:
// - WorkerManager *WorkerManager[T]: Manages workers that process units of work.
// - Processor processor.Processor[T]: Processes units of work.
// - TorFacade *tor.Facade: Interface for interacting with Tor network features.
// - config DispatcherConfig: Holds configuration settings for the Dispatcher.
// - ctx context.Context: Context for managing cancellation.
// - cancelFunc context.CancelFunc: Function to call to initiate cancellation.
type Dispatcher[T any] struct {
	WorkerManager *WorkerManager[T]
	Processor     processor.Processor[T]
	TorFacade     *tor.Facade
	config        DispatcherConfig
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

// NewDispatcher creates a new Dispatcher instance with specified configuration settings, a processor, and a Tor facade.
// It initializes a WorkerManager to handle worker-related operations based on the provided configuration.
//
// Parameters:
// - ctx context.Context: Parent context for the dispatcher.
// - config DispatcherConfig: Configuration settings for the Dispatcher, including the number of workers and batch limit.
// - processor processor.Processor[T]: The processor used for handling units of work.
// - torFacade *tor.Facade: Facade for interacting with the Tor network.
//
// Returns:
// - *Dispatcher[T]: A pointer to an instance of Dispatcher.
func NewDispatcher[T any](ctx context.Context, config DispatcherConfig, processor processor.Processor[T],
	torFacade *tor.Facade) *Dispatcher[T] {

	ctx, cancel := context.WithCancel(ctx)
	workerManager := NewWorkerManager[T](ctx, config, processor)
	return &Dispatcher[T]{
		WorkerManager: workerManager,
		Processor:     processor,
		TorFacade:     torFacade,
		config:        config,
		ctx:           ctx,
		cancelFunc:    cancel,
	}
}

// Run starts the dispatcher's operation, which includes starting the worker manager and the dispatch loop.
func (dispatcher *Dispatcher[T]) Run() {
	dispatcher.WorkerManager.Start()
	go dispatcher.dispatch()
}

// dispatch manages the distribution of units to workers and controls the batch processing logic based on the configuration settings.
func (dispatcher *Dispatcher[T]) dispatch() {
	for {
		select {
		case <-dispatcher.ctx.Done():
			return // Exit dispatch loop if context is cancelled
		default:
			batchCount := 0
			for batchCount < dispatcher.config.BatchLimit {
				unit, more := <-dispatcher.WorkerManager.UnitQueue
				if !more {
					return // Exit dispatch loop if no more units are available.
				}
				log.Printf("Dispatching unit: %s", unit.URL)
				err := dispatcher.WorkerManager.AssignUnit(unit)
				if err != nil {
					log.Printf("Failed to assign unit: %v", err)
					continue // Skip to the next unit if there was an error assigning the current one
				}
				batchCount++
			}
			dispatcher.WorkerManager.WaitForBatchCompletion()
		}
	}
}

// Stop halts the dispatcher's operations, including all worker processes managed by the WorkerManager.
func (dispatcher *Dispatcher[T]) Stop() {
	dispatcher.cancelFunc() // Trigger the cancellation
	dispatcher.WorkerManager.Stop()
}

// GetContext returns the context associated with the dispatcher.
func (dispatcher *Dispatcher[T]) GetContext() context.Context {
	return dispatcher.ctx
}
