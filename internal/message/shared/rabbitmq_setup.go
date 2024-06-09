package shared

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

// SetupRabbitMQ handles the connection and channel setup for RabbitMQ.
//
// Parameters:
// - connString string: The connection string for the RabbitMQ server.
// - queueName string: The name of the queue to declare.
//
// Returns:
// - *amqp091.Connection: A connection to the RabbitMQ server.
// - *amqp091.Channel: A channel for interacting with RabbitMQ.
// - error: An error if there is an issue establishing the connection, creating the channel, or declaring the queue.
func SetupRabbitMQ(connString string, queueName string) (*amqp091.Connection, *amqp091.Channel, error) {
	conn, connErr := amqp091.Dial(connString)
	if connErr != nil {
		defer closeRabbitMQResources(conn, nil)
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ at %s: %w", connString, connErr)
	}

	channel, channelErr := conn.Channel()
	if channelErr != nil {
		defer closeRabbitMQResources(nil, channel)
		return nil, nil, fmt.Errorf("failed to open a channel: %w", channelErr)
	}

	_, queueDeclErr := channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if queueDeclErr != nil {
		return nil, nil, fmt.Errorf("failed to declare queue %s: %w", queueName, queueDeclErr)
	}

	return conn, channel, nil
}

// closeRabbitMQResources closes the RabbitMQ connection and channel, logging any errors encountered.
//
// Parameters:
// - conn *amqp091.Connection: The RabbitMQ connection to close.
// - channel *amqp091.Channel: The RabbitMQ channel to close.
func closeRabbitMQResources(conn *amqp091.Connection, channel *amqp091.Channel) {
	if channel != nil {
		if err := channel.Close(); err != nil {
			log.Printf("failed to close RabbitMQ channel: %s", err)
		}
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close RabbitMQ connection: %s", err)
		}
	}
}
