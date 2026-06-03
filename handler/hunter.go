package handler

import "wallet-hunter/service"

type Hunter struct {
	ETHClient      service.ETH
	TelegramClient service.Telegram
}
