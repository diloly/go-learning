/// 商品服务 — Handler 实现
/// 使用前需生成代码:
///   cd product && kitex -module product -service product ../idl/product.thrift

package main

import (
	"context"
	"fmt"
	"sync"

	"product/kitex_gen/product"
)

type ProductStore struct {
	mu     sync.RWMutex
	byID   map[int64]*product.ProductResp
	nextID int64
}

var store = &ProductStore{
	byID:   make(map[int64]*product.ProductResp),
	nextID: 1,
}

var _ = func() []*product.ProductResp {
	// 预置数据
	store.create("iPhone 16", "最新款苹果手机", 7999, 100)
	store.create("MacBook Pro", "M4 芯片笔记本", 14999, 50)
	store.create("AirPods Pro", "主动降噪耳机", 1999, 200)
	return nil
}()

func (s *ProductStore) create(name, desc string, price float64, stock int64) *product.ProductResp {
	id := s.nextID
	s.nextID++
	p := &product.ProductResp{
		Id: id, Name: name, Description: desc,
		Price: price, Stock: stock,
	}
	s.byID[id] = p
	return p
}

type ProductServiceImpl struct{}

func (s *ProductServiceImpl) CreateProduct(ctx context.Context, req *product.CreateProductReq) (*product.ProductResp, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.create(req.Name, req.Description, req.Price, req.Stock), nil
}

func (s *ProductServiceImpl) GetProduct(ctx context.Context, req *product.GetProductReq) (*product.ProductResp, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	p, ok := store.byID[req.ProductId]
	if !ok {
		return nil, fmt.Errorf("商品不存在")
	}
	return p, nil
}

func (s *ProductServiceImpl) ListProducts(ctx context.Context, req *product.ListProductsReq) (*product.ListProductsResp, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var all []*product.ProductResp
	for _, p := range store.byID {
		all = append(all, p)
	}
	total := int64(len(all))

	page := int(req.Page)
	size := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	start := (page - 1) * size
	if start >= len(all) {
		return &product.ListProductsResp{Total: total}, nil
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}

	return &product.ListProductsResp{
		Products: all[start:end],
		Total:    total,
	}, nil
}

func (s *ProductServiceImpl) DeductStock(ctx context.Context, req *product.DeductStockReq) (*product.DeductStockResp, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	p, ok := store.byID[req.ProductId]
	if !ok {
		return &product.DeductStockResp{Success: false}, fmt.Errorf("商品不存在")
	}
	if p.Stock < req.Quantity {
		return &product.DeductStockResp{Success: false}, fmt.Errorf("库存不足")
	}
	p.Stock -= req.Quantity
	return &product.DeductStockResp{Success: true}, nil
}
