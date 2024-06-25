package main

import (
	"context"
	"flag"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/crawler"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/mguley/web-scraper-v1/internal/parser"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"github.com/mguley/web-scraper-v1/internal/tor"
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

	receiverParser := parser.NewReceiverResponseParser()
	receiverProcessor, createErr := processor.NewJobProcessor[model.ReceiverResponse](httpClient,
		cfg.RabbitMQ, receiverParser)
	if createErr != nil {
		log.Fatalf("Failed to create receiver processor: %v", createErr)
	}

	dispatcherConfig := crawler.DispatcherConfig{MaxWorkers: 1, BatchLimit: 5}
	dispatcher := crawler.NewDispatcher[model.ReceiverResponse](ctx, dispatcherConfig, receiverProcessor, facade)
	dispatcher.Run()
	log.Println("\nDispatcher started.")
	defer dispatcher.Stop()

	// Enqueue URLs for processing
	log.Println("Enqueuing jobs...")
	// fixme found defects and it requires refactoring
	/*if enqueueErr := enqueueJobs(dispatcher, receiverURL, 5); enqueueErr != nil {
		log.Fatalf("Failed to enqueue a job: %v", enqueueErr)
	}*/

	// Block to keep the application running
	select {}
}

func enqueueJobs(dispatcher *crawler.Dispatcher[model.ReceiverResponse], receiverURL string, unitCount int) error {
	for i := 0; i < unitCount; i++ {
		if err := dispatcher.WorkerManager.AssignUnit(crawler.Unit{URL: receiverURL}); err != nil {
			return err
		}
		log.Printf("Enqueued job %d with URL: %s", i+1, receiverURL)
	}
	return nil
}
