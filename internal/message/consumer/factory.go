package consumer

import "github.com/mguley/web-scraper-v1/config"

// Factory creates consumers based on configuration.
type Factory[T any] struct{}

// NewConsumer creates a new instance of a Consumer based on the provided RabbitMQ configuration.
//
// Parameters:
// - appConfig config.RabbitMQ: The RabbitMQ configuration settings.
//
// Returns:
// - Consumer[T]: A generic Consumer interface instance.
// - error: An error if there is an issue creating the consumer instance.
func (consumerFactory *Factory[T]) NewConsumer(appConfig config.RabbitMQ) (Consumer[T], error) {
	return NewRabbitMQConsumer[T](appConfig)
}
