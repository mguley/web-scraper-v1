package message

import "github.com/mguley/web-scraper-v1/internal/model"

// Publisher defines the interface for publishing messages to RabbitMQ.
// This abstraction ensures that job postings are correctly sent to the message broker.
type Publisher interface {
	// Publish sends a job posting message to RabbitMQ.
	// It takes a Job object and returns an error if the publishing fails.
	//
	// Parameters:
	// - message model.Job: The job posting to be published to RabbitMQ.
	//
	// Returns:
	// - error: An error object if there is a failure in publishing the message, otherwise nil.
	Publish(message model.Job) error
}
