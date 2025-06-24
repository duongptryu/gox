package messaging

import (
	"context"
)

// Command represents a CQRS command.
type Command interface{}

// Event represents a CQRS event.
type Event interface{}

// CommandHandler handles a command.
type CommandHandler interface {
	Handle(ctx context.Context, cmd *Command) error
}

// EventHandler handles an event.
type EventHandler interface {
	Handle(ctx context.Context, evt *Event) error
}

// Bus is the interface for publishing and subscribing to commands/events.
type Dispatcher interface {
	RegisterCommandHandler(commandName string, handler CommandHandler) error
	RegisterEventHandler(eventName string, handler EventHandler) error
	Run(ctx context.Context) error
}

type CommandBus interface {
	PublishCommand(ctx context.Context, cmd Command) error
}

type EventBus interface {
	PublishEvent(ctx context.Context, evt Event) error
}
