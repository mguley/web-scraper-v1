package crawler

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/processor"
)

// Unit represents a unit of work to be processed, in this context a URL.
//
// Fields:
// - URL string: The URL to be processed.
type Unit struct {
	URL string
}

// Worker represents a single worker that processes units from the UnitQueue.
// Each worker registers itself with the WorkerQueue and retrieves units to process.
//
// Type parameter:
//   - T any: The type of data that the Worker processes. This ensures that the processor operates on the same
//     data type, providing type safety throughout the processing pipeline.
//
// Fields:
// - ID int: Unique identifier for the worker.
// - UnitQueue chan Unit: Channel for receiving units to process.
// - WorkerQueue chan chan Unit: Channel of channels for managing available workers.
// - QuitChan chan bool: Channel for signaling the worker to stop.
// - Processor processor.Processor[T]: Job processor to handle job processing logic.
// - BatchDone chan bool: Channel for signaling the completion of a unit.
//
// Methods:
// - NewWorker: Constructs a new Worker with a specified ID, worker queue, job processor, and batch done channel.
// - Start: Begins the worker's job processing loop.
// - Stop: Signals the worker to stop processing units.
type Worker[T any] struct {
	ID          int
	UnitQueue   chan Unit
	WorkerQueue chan chan Unit
	QuitChan    chan bool
	Processor   processor.Processor[T]
	BatchDone   chan bool
}

// NewWorker creates a new Worker instance with the given ID, worker queue, batch done channel, and job processor.
//
// Parameters:
// - id int: Unique identifier for the worker.
// - workerQueue chan chan Unit: Channel of channels for managing available workers.
// - processor processor.Processor[T]: Job processor to handle job processing logic.
// - batchDone chan bool: Channel for signaling the completion of a unit.
//
// Returns:
// - *Worker[T]: A pointer to a new Worker instance.
func NewWorker[T any](id int, workerQueue chan chan Unit, processor processor.Processor[T],
	batchDone chan bool) *Worker[T] {
	return &Worker[T]{
		ID:          id,
		UnitQueue:   make(chan Unit),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		Processor:   processor,
		BatchDone:   batchDone,
	}
}

// Start begins the worker's job processing loop. The worker continuously registers itself in the WorkerQueue
// and waits for units to process. Upon receiving a unit, it processes the unit using the job processor.
// The worker can exit the loop if it receives a signal on the QuitChan.
func (worker *Worker[T]) Start() {
	go func() {
		for {
			// Register the worker's unit channel in the WorkerQueue.
			worker.WorkerQueue <- worker.UnitQueue
			select {
			case unit := <-worker.UnitQueue:
				// When a unit is received, process it using the job processor.
				_, processErr := worker.Processor.Process(unit.URL)
				if processErr != nil {
					fmt.Printf("Worker %d failed to process unit: %v\n", worker.ID, processErr)
				}
				worker.BatchDone <- true // Signal the dispatcher that the unit is done
			case <-worker.QuitChan:
				// Exit the loop if a quit signal is received.
				return
			}
		}
	}()
}

// Stop signals the worker to stop processing units by sending a message on the QuitChan.
// This will cause the worker's job processing loop to exit.
func (worker *Worker[T]) Stop() {
	go func() {
		worker.QuitChan <- true
	}()
}
