package eventbus

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
type Bus interface {
	RegisterCommandHandler(commandName string, handler CommandHandler) error
	RegisterEventHandler(eventName string, handler EventHandler) error
	Run(ctx context.Context) error
}

type BusCommand interface {
	PublishCommand(ctx context.Context, cmd Command) error
}

type BusEvent interface {
	PublishEvent(ctx context.Context, evt Event) error
}
