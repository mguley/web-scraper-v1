package message

import (
	"encoding/json"
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/message/publisher"
	"github.com/mguley/web-scraper-v1/internal/message/shared"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

// setupRabbitMQ initializes a RabbitMQ connection, purges the queue, and creates a new RabbitMQPublisher instance.
// It returns the RabbitMQPublisher instance and a cleanup function to close the connection and channel.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
//
// Returns:
// - *publisher.RabbitMQPublisher[model.Job]: The initialized RabbitMQPublisher instance.
// - func(): A cleanup function to close the RabbitMQ connection and channel.
func setupRabbitMQ(t *testing.T) (*publisher.RabbitMQPublisher[model.Job], func()) {
	config := cfg.RabbitMQ
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s", config.User, config.Pass, config.Host, config.Port)
	connection, channel, err := shared.SetupRabbitMQ(connString, config.QueueName)
	require.NoError(t, err, "Failed to setup RabbitMQ")

	cleanup := func() {
		require.NoError(t, connection.Close(), "Failed to close connection")
		require.NoError(t, channel.Close(), "Failed to close channel")
	}

	// clean up queue
	_, purgeErr := channel.QueuePurge(config.QueueName, false)
	require.NoError(t, purgeErr, "Failed to purge RabbitMQ queue")

	// create a new publisher instance
	factory := publisher.Factory[model.Job]{}
	pub, err := factory.NewPublisher(cfg.RabbitMQ)
	require.NoError(t, err)
	require.NotNil(t, pub)

	return pub.(*publisher.RabbitMQPublisher[model.Job]), cleanup
}

// TestPublishMessage verifies that a message is published to RabbitMQ successfully.
// It publishes test messages and checks if they are correctly added to the queue.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestPublishMessage(t *testing.T) {
	pub, cleanup := setupRabbitMQ(t)
	defer cleanup()

	tests := []struct {
		name    string
		message model.Job
	}{
		{"Test Message", model.Job{Title: "Test Message", URL: "https://www.example.com"}},
		{"Test Message 123", model.Job{Title: "Test Message 123", URL: "https://www.example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publishErr := pub.Publish(tt.message)
			assert.NoError(t, publishErr)

			channel, channelErr := pub.Connection.Channel()
			require.NoError(t, channelErr)
			defer func() {
				require.NoError(t, channel.Close(), "Failed to close channel")
			}()

			msg, ok, getErr := channel.Get(cfg.RabbitMQ.QueueName, true)
			require.NoError(t, getErr)
			require.True(t, ok)

			var receivedMessage model.Job
			unmarshalErr := json.Unmarshal(msg.Body, &receivedMessage)
			require.NoError(t, unmarshalErr)
			assert.Equal(t, tt.message, receivedMessage)
		})
	}
}

// TestConcurrency verifies that messages are published to RabbitMQ concurrently.
// It publishes a specified number of messages concurrently and verifies the total count of messages in the queue.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestConcurrency(t *testing.T) {
	pub, cleanup := setupRabbitMQ(t)
	defer cleanup()

	var waitGroup sync.WaitGroup
	messageCount := 100 // Number of messages to publish concurrently
	waitGroup.Add(messageCount)
	for i := 0; i < messageCount; i++ {
		go func(msgNumber int) {
			defer waitGroup.Done()
			testMessage := model.Job{Title: "Test Message", URL: "https://www.example.com"}
			publishErr := pub.Publish(testMessage)
			assert.NoError(t, publishErr)
		}(i)
	}
	waitGroup.Wait()

	// Verify the number of messages in the queue
	channel, channelErr := pub.Connection.Channel()
	require.NoError(t, channelErr)
	defer func() {
		require.NoError(t, channel.Close(), "Failed to close channel")
	}()

	messages, consumeErr := channel.Consume(
		cfg.RabbitMQ.QueueName,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	require.NoError(t, consumeErr)

	var receivedMessages []model.Job
	for i := 0; i < messageCount; i++ {
		select {
		case message := <-messages:
			var receivedMessage model.Job
			unmarshalErr := json.Unmarshal(message.Body, &receivedMessage)
			require.NoError(t, unmarshalErr)
			receivedMessages = append(receivedMessages, receivedMessage)
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout while waiting for messages")
		}
	}

	require.Len(t, receivedMessages, messageCount, "The number of messages received should be equal to the number of messages sent")
}
