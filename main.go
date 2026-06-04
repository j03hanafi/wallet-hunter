package main

import (
	"context"
	"wallet-hunter/config"
	"wallet-hunter/handler"
	"wallet-hunter/service"
	"wallet-hunter/util"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

	f, err := util.InitLogger()
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ethClient, err := service.NewETH(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer ethClient.Client.Close()

	telegramClient, err := service.NewTelegram(cfg)
	if err != nil {
		panic(err)
	}

	geminiClient, err := service.NewGemini(ctx, cfg)
	if err != nil {
		panic(err)
	}

	hunter := handler.NewHunter(cfg, ethClient, telegramClient, geminiClient)
	hunter.Start(ctx)
}
