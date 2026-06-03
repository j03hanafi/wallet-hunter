package config

import (
	"fmt"
	"os"
	"sync"
	"wallet-hunter/util"

	"gopkg.in/yaml.v3"
)

type ConfigReadOnly struct {
	RPCURL           string
	ReceiverAddress  string
	GeminiModel      string
	GasBufferPercent int64
}

type Config struct {
	mu sync.RWMutex

	RPCURL           string `yaml:"rpc_url"`
	ReceiverAddress  string `yaml:"receiver_address"`
	GasBufferPercent int64  `yaml:"gas_buffer_percent"`

	GeminiAPIKey string `yaml:"gemini_api_key"`
	GeminiModel  string `yaml:"gemini_model"`

	TargetChatID string `yaml:"target_chat_id"`
	BotToken     string `yaml:"bot_token"`
}

func NewConfig(path string) (*Config, error) {
	cfg, err := loadConfig(path)
	if err != nil {
		return nil, err
	}

	util.InitLogger()

	return cfg, nil
}

func (c *Config) Snapshot() ConfigReadOnly {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return ConfigReadOnly{
		RPCURL:           c.RPCURL,
		ReceiverAddress:  c.ReceiverAddress,
		GeminiModel:      c.GeminiModel,
		GasBufferPercent: c.GasBufferPercent,
	}
}

func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch key {
	case util.RPCURLKey:
		c.RPCURL = value.(string)
	case util.ReceiverAddressKey:
		c.ReceiverAddress = value.(string)
	case util.GeminiModelKey:
		c.GeminiModel = value.(string)
	case util.GasBufferPercentKey:
		c.GasBufferPercent = value.(int64)
	}
}

func loadConfig(path string) (*Config, error) {
	var cfg *Config

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// defaults
	if cfg.GeminiModel == "" {
		cfg.GeminiModel = "gemini-2.5-flash-lite"
	}
	if cfg.GasBufferPercent == 0 {
		cfg.GasBufferPercent = 10
	}

	if cfg.RPCURL == "" {
		return nil, fmt.Errorf("missing required config: rpc_url")
	}
	if cfg.ReceiverAddress == "" {
		return nil, fmt.Errorf("missing required config: receiver_address")
	}
	if !util.AddrRe.MatchString(cfg.ReceiverAddress) {
		return nil, fmt.Errorf("invalid receiver_address: %q", cfg.ReceiverAddress)
	}
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("missing required config: gemini_api_key")
	}
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("missing required config: bot_token")
	}
	if cfg.TargetChatID == "" {
		return nil, fmt.Errorf("missing required config: target_chat_id")
	}

	return cfg, nil
}
