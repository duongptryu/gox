# EventRouter Examples

Thư mục này chứa các ví dụ về cách sử dụng package `eventrouter`.

## Ví dụ đơn giản - Simple Example

File: `simple_example.go`

Ví dụ này minh họa một luồng xử lý đơn giản:
1. **Order Created** → **Order Processed** → **Payment Request** → **Payment Executed**

### Cách chạy

```bash
cd eventrouter/example
go run simple_example.go
```

### Luồng xử lý

1. **Tạo đơn hàng** (`orders.created`):
   ```json
   {
     "order_id": "order-001",
     "user_id": "user-123", 
     "amount": 99.99,
     "product": "Laptop",
     "created_at": "2024-01-01T10:00:00Z"
   }
   ```

2. **Xử lý đơn hàng** (`orders.processed`):
   - Handler `processOrder` nhận message từ topic `orders.created`
   - Transform thành `OrderProcessed` và publish đến `orders.processed`

3. **Tạo yêu cầu thanh toán** (`payments.requests`):
   - Handler `processPayment` nhận từ `orders.processed`
   - Transform thành `PaymentRequest` và publish đến `payments.requests`

4. **Thực hiện thanh toán**:
   - Handler `executePayment` chỉ consume từ `payments.requests`
   - Không publish message nào (consumer-only handler)

### Các tính năng được minh họa

- ✅ **Message Routing**: Route messages giữa các topics
- ✅ **Message Transformation**: Transform message payload khi route
- ✅ **Consumer-only Handlers**: Handlers chỉ consume không publish
- ✅ **Publisher/Subscriber Wrappers**: Sử dụng wrapper cho pub/sub
- ✅ **Structured Logging**: Log với context và metadata
- ✅ **Graceful Shutdown**: Dừng router một cách an toàn

### Output mẫu

```
INFO === Simple EventRouter Example ===
INFO Starting router...
INFO Published order order_id=order-001 product=Laptop amount=99.99
INFO Processing order message_id=abc123
INFO Order processed order_id=order-001 amount=99.99
INFO Processing payment message_id=def456  
INFO Payment request created order_id=order-001 amount=99.99
INFO Executing payment message_id=ghi789
INFO Payment executed successfully order_id=order-001 user_id=user-123 amount=99.99
INFO Published order order_id=order-002 product=Book amount=29.99
INFO === Example completed ===
```

## Kiến trúc ví dụ

```
orders.created ──→ [processOrder] ──→ orders.processed ──→ [processPayment] ──→ payments.requests ──→ [executePayment]
```

### Handlers

- **processOrder**: Transform `OrderCreated` → `OrderProcessed`
- **processPayment**: Transform `OrderProcessed` → `PaymentRequest`  
- **executePayment**: Consumer-only, xử lý payment cuối cùng

### Topics

- `orders.created`: Đơn hàng mới được tạo
- `orders.processed`: Đơn hàng đã được xử lý
- `payments.requests`: Yêu cầu thanh toán 