package message

import (
	"encoding/json"
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/message/shared"
	"github.com/mguley/web-scraper-v1/internal/utils"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQConsumer implements the Consumer interface for RabbitMQ.
// It provides functionality to consume messages from a RabbitMQ queue.
type RabbitMQConsumer[T any] struct {
	Connection *amqp091.Connection // Connection to the RabbitMQ server.
	Channel    *amqp091.Channel    // Channel for interacting with RabbitMQ.
	QueueName  string              // Name of the queue to consume messages from.
	logger     utils.Logger        // Logger instance for logging messages and errors.
}

// NewRabbitMQConsumer creates a new instance of RabbitMQConsumer.
// It establishes a connection and a channel to the RabbitMQ server and declares the specified queue.
//
// Parameters:
// - appConfig config.RabbitMQ: The RabbitMQ configuration settings.
//
// Returns:
// - *RabbitMQConsumer[T]: A pointer to an instance of RabbitMQConsumer.
// - error: An error if there is an issue establishing the connection, creating the channel, or declaring the queue.
func NewRabbitMQConsumer[T any](appConfig config.RabbitMQ) (*RabbitMQConsumer[T], error) {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s", appConfig.User, appConfig.Pass, appConfig.Host, appConfig.Port)
	connection, channel, err := shared.SetupRabbitMQ(connString, appConfig.QueueName)
	if err != nil {
		return nil, fmt.Errorf("failed to set up RabbitMQ: %w", err)
	}

	return &RabbitMQConsumer[T]{
		Connection: connection,
		Channel:    channel,
		QueueName:  appConfig.QueueName,
		logger:     utils.GetLogger(),
	}, nil
}

// Consume starts consuming messages from the RabbitMQ queue.
//
// Returns:
// - <-chan T: A receive-only channel from which T objects can be read.
// - error: An error if there is an issue setting up the consumer.
func (consumer *RabbitMQConsumer[T]) Consume() (<-chan T, error) {
	messages, err := consumer.Channel.Consume(
		consumer.QueueName, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		consumer.logError(fmt.Errorf("failed to start consuming messages: %w", err))
		return nil, err
	}

	out := make(chan T)
	go consumer.handleMessages(messages, out)
	return out, nil
}

// handleMessages processes incoming messages and sends them to the output channel.
//
// Parameters:
// - messages <-chan amqp091.Delivery: Channel of messages to be consumed.
// - out chan<- T: Channel to output unmarshalled messages.
func (consumer *RabbitMQConsumer[T]) handleMessages(messages <-chan amqp091.Delivery, out chan<- T) {
	defer close(out)
	for msg := range messages {
		var item T
		if err := json.Unmarshal(msg.Body, &item); err != nil {
			consumer.logError(fmt.Errorf("failed to unmarshal message: %w", err))
			continue
		}
		out <- item
		consumer.logInfo("Message pushed to channel")
	}
}

// Close gracefully closes the RabbitMQ connection and channel.
// It ensures that all resources are properly released.
//
// This method should be called when the RabbitMQConsumer instance is no longer needed, to ensure
// that connections are properly closed and resources are freed.
//
// Returns:
// - error: An error object if there is a failure in closing the connection or channel, otherwise nil.
func (consumer *RabbitMQConsumer[T]) Close() error {
	var errors []error

	if closeChannelErr := consumer.Channel.Close(); closeChannelErr != nil {
		consumer.logError(fmt.Errorf("failed to close RabbitMQ channel: %w", closeChannelErr))
		errors = append(errors, closeChannelErr)
	}

	if closeConnErr := consumer.Channel.Close(); closeConnErr != nil {
		consumer.logError(fmt.Errorf("failed to close RabbitMQ connection: %w", closeConnErr))
		errors = append(errors, closeConnErr)
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred while closing resources: %v", errors)
	}

	return nil
}

// logError logs an error message using the consumer's logger.
//
// Parameters:
// - err error: The error to log.
func (consumer *RabbitMQConsumer[T]) logError(err error) {
	consumer.logger.LogError(err)
}

// logInfo logs an informational message using the consumer's logger.
//
// Parameters:
// - message string: The message to log.
func (consumer *RabbitMQConsumer[T]) logInfo(message string) {
	consumer.logger.LogInfo(message)
}
