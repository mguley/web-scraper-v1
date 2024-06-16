package crawler

import (
	"context"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/crawler"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
	"time"
)

// DummyWorkerProcessor implements the Processor[T any] interface for testing purposes.
// This processor simulates processing by introducing a delay without performing any real operations.
type DummyWorkerProcessor struct{}

// Process mocks the processing of a unit and always returns nil.
// It introduces a delay to simulate processing time.
//
// Parameters:
// - url string: The URL to be processed.
//
// Returns:
// - *model.Job: Always returns nil for this dummy processor.
// - error: Always returns nil for this dummy processor.
func (dummyWorkerProcessor *DummyWorkerProcessor) Process(url string) (*model.Job, error) {
	time.Sleep(500 * time.Millisecond) // Simulate processing time
	return nil, nil
}

// TestWorkerStart tests the Start method of the Worker.
// It verifies that the worker can process a small number of units correctly.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerStart(t *testing.T) {
	t.Parallel()
	unitsToProcess := 10
	workerCount := 5
	ctx := context.Background()
	workerQueue := make(chan chan crawler.Unit, workerCount)
	batchDone := make(chan bool, unitsToProcess)
	processor := &DummyWorkerProcessor{}
	workers := make([]*crawler.Worker[model.Job], 0, workerCount)

	// Create and start multiple workers
	for i := 0; i < workerCount; i++ {
		worker := crawler.NewWorker[model.Job](ctx, i, workerQueue, processor, batchDone)
		go worker.Start()
		workers = append(workers, worker)
	}

	// Process multiple units
	for i := 0; i < unitsToProcess; i++ {
		unit := crawler.Unit{URL: fmt.Sprintf("https://ex.com/%d", i)}
		select {
		case workerChannel := <-workerQueue:
			workerChannel <- unit
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out waiting to send unit %d", i)
		}
	}

	// Check all units are processed
	var processedUnits int
	for i := 0; i < unitsToProcess; i++ {
		select {
		case <-batchDone:
			processedUnits++
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for unit to finish")
		}
	}

	require.Equal(t, unitsToProcess, processedUnits, "All units should be processed")

	// Stop all workers
	for _, worker := range workers {
		worker.Stop()
	}
}

// TestWorkerMemoryUsage tests the memory usage of the Worker.
// It measures memory allocation before and after creating workers, and after processing units.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerMemoryUsage(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	unitsToProcess := 150_000
	workerCount := 25_000
	ctx := context.Background()
	workerQueue := make(chan chan crawler.Unit, workerCount)
	batchDone := make(chan bool, unitsToProcess)
	processor := &DummyWorkerProcessor{}
	workers := make([]*crawler.Worker[model.Job], 0, workerCount)

	// Measure memory usage before creating workers
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	fmt.Printf("Memory Usage Before: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v\n",
		bToMb(memStatsBefore.Alloc), bToMb(memStatsBefore.TotalAlloc), bToMb(memStatsBefore.Sys), memStatsBefore.NumGC)

	// Create and start multiple workers
	for i := 0; i < workerCount; i++ {
		worker := crawler.NewWorker[model.Job](ctx, i, workerQueue, processor, batchDone)
		go worker.Start()
		workers = append(workers, worker)
	}

	// Measure memory usage after creating workers
	var memStatsAfterWorkers runtime.MemStats
	runtime.ReadMemStats(&memStatsAfterWorkers)
	fmt.Printf("Memory Usage After Workers Created: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v\n",
		bToMb(memStatsAfterWorkers.Alloc), bToMb(memStatsAfterWorkers.TotalAlloc), bToMb(memStatsAfterWorkers.Sys), memStatsAfterWorkers.NumGC)

	// Process multiple units
	for i := 0; i < unitsToProcess; i++ {
		unit := crawler.Unit{URL: fmt.Sprintf("https://ex.com/%d", i)}
		select {
		case workerChannel := <-workerQueue:
			workerChannel <- unit
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out waiting to send unit %d", i)
		}
	}

	// Check all units are processed
	var processedUnits int
	for i := 0; i < unitsToProcess; i++ {
		select {
		case <-batchDone:
			processedUnits++
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for unit to finish")
		}
	}

	require.Equal(t, unitsToProcess, processedUnits, "All units should be processed")

	// Stop all workers
	for _, worker := range workers {
		worker.Stop()
	}

	// Measure memory usage after stopping workers
	var memStatsAfterStopping runtime.MemStats
	runtime.ReadMemStats(&memStatsAfterStopping)
	fmt.Printf("Memory Usage After Stopping Workers: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v\n",
		bToMb(memStatsAfterStopping.Alloc), bToMb(memStatsAfterStopping.TotalAlloc), bToMb(memStatsAfterStopping.Sys), memStatsAfterStopping.NumGC)
}

// TestWorkerStop tests the Stop method of the Worker.
// It verifies that the worker stops processing correctly.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerStop(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	ctx := context.Background()
	workerQueue := make(chan chan crawler.Unit)
	batchDone := make(chan bool)
	processor := &DummyWorkerProcessor{}
	worker := crawler.NewWorker[model.Job](ctx, 1, workerQueue, processor, batchDone)

	go worker.Start()

	// Send a unit to ensure the worker is running
	unit := crawler.Unit{URL: fmt.Sprintf("https://ex.com/%d", 1)}
	unitQueue := <-workerQueue
	unitQueue <- unit

	// Ensure the unit is being processed
	select {
	case <-batchDone:
		t.Log("Worker processed the unit before stopping")
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not start processing in time")
	}

	// Stop the worker
	worker.Stop()

	// Ensure QuitChan received the 'true' value
	select {
	case quitSignal := <-worker.QuitChan:
		t.Log("Worker stopped by signal", quitSignal)
		require.Equal(t, true, quitSignal, "QuitChan should receive 'true'")
	case <-time.After(2 * time.Second):
		t.Fatal("QuitChan did not receive 'true' in time")
	}

	// Ensure no more units can be processed
	select {
	case unitQueue <- unit:
		t.Fatal("Worker should not accept units after stopping")
	default:
		t.Log("Worker correctly not accepting units after stop")
	}
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
