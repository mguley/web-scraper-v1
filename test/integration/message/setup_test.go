package message

import (
	"flag"
	"github.com/mguley/web-scraper-v1/config"
	"log"
	"os"
	"testing"
)

var (
	// envPath holds the path to the environment configuration file.
	envPath string
	// cfg holds the global configuration settings loaded from the environment file.
	cfg *config.Config
)

// init initializes the envPath flag for specifying the path to the environment file.
// This function is called automatically when the package is initialized.
func init() {
	flag.StringVar(&envPath, "env", "../config/.env", "path to env file")
}

// setup initializes the global configuration by loading settings from the specified environment file.
// If there is an error during the initialization or configuration retrieval, it logs a fatal error.
func setup() {
	// Initialize configuration
	if initConfigErr := config.InitConfig(envPath); initConfigErr != nil {
		log.Fatalf("Failed to initialize configuration: %v", initConfigErr)
		return
	}

	var configErr error
	cfg, configErr = config.GetConfig()
	if configErr != nil {
		log.Fatalf("Failed to get config: %v", configErr)
	}
}

// teardown performs any necessary cleanup after tests are executed.
func teardown() {}

// TestMain is the entry point for the test suite.
// It parses command-line flags, sets up the global configuration, runs the tests, and performs cleanup.
func TestMain(m *testing.M) {
	flag.Parse()
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
