package config

import (
	"fmt"
	"os"
	"time"

	"github.com/DanRulev/vocabot.git/pkg/validator"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig `mapstructure:"app" validate:"required"`
	BotToken string    `mapstructure:"bot_token" validate:"required"`
	DB       DBConfig  `mapstructure:"db" validate:"required"`
	Env      string    `mapstructure:"env" validate:"oneof=development production staging"`
}

type AppConfig struct {
	Timeout time.Duration `mapstructure:"timeout" validate:"min=1"`
}

type DBConfig struct {
	Conn DBConn `mapstructure:"conn"`
	Cfg  DBCfg  `mapstructure:"cfg"`
}

type DBConn struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     string `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Name     string `mapstructure:"name" validate:"required"`
	SSL      string `mapstructure:"ssl" validate:"oneof=disable require verify-full"`
}

type DBCfg struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1,max=1000"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=0,max=100"`
	ConnMaxLifeTime time.Duration `mapstructure:"conn_max_life_time" validate:"min=0"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" validate:"min=0"`
}

func Init() (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()

	configName := os.Getenv("CONFIG_NAME")
	if configName == "" {
		configName = "default"
	}

	v.AddConfigPath("configs")
	v.SetConfigName(configName)

	if err := v.BindEnv("bot_token", "BOT_TOKEN"); err != nil {
		return nil, fmt.Errorf("failed to bind BOT_TOKEN: %w", err)
	}
	if err := v.BindEnv("db.conn.host", "DB_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind DB_HOST: %w", err)
	}
	if err := v.BindEnv("db.conn.port", "DB_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind DB_PORT: %w", err)
	}
	if err := v.BindEnv("db.conn.user", "DB_USER"); err != nil {
		return nil, fmt.Errorf("failed to bind DB_USER: %w", err)
	}
	if err := v.BindEnv("db.conn.password", "DB_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind DB_PASSWORD: %w", err)
	}
	if err := v.BindEnv("db.conn.name", "DB_NAME"); err != nil {
		return nil, fmt.Errorf("failed to bind DB_NAME: %w", err)
	}
	if err := v.BindEnv("db.conn.ssl", "DB_SSL"); err != nil {
		return nil, fmt.Errorf("failed to bind DB_SSL: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := Config{}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validator.ValidateStruct(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
