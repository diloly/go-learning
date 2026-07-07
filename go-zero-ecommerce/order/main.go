package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// ───────────── Config ─────────────

type Config struct {
	rest.RestConf
	RabbitMQ struct {
		Host     string
		Port     int
		Username string
		Password string
		VHost    string
	}
}

// ───────────── Model ─────────────

type Order struct {
	OrderID    string  `json:"order_id"`
	UserID     int64   `json:"user_id"`
	ProductID  int64   `json:"product_id"`
	Quantity   int64   `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"` // pending / success / failed
	Address    string  `json:"address"`
	CreatedAt  string  `json:"created_at"`
}

type OrderStore struct {
	mu     sync.RWMutex
	byID   map[string]*Order
	orders []*Order
}

func NewOrderStore() *OrderStore {
	return &OrderStore{
		byID: make(map[string]*Order),
	}
}

func (s *OrderStore) Save(o *Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[o.OrderID] = o
	s.orders = append(s.orders, o)
}

func (s *OrderStore) GetByID(id string) (*Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("订单不存在")
	}
	return o, nil
}

func (s *OrderStore) ListByUser(uid int64, page, size int) ([]*Order, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var filtered []*Order
	for _, o := range s.orders {
		if o.UserID == uid {
			filtered = append(filtered, o)
		}
	}
	total := int64(len(filtered))
	start := (page - 1) * size
	if start >= len(filtered) {
		return nil, total
	}
	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total
}

func (s *OrderStore) UpdateStatus(id, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o, ok := s.byID[id]; ok {
		o.Status = status
	}
}

// ───────────── RabbitMQ ─────────────

type MQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   amqp091.Queue
}

func NewMQClient(cfg Config) (*MQClient, error) {
	addr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.RabbitMQ.Username, cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host, cfg.RabbitMQ.Port, cfg.RabbitMQ.VHost)

	conn, err := amqp091.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQ 连接失败: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建 channel 失败: %w", err)
	}

	q, err := ch.QueueDeclare(
		"order_queue", true, false, false, false, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("声明队列失败: %w", err)
	}

	log.Println("✅ RabbitMQ 连接成功")
	return &MQClient{conn: conn, channel: ch, queue: q}, nil
}

func (m *MQClient) Publish(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.channel.PublishWithContext(ctx,
		"", m.queue.Name, false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        data,
		})
}

func (m *MQClient) Consume(orderStore *OrderStore) {
	msgs, err := m.channel.Consume(m.queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("消费队列失败: %v", err)
	}

	log.Println("📦 订单消费者启动，等待消息...")
	for msg := range msgs {
		var body struct {
			OrderID    string  `json:"order_id"`
			UserID     int64   `json:"user_id"`
			ProductID  int64   `json:"product_id"`
			Quantity   int64   `json:"quantity"`
			TotalPrice float64 `json:"total_price"`
			Address    string  `json:"address"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		log.Printf("🔄 处理订单: %s (商品 %d x %d)", body.OrderID, body.ProductID, body.Quantity)

		// 模拟异步处理：扣库存（实际应调用商品服务 HTTP API）
		time.Sleep(500 * time.Millisecond)

		// 更新订单状态
		orderStore.UpdateStatus(body.OrderID, "success")
		log.Printf("✅ 订单完成: %s", body.OrderID)
	}
}

func (m *MQClient) Close() {
	m.channel.Close()
	m.conn.Close()
}

// ───────────── Handlers ─────────────

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": data})
}
func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"code": -1, "message": msg})
}

// ───────────── Main ─────────────

func main() {
	var c Config
	conf.MustLoad("etc/config.yaml", &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	orderStore := NewOrderStore()
	var orderMu sync.Mutex
	orderSeq := int64(0)

	// 连接 RabbitMQ
	mq, err := NewMQClient(c)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}
	defer mq.Close()

	// 启动消费者 goroutine（异步处理订单）
	go mq.Consume(orderStore)

	// 预置商品价格表（简化：实际应调商品服务）
	productPrices := map[int64]float64{
		1: 7999, 2: 14999, 3: 1999,
	}

	server.AddRoutes([]rest.Route{
		{
			Method: http.MethodPost, Path: "/api/order/create",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					UserID    int64  `json:"user_id"`
					ProductID int64  `json:"product_id"`
					Quantity  int64  `json:"quantity"`
					Address   string `json:"address"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					writeError(w, 400, "参数错误")
					return
				}

				price, ok := productPrices[body.ProductID]
				if !ok {
					writeError(w, 400, "商品不存在")
					return
				}

				// 生成订单
				orderMu.Lock()
				orderSeq++
				orderID := fmt.Sprintf("ORD%d%05d", time.Now().Unix(), orderSeq)
				orderMu.Unlock()

				totalPrice := price * float64(body.Quantity)

				order := &Order{
					OrderID:    orderID,
					UserID:     body.UserID,
					ProductID:  body.ProductID,
					Quantity:   body.Quantity,
					TotalPrice: totalPrice,
					Status:     "pending",
					Address:    body.Address,
					CreatedAt:  time.Now().Format(time.RFC3339),
				}
				orderStore.Save(order)

				// 发送消息到 RabbitMQ 异步处理
				msgData, _ := json.Marshal(map[string]interface{}{
					"order_id":    orderID,
					"user_id":     body.UserID,
					"product_id":  body.ProductID,
					"quantity":    body.Quantity,
					"total_price": totalPrice,
					"address":     body.Address,
				})
				if err := mq.Publish(msgData); err != nil {
					log.Printf("⚠️ 消息发送失败: %v", err)
					orderStore.UpdateStatus(orderID, "failed")
					writeError(w, 500, "下单失败，请重试")
					return
				}

				writeJSON(w, map[string]string{
					"order_id": orderID,
					"status":   "pending",
					"message":  "订单已提交，正在异步处理",
				})
			},
		},
		{
			Method: http.MethodGet, Path: "/api/order/:id",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				orderID := r.PathValue("id")
				o, err := orderStore.GetByID(orderID)
				if err != nil {
					writeError(w, 404, err.Error())
					return
				}
				writeJSON(w, o)
			},
		},
		{
			Method: http.MethodGet, Path: "/api/order/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
				page, _ := strconv.Atoi(r.URL.Query().Get("page"))
				size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
				if page < 1 {
					page = 1
				}
				if size < 1 || size > 100 {
					size = 10
				}
				orders, total := orderStore.ListByUser(userID, page, size)
				writeJSON(w, map[string]interface{}{
					"list":  orders,
					"total": total,
				})
			},
		},
	})

	log.Printf("🚀 订单服务启动: :%d\n", c.Port)
	server.Start()
}
