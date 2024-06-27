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
)

var (
	cfg         *config.Config // Pointer to configuration loaded from the environment
	receiverURL string         // URL of the receiver service to be scraped
	poolSize    int            // Number of simultaneous connections in the pool
)

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
	poolSize = 2
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("Application panicked: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Tor network connection facade
	facade, err := tor.NewTorFacade(&cfg.TorProxy, poolSize)
	if err != nil {
		log.Fatalf("Failed to create facade: %v", err)
	}

	// Establish an HTTP client through the Tor facade
	httpClient, connErr := facade.EstablishConnection()
	if connErr != nil {
		log.Fatalf("Failed to establish connection: %v", connErr)
	}
	log.Println("\nNetwork connection established.")

	availableUserAgentGenerators := []useragent.UserAgentGenerator{
		useragent.NewChromeUserAgentGenerator(),
	}
	compositeUserAgentGen := useragent.NewCompositeUserAgentGenerator(availableUserAgentGenerators)

	receiverParser := parser.NewReceiverResponseParser()
	receiverProcessor, createErr := processor.NewJobProcessor[model.ReceiverResponse](httpClient,
		cfg.RabbitMQ, receiverParser, compositeUserAgentGen)
	if createErr != nil {
		log.Fatalf("Failed to create receiver processor: %v", createErr)
	}

	queueManager := taskqueue.NewTaskQueueManager[model.ReceiverResponse](ctx, 3, receiverProcessor)
	defer queueManager.Stop()

	log.Println("\nTask queue manager started.")

	// Enqueue URLs for processing
	log.Println("Enqueuing jobs...")

	unitCount := 3
	// Add tasks to the queue
	for i := 0; i < unitCount; i++ {
		queueManager.AddTask(taskqueue.Task{ID: fmt.Sprintf("task-%d", i+1), URL: receiverURL})
		log.Printf("Enqueued job %d with URL: %s", i+1, receiverURL)
	}

	// Process all existing tasks
	log.Println("Waiting for jobs to finish...")
	queueManager.ProcessExistingTasks()
	log.Println("We are done.")
}
