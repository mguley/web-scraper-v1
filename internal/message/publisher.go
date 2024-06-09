package message

// Publisher defines the interface for publishing messages to RabbitMQ.
// This abstraction ensures that job postings are correctly sent to the message broker.
type Publisher[T any] interface {
	// Publish sends a message to RabbitMQ.
	// It takes an object of type T and returns an error if the publishing fails.
	//
	// Parameters:
	// - message T: The message to be published to RabbitMQ.
	//
	// Returns:
	// - error: An error object if there is a failure in publishing the message, otherwise nil.
	Publish(message T) error

	// Close gracefully closes the connection and channel.
	// It ensures that any resources associated with the publisher are properly released.
	//
	// Returns:
	// - error: An error object if there is a failure in closing the publisher, otherwise nil.
	Close() error
}
