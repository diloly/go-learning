package store

import (
	"errors"
	"fmt"

	"shortlink/config"
	"shortlink/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─────────────────────────────────────────────────
// DBStore — MySQL / PostgreSQL 数据持久层
// ─────────────────────────────────────────────────

type DBStore struct {
	DB *gorm.DB
}

// NewDBStore 根据配置创建数据库连接
func NewDBStore(cfg config.DatabaseConfig) (*DBStore, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
		)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
		)
		dialector = mysql.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.ShortLink{}); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 连接池设置
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)

	fmt.Println("✅ 数据库连接成功")
	return &DBStore{DB: db}, nil
}

// ── CRUD ─────────────────────────────────────────

// Create 创建短链接，写入数据库后返回带 ID 的记录
func (s *DBStore) Create(originalURL string) (*model.ShortLink, error) {
	link := &model.ShortLink{
		OriginalURL: originalURL,
	}

	if err := s.DB.Create(link).Error; err != nil {
		return nil, err
	}

	// 根据自增 ID 生成短码
	link.ShortCode = model.EncodeBase62(link.ID)
	if err := s.DB.Model(link).Update("short_code", link.ShortCode).Error; err != nil {
		return nil, err
	}

	return link, nil
}

// GetByShortCode 按短码查询
func (s *DBStore) GetByShortCode(code string) (*model.ShortLink, error) {
	var link model.ShortLink
	err := s.DB.Where("short_code = ?", code).First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到返回 nil, nil
		}
		return nil, err
	}
	return &link, nil
}

// IncrementCount 增加访问计数
func (s *DBStore) IncrementCount(code string) error {
	return s.DB.Model(&model.ShortLink{}).
		Where("short_code = ?", code).
		UpdateColumn("access_count", gorm.Expr("access_count + 1")).
		Error
}

// GetStats 获取短链接统计
func (s *DBStore) GetStats(code string) (*model.ShortLink, error) {
	return s.GetByShortCode(code)
}

// Close 关闭数据库连接
func (s *DBStore) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
