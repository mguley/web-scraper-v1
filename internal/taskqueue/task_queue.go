package taskqueue

import (
	"context"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"log"
	"sync"
)

// QueueManager manages a task queue and a pool of workers to process the tasks.
// It ensures thread safety and proper lifecycle management of workers.
type QueueManager[T any] struct {
	taskQueue *TaskQueue         // Task queue for managing tasks.
	workers   []*Worker[T]       // Slice of workers processing tasks.
	ctx       context.Context    // Context for managing cancellation.
	cancel    context.CancelFunc // Function to cancel the context.
	wg        sync.WaitGroup     // Wait group for synchronizing worker completion.
}

// NewTaskQueueManager creates a new instance of QueueManager.
// It initializes the task queue and starts the specified number of workers.
//
// Parameters:
// - ctx context.Context: Parent context for the queue manager.
// - workerCount int: Number of workers to start.
// - processor processor.Processor[T]: Processor for handling task processing.
//
// Returns:
// - *QueueManager[T]: A pointer to an instance of QueueManager.
func NewTaskQueueManager[T any](ctx context.Context, workerCount int, processor processor.Processor[T]) *QueueManager[T] {
	ctx, cancel := context.WithCancel(ctx)
	taskQueue := NewTaskQueue()
	manager := &QueueManager[T]{
		taskQueue: taskQueue,
		workers:   make([]*Worker[T], workerCount),
		ctx:       ctx,
		cancel:    cancel,
	}
	manager.initializeWorkers(workerCount, processor)
	return manager
}

// initializeWorkers starts the specified number of workers to process tasks.
//
// Parameters:
// - workerCount int: Number of workers to start.
// - processor processor.Processor[T]: Processor for handling task processing.
func (manager *QueueManager[T]) initializeWorkers(workerCount int, processor processor.Processor[T]) {
	for i := 0; i < workerCount; i++ {
		worker := NewWorker[T](manager.ctx, i+1, manager.taskQueue, processor)
		manager.workers[i] = worker
		manager.wg.Add(1)
		go worker.Start(&manager.wg)
	}
}

// AddTask adds a new task to the task queue.
//
// Parameters:
// - task Task: The task to be added to the queue.
func (manager *QueueManager[T]) AddTask(task Task) {
	manager.taskQueue.AddTask(task)
	log.Printf("Added new task to queue: %v", task)
}

// ProcessExistingTasks waits until all tasks in the queue are processed.
func (manager *QueueManager[T]) ProcessExistingTasks() {
	manager.taskQueue.ProcessExistingTasks()
}

// Stop stops all workers and closes the task queue.
func (manager *QueueManager[T]) Stop() {
	log.Printf("Stopping task queue manager")
	manager.cancel()
	manager.taskQueue.Close()
	manager.wg.Wait()
	for _, worker := range manager.workers {
		worker.Stop()
	}
}
