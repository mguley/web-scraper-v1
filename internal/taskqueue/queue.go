package taskqueue

import "sync"

// Task represents a unit of work to be processed.
// Fields:
// - ID string: Unique identifier for the task.
// - URL string: URL to be processed by the task.
type Task struct {
	ID  string
	URL string
}

// TaskQueue provides a thread-safe queue for managing tasks.
// It uses a mutex for mutual exclusion and a condition variable for synchronization.
type TaskQueue struct {
	tasks []Task     // Slice to hold the tasks.
	mu    sync.Mutex // Mutex for ensuring thread safety.
	cond  *sync.Cond // Condition variable for synchronization.
}

// NewTaskQueue creates a new instance of TaskQueue.
// It initializes the task slice and the condition variable.
//
// Returns:
// - *TaskQueue: A pointer to an instance of TaskQueue.
func NewTaskQueue() *TaskQueue {
	queue := &TaskQueue{
		tasks: make([]Task, 0),
	}
	queue.cond = sync.NewCond(&queue.mu)
	return queue
}

// AddTask adds a new task to the queue and signals a waiting goroutine.
//
// Parameters:
// - task Task: The task to be added to the queue.
func (queue *TaskQueue) AddTask(task Task) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.tasks = append(queue.tasks, task)
	queue.cond.Signal()
}

// GetTask retrieves a task from the queue, waiting if necessary until a task is available.
//
// Returns:
// - Task: The task retrieved from the queue.
func (queue *TaskQueue) GetTask() Task {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	for len(queue.tasks) == 0 {
		queue.cond.Wait()
	}
	task := queue.tasks[0]
	queue.tasks = queue.tasks[1:]
	return task
}

// Close clears the queue and broadcasts to all waiting goroutines.
func (queue *TaskQueue) Close() {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.tasks = nil
	queue.cond.Broadcast()
}
