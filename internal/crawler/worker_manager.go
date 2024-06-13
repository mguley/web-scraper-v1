package crawler

import (
	"github.com/mguley/web-scraper-v1/internal/processor"
	"sync"
)

// WorkerManager manages the lifecycle and operations of a pool of workers that process units of work.
//
// Type parameter:
//   - T any: The type of data that the workers process. This ensures type safety across the worker processes.
//
// Fields:
// - WorkerQueue chan chan Unit: Channel of worker channels, used to distribute available workers.
// - config DispatcherConfig: Configuration settings for managing worker operations.
// - UnitQueue chan Unit: Channel for receiving units of work to be processed.
// - BatchDone chan bool: Channel for signaling the completion of processing a batch of units.
// - batchWaitGroup sync.WaitGroup: Synchronizes the completion of processing batches.
// - workers []*Worker[T]: List of worker instances managed by this manager.
// - Processor processor.Processor[T]: Processor used for handling units of work.
type WorkerManager[T any] struct {
	WorkerQueue    chan chan Unit
	config         DispatcherConfig
	UnitQueue      chan Unit
	BatchDone      chan bool
	batchWaitGroup sync.WaitGroup
	workers        []*Worker[T]
	Processor      processor.Processor[T]
}

// NewWorkerManager creates a new WorkerManager instance with a specified configuration and a processor.
// This constructor sets up the necessary channels and initializes the list of workers based on the configuration.
//
// Parameters:
// - config DispatcherConfig: Configuration settings including the maximum number of workers and batch limit.
// - processor processor.Processor[T]: The processor used for handling units of work.
//
// Returns:
// - *WorkerManager[T]: A pointer to an instance of WorkerManager.
func NewWorkerManager[T any](config DispatcherConfig, processor processor.Processor[T]) *WorkerManager[T] {
	return &WorkerManager[T]{
		WorkerQueue: make(chan chan Unit, config.MaxWorkers),
		config:      config,
		UnitQueue:   make(chan Unit),
		BatchDone:   make(chan bool, config.MaxWorkers),
		workers:     make([]*Worker[T], 0, config.MaxWorkers),
		Processor:   processor,
	}
}

// Start initializes and starts all worker instances under management.
func (workerManager *WorkerManager[T]) Start() {
	for i := 0; i < workerManager.config.MaxWorkers; i++ {
		worker := NewWorker[T](i, workerManager.WorkerQueue, workerManager.Processor, workerManager.BatchDone)
		worker.Start()
		workerManager.workers = append(workerManager.workers, worker)
	}
}

// AssignUnit sends a unit of work to an available worker for processing.
func (workerManager *WorkerManager[T]) AssignUnit(unit Unit) {
	workerChannel := <-workerManager.WorkerQueue
	workerChannel <- unit
}

// WaitForBatchCompletion waits for the completion of processing for all units in the current batch.
// This method synchronizes on the batch limit defined in the configuration.
func (workerManager *WorkerManager[T]) WaitForBatchCompletion() {
	workerManager.batchWaitGroup.Add(workerManager.config.BatchLimit)
	for i := 0; i < workerManager.config.BatchLimit; i++ {
		<-workerManager.BatchDone
		workerManager.batchWaitGroup.Done()
	}
	workerManager.batchWaitGroup.Wait()
}

// Stop terminates all workers managed by this manager and cleans up resources.
func (workerManager *WorkerManager[T]) Stop() {
	for _, worker := range workerManager.workers {
		worker.Stop()
	}
	close(workerManager.UnitQueue) // Signal that no more units will be dispatched.
}
