package main

import (
	"context"
	"log"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/serhi/tiger-broker-parser/internal/config"
	"github.com/serhi/tiger-broker-parser/internal/display"
	"github.com/serhi/tiger-broker-parser/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	spotClient := binance.NewClient(cfg.APIKey, cfg.APISecret)
	futuresClient := futures.NewClient(cfg.APIKey, cfg.APISecret)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	display.PrintHeader("Tiger Broker (Binance) — Account Overview")
	service.PrintAccountBalance(ctx, spotClient)
	service.PrintRecentTrades(ctx, spotClient)
	service.PrintOpenOrders(ctx, spotClient)

	display.PrintHeader("Futures Account")
	service.PrintFuturesAccount(ctx, futuresClient)
	service.PrintFuturesPositions(ctx, futuresClient)
	service.PrintFuturesOpenOrders(ctx, futuresClient)
}
