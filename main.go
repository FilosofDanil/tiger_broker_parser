package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	apiKey := "z6HwUEkbb2pYk0SKeJOccJ5Q7AlDBZOMUCeFx0jGmkRx6u4S8Vd1ohrQBD9ApHit"
	apiSecret := "dMmRAyc2zVABZjqjXB6tIdp46HI5CwZlfUuYexZ19iVP8OOEpONjUtp8WtP5T6IV"

	if apiKey == "" || apiSecret == "" {
		log.Fatal("BINANCE_API_KEY and BINANCE_API_SECRET must be set in .env or environment")
	}

	client := binance.NewClient(apiKey, apiSecret)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	printHeader("Tiger Broker (Binance) — Account Overview")
	printAccountBalance(ctx, client)
	printRecentTrades(ctx, client)
	printOpenOrders(ctx, client)
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

func printHeader(title string) {
	border := "═══════════════════════════════════════════════════════"
	fmt.Printf("\n%s\n  %s\n  %s\n%s\n\n", border, title, time.Now().UTC().Format("2006-01-02 15:04:05 UTC"), border)
}

func printSection(title string) {
	fmt.Printf("┌─ %s\n", title)
}
