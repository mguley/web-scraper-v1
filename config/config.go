package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"sync"
)

// TorProxyConfig holds the configuration settings for the Tor proxy.
type TorProxyConfig struct {
	Host            string // Host is the hostname or IP address of the Tor proxy server.
	Port            string // Port is the port number of the Tor proxy server.
	PingUrl         string // PingUrl is the URL to ping the Tor proxy server.
	VerifyUrl       string // VerifyUrl is the URL to verify the IP through the Tor proxy.
	ControlPassword string // ControlPassword is the password for the Tor control port.
	ControlPort     string // ControlPort is the port number of the Tor control port.
}

// Sitemap represents the sitemap configuration with a URL.
type Sitemap struct {
	URL string // URL is the address of the sitemap to fetch.
}

// MongoDB holds the MongoDB configuration settings.
type MongoDB struct {
	Host       string // Host is the hostname or IP address of the MongoDB server.
	Port       string // Port is the port number of the MongoDB server.
	User       string // User is the username for connecting to the MongoDB server.
	Pass       string // Pass is the password for connecting to the MongoDB server.
	Database   string // Database is the name of the MongoDB database.
	Collection string // Collection is the name of the MongoDB collection.
}

// RabbitMQ holds the RabbitMQ configuration settings.
type RabbitMQ struct {
	Host         string // Host is the hostname or IP address of the RabbitMQ server.
	Port         string // Port is the port number of the RabbitMQ server.
	User         string // User is the username for connecting to the RabbitMQ server.
	Pass         string // Pass is the password for connecting to the RabbitMQ server.
	ExchangeName string // ExchangeName is the name of the RabbitMQ exchange.
	QueueName    string // QueueName is the name of the RabbitMQ queue.
}

// Config aggregates all the configuration required by the application.
type Config struct {
	MongoDB  MongoDB        // MongoDB is the configuration for the MongoDB server.
	RabbitMQ RabbitMQ       // RabbitMQ is the configuration for the RabbitMQ server.
	Keywords []string       // Keywords are the search keywords.
	Sitemap  Sitemap        // Sitemap holds the sitemap URL configuration.
	TorProxy TorProxyConfig // TorProxy is the configuration for the TorProxy.
}

var (
	configInstance *Config
	once           sync.Once
)

// loadEnv attempts to load environment variables from a .env file located at the provided path.
// It returns an error if the .env file cannot be loaded.
func loadEnv(path string) error {
	return godotenv.Load(path)
}

// initializeConfig initializes the application's configuration by loading environment variables.
// It accepts a path to the .env file and returns a pointer to a Config struct populated with all the configurations.
func initializeConfig(envPath string) (*Config, error) {
	if err := loadEnv(envPath); err != nil {
		currentDir, _ := os.Getwd()
		return nil, fmt.Errorf("failed to load .env file from directory: %s, error: %v", currentDir, err)
	}

	return &Config{
		MongoDB: MongoDB{
			Host:       os.Getenv("MONGODB_HOST"),
			Port:       os.Getenv("MONGODB_PORT"),
			User:       os.Getenv("MONGODB_USER"),
			Pass:       os.Getenv("MONGODB_PASS"),
			Database:   os.Getenv("MONGODB_DATABASE"),
			Collection: os.Getenv("MONGODB_COLLECTION"),
		},
		RabbitMQ: RabbitMQ{
			Host:         os.Getenv("RABBIT_HOST"),
			Port:         os.Getenv("RABBIT_PORT"),
			User:         os.Getenv("RABBIT_USER"),
			Pass:         os.Getenv("RABBIT_PASS"),
			ExchangeName: os.Getenv("RABBIT_EXCHANGE_NAME"),
			QueueName:    os.Getenv("RABBIT_QUEUE_NAME"),
		},
		Keywords: getSearchKeywordsFromEnv(),
		Sitemap: Sitemap{
			URL: os.Getenv("SITEMAP_URL"),
		},
		TorProxy: TorProxyConfig{
			Host:            os.Getenv("TOR_PROXY_HOST"),
			Port:            os.Getenv("TOR_PROXY_PORT"),
			PingUrl:         os.Getenv("TOR_PROXY_PING_URL"),
			VerifyUrl:       os.Getenv("TOR_PROXY_VERIFY_URL"),
			ControlPassword: os.Getenv("TOR_PROXY_CONTROL_PASSWORD"),
			ControlPort:     os.Getenv("TOR_PROXY_CONTROL_PORT"),
		},
	}, nil
}

// InitConfig initializes the configuration if it hasn't been initialized yet.
// This function should be called once at the start of the application.
func InitConfig(envPath string) error {
	var err error
	once.Do(func() {
		configInstance, err = initializeConfig(envPath)
	})
	return err
}

// GetConfig returns the singleton instance of the application's configuration.
// It returns an error if the configuration hasn't been initialized yet.
func GetConfig() (*Config, error) {
	if configInstance == nil {
		return nil, fmt.Errorf("configuration not initialized, please call InitConfig first")
	}
	return configInstance, nil
}

// getSearchKeywordsFromEnv retrieves the search keywords from  environment variables.
// It returns a slice of keywords.
func getSearchKeywordsFromEnv() []string {
	keywords := os.Getenv("SEARCH_KEYWORDS")
	if keywords == "" {
		return []string{}
	}
	return split(keywords, ",")
}

// split it splits a string by the provided separator and trims whitespace from each part.
// It returns a slice of strings.
func split(input string, separator string) []string {
	parts := strings.Split(input, separator)
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}
