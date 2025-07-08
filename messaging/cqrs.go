package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/duongptryu/gox/logger"
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
func NewBus(cfg Config) (*cqrsBus, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	generateEventTopic := func(eventName string) string {
		return fmt.Sprintf("events.%s", eventName)
	}

	generateCommandTopic := func(commandName string) string {
		return fmt.Sprintf("commands.%s", commandName)
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
			return generateCommandTopic(params.CommandName), nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
		OnSend: func(params cqrs.CommandBusOnSendParams) error {
			logger.Info(params.Message.Context(), "Sending command", logger.F("command_name", params.CommandName))
			params.Message.Metadata.Set("sent_at", time.Now().String())
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	eventBus, err := cqrs.NewEventBusWithConfig(cfg.Publisher, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return generateEventTopic(params.EventName), nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
		OnPublish: func(params cqrs.OnEventSendParams) error {
			logger.Info(params.Message.Context(), "Publishing event", logger.F("event_name", params.EventName))
			params.Message.Metadata.Set("published_at", time.Now().String())
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	commandProcessor, err := cqrs.NewCommandProcessorWithConfig(router, cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return generateCommandTopic(params.CommandName), nil
		},
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return cfg.Subscriber, nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
		OnHandle: func(params cqrs.CommandProcessorOnHandleParams) error {
			start := time.Now()

			err := params.Handler.Handle(params.Message.Context(), params.Command)

			logger.Info(params.Message.Context(), "Command handled",
				logger.F("command_name", params.CommandName),
				logger.F("duration", time.Since(start)),
				logger.F("err", err),
			)

			return err
		},
	})
	if err != nil {
		return nil, err
	}

	eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
			return generateEventTopic(params.EventName), nil
		},
		SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return cfg.Subscriber, nil
		},
		Marshaler: marshaler,
		Logger:    wmLogger,
		OnHandle: func(params cqrs.EventProcessorOnHandleParams) error {
			start := time.Now()

			err := params.Handler.Handle(params.Message.Context(), params.Event)

			logger.Info(params.Message.Context(), "Event handled",
				logger.F("event_name", params.EventName),
				logger.F("duration", time.Since(start)),
				logger.F("err", err),
			)

			return err
		},
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

func (b *cqrsBus) GetCommandBus() CommandBus {
	return b
}

func (b *cqrsBus) GetEventBus() EventBus {
	return b
}

func (b *cqrsBus) GetCommandProcessor() *cqrs.CommandProcessor {
	return b.commandProcessor
}

func (b *cqrsBus) GetEventProcessor() *cqrs.EventProcessor {
	return b.eventProcessor
}

func (b *cqrsBus) PublishCommand(ctx context.Context, cmd any) error {
	return b.commandBus.Send(ctx, cmd)
}

func (b *cqrsBus) PublishEvent(ctx context.Context, evt any) error {
	return b.eventBus.Publish(ctx, evt)
}

func (b *cqrsBus) RegisterCommandHandler(commandName string, handler CommandHandler) error {
	_, err := b.commandProcessor.AddHandler(cqrs.NewCommandHandler(commandName, handler))
	if err != nil {
		return err
	}

	return nil
}

func (b *cqrsBus) RegisterEventHandler(eventName string, handler EventHandler) error {
	_, err := b.eventProcessor.AddHandler(cqrs.NewEventHandler(eventName, handler))
	if err != nil {
		return err
	}

	return nil
}

func (b *cqrsBus) Run(ctx context.Context) error {
	return b.router.Run(ctx)
}
