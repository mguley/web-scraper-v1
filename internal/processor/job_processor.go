package processor

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/message/publisher"
	"github.com/mguley/web-scraper-v1/internal/parser"
	"io"
	"net/http"
)

// JobProcessor is a generic type that orchestrates the process of fetching, parsing and publishing data.
// It utilizes a configurable HTTP client for making web requests, a parser to transform HTML content into
// structured data, and a publisher to send the data to a RabbitMQ message broker.
//
// Type parameter:
//   - T any: The type of data that the JobProcessor works with. This ensures that the processor, parser, and publisher
//     all operate on the same data type, providing type safety throughout the processing pipeline.
//
// Fields:
// - HttpClient *http.Client: The HTTP client used to fetch web pages.
// - Publisher publisher.Publisher[T]: The publisher used to send data to RabbitMQ.
// - Parser parser.Parser[T]: The parser used to convert raw HTML into a structured data type T.
//
// Methods:
// - NewJobProcessor: Constructs a new JobProcessor with a specified HTTP client, RabbitMQ configuration, and parser.
// - Process: Handles the complete processing pipeline from fetching a URL to publishing the parsed data.
type JobProcessor[T any] struct {
	HttpClient *http.Client
	Publisher  publisher.Publisher[T]
	Parser     parser.Parser[T]
}

// NewJobProcessor creates a new instance of JobProcessor. It sets up the necessary components for the data processing
// pipeline, including a configured HTTP client, a publisher set up via the publisher.Factory, and a parser.
//
// Parameters:
// - client *http.Client: A pre-configured HTTP client used for making web requests.
// - brokerConfig config.RabbitMQ: Configuration settings for RabbitMQ to set up the publisher.
// - parser parser.Parser[T]: A parser that converts HTML content to a data type T.
//
// Returns:
// - *JobProcessor[T]: A pointer to an instance of JobProcessor.
// - error: An error that might occur during the setup of the publisher.
func NewJobProcessor[T any](client *http.Client, brokerConfig config.RabbitMQ, parser parser.Parser[T]) (*JobProcessor[T], error) {
	factory := publisher.Factory[T]{}
	pub, err := factory.NewPublisher(brokerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &JobProcessor[T]{
		HttpClient: client,
		Publisher:  pub,
		Parser:     parser,
	}, nil
}

// Process fetches data from a specified URL, parses the HTML content into a structured form of type T, and
// publishes it to a RabbitMQ broker. It manages error handling throughout the process, ensuring each
// step's success before proceeding to the next.
//
// Parameters:
// - url string: The URL of the web page to fetch and process.
//
// Returns:
// - *T: A pointer to the structured data type T if processing is successful.
// - error: An error detailing any issue that occurs during fetching, parsing, or publishing.
func (jobProcessor *JobProcessor[T]) Process(url string) (*T, error) {
	data, err := jobProcessor.fetchData(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	parsedData, parseErr := jobProcessor.Parser.Parse(string(data))
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse data: %w", parseErr)
	}

	if publishErr := jobProcessor.Publisher.Publish(*parsedData); publishErr != nil {
		return nil, fmt.Errorf("failed to publish data: %w", publishErr)
	}

	return parsedData, nil
}

// fetchData handles the HTTP GET request to fetch the webpage content. It checks the HTTP status code and reads
// the response body if the request is successful.
//
// Parameters:
// - url string: The URL to fetch the data from.
//
// Returns:
// - []byte: The raw HTML content of the fetched web page.
// - error: An error object indicating a failure in the fetch operation.
func (jobProcessor *JobProcessor[T]) fetchData(url string) ([]byte, error) {
	response, err := jobProcessor.HttpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			return
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %d", response.StatusCode)
	}
	return io.ReadAll(response.Body)
}
