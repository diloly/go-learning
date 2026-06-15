package main

import (
	"fmt"
	"log"
	"os"

	"mini-blog/config"
	"mini-blog/handler"
	"mini-blog/model"
	"mini-blog/router"
	"mini-blog/service"
	"mini-blog/utils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// mysql 专用驱动（用于检测建库）
	_ "github.com/go-sql-driver/mysql"
)

// ─────────────────────────────────────────────────
// 简易博客后台 API — 入口
//
// 启动流程:
//   1. 加载配置 (viper)
//   2. 初始化 JWT
//   3. 连接数据库 (MySQL / PostgreSQL)
//   4. 自动迁移表结构
//   5. 初始化 Service → Handler
//   6. 注册路由 → 启动 HTTP 服务
// ─────────────────────────────────────────────────

func main() {
	// ── 1. 加载配置 ──
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// ── 2. 初始化 JWT ──
	utils.InitJWT(cfg.JWT.Secret)

	// ── 3. 连接数据库（支持 MySQL / PostgreSQL） ──
	dbCfg := cfg.Database

	var dialector gorm.Dialector
	switch dbCfg.Driver {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DBName,
		)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.DBName,
		)
		dialector = mysql.Open(dsn)
	default:
		log.Fatalf("❌ 不支持的数据库驱动: %s（仅支持 mysql / postgres）", dbCfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开发阶段打印 SQL
	})
	if err != nil {
		log.Fatalf("❌ 连接 %s 失败: %v", dbCfg.Driver, err)
	}

	// ── 4. 自动迁移（建表） ──
	if err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Article{},
	); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库迁移完成")

	// ── 5. 初始化 Service 层 ──
	userSvc := &service.UserService{DB: db}
	articleSvc := &service.ArticleService{DB: db}
	categorySvc := &service.CategoryService{DB: db}

	// ── 6. 初始化 Handler 层 ──
	userHandler := &handler.UserHandler{Svc: userSvc}
	articleHandler := &handler.ArticleHandler{Svc: articleSvc}
	categoryHandler := &handler.CategoryHandler{Svc: categorySvc}

	// ── 7. 设置路由 ──
	r := router.Setup(userHandler, articleHandler, categoryHandler, cfg.Server.Mode)

	// ── 8. 启动 HTTP 服务器 ──
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🚀 服务启动: http://localhost%s\n", addr)
	fmt.Println("📌 接口前缀: /api/v1")
	fmt.Println("────────────────────────────────────────────")

	if err := r.Run(addr); err != nil {
		fmt.Printf("❌ 启动服务失败: %v\n", err)
		os.Exit(1)
	}
}
