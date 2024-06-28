package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/mguley/web-scraper-v1/internal/parser"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"github.com/mguley/web-scraper-v1/internal/taskqueue"
	"github.com/mguley/web-scraper-v1/internal/tor"
	"github.com/mguley/web-scraper-v1/internal/useragent"
	"log"
	"os"
	"time"
)

var (
	cfg         *config.Config // Pointer to configuration loaded from the environment
	receiverURL string         // URL of the receiver service to be scraped
	poolSize    int            // Number of simultaneous connections in the pool
)

// init initializes the configuration by reading environment variables and command-line flags.
func init() {
	flag.StringVar(&receiverURL, "receiver-url", "", "URL of the receiver service")
	flag.Parse()

	// Obtain receiver URL from environment variables if not set via flags
	if receiverURL == "" {
		receiverURL = os.Getenv("RECEIVER_URL")
	}
	if receiverURL == "" {
		log.Fatalf("Receiver URL must be specified")
	}

	// Load configuration from environment variables (transmitter pod)
	cfg = &config.Config{
		RabbitMQ: config.RabbitMQ{
			Host:         os.Getenv("RABBIT_HOST"),
			Port:         os.Getenv("RABBIT_PORT"),
			User:         os.Getenv("RABBIT_USER"),
			Pass:         os.Getenv("RABBIT_PASS"),
			ExchangeName: os.Getenv("RABBIT_EXCHANGE_NAME"),
			QueueName:    os.Getenv("RABBIT_QUEUE_NAME"),
		},
		TorProxy: config.TorProxyConfig{
			Host:            os.Getenv("TOR_PROXY_HOST"),
			Port:            os.Getenv("TOR_PROXY_PORT"),
			ControlPort:     os.Getenv("TOR_PROXY_CONTROL_PORT"),
			ControlPassword: os.Getenv("TOR_PROXY_CONTROL_PASSWORD"),
			PingUrl:         os.Getenv("TOR_PROXY_PING_URL"),
			VerifyUrl:       os.Getenv("TOR_PROXY_VERIFY_URL"),
		},
	}

	// Number of connections to maintain in the pool.
	poolSize = 1
}

// main is the entry point of the application.
// It initializes the Tor facade, user agent generator, and receiver processor,
// then processes tasks in batches, changing the Tor identity between batches.
func main() {
	defer handlePanic()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	facade := initializeTorFacade()
	compositeUserAgentGen := initializeUserAgentGenerator()
	receiverProcessor := initializeReceiverProcessor(facade, compositeUserAgentGen)

	queueManager := initializeQueueManager(ctx, receiverProcessor)
	defer queueManager.Stop()

	processTasks(queueManager, facade, receiverProcessor)
}

// handlePanic handles any panic that occurs in the main function, logging the error and exiting the program.
func handlePanic() {
	if err := recover(); err != nil {
		log.Fatalf("Application panicked: %v", err)
	}
}

// borrowAndUseConnection borrows a connection from the Tor facade, uses it in the provided function, and returns
// the connection to the pool afterward.
func borrowAndUseConnection(facade *tor.Facade, use func(conn *tor.Connection)) {
	conn, err := facade.BorrowConnection()
	if err != nil {
		log.Fatalf("Failed to borrow connection: %v", err)
	}
	defer facade.ReturnConnection(conn)
	use(conn)
}

// initializeTorFacade initializes the Tor facade.
//
// Returns:
// - *tor.Facade: The initialized Tor facade.
func initializeTorFacade() *tor.Facade {
	facade, err := tor.NewTorFacade(&cfg.TorProxy, poolSize)
	if err != nil {
		log.Fatalf("Failed to initialize tor: %v", err)
	}
	log.Println("\nTor facade initialized.")
	return facade
}

// initializeUserAgentGenerator initializes the user agent generator.
//
// Returns:
// - *useragent.CompositeUserAgentGenerator: The composite user agent generator.
func initializeUserAgentGenerator() *useragent.CompositeUserAgentGenerator {
	availableUserAgentGenerators := []useragent.UserAgentGenerator{
		useragent.NewChromeUserAgentGenerator(),
	}
	return useragent.NewCompositeUserAgentGenerator(availableUserAgentGenerators)
}

// initializeReceiverProcessor initializes the receiver processor with a borrowed HTTP client and user agent generator.
//
// Parameters:
// - facade *tor.Facade: The Tor facade for borrowing a connection.
// - compositeUserAgentGen *useragent.CompositeUserAgentGenerator: The user agent generator.
//
// Returns:
// - *processor.JobProcessor[model.ReceiverResponse]: The initialized receiver processor.
func initializeReceiverProcessor(facade *tor.Facade,
	compositeUserAgentGen *useragent.CompositeUserAgentGenerator) *processor.JobProcessor[model.ReceiverResponse] {

	var receiverProcessor *processor.JobProcessor[model.ReceiverResponse]

	borrowAndUseConnection(facade, func(conn *tor.Connection) {
		receiverParser := parser.NewReceiverResponseParser()
		var createErr error
		receiverProcessor, createErr = processor.NewJobProcessor[model.ReceiverResponse](conn.HttpClient,
			cfg.RabbitMQ, receiverParser, compositeUserAgentGen)
		if createErr != nil {
			log.Fatalf("Failed to create receiver processor: %v", createErr)
		}
	})

	return receiverProcessor
}

// initializeQueueManager initializes the task queue manager with the specified context and receiver processor.
//
// Parameters:
// - ctx context.Context: The context for managing cancellation.
// - receiverProcessor *processor.JobProcessor[model.ReceiverResponse]: The job processor for receiver responses.
//
// Returns:
// - *taskqueue.QueueManager[model.ReceiverResponse]: The initialized task queue manager.
func initializeQueueManager(ctx context.Context,
	receiverProcessor *processor.JobProcessor[model.ReceiverResponse]) *taskqueue.QueueManager[model.ReceiverResponse] {

	workerConfig := &taskqueue.WorkerConfig{RetryLimit: 3, RetryDelay: 2 * time.Second}
	workerCount := 3
	queueManager := taskqueue.NewTaskQueueManager[model.ReceiverResponse](ctx, workerCount, receiverProcessor, workerConfig)
	log.Println("\nTask queue manager started.")
	return queueManager
}

// processTasks processes the tasks in batches, changing the Tor identity between batches.
//
// Parameters:
// - queueManager *taskqueue.QueueManager[model.ReceiverResponse]: The queue manager managing the tasks.
// - facade *tor.Facade: The Tor facade for changing the Tor identity.
// - receiverProcessor *processor.JobProcessor[model.ReceiverResponse]: The job processor for receiver responses.
func processTasks(queueManager *taskqueue.QueueManager[model.ReceiverResponse],
	facade *tor.Facade, receiverProcessor *processor.JobProcessor[model.ReceiverResponse]) {

	totalRequests := 11
	batchSize := 3

	for i := 0; i < totalRequests; i++ {
		if i > 0 && i%batchSize == 0 {
			processBatch(queueManager, facade, receiverProcessor)
		}

		taskId := i + 1
		queueManager.AddTask(taskqueue.Task{ID: fmt.Sprintf("task-%d", taskId), URL: receiverURL})
		log.Printf("Enqueued job %d with URL: %s", taskId, receiverURL)
	}

	log.Println("Waiting for last batch to finish...")
	queueManager.ProcessExistingTasks()

	log.Println("All tasks completed.")
}

// processBatch processes the current batch of tasks and changes the Tor identity.
//
// Parameters:
// - queueManager *taskqueue.QueueManager[model.ReceiverResponse]: The queue manager managing the tasks.
// - facade *tor.Facade: The Tor facade for changing the Tor identity.
// - receiverProcessor *processor.JobProcessor[model.ReceiverResponse]: The job processor for receiver responses.
func processBatch(queueManager *taskqueue.QueueManager[model.ReceiverResponse], facade *tor.Facade,
	receiverProcessor *processor.JobProcessor[model.ReceiverResponse]) {

	log.Println("Waiting for batch to finish...")
	queueManager.ProcessExistingTasks()

	log.Println("Changing Tor identity...")
	err := facade.ChangeIdentity()
	if err != nil {
		log.Fatalf("Failed to change Tor identity: %v", err)
	}
	log.Println("Tor identity changed.")

	// Borrow a new connection and set the processor's HTTP client
	borrowAndUseConnection(facade, func(conn *tor.Connection) {
		receiverProcessor.SetHttpClient(conn.HttpClient)
	})
}
