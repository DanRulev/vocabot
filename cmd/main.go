package main

import (
	"log"

	"github.com/DanRulev/vocabot.git/internal/bot"
	"github.com/DanRulev/vocabot.git/internal/client"
	"github.com/DanRulev/vocabot.git/internal/config"
	"github.com/DanRulev/vocabot.git/internal/repository"
	"github.com/DanRulev/vocabot.git/internal/service"
	"github.com/DanRulev/vocabot.git/internal/storage/cache"
	"github.com/DanRulev/vocabot.git/internal/storage/db"

	"go.uber.org/zap"
)

func setupLogger(env string) *zap.Logger {
	var logger *zap.Logger
	if env == "development" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	return logger
}

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatal("failed load config " + err.Error())
		return
	}

	logger := setupLogger(cfg.Env)

	db, err := db.InitDB(cfg.DB)
	if err != nil {
		logger.Fatal("failed init db", zap.Error(err))
	}

	repos := repository.NewRepository(db)

	clients := client.InitClients()
	services := service.InitServices(clients, repos, logger)
	cache := cache.NewCache()

	handler, err := bot.NewTelegramAPI(cfg.BotToken, cfg.Env, services, cache)
	if err != nil {
		logger.Fatal(err.Error())
		return
	}

	handler.Start()
}
