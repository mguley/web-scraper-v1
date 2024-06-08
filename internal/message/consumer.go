package message

import "github.com/mguley/web-scraper-v1/internal/model"

// Consumer defines the interface for consuming messages from RabbitMQ.
// This abstraction ensures that job postings are correctly received from the message broker.
type Consumer interface {
	// Consume starts the consumption of job posting messages from RabbitMQ.
	// It returns a channel of Job objects and an error if the consumption setup fails.
	//
	// Returns:
	// - <-chan model.Job: A receive-only channel from which Job objects can be read.
	// - error: An error object if there is a failure in setting up the consumer, otherwise nil.
	Consume() (<-chan model.Job, error)
}
