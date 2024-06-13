package crawler

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"sync"
	"time"
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
// It ensures that a worker is available before assigning the unit and logs the process.
// It uses a timeout to avoid blocking indefinitely if no worker is available.
//
// Parameters:
// - unit Unit: The unit of work to be processed, containing the URL to be processed.
//
// Returns:
// - error: An error if no available worker is found within the timeout period or if the unit cannot be assigned.
func (workerManager *WorkerManager[T]) AssignUnit(unit Unit) error {
	select {
	case workerChannel, ok := <-workerManager.WorkerQueue:
		if !ok {
			return fmt.Errorf("worker queue closed, cannot assign unit: %s", unit.URL)
		}
		select {
		case workerChannel <- unit:
			fmt.Printf("assigned unit to worker: %s\n", unit.URL)
		case <-time.After(time.Second * 3):
			return fmt.Errorf("timeout while assigning unit to worker: %s", unit.URL)
		}
	case <-time.After(time.Second * 3):
		return fmt.Errorf("no available workers to assign unit: %s", unit.URL)
	}
	return nil
}

// GetWorkers returns a copy of the slice of workers managed by this WorkerManager.
// It ensures that the returned slice is immutable from the caller's perspective.
//
// By providing a copy of the workers slice, this method ensures that any modifications
// made to the returned slice do not affect the original workers slice within the WorkerManager.
//
// Returns:
// - []*Worker[T]: A copy of the slice of workers managed by this WorkerManager.
func (workerManager *WorkerManager[T]) GetWorkers() []*Worker[T] {
	workersCopy := make([]*Worker[T], len(workerManager.workers))
	copy(workersCopy, workerManager.workers)
	return workersCopy
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
	// Wait for all workers to finish their tasks
	workerManager.batchWaitGroup.Wait()

	for _, worker := range workerManager.workers {
		worker.Stop()
	}
}
