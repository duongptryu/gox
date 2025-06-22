package eventrouter

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
)

// MessageHandler handles incoming messages
type MessageHandler interface {
	Handle(ctx context.Context, msg *message.Message) error
}

// MessageHandlerFunc is a function type that implements MessageHandler
type MessageHandlerFunc func(ctx context.Context, msg *message.Message) error

// Handle implements MessageHandler interface
func (f MessageHandlerFunc) Handle(ctx context.Context, msg *message.Message) error {
	return f(ctx, msg)
}

// Router interface for message routing
type Router interface {
	// AddHandler adds a message handler for a specific topic
	AddHandler(handlerName, subscribeTopic, publishTopic string, subscriber message.Subscriber, publisher message.Publisher, handler MessageHandler) error

	// AddNoPublisherHandler adds a handler that doesn't publish to another topic
	AddNoPublisherHandler(handlerName, subscribeTopic string, subscriber message.Subscriber, handler MessageHandler) error

	// AddPlugin adds a router plugin
	AddPlugin(plugin message.RouterPlugin)

	// AddMiddleware adds middleware to the router
	AddMiddleware(middleware message.HandlerMiddleware)

	// Run starts the router (blocking)
	Run(ctx context.Context) error

	// Running returns true if the router is running
	Running() bool

	// Close closes the router
	Close() error
}

// Publisher interface for publishing messages
type Publisher interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, msg *message.Message) error

	// Close closes the publisher
	Close() error
}

// Subscriber interface for subscribing to messages
type Subscriber interface {
	// Subscribe subscribes to a topic
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)

	// Close closes the subscriber
	Close() error
}
