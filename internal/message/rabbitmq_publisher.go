package message

import (
	"encoding/json"
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/message/shared"
	"github.com/mguley/web-scraper-v1/internal/utils"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implements the Publisher interface for RabbitMQ.
// It provides functionality to publish messages to a RabbitMQ queue.
type RabbitMQPublisher[T any] struct {
	Connection *amqp091.Connection // Connection to the RabbitMQ server.
	Channel    *amqp091.Channel    // Channel for interacting with RabbitMQ.
	QueueName  string              // Name of the queue to publish messages to.
	logger     utils.Logger        // Logger instance for logging messages and errors.
}

// NewRabbitMQPublisher creates a new instance of RabbitMQPublisher.
// It establishes a connection and a channel to the RabbitMQ server and declares the specified queue.
//
// Parameters:
// - appConfig config.RabbitMQ: The RabbitMQ configuration settings.
//
// Returns:
// - *RabbitMQPublisher[T]: A pointer to an instance of RabbitMQPublisher.
// - error: An error if there is an issue establishing the connection, creating the channel, or declaring the queue.
func NewRabbitMQPublisher[T any](appConfig config.RabbitMQ) (*RabbitMQPublisher[T], error) {
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s", appConfig.User, appConfig.Pass, appConfig.Host, appConfig.Port)
	connection, channel, err := shared.SetupRabbitMQ(connString, appConfig.QueueName)
	if err != nil {
		return nil, fmt.Errorf("failed to set up RabbitMQ: %w", err)
	}

	return &RabbitMQPublisher[T]{
		Connection: connection,
		Channel:    channel,
		QueueName:  appConfig.QueueName,
		logger:     utils.GetLogger(),
	}, nil
}

// Publish sends a message to the RabbitMQ queue.
// It marshals the message to JSON and publishes it to the queue.
//
// Parameters:
// - message T: The message to be published to RabbitMQ.
//
// Returns:
// - error: An error object if there is a failure in publishing the message, otherwise nil.
func (publisher *RabbitMQPublisher[T]) Publish(message T) error {
	body, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		publisher.logError(fmt.Errorf("failed to marshal message: %w", marshalErr))
		return marshalErr
	}

	clientMessage := amqp091.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	publishErr := publisher.Channel.Publish(
		"",                  // exchange
		publisher.QueueName, // routing key
		false,               // mandatory
		false,               // immediate
		clientMessage,
	)

	if publishErr != nil {
		publisher.logError(fmt.Errorf("failed to publish message: %w", publishErr))
		return publishErr
	}

	publisher.logInfo("Message published to queue")
	return nil
}

// Close gracefully closes the RabbitMQ connection and channel.
// It ensures that all resources are properly released.
//
// This method should be called when the RabbitMQPublisher instance is no longer needed, to ensure
// that connections are properly closed and resources are freed.
//
// Returns:
// - error: An error object if there is a failure in closing the connection or channel, otherwise nil.
func (publisher *RabbitMQPublisher[T]) Close() error {
	var errors []error

	if closeChannelErr := publisher.Channel.Close(); closeChannelErr != nil {
		publisher.logError(fmt.Errorf("failed to close RabbitMQ channel: %w", closeChannelErr))
		errors = append(errors, closeChannelErr)
	}

	if closeConnErr := publisher.Connection.Close(); closeConnErr != nil {
		publisher.logError(fmt.Errorf("failed to close RabbitMQ connection: %w", closeConnErr))
		errors = append(errors, closeConnErr)
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple errors occurred while closing resources: %v", errors)
	}

	return nil
}

// logError logs an error message using the publisher's logger.
//
// Parameters:
// - err error: The error to log.
func (publisher *RabbitMQPublisher[T]) logError(err error) {
	publisher.logger.LogError(err)
}

// logInfo logs an informational message using the publisher's logger.
//
// Parameters:
// - message string: The message to log.
func (publisher *RabbitMQPublisher[T]) logInfo(message string) {
	publisher.logger.LogInfo(message)
}
