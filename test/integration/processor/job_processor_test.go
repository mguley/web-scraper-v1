package processor

import (
	"context"
	"github.com/mguley/web-scraper-v1/internal/message/publisher"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/mguley/web-scraper-v1/internal/processor"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

// DummyParser is a mock parser for testing purposes.
// It simulates the behavior of a real parser by returning a predefined model.Job object.
type DummyParser struct{}

// Parse simulates parsing of HTML content and returns a dummy model.Job object.
// In a real scenario, this method would extract data from HTML using parsing logic.
//
// Parameters:
// - html string: The HTML content to be parsed.
//
// Returns:
// - *model.Job: A pointer to the resulting model.Job object with predefined or extracted data.
// - error: Always returns nil in this dummy implementation.
func (parser *DummyParser) Parse(html string) (*model.Job, error) {
	// For demonstration, we return a static job object. In practice, you'd parse the HTML content.
	return &model.Job{
		ID:          "12345",
		Title:       "Software Engineer",
		Company:     "Tech Corp",
		Description: "Join our tech team to develop innovative solutions.",
		URL:         "https://ex.com/job/12345",
	}, nil
}

// setupJobProcessor sets up a JobProcessor instance with a dummy parser and a real RabbitMQ publisher.
// It also returns cleanup functions and test URLs.
//
// Parameters:
// - t *testing.T: The testing context.
//
// Returns:
// - *processor.JobProcessor[model.Job]: The initialized JobProcessor instance.
// - func(): A cleanup function to delete the queue and close the publisher connection.
// - string: A URL for testing successful processing.
// - string: A URL for testing request timeout.
func setupJobProcessor(t *testing.T) (*processor.JobProcessor[model.Job], func(), string, string) {
	pubFactory := publisher.Factory[model.Job]{}
	pub, pubErr := pubFactory.NewPublisher(cfg.RabbitMQ)
	require.NoError(t, pubErr, "Failed to create publisher")
	require.NotNil(t, pub)

	dummyParser := &DummyParser{}
	publisherRabbitMQ := pub.(*publisher.RabbitMQPublisher[model.Job])
	client := &http.Client{}

	jobProcessor, err := processor.NewJobProcessor[model.Job](client, cfg.RabbitMQ, dummyParser)
	require.NoError(t, err)
	require.NotNil(t, jobProcessor)

	cleanup := func() {
		_, delErr := publisherRabbitMQ.Channel.QueueDelete(cfg.RabbitMQ.QueueName, false, false, false)
		require.NoError(t, delErr, "Failed to delete queue")
		require.NoError(t, publisherRabbitMQ.Close())
	}

	return jobProcessor, cleanup, "https://example.com/", "https://httpbin.org/delay/5"
}

// TestJobProcessorIntegration tests the complete processing pipeline from fetching data, parsing it, and publishing it.
// This test uses a real URL to validate the integration of different components.
//
// Parameters:
// - t *testing.T: The testing context.
func TestJobProcessorIntegration(t *testing.T) {
	jobProcessor, cleanup, url, _ := setupJobProcessor(t)
	defer cleanup()
	require.NotNil(t, jobProcessor)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	processedData, processErr := jobProcessor.Process(ctx, url)
	require.NoError(t, processErr)
	require.NotNil(t, processedData)

	// Asserted the parsed data
	require.Equal(t, "12345", processedData.ID)
	require.Equal(t, "Software Engineer", processedData.Title)
}

// TestJobProcessorCancellation tests the JobProcessor's response to context cancellation.
// It ensures the JobProcessor stops processing when the context is cancelled.
//
// Parameters:
// - t *testing.T: The testing context.
func TestJobProcessorCancellation(t *testing.T) {
	jobProcessor, cleanup, url, _ := setupJobProcessor(t)
	defer cleanup()
	require.NotNil(t, jobProcessor)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, processErr := jobProcessor.Process(ctx, url)
	require.Error(t, processErr)
	require.Equal(t, context.DeadlineExceeded, ctx.Err())
}

// TestJobProcessorTimeout tests the JobProcessor's handling of HTTP request timeouts.
// It ensures the JobProcessor properly handles timeouts when fetching data.
//
// Parameters:
// - t *testing.T: The testing context.
func TestJobProcessorTimeout(t *testing.T) {
	jobProcessor, cleanup, _, delayUrl := setupJobProcessor(t)
	defer cleanup()
	require.NotNil(t, jobProcessor)
	require.NotEmpty(t, delayUrl)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, processErr := jobProcessor.Process(ctx, delayUrl)
	require.Error(t, processErr)
	require.Equal(t, context.DeadlineExceeded, ctx.Err())
}
