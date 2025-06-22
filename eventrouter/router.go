package eventrouter

import (
	"context"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/duongptryu/gox/syserr"
)

// Config holds router configuration
type Config struct {
	Logger       *slog.Logger
	CloseTimeout *time.Duration
}

// messageRouter implements the Router interface using Watermill's message.Router
type messageRouter struct {
	router *message.Router
	logger *slog.Logger
	config Config
}

// NewRouter creates a new message router
func NewRouter(config Config) (Router, error) {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	wmLogger := watermill.NewSlogLogger(config.Logger)

	routerConfig := message.RouterConfig{}
	if config.CloseTimeout != nil {
		routerConfig.CloseTimeout = *config.CloseTimeout
	}

	router, err := message.NewRouter(routerConfig, wmLogger)
	if err != nil {
		return nil, syserr.WrapAsIs(err, "failed to create watermill router")
	}

	return &messageRouter{
		router: router,
		logger: config.Logger,
		config: config,
	}, nil
}

// AddHandler adds a message handler that can transform and publish messages to another topic
func (r *messageRouter) AddHandler(handlerName, subscribeTopic, publishTopic string, subscriber message.Subscriber, publisher message.Publisher, handler MessageHandler) error {
	if r.router.IsRunning() {
		return syserr.New(syserr.InvalidArgumentCode, "cannot add handler to running router")
	}

	// Convert our MessageHandler to Watermill's HandlerFunc
	handlerFunc := func(msg *message.Message) ([]*message.Message, error) {
		err := handler.Handle(msg.Context(), msg)
		if err != nil {
			return nil, err
		}
		// For handlers that transform messages, they should create new messages and return them
		// This is a simple pass-through, but users can implement their own logic
		return []*message.Message{msg}, nil
	}

	r.router.AddHandler(handlerName, subscribeTopic, subscriber, publishTopic, publisher, handlerFunc)
	return nil
}

// AddNoPublisherHandler adds a handler that only consumes messages without publishing
func (r *messageRouter) AddNoPublisherHandler(handlerName, subscribeTopic string, subscriber message.Subscriber, handler MessageHandler) error {
	if r.router.IsRunning() {
		return syserr.New(syserr.InvalidArgumentCode, "cannot add handler to running router")
	}

	// Convert our MessageHandler to Watermill's NoPublishHandlerFunc
	handlerFunc := func(msg *message.Message) error {
		return handler.Handle(msg.Context(), msg)
	}

	r.router.AddNoPublisherHandler(handlerName, subscribeTopic, subscriber, handlerFunc)
	return nil
}

// AddPlugin adds a router plugin
func (r *messageRouter) AddPlugin(plugin message.RouterPlugin) {
	r.router.AddPlugin(plugin)
}

// AddMiddleware adds middleware to the router
func (r *messageRouter) AddMiddleware(middleware message.HandlerMiddleware) {
	r.router.AddMiddleware(middleware)
}

// Run starts the router (blocking)
func (r *messageRouter) Run(ctx context.Context) error {
	r.logger.Info("Starting message router")

	err := r.router.Run(ctx)
	if err != nil {
		return syserr.WrapAsIs(err, "router run failed")
	}

	r.logger.Info("Message router stopped")
	return nil
}

// Running returns true if the router is running
func (r *messageRouter) Running() bool {
	return r.router.IsRunning()
}

// Close closes the router
func (r *messageRouter) Close() error {
	r.logger.Info("Closing message router")

	err := r.router.Close()
	if err != nil {
		return syserr.WrapAsIs(err, "failed to close router")
	}

	r.logger.Info("Message router closed")
	return nil
}

// Publisher wraps a Watermill publisher
type messagePublisher struct {
	publisher message.Publisher
	logger    *slog.Logger
}

// NewPublisher creates a new publisher wrapper
func NewPublisher(publisher message.Publisher, logger *slog.Logger) Publisher {
	if logger == nil {
		logger = slog.Default()
	}

	return &messagePublisher{
		publisher: publisher,
		logger:    logger,
	}
}

// Publish publishes a message to a topic
func (p *messagePublisher) Publish(ctx context.Context, topic string, msg *message.Message) error {
	p.logger.Debug("Publishing message",
		slog.String("topic", topic),
		slog.String("message_id", msg.UUID))

	err := p.publisher.Publish(topic, msg)
	if err != nil {
		return syserr.WrapAsIs(err, "failed to publish message",
			syserr.F("topic", topic),
			syserr.F("message_id", msg.UUID))
	}

	return nil
}

// Close closes the publisher
func (p *messagePublisher) Close() error {
	return p.publisher.Close()
}

// Subscriber wraps a Watermill subscriber
type messageSubscriber struct {
	subscriber message.Subscriber
	logger     *slog.Logger
}

// NewSubscriber creates a new subscriber wrapper
func NewSubscriber(subscriber message.Subscriber, logger *slog.Logger) Subscriber {
	if logger == nil {
		logger = slog.Default()
	}

	return &messageSubscriber{
		subscriber: subscriber,
		logger:     logger,
	}
}

// Subscribe subscribes to a topic
func (s *messageSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	s.logger.Debug("Subscribing to topic", slog.String("topic", topic))

	messages, err := s.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return nil, syserr.WrapAsIs(err, "failed to subscribe to topic", syserr.F("topic", topic))
	}

	return messages, nil
}

// Close closes the subscriber
func (s *messageSubscriber) Close() error {
	return s.subscriber.Close()
}
