package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("BINANCE_API_KEY and BINANCE_API_SECRET must be set in .env or environment")
	}

	spotClient := binance.NewClient(apiKey, apiSecret)
	futuresClient := futures.NewClient(apiKey, apiSecret)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	printHeader("Tiger Broker (Binance) — Account Overview")

	printAccountBalance(ctx, spotClient)
	printRecentTrades(ctx, spotClient)
	printOpenOrders(ctx, spotClient)

	printHeader("Futures Account")
	printFuturesAccount(ctx, futuresClient)
	printFuturesPositions(ctx, futuresClient)
	printFuturesOpenOrders(ctx, futuresClient)
}

func printAccountBalance(ctx context.Context, client *binance.Client) {
	printSection("Account Balances")

	account, err := client.NewGetAccountService().Do(ctx)
	if err != nil {
		fmt.Printf("  Error fetching account: %v\n\n", err)
		return
	}

	type nonZeroBalance struct {
		asset  string
		free   float64
		locked float64
		total  float64
	}

	var balances []nonZeroBalance
	for _, b := range account.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		total := free + locked
		if total > 0 {
			balances = append(balances, nonZeroBalance{b.Asset, free, locked, total})
		}
	}

	sort.Slice(balances, func(i, j int) bool {
		return balances[i].total > balances[j].total
	})

	if len(balances) == 0 {
		fmt.Println("  No non-zero balances found.")
	} else {
		fmt.Printf("  %-10s %16s %16s %16s\n", "Asset", "Free", "Locked", "Total")
		fmt.Printf("  %-10s %16s %16s %16s\n", "-----", "----", "------", "-----")
		for _, b := range balances {
			fmt.Printf("  %-10s %16.6f %16.6f %16.6f\n", b.asset, b.free, b.locked, b.total)
		}
	}

	fmt.Printf("\n  Maker commission: %d bps | Taker commission: %d bps\n",
		account.MakerCommission, account.TakerCommission)
	fmt.Printf("  Can trade: %v | Can deposit: %v | Can withdraw: %v\n\n",
		account.CanTrade, account.CanDeposit, account.CanWithdraw)
}

func printRecentTrades(ctx context.Context, client *binance.Client) {
	printSection("Recent Trade Activity (BTCUSDT, ETHUSDT, SOLUSDT)")

	symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	totalTrades := 0

	for _, symbol := range symbols {
		trades, err := client.NewListTradesService().
			Symbol(symbol).
			Limit(5).
			Do(ctx)
		if err != nil {
			fmt.Printf("  [%s] Error: %v\n", symbol, err)
			continue
		}
		if len(trades) == 0 {
			continue
		}

		fmt.Printf("  ── %s ──\n", symbol)
		fmt.Printf("  %-22s %-6s %-14s %-14s %-14s\n", "Time", "Side", "Price", "Qty", "Commission")
		for _, t := range trades {
			ts := time.Unix(t.Time/1000, 0).UTC().Format("2006-01-02 15:04:05")
			side := "BUY"
			if !t.IsBuyer {
				side = "SELL"
			}
			price, _ := strconv.ParseFloat(t.Price, 64)
			qty, _ := strconv.ParseFloat(t.Quantity, 64)
			commission, _ := strconv.ParseFloat(t.Commission, 64)
			fmt.Printf("  %-22s %-6s %14.4f %14.6f %14.8f %s\n",
				ts, side, price, qty, commission, t.CommissionAsset)
		}
		fmt.Println()
		totalTrades += len(trades)
	}

	if totalTrades == 0 {
		fmt.Print("  No recent trades found for tracked symbols.\n\n")
	}
}

func printOpenOrders(ctx context.Context, client *binance.Client) {
	printSection("Open Orders")

	orders, err := client.NewListOpenOrdersService().Do(ctx)
	if err != nil {
		fmt.Printf("  Error fetching open orders: %v\n\n", err)
		return
	}

	if len(orders) == 0 {
		fmt.Print("  No open orders.\n\n")
		return
	}

	fmt.Printf("  %-10s %-6s %-10s %-14s %-14s %-16s\n",
		"Symbol", "Side", "Type", "Price", "Qty", "Time")
	fmt.Printf("  %-10s %-6s %-10s %-14s %-14s %-16s\n",
		"------", "----", "----", "-----", "---", "----")
	for _, o := range orders {
		ts := time.Unix(o.Time/1000, 0).UTC().Format("2006-01-02 15:04")
		price, _ := strconv.ParseFloat(o.Price, 64)
		qty, _ := strconv.ParseFloat(o.OrigQuantity, 64)
		fmt.Printf("  %-10s %-6s %-10s %14.4f %14.6f %-16s\n",
			o.Symbol, o.Side, o.Type, price, qty, ts)
	}
	fmt.Println()
}

func printFuturesAccount(ctx context.Context, client *futures.Client) {
	printSection("Futures Wallet Summary")

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

func printFuturesPositions(ctx context.Context, client *futures.Client) {
	printSection("Open Futures Positions")

	risks, err := client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		fmt.Printf("  Error fetching positions: %v\n\n", err)
		return
	}

	type pos struct {
		symbol     string
		side       string
		amt        float64
		entryPrice float64
		markPrice  float64
		liqPrice   float64
		unPnL      float64
		leverage   string
		marginType string
	}

	var open []pos
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
		open = append(open, pos{r.Symbol, side, amt, entry, mark, liq, unPnL, r.Leverage, r.MarginType})
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
			p.symbol, p.side, p.amt, p.entryPrice, p.markPrice, p.liqPrice, p.unPnL, p.leverage, p.marginType)
	}
	fmt.Println()
}

func printFuturesOpenOrders(ctx context.Context, client *futures.Client) {
	printSection("Open Futures Orders")

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

func printHeader(title string) {
	border := "═══════════════════════════════════════════════════════"
	fmt.Printf("\n%s\n  %s\n  %s\n%s\n\n", border, title, time.Now().UTC().Format("2006-01-02 15:04:05 UTC"), border)
}

func printSection(title string) {
	fmt.Printf("┌─ %s\n", title)
}
