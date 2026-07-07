package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var errNotFound = fmt.Errorf("商品不存在")
var errInsufficientStock = fmt.Errorf("库存不足")

// ───────────── Config ─────────────

type Config struct {
	rest.RestConf
}

// ───────────── Model ─────────────

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int64   `json:"stock"`
}

// ───────────── Store ─────────────

type ProductStore struct {
	mu     sync.RWMutex
	byID   map[int64]*Product
	nextID int64
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		byID:   make(map[int64]*Product),
		nextID: 1,
	}
}

func (s *ProductStore) Create(name, desc string, price float64, stock int64) *Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := &Product{
		ID: s.nextID, Name: name, Description: desc,
		Price: price, Stock: stock,
	}
	s.byID[p.ID] = p
	s.nextID++
	return p
}

func (s *ProductStore) GetByID(id int64) (*Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byID[id]
	if !ok {
		return nil, errNotFound
	}
	return p, nil
}

func (s *ProductStore) List(page, size int) ([]*Product, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := int64(len(s.byID))
	start := (page - 1) * size
	if start >= len(s.byID) {
		return nil, total
	}
	end := start + size
	if end > len(s.byID) {
		end = len(s.byID)
	}
	result := make([]*Product, 0, end-start)
	// 稳定遍历
	ids := make([]int64, 0, len(s.byID))
	for id := range s.byID {
		ids = append(ids, id)
	}
	// 按 ID 排序
	for i := 0; i < len(ids)-1; i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] > ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	for _, id := range ids[start:end] {
		result = append(result, s.byID[id])
	}
	return result, total
}

func (s *ProductStore) DeductStock(id, qty int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.byID[id]
	if !ok {
		return errNotFound
	}
	if p.Stock < qty {
		return errInsufficientStock
	}
	p.Stock -= qty
	return nil
}

// ───────────── Helpers ─────────────

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

	store := NewProductStore()
	// 预置测试数据
	store.Create("iPhone 16", "最新款苹果手机", 7999, 100)
	store.Create("MacBook Pro", "M4 芯片笔记本电脑", 14999, 50)
	store.Create("AirPods Pro", "主动降噪耳机", 1999, 200)

	server.AddRoutes([]rest.Route{
		{
			Method: http.MethodPost, Path: "/api/product/create",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Name string  `json:"name"`
					Desc string  `json:"description"`
					Px   float64 `json:"price"`
					Stk  int64   `json:"stock"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					writeError(w, 400, "参数错误")
					return
				}
				p := store.Create(body.Name, body.Desc, body.Px, body.Stk)
				writeJSON(w, p)
			},
		},
		{
			Method: http.MethodGet, Path: "/api/product/:id",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
				p, err := store.GetByID(id)
				if err != nil {
					writeError(w, 404, err.Error())
					return
				}
				writeJSON(w, p)
			},
		},
		{
			Method: http.MethodGet, Path: "/api/product/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				page, _ := strconv.Atoi(r.URL.Query().Get("page"))
				size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
				if page < 1 {
					page = 1
				}
				if size < 1 || size > 100 {
					size = 10
				}
				products, total := store.List(page, size)
				writeJSON(w, map[string]interface{}{
					"list":  products,
					"total": total,
				})
			},
		},
		{
			Method: http.MethodPost, Path: "/api/product/deduct_stock",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					ProductID int64 `json:"product_id"`
					Quantity  int64 `json:"quantity"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					writeError(w, 400, "参数错误")
					return
				}
				if err := store.DeductStock(body.ProductID, body.Quantity); err != nil {
					writeError(w, 400, err.Error())
					return
				}
				writeJSON(w, map[string]bool{"success": true})
			},
		},
	})

	log.Printf("🚀 商品服务启动: :%d\n", c.Port)
	server.Start()
}
