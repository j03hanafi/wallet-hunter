package service

import (
	"context"
	"wallet-hunter/config"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ETH struct {
	cfg    *config.Config
	Client ethclient.Client
}

func NewClient(ctx context.Context, cfg *config.Config) (*ETH, error) {
	client, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	return &ETH{
		cfg:    cfg,
		Client: *client,
	}, nil
}
