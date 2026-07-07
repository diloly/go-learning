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

// ───────────── Config ─────────────

type Config struct {
	rest.RestConf
}

// ───────────── Model ─────────────

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Nickname string `json:"nickname"`
}

// ───────────── Store ─────────────

type UserStore struct {
	mu     sync.RWMutex
	byID   map[int64]*User
	byName map[string]*User
	nextID int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		byID:   make(map[int64]*User),
		byName: make(map[string]*User),
		nextID: 1,
	}
}

func (s *UserStore) Create(username, password, nickname string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[username]; exists {
		return nil, fmt.Errorf("用户名已存在")
	}
	u := &User{ID: s.nextID, Username: username, Password: password, Nickname: nickname}
	s.byID[u.ID] = u
	s.byName[username] = u
	s.nextID++
	return u, nil
}

func (s *UserStore) FindByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byName[username]
	if !ok {
		return nil, fmt.Errorf("用户不存在")
	}
	return u, nil
}

func (s *UserStore) FindByID(id int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	if !ok {
		return nil, fmt.Errorf("用户不存在")
	}
	return u, nil
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

func main() {
	var c Config
	conf.MustLoad("etc/config.yaml", &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	store := NewUserStore()

	server.AddRoutes([]rest.Route{
		{
			Method: http.MethodPost,
			Path:   "/api/user/register",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Username string `json:"username"`
					Password string `json:"password"`
					Nickname string `json:"nickname"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					writeError(w, 400, "参数错误")
					return
				}
				u, err := store.Create(body.Username, body.Password, body.Nickname)
				if err != nil {
					writeError(w, 400, err.Error())
					return
				}
				writeJSON(w, u)
			},
		},
		{
			Method: http.MethodPost,
			Path:   "/api/user/login",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					writeError(w, 400, "参数错误")
					return
				}
				u, err := store.FindByUsername(body.Username)
				if err != nil || u.Password != body.Password {
					writeError(w, 401, "用户名或密码错误")
					return
				}
				writeJSON(w, map[string]interface{}{
					"user_id": u.ID,
					"token":   fmt.Sprintf("token_%d", u.ID), // 简化版 JWT
				})
			},
		},
		{
			Method: http.MethodGet,
			Path:   "/api/user/:id",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				idStr := r.PathValue("id")
				id, _ := strconv.ParseInt(idStr, 10, 64)
				u, err := store.FindByID(id)
				if err != nil {
					writeError(w, 404, err.Error())
					return
				}
				writeJSON(w, u)
			},
		},
	})

	log.Printf("🚀 用户服务启动: :%d\n", c.Port)
	server.Start()
}
