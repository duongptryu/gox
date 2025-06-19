# eventbus

A CQRS (Command Query Responsibility Segregation) event bus package for Go, built on top of [Watermill](https://watermill.io/) to facilitate command and event handling in distributed systems.

## Features
- Publish and subscribe to commands and events using the CQRS pattern
- Register command and event handlers
- Pluggable with any Watermill-compatible message broker (e.g., Kafka, RabbitMQ, Google Pub/Sub)
- Simple integration with context-aware handlers

## Interfaces

### Command & Event
- `Command`: Marker interface for CQRS commands
- `Event`: Marker interface for CQRS events

### Handlers
- `CommandHandler`: Interface with `Handle(ctx context.Context, cmd Command) error`
- `EventHandler`: Interface with `Handle(ctx context.Context, evt Event) error`

### Bus
- `Bus`: Main interface for registering handlers and running the bus
  - `RegisterCommandHandler(commandName string, handler CommandHandler) error`
  - `RegisterEventHandler(eventName string, handler EventHandler) error`
  - `Run(ctx context.Context) error`
- `BusCommand`: For publishing commands
  - `PublishCommand(ctx context.Context, cmd Command) error`
- `BusEvent`: For publishing events
  - `PublishEvent(ctx context.Context, evt Event) error`

## Implementation
The default implementation uses Watermill's CQRS components. Topics are auto-generated as `commands.<CommandName>` and `events.<EventName>`.

### Configuration
Create a bus using:

```go
cfg := eventbus.Config{
    Publisher:  publisher,   // Watermill message.Publisher
    Subscriber: subscriber, // Watermill message.Subscriber
    Logger:     logger,     // *slog.Logger (optional)
}
bus, err := eventbus.NewBus(cfg)
```

## Usage Example

### Command Example
```go
// Define your command and handler
type MyCommand struct {
    Data string
}

func (c MyCommand) String() string { return "MyCommand" }
func (c MyCommand) DoSomething() {}
// Ensure MyCommand implements eventbus.Command
var _ eventbus.Command = (*MyCommand)(nil)

type MyCommandHandler struct{}

func (h *MyCommandHandler) Handle(ctx context.Context, cmd eventbus.Command) error {
    // handle command
    return nil
}

// Register handler
bus.RegisterCommandHandler("MyCommand", &MyCommandHandler{})

// Publish command
cmd := &MyCommand{Data: "hello"}
bus.PublishCommand(ctx, cmd)
```

### Event Example
```go
// Define your event and handler
type MyEvent struct {
    Message string
}

func (e MyEvent) String() string { return "MyEvent" }
func (e MyEvent) DoSomething() {}
// Ensure MyEvent implements eventbus.Event
var _ eventbus.Event = (*MyEvent)(nil)

type MyEventHandler struct{}

func (h *MyEventHandler) Handle(ctx context.Context, evt eventbus.Event) error {
    // handle event
    return nil
}

// Register handler
bus.RegisterEventHandler("MyEvent", &MyEventHandler{})

// Publish event
evt := &MyEvent{Message: "event fired!"}
bus.PublishEvent(ctx, evt)
```

### Running the Bus
```go
// Run the bus (blocking)
go bus.Run(ctx)
```

## Dependencies
- [Watermill](https://github.com/ThreeDotsLabs/watermill)
- Go 1.18+

## License
MIT 