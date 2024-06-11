package message

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/message/consumer"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

// setupConsumer initializes a RabbitMQConsumer for consuming messages.
// It optionally deletes the queue if specified.
// Returns the consumer instance along with a cleanup function to properly release resources after tests.
//
// Parameters:
// - t *testing.T: The testing framework's instance, used for making assertions.
// - deleteQueue bool: Flag indicating whether to delete the queue during cleanup.
//
// Returns:
// - *consumer.RabbitMQConsumer[model.Job]: The initialized consumer.
// - func(): A cleanup function that deletes the queue if required and closes the consumer.
func setupConsumer(t *testing.T, deleteQueue bool) (*consumer.RabbitMQConsumer[model.Job], func()) {
	factory := consumer.Factory[model.Job]{}
	cons, consErr := factory.NewConsumer(cfg.RabbitMQ)
	require.NoError(t, consErr)
	require.NotNil(t, cons)

	consumerRabbitMQ := cons.(*consumer.RabbitMQConsumer[model.Job])

	releaseResources := func() {
		if deleteQueue {
			_, delErr := consumerRabbitMQ.Channel.QueueDelete(cfg.RabbitMQ.QueueName, false, false, false)
			require.NoError(t, delErr)
		}
		require.NoError(t, consumerRabbitMQ.Close())
	}

	return consumerRabbitMQ, releaseResources
}

// TestConsumeMessages tests the end-to-end process of publishing and then consuming messages.
// It verifies that all published messages are successfully consumed and that the counts match.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestConsumeMessages(t *testing.T) {
	pub, releasePubResources := setupPublisher(t, true)
	defer releasePubResources()
	require.NotNil(t, pub)

	var waitGroup sync.WaitGroup
	messageCount := 5_000 // Number of messages to publish concurrently
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
	require.Equal(t, messageCount, len(list), "All messages should be published")

	cons, releaseConsumerResources := setupConsumer(t, true)
	defer releaseConsumerResources()
	require.NotNil(t, cons)

	messageChannel, err := cons.Consume()
	require.NoError(t, err)
	require.NotNil(t, messageChannel)
	var receivedItems []model.Job
	var lock sync.Mutex

	// Start consuming messages
	waitGroup.Add(messageCount)
	for i := 0; i < messageCount; i++ {
		go func() {
			defer waitGroup.Done()
			select {
			case item := <-messageChannel:
				lock.Lock()
				receivedItems = append(receivedItems, item)
				lock.Unlock()
				if len(receivedItems) == messageCount {
					return
				}
			case <-time.After(time.Second * 10):
				require.Fail(t, "Timeout waiting for message")
			}
		}()
	}
	waitGroup.Wait()
	require.Equal(t, messageCount, len(receivedItems), "The received count should be the same as the sent count of messages")
}

// TestCloseFunctionConsumer verifies that the Close function properly cleans up and closes the RabbitMQ connection and channel.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestCloseFunctionConsumer(t *testing.T) {
	cons, _ := setupConsumer(t, true)
	require.NotNil(t, cons)

	// Ensure that the RabbitMQ connection and channel are properly opened
	require.NotNil(t, cons.Connection, "RabbitMQ connection should be established")
	require.NotNil(t, cons.Channel, "RabbitMQ channel should be established")

	// Attempt to delete the queue and verify it is handled without errors
	_, delErr := cons.Channel.QueueDelete(cfg.RabbitMQ.QueueName, false, false, false)
	require.NoError(t, delErr, "Failed to delete queue")

	// Close the consumer and ensure there are no errors
	closeErr := cons.Close()
	require.NoError(t, closeErr, "Closing the consumer should not produce an error")

	// Check that the connection and channel are no longer operational
	connErr := cons.Connection.Close()
	require.Nil(t, connErr, "The connection should already be closed")
	channelErr := cons.Channel.Close()
	require.Nil(t, channelErr, "The channel should already be closed")
}
