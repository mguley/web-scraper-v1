package crawler

import (
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
// - url string: The URL to be processed.
//
// Returns:
// - *model.Job: Always returns nil for this dummy processor.
// - error: Always returns nil for this dummy processor.
func (simpleProcessor *SimpleDummyProcessor) Process(url string) (*model.Job, error) {
	time.Sleep(1 * time.Second) // Simulate processing time
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
	config := crawler.DispatcherConfig{MaxWorkers: 5, BatchLimit: 10}
	dummyProcessor := &SimpleDummyProcessor{}
	manager := crawler.NewWorkerManager[model.Job](config, dummyProcessor)
	require.NotNil(t, manager)
	return manager, &config
}

// TestWorkerManagerInitialization tests the initialization of a WorkerManager instance.
// This test ensures that the WorkerManager and its configuration are correctly initialized.
//
// Parameters:
// - t *testing.T: The testing context.
func TestWorkerManagerInitialization(t *testing.T) {
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
