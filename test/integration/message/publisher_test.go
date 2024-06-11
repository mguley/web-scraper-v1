package message

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/message/publisher"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

// setupPublisher initializes a RabbitMQPublisher for publishing messages.
// It also purges any messages that might be lingering in the queue to ensure tests start clean.
// Returns the publisher instance along with a cleanup function to properly release resources after tests.
//
// Parameters:
// - t *testing.T: The testing framework's instance, used for making assertions.
// - deleteQueue bool: A flag that indicates if we need to drop a queue.
//
// Returns:
// - *publisher.RabbitMQPublisher[model.Job]: The initialized publisher.
// - func(): A cleanup function to close the connection and delete the queue.
func setupPublisher(t *testing.T, deleteQueue bool) (*publisher.RabbitMQPublisher[model.Job], func()) {
	factory := publisher.Factory[model.Job]{}
	pub, pubErr := factory.NewPublisher(cfg.RabbitMQ)
	require.NoError(t, pubErr)
	require.NotNil(t, pub)

	publisherRabbitMQ := pub.(*publisher.RabbitMQPublisher[model.Job])

	// Clean up queue
	_, purgeErr := publisherRabbitMQ.Channel.QueuePurge(cfg.RabbitMQ.QueueName, false)
	require.NoError(t, purgeErr)

	releaseResources := func() {
		if deleteQueue {
			_, delErr := publisherRabbitMQ.Channel.QueueDelete(cfg.RabbitMQ.QueueName, false, false, false)
			require.NoError(t, delErr)
		}
		require.NoError(t, publisherRabbitMQ.Close())
	}

	return publisherRabbitMQ, releaseResources
}

// TestPublishMessages verifies that messages can be published concurrently to a RabbitMQ queue.
// It sets up a publisher, sends a large number of messages, and ensures all are successfully added.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestPublishMessages(t *testing.T) {
	pub, releaseResources := setupPublisher(t, true)
	defer releaseResources()
	require.NotNil(t, pub)

	var waitGroup sync.WaitGroup
	messageCount := 10_000 // Number of messages to publish concurrently
	list := make([]model.Job, messageCount)

	// Start publishing messages concurrently
	waitGroup.Add(messageCount)
	for i := 0; i < messageCount; i++ {
		go func(id int) {
			defer waitGroup.Done()
			item := model.Job{Title: fmt.Sprintf("message payload %d", id), URL: "https://www.ex.com"}
			pubErr := pub.Publish(item)
			require.NoError(t, pubErr)
			list[id] = item
		}(i)
	}
	waitGroup.Wait()
	require.Equal(t, messageCount, len(list))
}

// TestCloseFunctionPublisher verifies the proper closure of RabbitMQ connections and channels used by the publisher.
// It checks that the publisher's cleanup operations (closing connections and deleting queues) are error-free.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCloseFunctionPublisher(t *testing.T) {
	pub, _ := setupPublisher(t, true)
	require.NotNil(t, pub)

	// Ensure that the RabbitMQ connection and channel are properly opened
	require.NotNil(t, pub.Connection, "RabbitMQ connection should be established")
	require.NotNil(t, pub.Channel, "RabbitMQ channel should be established")

	// Attempt to delete the queue and verify it is handled without errors
	_, delErr := pub.Channel.QueueDelete(cfg.RabbitMQ.QueueName, false, false, false)
	require.NoError(t, delErr, "Failed to delete queue")

	// Close the publisher and ensure there are no errors
	closeErr := pub.Close()
	assert.NoError(t, closeErr, "Closing the publisher should not produce an error")

	// Check that the connection and channel are no longer operational
	connErr := pub.Connection.Close()
	assert.Error(t, connErr, "The connection should already be closed")
	channelErr := pub.Channel.Close()
	assert.Nil(t, channelErr, "The channel should already be closed")
}
