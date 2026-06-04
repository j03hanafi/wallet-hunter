package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

func (t *Telegram) Updates(timeout int) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	return t.Client.GetUpdatesChan(u)
}

func (t *Telegram) ListenImageMessage(update tgbotapi.Update) (tgbotapi.PhotoSize, bool) {
	if update.Message == nil {
		return tgbotapi.PhotoSize{}, false
	}

	if update.Message.Chat.ID != t.cfg.TargetChatID {
		return tgbotapi.PhotoSize{}, false
	}

	if update.Message.IsCommand() {
		t.HandleCommand(update.Message)
		return tgbotapi.PhotoSize{}, false
	}

	if len(update.Message.Photo) == 0 {
		return tgbotapi.PhotoSize{}, false
	}

	return update.Message.Photo[len(update.Message.Photo)-1], true
}

func (t *Telegram) FetchImage(image tgbotapi.PhotoSize) ([]byte, error) {
	fileURL, err := t.Client.GetFileDirectURL(image.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file URL: %w", err)
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	return imgBytes, nil
}

func (t *Telegram) HandleCommand(m *tgbotapi.Message) {
	cmd := m.Command()
	arg := strings.TrimSpace(m.CommandArguments())

	var reply string
	reply = fmt.Sprintf("Command: %q\nArgument: %q\n", cmd, arg)

	switch cmd {
	case util.RPCURLKey:
		t.cfg.Set(util.RPCURLKey, arg)
	case util.ReceiverAddressKey:
		t.cfg.Set(util.ReceiverAddressKey, arg)
	case util.GeminiModelKey:
		t.cfg.Set(util.GeminiModelKey, arg)
	case util.GasBufferPercentKey:
		t.cfg.Set(util.GasBufferPercentKey, arg)
	case util.StatusKey:
		reply = "Wallet Hunter is running with the following config:\n\n"

		cfg := t.cfg.Snapshot()
		var cfgStr string
		cfgStr = fmt.Sprintf("RPC URL: %s\n", cfg.RPCURL)
		cfgStr += fmt.Sprintf("Receiver Address: %s\n", cfg.ReceiverAddress)
		cfgStr += fmt.Sprintf("Gemini Model: %s\n", cfg.GeminiModel)
		cfgStr += fmt.Sprintf("Gas Buffer Percent: %d\n", cfg.GasBufferPercent)

		reply += util.Pre(cfgStr)
		reply += "\n"
	default:
		reply = "Unknown command\n"
	}

	if reply != "" {
		_ = t.SendHTML(reply)
	}
}

func (t *Telegram) SendHTML(text string) error {
	msg := tgbotapi.NewMessage(t.cfg.TargetChatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	_, err := t.Client.Send(msg)
	return err
}
