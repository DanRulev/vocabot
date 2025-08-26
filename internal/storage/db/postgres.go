package db

import (
	"context"
	"fmt"
	"time"

	"github.com/DanRulev/vocabot.git/internal/config"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

func InitDB(cfg config.DBConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%v port=%v dbname=%v user=%v password=%v sslmode=%v",
		cfg.Conn.Host, cfg.Conn.Port, cfg.Conn.Name, cfg.Conn.User, cfg.Conn.Password, cfg.Conn.SSL)
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed open db connect: %w", err)
	}

	db.SetMaxOpenConns(cfg.Cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Cfg.ConnMaxLifeTime)
	db.SetConnMaxIdleTime(cfg.Cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed db ping: %w", err)
	}

	return db, nil
}
