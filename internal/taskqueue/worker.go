package taskqueue

import (
	"context"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"log"
	"sync"
)

// Worker represents a single worker that processes tasks from the TaskQueue.
// It uses a processor to handle the task processing logic.
type Worker[T any] struct {
	workerId  int                    // Unique identifier for the worker.
	taskQueue *TaskQueue             // Task queue from which tasks are retrieved.
	processor processor.Processor[T] // Processor for handling task processing.
	ctx       context.Context        // Context for managing cancellation.
	quit      chan struct{}          // Channel to signal the worker to stop.
}

// NewWorker creates a new instance of Worker.
// It initializes the worker with the specified ID, task queue, and processor.
//
// Parameters:
// - ctx context.Context: Context for managing cancellation.
// - workerId int: Unique identifier for the worker.
// - taskQueue *TaskQueue: Task queue from which tasks are retrieved.
// - processor processor.Processor[T]: Processor for handling task processing.
//
// Returns:
// - *Worker[T]: A pointer to an instance of Worker.
func NewWorker[T any](ctx context.Context, workerId int, taskQueue *TaskQueue,
	processor processor.Processor[T]) *Worker[T] {
	return &Worker[T]{
		workerId:  workerId,
		taskQueue: taskQueue,
		processor: processor,
		ctx:       ctx,
		quit:      make(chan struct{}),
	}
}

// Start begins the worker's task processing loop.
// It processes tasks from the task queue until the context is cancelled or the worker is signaled to stop.
//
// Parameters:
// - wg *sync.WaitGroup: Wait group for synchronizing worker completion.
func (worker *Worker[T]) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-worker.ctx.Done():
			log.Printf("Worker %d context cancelled", worker.workerId)
			return
		case <-worker.quit:
			log.Printf("Worker %d quitting", worker.workerId)
			return
		default:
			task := worker.taskQueue.GetTask()
			log.Printf("Worker %d processing task: %v", worker.workerId, task)
			_, err := worker.processor.Process(worker.ctx, task.URL)
			if err != nil {
				log.Printf("Worker %d failed to process task: %v", worker.workerId, err)
			}
			log.Printf("Worker %d completed task: %v", worker.workerId, task)
		}
	}
}

// Stop signals the worker to stop processing tasks.
func (worker *Worker[T]) Stop() {
	log.Println("Worker stopped")
	close(worker.quit)
}
