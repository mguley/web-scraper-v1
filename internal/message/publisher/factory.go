package publisher

import "github.com/mguley/web-scraper-v1/config"

// Factory creates publishers based on configuration.
type Factory[T any] struct{}

// NewPublisher creates a new instance of a Publisher based on the provided RabbitMQ configuration.
//
// Parameters:
// - appConfig config.RabbitMQ: The RabbitMQ configuration settings.
//
// Returns:
// - Publisher[T]: A generic Publisher interface instance.
// - error: An error if there is an issue creating the publisher instance.
func (publisherFactory *Factory[T]) NewPublisher(appConfig config.RabbitMQ) (Publisher[T], error) {
	return NewRabbitMQPublisher[T](appConfig)
}
