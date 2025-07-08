package messaging

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

// CommandHandler handles a command.
type CommandHandler func(context.Context, *any) error

// EventHandler handles an event.
type EventHandler func(context.Context, *any) error

// Bus is the interface for publishing and subscribing to commands/events.
type Dispatcher interface {
	RegisterCommandHandler(commandName string, handler CommandHandler) error
	RegisterEventHandler(eventName string, handler EventHandler) error
	GetCommandProcessor() *cqrs.CommandProcessor
	GetEventProcessor() *cqrs.EventProcessor
	Run(ctx context.Context) error
}

type CommandBus interface {
	PublishCommand(ctx context.Context, cmd any) error
}

type EventBus interface {
	PublishEvent(ctx context.Context, evt any) error
}
