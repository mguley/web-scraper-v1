package message

// Consumer defines the interface for consuming messages from RabbitMQ.
// This interface abstracts the process of receiving job postings from the message broker.
type Consumer[T any] interface {
	// Consume starts the consumption of messages from RabbitMQ.
	// It returns a channel of T objects and an error if the consumption setup fails.
	//
	// Returns:
	// - <-chan T: A receive-only channel from which T objects can be read.
	// - error: An error object if there is a failure in setting up the consumer, otherwise nil.
	Consume() (<-chan T, error)

	// Close gracefully closes the connection and channel.
	// It ensures that any resources associated with the consumer are properly released.
	//
	// Returns:
	// - error: An error object if there is a failure in closing the consumer, otherwise nil.
	Close() error
}
