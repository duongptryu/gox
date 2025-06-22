package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/duongptryu/gox/eventrouter"
)

// Simple event structures
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

// Order processor handler
func processOrder(ctx context.Context, msg *message.Message) error {
	slog.Info("Processing order", slog.String("message_id", msg.UUID))

	// Parse order
	var order OrderCreated
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	// Simulate processing
	time.Sleep(100 * time.Millisecond)

	// Create processed order
	processed := OrderProcessed{
		OrderID:     order.OrderID,
		UserID:      order.UserID,
		Amount:      order.Amount,
		Product:     order.Product,
		Status:      "processed",
		ProcessedAt: time.Now().Format(time.RFC3339),
	}

	// Update message
	data, err := json.Marshal(processed)
	if err != nil {
		return fmt.Errorf("failed to marshal processed order: %w", err)
	}

	msg.Payload = data
	msg.Metadata.Set("status", "processed")

	slog.Info("Order processed",
		slog.String("order_id", order.OrderID),
		slog.Float64("amount", order.Amount))

	return nil
}

// Payment processor handler
func processPayment(ctx context.Context, msg *message.Message) error {
	slog.Info("Processing payment", slog.String("message_id", msg.UUID))

	// Parse processed order
	var order OrderProcessed
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return fmt.Errorf("failed to unmarshal processed order: %w", err)
	}

	// Create payment request
	payment := PaymentRequest{
		OrderID: order.OrderID,
		UserID:  order.UserID,
		Amount:  order.Amount,
	}

	// Update message
	data, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request: %w", err)
	}

	msg.Payload = data
	msg.Metadata.Set("payment_amount", fmt.Sprintf("%.2f", order.Amount))

	slog.Info("Payment request created",
		slog.String("order_id", order.OrderID),
		slog.Float64("amount", order.Amount))

	return nil
}

// Payment execution handler (consumer only)
func executePayment(ctx context.Context, msg *message.Message) error {
	slog.Info("Executing payment", slog.String("message_id", msg.UUID))

	// Parse payment request
	var payment PaymentRequest
	if err := json.Unmarshal(msg.Payload, &payment); err != nil {
		return fmt.Errorf("failed to unmarshal payment request: %w", err)
	}

	// Simulate payment execution
	time.Sleep(200 * time.Millisecond)

	slog.Info("Payment executed successfully",
		slog.String("order_id", payment.OrderID),
		slog.String("user_id", payment.UserID),
		slog.Float64("amount", payment.Amount))

	return nil
}

func runSimpleExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("=== Simple EventRouter Example ===")

	// Create in-memory pub/sub
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewSlogLogger(slog.Default()),
	)

	// Create router
	router, err := eventrouter.NewRouter(eventrouter.Config{
		Logger: slog.Default(),
	})
	if err != nil {
		slog.Error("Failed to create router", slog.Any("error", err))
		return
	}
	defer router.Close()

	// Add handlers with message transformation
	err = router.AddHandler(
		"order_processor",
		"orders.created",   // Subscribe to this topic
		"orders.processed", // Publish transformed message to this topic
		pubSub,
		pubSub,
		eventrouter.MessageHandlerFunc(processOrder),
	)
	if err != nil {
		slog.Error("Failed to add order processor", slog.Any("error", err))
		return
	}

	err = router.AddHandler(
		"payment_processor",
		"orders.processed",  // Subscribe to processed orders
		"payments.requests", // Publish payment requests
		pubSub,
		pubSub,
		eventrouter.MessageHandlerFunc(processPayment),
	)
	if err != nil {
		slog.Error("Failed to add payment processor", slog.Any("error", err))
		return
	}

	// Add consumer-only handler
	err = router.AddNoPublisherHandler(
		"payment_executor",
		"payments.requests", // Only consume payment requests
		pubSub,
		eventrouter.MessageHandlerFunc(executePayment),
	)
	if err != nil {
		slog.Error("Failed to add payment executor", slog.Any("error", err))
		return
	}

	// Start router
	go func() {
		slog.Info("Starting router...")
		if err := router.Run(ctx); err != nil {
			slog.Error("Router error", slog.Any("error", err))
		}
	}()

	// Wait for router to start
	time.Sleep(500 * time.Millisecond)

	// Create publisher
	publisher := eventrouter.NewPublisher(pubSub, slog.Default())

	// Create some test orders
	orders := []OrderCreated{
		{
			OrderID:   "order-001",
			UserID:    "user-123",
			Amount:    99.99,
			Product:   "Laptop",
			CreatedAt: time.Now().Format(time.RFC3339),
		},
		{
			OrderID:   "order-002",
			UserID:    "user-456",
			Amount:    29.99,
			Product:   "Book",
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	}

	// Publish orders
	for i, order := range orders {
		// Add delay between orders
		if i > 0 {
			time.Sleep(2 * time.Second)
		}

		orderData, err := json.Marshal(order)
		if err != nil {
			slog.Error("Failed to marshal order", slog.Any("error", err))
			continue
		}

		msg := message.NewMessage(watermill.NewUUID(), orderData)
		msg.Metadata.Set("order_id", order.OrderID)
		msg.Metadata.Set("user_id", order.UserID)

		err = publisher.Publish(ctx, "orders.created", msg)
		if err != nil {
			slog.Error("Failed to publish order",
				slog.String("order_id", order.OrderID),
				slog.Any("error", err))
		} else {
			slog.Info("Published order",
				slog.String("order_id", order.OrderID),
				slog.String("product", order.Product),
				slog.Float64("amount", order.Amount))
		}
	}

	// Wait for processing to complete
	time.Sleep(5 * time.Second)

	slog.Info("=== Example completed ===")
}

func main() {
	runSimpleExample()
}
