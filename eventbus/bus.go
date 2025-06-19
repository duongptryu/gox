package eventbus

import (
	"context"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

// Config holds configuration for the event bus.
type Config struct {
	Publisher  message.Publisher
	Subscriber message.Subscriber
	Logger     *slog.Logger
}

// cqrsBus implements the Bus interface using Watermill CQRS.
type cqrsBus struct {
	commandBus       *cqrs.CommandBus
	eventBus         *cqrs.EventBus
	commandProcessor *cqrs.CommandProcessor
	eventProcessor   *cqrs.EventProcessor
	router           *message.Router
	logger           *slog.Logger
	marshaler        cqrs.CommandEventMarshaler
}

// NewBus creates a new CQRS event bus.
func NewBus(cfg Config) (Bus, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	wmLogger := watermill.NewSlogLogger(cfg.Logger)
	marshaler := cqrs.JSONMarshaler{
		GenerateName: cqrs.StructName,
	}

	router, err := message.NewRouter(message.RouterConfig{}, wmLogger)
	if err != nil {
		return nil, err
	}

	commandBus, err := cqrs.NewCommandBusWithConfig(cfg.Publisher, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
	})
	if err != nil {
		return nil, err
	}

	eventBus, err := cqrs.NewEventBusWithConfig(cfg.Publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
	})
	if err != nil {
		return nil, err
	}

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(router, cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return cfg.Subscriber, nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
	})
	if err != nil {
		return nil, err
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return "events." + params.EventName, nil
		},
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return cfg.Subscriber, nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
	})
	if err != nil {
		return nil, err
	}

	return &cqrsBus{
		commandBus:       commandBus,
		eventBus:         eventBus,
		commandProcessor: commandProcessor,
		eventProcessor:   eventProcessor,
		router:           router,
		logger:           cfg.Logger,
		marshaler:        marshaler,
	}, nil
}

func (b *cqrsBus) PublishCommand(ctx context.Context, cmd Command) error {
	return b.commandBus.Send(ctx, cmd)
}

func (b *cqrsBus) PublishEvent(ctx context.Context, evt Event) error {
	return b.eventBus.Publish(ctx, evt)
}

func (b *cqrsBus) RegisterCommandHandler(commandName string, handler CommandHandler) error {
	_, err := b.commandProcessor.AddHandler(cqrs.NewCommandHandler(commandName, handler.Handle))
	if err != nil {
		return err
	}

	return nil
}

func (b *cqrsBus) RegisterEventHandler(eventName string, handler EventHandler) error {
	_, err := b.eventProcessor.AddHandler(cqrs.NewEventHandler(eventName, handler.Handle))
	if err != nil {
		return err
	}

	return nil
}

func (b *cqrsBus) Run(ctx context.Context) error {
	return b.router.Run(ctx)
}
