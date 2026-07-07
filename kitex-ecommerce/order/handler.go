/// 订单服务 — Handler 实现
/// 使用前需生成代码:
///   cd order && kitex -module order -service order ../idl/order.thrift

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"order/kitex_gen/order"
)

type Order struct {
	OrderID    string
	UserID     int64
	ProductID  int64
	Quantity   int64
	TotalPrice float64
	Status     string
	Address    string
	CreatedAt  string
}

type OrderStore struct {
	mu     sync.RWMutex
	byID   map[string]*Order
	orders []*Order
}

var orderStore = &OrderStore{
	byID: make(map[string]*Order),
}

var orderMu sync.Mutex
var orderSeq int64

var productPrices = map[int64]float64{
	1: 7999, 2: 14999, 3: 1999,
}

// ───────────── RocketMQ 生产者 ─────────────

type MQProducer interface {
	SendOrderMsg(data []byte) error
	Start() error
	Stop() error
}

// 模拟 RocketMQ 生产者（实际接入时替换为 rocketmq-go 客户端）
type MockRocketMQ struct {
	messages [][]byte
}

func (m *MockRocketMQ) Start() error                          { return nil }
func (m *MockRocketMQ) Stop() error                           { return nil }
func (m *MockRocketMQ) SendOrderMsg(data []byte) error        { m.messages = append(m.messages, data); return nil }

var mqProducer MQProducer = &MockRocketMQ{}

// ───────────── 异步下单消费者 ─────────────

func startOrderConsumer() {
	// 实际项目中这里启动 RocketMQ PushConsumer
	// 此处用 goroutine + 定时轮询模拟
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if producer, ok := mqProducer.(*MockRocketMQ); ok && len(producer.messages) > 0 {
				for _, msg := range producer.messages {
					fmt.Printf("🔄 [RocketMQ] 处理订单: %s\n", string(msg))
					// 模拟异步处理
					time.Sleep(300 * time.Millisecond)
				}
				producer.messages = nil
			}
		}
	}()
}

// ───────────── Service 实现 ─────────────

type OrderServiceImpl struct{}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	price, ok := productPrices[req.ProductId]
	if !ok {
		return nil, fmt.Errorf("商品不存在")
	}

	orderMu.Lock()
	orderSeq++
	orderID := fmt.Sprintf("KTE%d%05d", time.Now().Unix(), orderSeq)
	orderMu.Unlock()

	o := &Order{
		OrderID:    orderID,
		UserID:     req.UserId,
		ProductID:  req.ProductId,
		Quantity:   req.Quantity,
		TotalPrice: price * float64(req.Quantity),
		Status:     "pending",
		Address:    req.Address,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	orderStore.mu.Lock()
	orderStore.byID[orderID] = o
	orderStore.orders = append(orderStore.orders, o)
	orderStore.mu.Unlock()

	// 发送到 RocketMQ 异步处理
	msg := []byte(fmt.Sprintf(`{"order_id":"%s","product_id":%d,"quantity":%d}`,
		orderID, req.ProductId, req.Quantity))
	if err := mqProducer.SendOrderMsg(msg); err != nil {
		return nil, fmt.Errorf("消息发送失败: %w", err)
	}

	return &order.CreateOrderResp{
		OrderId: orderID,
		Status:  "pending",
	}, nil
}

func (s *OrderServiceImpl) GetOrder(ctx context.Context, req *order.GetOrderReq) (*order.OrderResp, error) {
	orderStore.mu.RLock()
	defer orderStore.mu.RUnlock()

	o, ok := orderStore.byID[req.OrderId]
	if !ok {
		return nil, fmt.Errorf("订单不存在")
	}
	return &order.OrderResp{
		OrderId:    o.OrderID,
		UserId:     o.UserID,
		ProductId:  o.ProductID,
		Quantity:   o.Quantity,
		TotalPrice: o.TotalPrice,
		Status:     o.Status,
		Address:    o.Address,
	}, nil
}

func (s *OrderServiceImpl) ListOrders(ctx context.Context, req *order.ListOrdersReq) (*order.ListOrdersResp, error) {
	orderStore.mu.RLock()
	defer orderStore.mu.RUnlock()

	var filtered []*order.OrderResp
	for _, o := range orderStore.orders {
		if o.UserID == req.UserId {
			filtered = append(filtered, &order.OrderResp{
				OrderId:    o.OrderID,
				UserId:     o.UserID,
				ProductId:  o.ProductID,
				Quantity:   o.Quantity,
				TotalPrice: o.TotalPrice,
				Status:     o.Status,
				Address:    o.Address,
			})
		}
	}
	total := int64(len(filtered))

	page, size := int(req.Page), int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	start := (page - 1) * size
	if start >= len(filtered) {
		return &order.ListOrdersResp{Total: total}, nil
	}
	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}

	return &order.ListOrdersResp{
		Orders: filtered[start:end],
		Total:  total,
	}, nil
}
