package crawler

import (
	"github.com/mguley/web-scraper-v1/internal/processor"
	"github.com/mguley/web-scraper-v1/internal/tor"
)

const (
	batchLimit = 25 // Maximum number of units to process in one batch.
)

// Dispatcher orchestrates the distribution of work units to workers.
// It manages the lifecycle and operations of workers through the WorkerManager.
//
// Type parameter:
//   - T any: The type of data that the Dispatcher processes. This ensures type safety throughout the processing pipeline.
//
// Fields:
// - WorkerManager *WorkerManager[T]: Manages workers that process units of work.
// - Processor processor.Processor[T]: Processes units of work.
// - TorFacade *tor.Facade: Interface for interacting with Tor network features.
type Dispatcher[T any] struct {
	WorkerManager *WorkerManager[T]
	Processor     processor.Processor[T]
	TorFacade     *tor.Facade
}

// NewDispatcher creates a new Dispatcher instance with a specified number of workers, processor, and Tor facade.
// It initializes a WorkerManager to handle worker-related operations.
//
// Parameters:
// - maxWorkers int: The maximum number of workers the dispatcher can manage.
// - processor processor.Processor[T]: The processor used for handling units of work.
// - torFacade *tor.Facade: Facade for interacting with the Tor network.
//
// Returns:
// - *Dispatcher[T]: A pointer to an instance of Dispatcher.
func NewDispatcher[T any](maxWorkers int, processor processor.Processor[T], torFacade *tor.Facade) *Dispatcher[T] {
	workerManager := NewWorkerManager[T](maxWorkers, processor)
	return &Dispatcher[T]{
		WorkerManager: workerManager,
		Processor:     processor,
		TorFacade:     torFacade,
	}
}

// Run starts the dispatcher's operation, which includes starting the worker manager and the dispatch loop.
func (dispatcher *Dispatcher[T]) Run() {
	dispatcher.WorkerManager.Start()
	go dispatcher.dispatch()
}

// dispatch manages the distribution of units to workers and controls the batch processing logic.
func (dispatcher *Dispatcher[T]) dispatch() {
	for {
		batchCount := 0
		for batchCount < batchLimit {
			unit, more := <-dispatcher.WorkerManager.UnitQueue
			if !more {
				return // Exit dispatch loop if no more units are available.
			}
			dispatcher.WorkerManager.AssignUnit(unit)
			batchCount++
		}

		dispatcher.WorkerManager.WaitForBatchCompletion()
	}
}

// Stop halts the dispatcher's operations, including all worker processes managed by the WorkerManager.
func (dispatcher *Dispatcher[T]) Stop() {
	dispatcher.WorkerManager.Stop()
}
