package crawler

import (
	"context"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/crawler"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

// DummyProcessor implements the Processor interface for testing purposes.
// This processor simulates processing by introducing a delay without performing any real operations.
type DummyProcessor struct{}

// Process mocks the processing of a unit and always returns nil.
// It introduces a delay to simulate processing time.
//
// Parameters:
// - url string: The URL to be processed.
//
// Returns:
// - *model.Job: Always returns nil for this dummy processor.
// - error: Always returns nil for this dummy processor.
func (dummyProcessor *DummyProcessor) Process(url string) (*model.Job, error) {
	time.Sleep(100 * time.Millisecond)
	return nil, nil
}

// TestDispatcherStartupAndShutdown verifies the correct startup and shutdown of the Dispatcher.
// It ensures the Dispatcher initializes, runs for a short period, and shuts down cleanly.
//
// Steps:
// - Initialize the Dispatcher with a context and configuration.
// - Start the Dispatcher.
// - Allow the Dispatcher to run for a short period.
// - Stop the Dispatcher and verify it shuts down correctly.
//
// Parameters:
// - t *testing.T: The testing context.
func TestDispatcherStartupAndShutdown(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := crawler.DispatcherConfig{
		MaxWorkers: 3,
		BatchLimit: 5,
	}
	processor := &DummyProcessor{}
	torFacade := new(tor.Facade)

	dispatcher := crawler.NewDispatcher[model.Job](ctx, config, processor, torFacade)
	go dispatcher.Run()

	// Allow the dispatcher to run for a short period
	time.Sleep(1 * time.Second)

	// Shut down the dispatcher and ensure it exits cleanly
	dispatcher.Stop()

	// Verify the context is cancelled
	select {
	case <-dispatcher.GetContext().Done():
		t.Log("Dispatcher stopped")
	default:
		t.Fatal("Dispatcher context should be cancelled")
	}

	require.Empty(t, dispatcher.WorkerManager.UnitQueue, "Channel for receiving units of work should be empty")
}

// TestDispatcherTaskProcessing verifies that the Dispatcher processes units correctly.
// It ensures that all units are processed and the unit count matches the processed units count.
//
// Steps:
// - Initialize the Dispatcher with a context and configuration.
// - Start the Dispatcher.
// - Create a number of units to be processed.
// - Push units into the WorkerManager's UnitQueue.
// - Monitor BatchDone signals and ensure all units are processed.
//
// Parameters:
// - t *testing.T: The testing context.
func TestDispatcherTaskProcessing(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := crawler.DispatcherConfig{
		MaxWorkers: 1_000,
		BatchLimit: 300,
	}
	processor := &DummyProcessor{}
	torFacade := new(tor.Facade)

	dispatcher := crawler.NewDispatcher[model.Job](ctx, config, processor, torFacade)
	go dispatcher.Run()
	defer dispatcher.Stop()

	processedUnits := 0
	unitCount := 5_000
	var wg sync.WaitGroup
	wg.Add(unitCount)

	units := make([]crawler.Unit, unitCount)
	for i := 0; i < unitCount; i++ {
		units[i] = crawler.Unit{URL: fmt.Sprintf("https://example.com/%d", i)}
	}

	// Push units into the WorkerManager's UnitQueue
	go func() {
		for _, unit := range units {
			dispatcher.WorkerManager.UnitQueue <- unit
		}
	}()

	// Monitor BatchDone signals and decrement WaitGroup
	go func() {
		for range units {
			<-dispatcher.WorkerManager.BatchDone
			processedUnits++
			wg.Done()
		}
	}()

	// Wait for all units to be processed
	wg.Wait()
	require.Equal(t, unitCount, processedUnits, "Initial unit count should match processed units")
	require.Empty(t, dispatcher.WorkerManager.UnitQueue, "UnitQueue should be empty after processing all units")
}

// TestDispatcherCancellation verifies the Dispatcher's response to context cancellation.
// It ensures the Dispatcher stops processing units when the context is cancelled.
//
// Steps:
// - Initialize the Dispatcher with a context and configuration.
// - Start the Dispatcher.
// - Create a large number of units to be processed.
// - Push units into the WorkerManager's UnitQueue.
// - Cancel the context after a short delay.
// - Verify that the Dispatcher stops processing and shuts down correctly.
//
// Parameters:
// - t *testing.T: The testing context.
func TestDispatcherCancellation(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := crawler.DispatcherConfig{
		MaxWorkers: 15,
		BatchLimit: 5,
	}
	processor := &DummyProcessor{}
	torFacade := new(tor.Facade)

	dispatcher := crawler.NewDispatcher[model.Job](ctx, config, processor, torFacade)
	go dispatcher.Run()

	processedUnits := 0
	unitCount := 100_000
	units := make([]crawler.Unit, unitCount)
	for i := 0; i < unitCount; i++ {
		units[i] = crawler.Unit{URL: fmt.Sprintf("https://example.com/%d", i)}
	}

	// Push units into the WorkerManager's UnitQueue
	go func() {
		for _, unit := range units {
			dispatcher.WorkerManager.UnitQueue <- unit
		}
	}()

	// Monitor BatchDone signals
	go func() {
		for range units {
			<-dispatcher.WorkerManager.BatchDone
			processedUnits++
		}
	}()

	// Cancel the context after a short delay
	time.AfterFunc(2*time.Second, cancel)

	// Wait to ensure the dispatcher has responded to the cancellation
	<-ctx.Done()

	// Verify that the dispatcher has stopped processing
	select {
	case <-dispatcher.GetContext().Done():
		t.Log("Dispatcher context cancelled as expected")
		t.Logf("Processed units %v", processedUnits)
	default:
		t.Fatal("Dispatcher context should be cancelled")
	}

	// Verify the worker manager has stopped
	dispatcher.Stop()
	require.NotEqual(t, unitCount, processedUnits, "The initial unit count should not match processed units due to the cancellation")
	require.Empty(t, dispatcher.WorkerManager.UnitQueue, "UnitQueue should be empty after dispatcher stops")
}
