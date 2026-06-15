package config

import "github.com/spf13/viper"

// ─────────────────────────────────────────────────
// Config — 全局配置，从 config.yaml 中读取
// ─────────────────────────────────────────────────

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`       // mysql / postgres
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	MaxOpenConn int    `mapstructure:"max_open_conns"` // 最大连接数
	MaxIdleConn int    `mapstructure:"max_idle_conns"` // 最大空闲连接数
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"` // 小时
}

// LoadConfig 读取 yaml 配置文件并解析到结构体
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
