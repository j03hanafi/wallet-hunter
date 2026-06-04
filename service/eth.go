package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"wallet-hunter/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ETH struct {
	cfg    *config.Config
	Client *ethclient.Client
}

type TX struct {
	GasLimit uint64
	GasPrice *big.Int
	Gas      *big.Int
	Amount   *big.Int
	Total    *big.Int
	Hash     string
	Sender   string
	Receiver string
}

func NewETH(ctx context.Context, cfg *config.Config) (*ETH, error) {
	client, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, err
	}

	return &ETH{
		cfg:    cfg,
		Client: client,
	}, nil
}

func (e *ETH) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	bal, err := e.Client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("balance query failed: %w", err)
	}
	return bal, nil
}

func (e *ETH) SendToAddress(ctx context.Context, privKey *ecdsa.PrivateKey, to common.Address, balance *big.Int) (TX, error) {
	pub, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return TX{}, fmt.Errorf("cannot assert public key to ECDSA")
	}
	sender := crypto.PubkeyToAddress(*pub)

	nonce, err := e.Client.PendingNonceAt(ctx, sender)
	if err != nil {
		return TX{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := e.Client.SuggestGasPrice(ctx)
	if err != nil {
		return TX{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas with Value=0, NOT full balance.
	// Estimating with full balance leaves no room for gas → node returns
	// "insufficient funds" during simulation. Native EOA transfer gas does
	// not depend on value, so 0 gives the correct limit (normally 21000).
	gasLimit, err := e.Client.EstimateGas(ctx, ethereum.CallMsg{
		From:  sender,
		To:    &to,
		Value: big.NewInt(0),
	})
	if err != nil {
		return TX{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Buffer price: gasPrice * (100+buf) / 100  (big.Int, no floats)
	bufferedGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(100+e.cfg.GasBufferPercent))
	bufferedGasPrice.Div(bufferedGasPrice, big.NewInt(100))

	// reserve = gasLimit * bufferedGasPrice
	reserve := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), bufferedGasPrice)

	tx := TX{
		GasLimit: gasLimit,
		GasPrice: bufferedGasPrice,
		Gas:      reserve,
		Total:    balance,
	}

	// amount = balance - reserve
	amount := new(big.Int).Sub(balance, reserve)
	if amount.Sign() <= 0 {
		return TX{}, fmt.Errorf("balance too low to cover gas: balance=%s reserve=%s wei", balance.String(), reserve.String())
	}

	tx.Amount = amount

	chainID, err := e.Client.NetworkID(ctx)
	if err != nil {
		return TX{}, fmt.Errorf("failed to get network ID: %w", err)
	}

	transaction := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: bufferedGasPrice,
		Data:     nil,
	})

	signedTx, err := types.SignTx(transaction, types.LatestSignerForChainID(chainID), privKey)
	if err != nil {
		return TX{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	if err = e.Client.SendTransaction(ctx, signedTx); err != nil {
		return TX{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	tx.Hash = signedTx.Hash().Hex()

	return tx, nil
}
