package service

import (
	"log"
	"wallet-hunter/config"
	"wallet-hunter/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram struct {
	cfg    *config.Config
	Client *tgbotapi.BotAPI
}

func NewTelegram(cfg *config.Config) (*Telegram, error) {
	client, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	telegram := &Telegram{
		cfg:    cfg,
		Client: client,
	}

	if err = telegram.registerCommands(); err != nil {
		return nil, err
	}

	return telegram, nil
}

func (t *Telegram) registerCommands() error {
	cmds := []tgbotapi.BotCommand{
		{Command: util.RPCURLKey, Description: "set RPC URL, e.g. /rpc_url http://localhost:8545"},
		{Command: util.ReceiverAddressKey, Description: "receive funds to this address, e.g. /receiver_address 0x..."},
		{Command: util.GeminiModelKey, Description: "set model, e.g. /gemini_model gemini-2.5-flash-lite"},
		{Command: util.GasBufferPercentKey, Description: "set gas buffer percentage, e.g. /gas_buffer_percent 10"},
		{Command: util.StatusKey, Description: "show current config"},
	}

	if _, err := t.Client.Request(tgbotapi.NewSetMyCommands(cmds...)); err != nil {
		log.Println("set commands:", err)
		return err
	}
	return nil
}
