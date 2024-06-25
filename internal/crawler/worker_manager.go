package crawler

import (
	"context"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"log"
	"sync"
	"time"
)

// Deprecated: WorkerManager is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// WorkerManager manages the lifecycle and operations of a pool of workers that process units of work.
// It utilizes a context to handle graceful shutdowns and manages synchronization of task completion across workers.
//
// Type parameter:
//   - T any: The type of data that the workers process. This ensures type safety across the worker processes.
//
// Fields:
// - ctx context.Context: Context for managing cancellation and shutdown.
// - WorkerQueue chan chan Unit: Channel of worker channels, used to distribute available workers.
// - config DispatcherConfig: Configuration settings for managing worker operations.
// - UnitQueue chan Unit: Channel for receiving units of work to be processed.
// - BatchDone chan bool: Channel for signaling the completion of processing a batch of units.
// - batchWaitGroup sync.WaitGroup: Synchronizes the completion of processing batches.
// - workers []*Worker[T]: List of worker instances managed by this manager.
// - Processor processor.Processor[T]: Processor used for handling units of work.
type WorkerManager[T any] struct {
	ctx            context.Context
	WorkerQueue    chan chan Unit
	config         DispatcherConfig
	UnitQueue      chan Unit
	BatchDone      chan bool
	batchWaitGroup sync.WaitGroup
	workers        []*Worker[T]
	Processor      processor.Processor[T]
}

// Deprecated: NewWorkerManager is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// NewWorkerManager creates a new WorkerManager instance with a specified configuration, a processor, and a context.
// It sets up the necessary channels and initializes the list of workers based on the configuration.
// The context is used to manage the lifecycle of workers, especially for graceful shutdown during cancellation.
//
// Parameters:
// - ctx context.Context: The context that controls cancellation and lifecycle.
// - config DispatcherConfig: Configuration settings including the maximum number of workers and batch limit.
// - processor processor.Processor[T]: The processor used for handling units of work.
//
// Returns:
// - *WorkerManager[T]: A pointer to an instance of WorkerManager.
func NewWorkerManager[T any](ctx context.Context, config DispatcherConfig,
	processor processor.Processor[T]) *WorkerManager[T] {
	return &WorkerManager[T]{
		ctx:         ctx,
		WorkerQueue: make(chan chan Unit, config.MaxWorkers),
		config:      config,
		UnitQueue:   make(chan Unit),
		BatchDone:   make(chan bool, config.MaxWorkers),
		workers:     make([]*Worker[T], 0, config.MaxWorkers),
		Processor:   processor,
	}
}

// Deprecated: Start is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// Start initializes and starts all worker instances under management.
// It monitors the context for cancellation to halt the startup process if necessary.
func (workerManager *WorkerManager[T]) Start() {
	for i := 0; i < workerManager.config.MaxWorkers; i++ {
		select {
		case <-workerManager.ctx.Done():
			return // Exit if the context is cancelled
		default:
			worker := NewWorker[T](workerManager.ctx, i, workerManager.WorkerQueue, workerManager.Processor, workerManager.BatchDone)
			worker.Start()
			workerManager.workers = append(workerManager.workers, worker)
		}
	}
}

// Deprecated: AssignUnit is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// AssignUnit sends a unit of work to an available worker for processing.
// It ensures that a worker is available before assigning the unit, and uses a timeout to avoid blocking indefinitely.
// Errors are returned if the worker queue is closed or if a timeout occurs.
//
// Parameters:
// - unit Unit: The unit of work to be processed.
//
// Returns:
// - error: An error if no available worker is found within the timeout period, or if the unit cannot be assigned.
func (workerManager *WorkerManager[T]) AssignUnit(unit Unit) error {
	select {
	case <-workerManager.ctx.Done():
		return fmt.Errorf("context cancelled, cannot assign unit: %s", unit.URL)
	case workerChannel, ok := <-workerManager.WorkerQueue:
		if !ok {
			return fmt.Errorf("worker queue closed, cannot assign unit: %s", unit.URL)
		}
		select {
		case workerChannel <- unit:
			log.Printf("Assigned unit to worker: %s", unit.URL)
		case <-time.After(time.Second * 3):
			return fmt.Errorf("timeout while assigning unit to worker: %s", unit.URL)
		}
	case <-time.After(time.Second * 3):
		return fmt.Errorf("no available workers to assign unit: %s", unit.URL)
	}
	return nil
}

// Deprecated: GetWorkers is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// GetWorkers returns a copy of the slice of workers managed by this WorkerManager.
// This method ensures that any modifications made to the returned slice do not affect the original workers slice.
//
// Returns:
// - []*Worker[T]: A copy of the slice of workers managed by this WorkerManager.
func (workerManager *WorkerManager[T]) GetWorkers() []*Worker[T] {
	workersCopy := make([]*Worker[T], len(workerManager.workers))
	copy(workersCopy, workerManager.workers)
	return workersCopy
}

// Deprecated: WaitForBatchCompletion is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// WaitForBatchCompletion waits for the completion of processing for all units in the current batch.
// This method uses the batchWaitGroup to synchronize the completion based on the batch limit defined in the configuration.
func (workerManager *WorkerManager[T]) WaitForBatchCompletion() {
	workerManager.batchWaitGroup.Add(workerManager.config.BatchLimit)
	for i := 0; i < workerManager.config.BatchLimit; i++ {
		<-workerManager.BatchDone
		workerManager.batchWaitGroup.Done()
	}
	workerManager.batchWaitGroup.Wait()
}

// Deprecated: Stop is deprecated and will be removed in a future release.
// Use the new task queue and worker pool mechanism instead.
// Stop terminates all workers managed by this manager and cleans up resources.
func (workerManager *WorkerManager[T]) Stop() {
	// Wait for all workers to finish their tasks
	workerManager.batchWaitGroup.Wait()

	for _, worker := range workerManager.workers {
		worker.Stop()
	}
}
