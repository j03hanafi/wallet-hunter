package main

import (
	"context"
	"wallet-hunter/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

}
