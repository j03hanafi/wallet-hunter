package util

import (
	"fmt"
	"math/big"
	"strings"
)

// WeiBreakdown formats a wei amount as ETH | gwei | wei.
func WeiBreakdown(label string, wei *big.Int) string {
	f := new(big.Float).SetInt(wei)
	eth := new(big.Float).Quo(f, big.NewFloat(1e18))
	gwei := new(big.Float).Quo(f, big.NewFloat(1e9))
	return fmt.Sprintf("%-8s %s ETH | %s gwei | %s wei",
		label, TrimZeros(eth.Text('f', 18)), TrimZeros(gwei.Text('f', 9)), wei.String())
}

// TrimZeros drops trailing zeros (and a trailing dot) from a decimal string.
func TrimZeros(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}
