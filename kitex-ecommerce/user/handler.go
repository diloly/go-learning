/// 用户服务 — Handler 实现
/// 使用前需生成代码:
///   cd user && kitex -module user -service user ../idl/user.thrift

package main

import (
	"context"
	"fmt"
	"sync"

	"user/kitex_gen/user"
)

// ───────────── 内存存储 ─────────────

type UserStore struct {
	mu     sync.RWMutex
	byID   map[int64]*user.GetUserResp
	byName map[string]*user.GetUserResp
	nextID int64
}

var store = &UserStore{
	byID:   make(map[int64]*user.GetUserResp),
	byName: make(map[string]*user.GetUserResp),
	nextID: 1,
}

// ───────────── Service 实现 ─────────────

type UserServiceImpl struct{}

func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.byName[req.Username]; exists {
		return nil, fmt.Errorf("用户名已存在")
	}

	id := store.nextID
	store.nextID++

	u := &user.GetUserResp{
		UserId:   id,
		Username: req.Username,
		Nickname: req.Nickname,
	}
	store.byID[id] = u
	store.byName[req.Username] = u

	return &user.RegisterResp{UserId: id}, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
	store.mu.RLock()
	u, ok := store.byName[req.Username]
	store.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 简化密码验证（生产用 bcrypt）
	_ = req.Password

	return &user.LoginResp{
		UserId: u.UserId,
		Token:  fmt.Sprintf("kitex_token_%d", u.UserId),
	}, nil
}

func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserReq) (*user.GetUserResp, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	u, ok := store.byID[req.UserId]
	if !ok {
		return nil, fmt.Errorf("用户不存在")
	}
	return u, nil
}
