//go:build deprecated
// +build deprecated

package crawler

import (
	"context"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/crawler"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/stretchr/testify/require"
	"runtime"
	"sync"
	"testing"
	"time"
)

// SimpleDummyProcessor implements the Processor[T any] interface for testing purposes.
// This processor simulates processing by introducing a delay without performing any real operations.
type SimpleDummyProcessor struct{}

// Process mocks the processing of a unit and always returns nil.
// It introduces a delay to simulate processing time.
//
// Parameters:
// - ctx context.Context: The context for managing cancellation and deadlines.
// - url string: The URL to be processed.
//
// Returns:
// - *model.Job: Always returns nil for this dummy processor.
// - error: Always returns nil for this dummy processor.
func (simpleProcessor *SimpleDummyProcessor) Process(ctx context.Context, url string) (*model.Job, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1 * time.Second):
	}
	return nil, nil
}

// setupManager sets up a WorkerManager with a dummy processor for testing purposes.
// This helper function initializes a WorkerManager with a specified configuration and ensures it is not nil.
//
// Parameters:
// - t *testing.T: The testing context.
//
// Returns:
// - *crawler.WorkerManager[model.Job]: The initialized WorkerManager instance.
// - *crawler.DispatcherConfig: The configuration used for the WorkerManager.
func setupManager(t *testing.T) (*crawler.WorkerManager[model.Job], *crawler.DispatcherConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	config := crawler.DispatcherConfig{MaxWorkers: 5, BatchLimit: 10}
	dummyProcessor := &SimpleDummyProcessor{}
	manager := crawler.NewWorkerManager[model.Job](ctx, config, dummyProcessor)
	require.NotNil(t, manager)
	return manager, &config
}

// TestWorkerManagerCancellation verifies the WorkerManager's response to context cancellation.
// It ensures the WorkerManager stops processing units when the context is cancelled.
//
// Steps:
// - Initialize the WorkerManager with a context and configuration.
// - Start the WorkerManager.
// - Create a large number of units to be processed.
// - Send units to available workers for processing.
// - Cancel the context after a short delay.
// - Verify that the WorkerManager stops processing and shuts down correctly.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerManagerCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	config := crawler.DispatcherConfig{MaxWorkers: 100, BatchLimit: 100}
	dummyProcessor := &SimpleDummyProcessor{}
	manager := crawler.NewWorkerManager[model.Job](ctx, config, dummyProcessor)
	require.NotNil(t, manager)

	manager.Start()

	processedUnits := 0
	unitCount := 10_000
	units := make([]crawler.Unit, unitCount)
	for i := 0; i < unitCount; i++ {
		units[i] = crawler.Unit{URL: fmt.Sprintf("https://example.com/%d", i)}
	}

	// Push units into the WorkerManager's UnitQueue
	go func() {
		for _, unit := range units {
			if assignErr := manager.AssignUnit(unit); assignErr != nil {
				t.Log(assignErr)
				return
			}
		}
	}()

	// Monitor BatchDone signals
	go func() {
		for range units {
			<-manager.BatchDone
			processedUnits++
		}
	}()

	// Cancel the context after a short delay
	time.AfterFunc(2*time.Second, cancel)

	// Wait to ensure the manager has responded to the cancellation
	<-ctx.Done()

	// Verify that the manager has stopped processing
	select {
	case <-ctx.Done():
		t.Log("Manager context cancelled as expected")
		t.Logf("Processed units %v", processedUnits)
	default:
		t.Fatal("Manager context should be cancelled")
	}

	// Verify the worker manager has stopped
	manager.Stop()
	require.NotEqual(t, unitCount, processedUnits, "The initial unit count should not match processed units due to the cancellation")
	require.Empty(t, manager.UnitQueue, "UnitQueue should be empty after manager stops")
}

// TestWorkerManagerInitialization tests the initialization of a WorkerManager instance.
// This test ensures that the WorkerManager and its configuration are correctly initialized.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerManagerInitialization(t *testing.T) {
	t.Parallel()
	manager, config := setupManager(t)
	require.NotNil(t, manager)
	require.NotNil(t, config)
}

// TestWorkerManagerStart tests the start functionality of the WorkerManager.
// This test ensures that the correct number of workers are created and verifies the actual number of goroutines.
//
// The test captures the initial number of goroutines, starts the WorkerManager, and then verifies that the number of
// goroutines has increased by the expected number of workers.
//
// Steps:
// - Capture the initial number of goroutines.
// - Start the WorkerManager.
// - Verify the number of created workers matches the configuration.
// - Wait for the workers to start their goroutines.
// - Capture the current number of goroutines.
// - Verify the current number of goroutines matches the expected count.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerManagerStart(t *testing.T) {
	manager, config := setupManager(t)
	require.NotNil(t, manager)
	require.NotNil(t, config)

	initialGoroutines := runtime.NumGoroutine()
	manager.Start()
	defer manager.Stop()

	require.Equal(t, config.MaxWorkers, len(manager.GetWorkers()),
		fmt.Sprintf("expected %d workers, got %d", config.MaxWorkers, len(manager.GetWorkers())))

	// Allow some time for goroutines to start
	time.Sleep(2 * time.Second)

	currentGoroutines := runtime.NumGoroutine()
	require.Equal(t, initialGoroutines+config.MaxWorkers, currentGoroutines,
		fmt.Sprintf("expected %d goroutines, got %d", initialGoroutines+config.MaxWorkers, currentGoroutines))
}

// TestAssignUnit tests the AssignUnit functionality of the WorkerManager.
// This test ensures that units of work are correctly assigned to workers and processed.
//
// It sets up a WorkerManager, assigns multiple units of work, and verifies that all units are processed by
// tracking the BatchDone channel.
//
// Steps:
// - Initialize the WorkerManager with a dummy processor.
// - Create multiple units of work to be processed.
// - Use a WaitGroup to track the processing of units.
// - Start a goroutine to listen for BatchDone signals and decrement the WaitGroup counter.
// - Assign units to the WorkerManager.
// - Wait for all units to be processed.
// - Verify that the BatchDone channel is empty, indicating all units were processed.
//
// Parameters:
// - t *testing.T: The testing context.
func TestAssignUnit(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	manager, config := setupManager(t)
	require.NotNil(t, manager)
	require.NotNil(t, config)

	manager.Start()
	defer manager.Stop()

	units := []crawler.Unit{
		{URL: "https://ex.com/1"},
		{URL: "https://ex.com/2"},
		{URL: "https://ex.com/3"},
		{URL: "https://ex.com/4"},
		{URL: "https://ex.com/5"},
		{URL: "https://ex.com/6"},
		{URL: "https://ex.com/7"},
	}

	var wg sync.WaitGroup
	wg.Add(len(units))

	go func() {
		for range units {
			<-manager.BatchDone
			wg.Done()
		}
	}()

	for _, unit := range units {
		err := manager.AssignUnit(unit)
		require.NoError(t, err, fmt.Sprintf("failed to assign unit: %v", unit.URL))
	}

	// Wait for all units to be processed
	wg.Wait()

	// Check that all units were processed
	require.Equal(t, 0, len(manager.BatchDone), "not all units were processed")
}

// TestWorkerManagerStop tests the stop functionality of the WorkerManager.
// This test ensures that all workers are terminated and verifies behavior after stopping the manager.
//
// Steps:
// - Initialize the WorkerManager with a dummy processor.
// - Assign multiple units of work.
// - Stop the WorkerManager and ensure all workers are terminated.
// - Verify that no more units can be assigned and workers are terminated.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerManagerStop(t *testing.T) {
	t.Parallel() // Run this test in parallel to ensure isolation
	manager, config := setupManager(t)
	require.NotNil(t, manager)
	require.NotNil(t, config)

	manager.Start()

	units := []crawler.Unit{
		{URL: "https://ex.com/1"},
		{URL: "https://ex.com/2"},
		{URL: "https://ex.com/3"},
	}

	var wg sync.WaitGroup
	wg.Add(len(units))

	go func() {
		for range units {
			<-manager.BatchDone
			wg.Done()
		}
	}()

	for _, unit := range units {
		err := manager.AssignUnit(unit)
		require.NoError(t, err, fmt.Sprintf("failed to assign unit: %v", unit.URL))
	}

	// Wait for all units to be processed
	wg.Wait()
	require.Equal(t, 0, len(manager.BatchDone), "not all units were processed")

	// Stop the WorkerManager and ensure all workers are terminated
	manager.Stop()
	time.Sleep(2 * time.Second)

	// Check that no more units can be assigned and workers are terminated
	err := manager.AssignUnit(crawler.Unit{URL: "https://ex.com/extra"})
	require.Error(t, err, "expected error when assigning unit after stopping the manager")
}
