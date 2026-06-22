package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shortlink/config"
	"shortlink/handler"
	"shortlink/router"
	"shortlink/service"
	"shortlink/store"
)

// ─────────────────────────────────────────────────
// 短链接服务 — 入口
//
// 启动流程:
//   1. 加载配置（支持热加载）
//   2. 连接 MySQL / PostgreSQL
//   3. 连接 Redis
//   4. 初始化 Service → Handler → Router
//   5. 启动 HTTP 服务
//   6. 等待退出信号 → 优雅关闭
// ─────────────────────────────────────────────────

func main() {
	// ── 1. 加载配置 ──
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  短链接服务 v1.0")
	fmt.Println("═══════════════════════════════════════")

	// ── 2. 连接数据库 ──
	dbStore, err := store.NewDBStore(cfg.Database)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	// ── 3. 连接 Redis ──
	redisStore, err := store.NewRedisStore(cfg.Redis, time.Duration(cfg.ShortLink.CacheTTL)*time.Second)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	// ── 4. 初始化业务层 ──
	svc := &service.ShortLinkService{
		DB:    dbStore,
		Cache: redisStore,
	}
	hdl := &handler.ShortLinkHandler{Svc: svc}

	// ── 5. 注册路由 ──
	r := router.Setup(hdl, cfg.Server.Mode)

	// ── 6. 启动 HTTP 服务 ──
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在 goroutine 中启动，不阻塞信号监听
	go func() {
		fmt.Printf("🚀 服务启动: http://localhost%s\n", addr)
		fmt.Println("📌 接口:")
		fmt.Println("   POST /api/v1/shorten    创建短链接")
		fmt.Println("   GET  /:code             访问短链接（302 重定向）")
		fmt.Println("   GET  /api/v1/:code/stats 查看统计")
		fmt.Println("────────────────────────────────────────────")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务启动失败: %v", err)
		}
	}()

	// ── 7. 等待退出信号，优雅关闭 ──
	// 捕获 SIGINT (Ctrl+C) 和 SIGTERM（kill）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit // 阻塞等待信号

	fmt.Printf("\n🛑 收到信号: %v，开始优雅退出...\n", sig)

	// 创建 5s 超时上下文，用于 http.Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 按依赖逆序关闭: HTTP → Redis → DB
	// ── 关闭 HTTP 服务（停止接受新请求，等待正在处理的请求完成） ──
	fmt.Println("⏳ 关闭 HTTP 服务...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ HTTP 关闭异常: %v", err)
	}

	// ── 关闭 Redis 连接 ──
	fmt.Println("⏳ 关闭 Redis 连接...")
	if err := redisStore.Close(); err != nil {
		log.Printf("⚠️ Redis 关闭异常: %v", err)
	}

	// ── 关闭数据库连接 ──
	fmt.Println("⏳ 关闭数据库连接...")
	if err := dbStore.Close(); err != nil {
		log.Printf("⚠️ 数据库关闭异常: %v", err)
	}

	// ── 优雅退出完成 ──
	// ★ 此处不使用 log.Fatal / os.Exit，因为 Fatal 底层调用了 os.Exit(1)
	//    os.Exit 会立刻终止进程，不会执行 defer，导致资源泄漏
	//    我们手动完成所有 cleanup 后自然退出
	fmt.Println("✅ 服务已安全关闭，再见！")
}
