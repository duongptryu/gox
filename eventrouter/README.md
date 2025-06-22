# eventrouter

A message routing package for Go, built on top of [Watermill](https://watermill.io/) to provide a simpler interface for message routing scenarios that don't necessarily follow CQRS patterns. This package complements the `eventbus` package by offering direct access to Watermill's router functionality.

## Features

- **Simple Message Routing**: Route messages between topics with custom handlers
- **Message Transformation**: Transform messages as they flow through the router
- **Middleware Support**: Add middleware for cross-cutting concerns (logging, metrics, etc.)
- **Plugin System**: Extend functionality with Watermill plugins
- **Multiple Execution Models**: Support for both single and parallel message processing
- **Framework Agnostic**: Works with any Watermill-compatible message broker
- **Context-Aware**: Full context support for tracing and cancellation

## Interfaces

### Core Interfaces

- `Router`: Main interface for message routing
  - `AddHandler(handlerName, subscribeTopic, publishTopic, subscriber, publisher, handler)` - Add handler that transforms and routes messages
  - `AddNoPublisherHandler(handlerName, subscribeTopic, subscriber, handler)` - Add handler that only consumes messages
  - `AddPlugin(plugin)` - Add router plugins
  - `AddMiddleware(middleware)` - Add middleware
  - `Run(ctx)` - Start the router (blocking)
  - `Running()` - Check if router is running
  - `Close()` - Close the router

- `MessageHandler`: Interface for handling messages
  - `Handle(ctx context.Context, msg *message.Message) error`

- `Publisher`: Interface for publishing messages
  - `Publish(ctx context.Context, topic string, msg *message.Message) error`

- `Subscriber`: Interface for subscribing to messages
  - `Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)`

## Usage Examples

### Basic Message Routing

```go
import (
    "context"
    "log/slog"
    
    "github.com/duongptryu/gox/eventrouter"
    "github.com/ThreeDotsLabs/watermill/message"
    "github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()
    
    // Create a simple in-memory pub/sub for example
    pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NewSlogLogger(logger))
    
    // Create router
    router, err := eventrouter.NewRouter(eventrouter.Config{
        Logger: logger,
    })
    if err != nil {
        panic(err)
    }
    defer router.Close()
    
    // Add a message transformation handler
    err = router.AddHandler(
        "user_events_processor",     // handler name
        "user.created",              // subscribe topic
        "user.processed",            // publish topic
        pubSub,                      // subscriber
        pubSub,                      // publisher
        eventrouter.MessageHandlerFunc(func(ctx context.Context, msg *message.Message) error {
            // Process the message
            logger.Info("Processing user created event", slog.String("user_id", string(msg.Payload)))
            
            // Transform the message payload if needed
            msg.Payload = []byte(`{"status": "processed", "user_id": "` + string(msg.Payload) + `"}`)
            
            return nil
        }),
    )
    if err != nil {
        panic(err)
    }
    
    // Start the router
    go router.Run(ctx)
}
```

### Message Consumer (No Publisher)

```go
// Add a handler that only consumes messages without publishing
err = router.AddNoPublisherHandler(
    "audit_logger",
    "user.processed",
    pubSub,
    eventrouter.MessageHandlerFunc(func(ctx context.Context, msg *message.Message) error {
        // Log the processed user event for audit purposes
        logger.Info("Audit log", 
            slog.String("topic", "user.processed"),
            slog.String("message_id", msg.UUID),
            slog.String("payload", string(msg.Payload)))
        
        return nil
    }),
)
```

### Using Middleware

```go
import "github.com/ThreeDotsLabs/watermill/message/router/middleware"

// Add middleware for retry and recovery
router.AddMiddleware(
    middleware.Retry{
        MaxRetries:      3,
        InitialInterval: time.Millisecond * 100,
        Logger:          watermill.NewSlogLogger(logger),
    }.Middleware,
)

router.AddMiddleware(middleware.Recoverer)
```

### Using Plugins

```go
import "github.com/ThreeDotsLabs/watermill/message/router/plugin"

// Add signal handling plugin for graceful shutdown
router.AddPlugin(plugin.SignalsHandler)
```

### Custom Message Handler

```go
type UserEventProcessor struct {
    userService UserService
    logger      *slog.Logger
}

func (p *UserEventProcessor) Handle(ctx context.Context, msg *message.Message) error {
    var event UserCreatedEvent
    if err := json.Unmarshal(msg.Payload, &event); err != nil {
        return fmt.Errorf("failed to unmarshal user event: %w", err)
    }
    
    // Process the user creation
    err := p.userService.ProcessUserCreation(ctx, event.UserID)
    if err != nil {
        return fmt.Errorf("failed to process user creation: %w", err)
    }
    
    p.logger.Info("User creation processed", slog.String("user_id", event.UserID))
    return nil
}

// Use the custom handler
processor := &UserEventProcessor{
    userService: userService,
    logger:      logger,
}

err = router.AddNoPublisherHandler(
    "user_processor",
    "user.created",
    pubSub,
    processor,
)
```

### Message Publishing

```go
// Wrap your Watermill publisher
publisher := eventrouter.NewPublisher(pubSub, logger)

// Create and publish a message
msg := message.NewMessage(watermill.NewUUID(), []byte(`{"user_id": "123", "email": "user@example.com"}`))
msg.Metadata.Set("content-type", "application/json")

err := publisher.Publish(ctx, "user.created", msg)
if err != nil {
    logger.Error("Failed to publish message", slog.Any("error", err))
}
```

### Message Subscription

```go
// Wrap your Watermill subscriber
subscriber := eventrouter.NewSubscriber(pubSub, logger)

// Subscribe to messages
messages, err := subscriber.Subscribe(ctx, "user.created")
if err != nil {
    panic(err)
}

// Process messages
go func() {
    for msg := range messages {
        logger.Info("Received message", 
            slog.String("topic", "user.created"),
            slog.String("message_id", msg.UUID),
            slog.String("payload", string(msg.Payload)))
        
        // Acknowledge the message
        msg.Ack()
    }
}()
```

## Ví dụ thực tế - Order Processing Pipeline

Đây là một ví dụ hoàn chỉnh về cách sử dụng eventrouter để xây dựng pipeline xử lý đơn hàng:

```go
package main

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"
    
    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill/message"
    "github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
    "github.com/duongptryu/gox/eventrouter"
)

// Event structures
type OrderCreated struct {
    OrderID   string  `json:"order_id"`
    UserID    string  `json:"user_id"`
    Amount    float64 `json:"amount"`
    Product   string  `json:"product"`
    CreatedAt string  `json:"created_at"`
}

type OrderProcessed struct {
    OrderID     string  `json:"order_id"`
    UserID      string  `json:"user_id"`
    Amount      float64 `json:"amount"`
    Product     string  `json:"product"`
    Status      string  `json:"status"`
    ProcessedAt string  `json:"processed_at"`
}

type PaymentRequest struct {
    OrderID string  `json:"order_id"`
    UserID  string  `json:"user_id"`
    Amount  float64 `json:"amount"`
}

// Handler xử lý đơn hàng
func processOrder(ctx context.Context, msg *message.Message) error {
    slog.Info("Processing order", slog.String("message_id", msg.UUID))
    
    var order OrderCreated
    if err := json.Unmarshal(msg.Payload, &order); err != nil {
        return err
    }
    
    // Simulate processing
    time.Sleep(100 * time.Millisecond)
    
    // Transform to processed order
    processed := OrderProcessed{
        OrderID:     order.OrderID,
        UserID:      order.UserID,
        Amount:      order.Amount,
        Product:     order.Product,
        Status:      "processed",
        ProcessedAt: time.Now().Format(time.RFC3339),
    }
    
    // Update message payload
    data, _ := json.Marshal(processed)
    msg.Payload = data
    msg.Metadata.Set("status", "processed")
    
    slog.Info("Order processed", slog.String("order_id", order.OrderID))
    return nil
}

// Handler tạo payment request
func processPayment(ctx context.Context, msg *message.Message) error {
    var order OrderProcessed
    json.Unmarshal(msg.Payload, &order)
    
    payment := PaymentRequest{
        OrderID: order.OrderID,
        UserID:  order.UserID,
        Amount:  order.Amount,
    }
    
    data, _ := json.Marshal(payment)
    msg.Payload = data
    
    slog.Info("Payment request created", slog.String("order_id", order.OrderID))
    return nil
}

// Handler thực hiện payment
func executePayment(ctx context.Context, msg *message.Message) error {
    var payment PaymentRequest
    json.Unmarshal(msg.Payload, &payment)
    
    // Simulate payment processing
    time.Sleep(200 * time.Millisecond)
    
    slog.Info("Payment executed", 
        slog.String("order_id", payment.OrderID),
        slog.Float64("amount", payment.Amount))
    return nil
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Create pub/sub
    pubSub := gochannel.NewGoChannel(
        gochannel.Config{},
        watermill.NewSlogLogger(slog.Default()),
    )
    
    // Create router
    router, _ := eventrouter.NewRouter(eventrouter.Config{
        Logger: slog.Default(),
    })
    defer router.Close()
    
    // Add handlers to create processing pipeline
    router.AddHandler("order_processor", "orders.created", "orders.processed", 
        pubSub, pubSub, eventrouter.MessageHandlerFunc(processOrder))
        
    router.AddHandler("payment_processor", "orders.processed", "payments.requests",
        pubSub, pubSub, eventrouter.MessageHandlerFunc(processPayment))
        
    router.AddNoPublisherHandler("payment_executor", "payments.requests",
        pubSub, eventrouter.MessageHandlerFunc(executePayment))
    
    // Start router
    go router.Run(ctx)
    time.Sleep(500 * time.Millisecond) // Wait for startup
    
    // Publish test order
    publisher := eventrouter.NewPublisher(pubSub, slog.Default())
    order := OrderCreated{
        OrderID:   "order-123",
        UserID:    "user-456",
        Amount:    99.99,
        Product:   "Laptop",
        CreatedAt: time.Now().Format(time.RFC3339),
    }
    
    orderData, _ := json.Marshal(order)
    msg := message.NewMessage(watermill.NewUUID(), orderData)
    
    publisher.Publish(ctx, "orders.created", msg)
    slog.Info("Published order", slog.String("order_id", order.OrderID))
    
    // Wait for processing
    time.Sleep(5 * time.Second)
    slog.Info("Example completed")
}
```

### Luồng xử lý:
```
orders.created → [processOrder] → orders.processed → [processPayment] → payments.requests → [executePayment]
```

**Chạy ví dụ:**
```bash
cd eventrouter/example
go run simple_example.go
```

## Differences from eventbus Package

| Feature | eventbus | eventrouter |
|---------|----------|-------------|
| **Purpose** | CQRS command/event handling | General message routing |
| **Abstraction Level** | High-level CQRS concepts | Direct Watermill router access |
| **Message Types** | Commands and Events | Raw Watermill messages |
| **Use Cases** | CQRS applications | Message transformation, routing |
| **Flexibility** | CQRS-focused | More flexible routing patterns |

## Configuration

```go
type Config struct {
    Logger       *slog.Logger    // Optional logger (defaults to slog.Default())
    CloseTimeout *time.Duration  // Optional timeout for graceful shutdown
}
```

## Integration with Other Packages

### With logger package

```go
import "github.com/duongptryu/gox/logger"

// Initialize logger
logger.Init(&logger.Config{
    Level:     slog.LevelInfo,
    AddSource: true,
})

// Use with eventrouter
router, err := eventrouter.NewRouter(eventrouter.Config{
    Logger: slog.Default(), // Uses the initialized logger
})
```

### With syserr package

The eventrouter package integrates with the `syserr` package for structured error handling:

```go
func (h *MyHandler) Handle(ctx context.Context, msg *message.Message) error {
    if msg.Payload == nil {
        return syserr.New(syserr.InvalidArgumentCode, "message payload is empty",
            syserr.F("message_id", msg.UUID))
    }
    
    // Handle message...
    return nil
}
```

## Best Practices

### Error Handling

- Always return meaningful errors from handlers
- Use the `syserr` package for structured error information
- Consider using retry middleware for transient errors

### Message Processing

- Keep handlers idempotent when possible
- Use message metadata for routing decisions
- Always acknowledge messages after successful processing

### Performance

- Use appropriate middleware for your use case
- Consider message batching for high-throughput scenarios
- Monitor handler performance and add timeouts if needed

### Graceful Shutdown

```go
// Use context for graceful shutdown
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Handle shutdown signals
go func() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    logger.Info("Shutdown signal received")
    cancel()
}()

// Run router with cancellable context
err := router.Run(ctx)
if err != nil {
    logger.Error("Router error", slog.Any("error", err))
}
```

## Dependencies

- [Watermill](https://github.com/ThreeDotsLabs/watermill) - Core message routing
- Go 1.18+

## License

MIT 