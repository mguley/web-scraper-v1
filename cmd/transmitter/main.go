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
	envPath     string         // Path to the environment variable file
	cfg         *config.Config // Pointer to configuration loaded from the environment
	receiverURL string         // URL of the receiver service to be scraped
	poolSize    int            // Number of simultaneous connections in the pool
)

func init() {
	flag.StringVar(&envPath, "env", ".env", "path to env file")
	flag.StringVar(&receiverURL, "receiver-url", "", "URL of the receiver service")
	flag.Parse()

	if initConfigErr := config.InitConfig(envPath); initConfigErr != nil {
		log.Fatalf("Failed to initialize configuration: %v", initConfigErr)
		return
	}

	var configErr error
	if cfg, configErr = config.GetConfig(); configErr != nil {
		log.Fatalf("Failed to get configuration: %v", configErr)
	}

	// Obtain receiver URL from environment variables if not set via flags
	receiverURL = os.Getenv("RECEIVER_URL")
	if receiverURL == "" {
		log.Fatalf("Receiver URL must be specified")
	}

	// Number of connections to maintain in the pool.
	poolSize = 2
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatcherConfig := crawler.DispatcherConfig{MaxWorkers: 2, BatchLimit: 5}

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

	receiverParser := parser.NewReceiverResponseParser()
	receiverProcessor, createErr := processor.NewJobProcessor[model.ReceiverResponse](httpClient, cfg.RabbitMQ, receiverParser)
	if createErr != nil {
		log.Fatalf("Failed to create receiver processor: %v", createErr)
	}

	dispatcher := crawler.NewDispatcher[model.ReceiverResponse](ctx, dispatcherConfig, receiverProcessor, facade)
	dispatcher.Run()
	defer dispatcher.Stop()

	// Enqueue an initial URL for processing
	dispatcher.WorkerManager.UnitQueue <- crawler.Unit{URL: receiverURL}
}
