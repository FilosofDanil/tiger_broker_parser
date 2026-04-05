package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/serhi/tiger-broker-parser/internal/display"
	"github.com/serhi/tiger-broker-parser/internal/model"
)

func PrintFuturesAccount(ctx context.Context, client *futures.Client) {
	display.PrintSection("Futures Wallet Summary")

	acc, err := client.NewGetAccountService().Do(ctx)
	if err != nil {
		fmt.Printf("  Error fetching futures account: %v\n\n", err)
		return
	}

	walletBal, _ := strconv.ParseFloat(acc.TotalWalletBalance, 64)
	marginBal, _ := strconv.ParseFloat(acc.TotalMarginBalance, 64)
	unrealizedPnL, _ := strconv.ParseFloat(acc.TotalUnrealizedProfit, 64)
	availableBal, _ := strconv.ParseFloat(acc.AvailableBalance, 64)
	initMargin, _ := strconv.ParseFloat(acc.TotalInitialMargin, 64)
	maintMargin, _ := strconv.ParseFloat(acc.TotalMaintMargin, 64)

	fmt.Printf("  Wallet balance:      %16.4f USDT\n", walletBal)
	fmt.Printf("  Margin balance:      %16.4f USDT\n", marginBal)
	fmt.Printf("  Unrealized PnL:      %16.4f USDT\n", unrealizedPnL)
	fmt.Printf("  Available balance:   %16.4f USDT\n", availableBal)
	fmt.Printf("  Initial margin:      %16.4f USDT\n", initMargin)
	fmt.Printf("  Maint. margin:       %16.4f USDT\n", maintMargin)
	fmt.Printf("  Fee tier: %d | Can trade: %v\n\n", acc.FeeTier, acc.CanTrade)
}

func PrintFuturesPositions(ctx context.Context, client *futures.Client) {
	display.PrintSection("Open Futures Positions")

	risks, err := client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		fmt.Printf("  Error fetching positions: %v\n\n", err)
		return
	}

	var open []model.Position
	for _, r := range risks {
		amt, _ := strconv.ParseFloat(r.PositionAmt, 64)
		if amt == 0 {
			continue
		}
		entry, _ := strconv.ParseFloat(r.EntryPrice, 64)
		mark, _ := strconv.ParseFloat(r.MarkPrice, 64)
		liq, _ := strconv.ParseFloat(r.LiquidationPrice, 64)
		unPnL, _ := strconv.ParseFloat(r.UnRealizedProfit, 64)
		side := "LONG"
		if amt < 0 {
			side = "SHORT"
		}
		open = append(open, model.Position{
			Symbol: r.Symbol, Side: side, Amt: amt,
			EntryPrice: entry, MarkPrice: mark, LiqPrice: liq,
			UnPnL: unPnL, Leverage: r.Leverage, MarginType: r.MarginType,
		})
	}

	if len(open) == 0 {
		fmt.Print("  No open positions.\n\n")
		return
	}

	fmt.Printf("  %-12s %-6s %10s %12s %12s %12s %12s %4s %8s\n",
		"Symbol", "Side", "Size", "Entry", "Mark", "Liq.", "UnPnL", "Lev", "Margin")
	fmt.Printf("  %-12s %-6s %10s %12s %12s %12s %12s %4s %8s\n",
		"------", "----", "----", "-----", "----", "----", "-----", "---", "------")
	for _, p := range open {
		fmt.Printf("  %-12s %-6s %10.4f %12.4f %12.4f %12.4f %12.4f %3sx %8s\n",
			p.Symbol, p.Side, p.Amt, p.EntryPrice, p.MarkPrice, p.LiqPrice, p.UnPnL, p.Leverage, p.MarginType)
	}
	fmt.Println()
}

func PrintFuturesOpenOrders(ctx context.Context, client *futures.Client) {
	display.PrintSection("Open Futures Orders")

	orders, err := client.NewListOpenOrdersService().Do(ctx)
	if err != nil {
		fmt.Printf("  Error fetching futures open orders: %v\n\n", err)
		return
	}

	if len(orders) == 0 {
		fmt.Print("  No open futures orders.\n\n")
		return
	}

	fmt.Printf("  %-12s %-6s %-10s %-6s %12s %12s %-16s\n",
		"Symbol", "Side", "Type", "PosSide", "Price", "Qty", "Time")
	fmt.Printf("  %-12s %-6s %-10s %-6s %12s %12s %-16s\n",
		"------", "----", "----", "-------", "-----", "---", "----")
	for _, o := range orders {
		ts := time.Unix(o.Time/1000, 0).UTC().Format("2006-01-02 15:04")
		price, _ := strconv.ParseFloat(o.Price, 64)
		qty, _ := strconv.ParseFloat(o.OrigQuantity, 64)
		fmt.Printf("  %-12s %-6s %-10s %-6s %12.4f %12.6f %-16s\n",
			o.Symbol, o.Side, o.Type, o.PositionSide, price, qty, ts)
	}
	fmt.Println()
}
