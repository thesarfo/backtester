package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/thesarfo/backtester/backtester"
	"github.com/thesarfo/backtester/binance"
	"github.com/thesarfo/backtester/dashboard"
	"github.com/thesarfo/backtester/strategies"
	"github.com/thesarfo/backtester/yahoo"
)

var reader = bufio.NewReader(os.Stdin)

const (
	bold   = "\033[1m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	grey   = "\033[90m"
	reset  = "\033[0m"
)

func header(text string) {
	fmt.Printf("\n%s%s▸ %s%s\n", bold, cyan, text, reset)
}

func hint(text string) {
	fmt.Printf("  %s%s%s\n", grey, text, reset)
}

func success(text string) {
	fmt.Printf("  %s✓ %s%s\n", green, text, reset)
}

func prompt(text string) string {
	fmt.Printf("  %s→ %s%s ", bold, text, reset)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}


func ask(question string, options []string) int {
	header(question)
	for i, opt := range options {
		fmt.Printf("  %s[%d]%s %s\n", yellow, i+1, reset, opt)
	}
	for {
		raw := prompt("Enter number")
		n, err := strconv.Atoi(raw)
		if err == nil && n >= 1 && n <= len(options) {
			fmt.Printf("  %s✓ Selected: %s%s\n", green, options[n-1], reset)
			return n
		}
		fmt.Printf("  Please enter a number between 1 and %d\n", len(options))
	}
}


func askFree(question, hint_, defaultVal string) string {
	header(question)
	if hint_ != "" {
		hint(hint_)
	}
	if defaultVal != "" {
		fmt.Printf("  %s(press enter to use default: %s)%s\n", grey, defaultVal, reset)
	}
	raw := prompt("Your answer")
	if raw == "" {
		fmt.Printf("  %s✓ Using default: %s%s\n", green, defaultVal, reset)
		return defaultVal
	}
	fmt.Printf("  %s✓ Set to: %s%s\n", green, raw, reset)
	return raw
}

func divider() {
	fmt.Printf("\n%s────────────────────────────────────────%s\n", grey, reset)
}


func main() {
	fmt.Printf("\n%s%s  Strategy Backtester  %s\n", bold, cyan, reset)
	fmt.Printf("%sTest trading strategies on real crypto & stock market data%s\n", grey, reset)
	divider()

	modeChoice := ask(
		"Step 1 of 4, What market do you want to backtest?",
		[]string{
			"Crypto  (fetches live data from Binance, no account needed)",
			"Stocks  (fetches live data from Yahoo Finance, no account needed)",
		},
	)
	mode := map[int]string{1: "crypto", 2: "stocks"}[modeChoice]

	var symbolHint, symbolDefault string
	switch mode {
	case "crypto":
		symbolHint = "Any Binance trading pair e.g. BTCUSDT, ETHUSDT, SOLUSDT, BNBUSDT"
		symbolDefault = "BTCUSDT"
	case "stocks":
		symbolHint = "Any stock ticker e.g. AAPL (Apple), TSLA (Tesla), NVDA (Nvidia), ^GSPC (S&P 500)"
		symbolDefault = "AAPL"
	}
	symbol := askFree("Step 2 of 4, Which symbol do you want to test?", symbolHint, symbolDefault)
	symbol = strings.ToUpper(symbol)

	var intervalOptions []string
	var intervalValues []string
	switch mode {
	case "crypto":
		intervalOptions = []string{
			"1h  , Hourly bars  (good balance of detail and history, recommended)",
			"4h  , 4-hour bars  (smoother trends, less noise)",
			"1d  , Daily bars   (long-term view, up to ~3 years of data)",
			"15m , 15-min bars  (short-term, last ~10 days only)",
		}
		intervalValues = []string{"1h", "4h", "1d", "15m"}
	case "stocks":
		intervalOptions = []string{
			"1d  , Daily bars   (recommended for stocks, goes back years)",
			"1wk , Weekly bars  (very long-term trends, reduced noise)",
			"1h  , Hourly bars  (last 60 days only, Yahoo Finance limit)",
		}
		intervalValues = []string{"1d", "1wk", "1h"}
	}
	intervalChoice := ask("Step 3 of 4, Which candle interval (timeframe) do you want?", intervalOptions)
	interval := intervalValues[intervalChoice-1]

	capitalStr := askFree(
		"Step 4 of 4, What is your starting capital? (in USD)",
		"This is the virtual amount the backtester starts with, no real money is used",
		"10000",
	)
	capital, err := strconv.ParseFloat(capitalStr, 64)
	if err != nil || capital <= 0 {
		fmt.Printf("\n  Invalid amount, using default of $10,000\n")
		capital = 10000
	}

	divider()
	fmt.Printf("\n%s%s  Your backtest is configured%s\n\n", bold, green, reset)
	fmt.Printf("  %-18s %s%s%s\n", "Market:", bold, strings.Title(mode), reset)
	fmt.Printf("  %-18s %s%s%s\n", "Symbol:", bold, symbol, reset)
	fmt.Printf("  %-18s %s%s%s\n", "Interval:", bold, interval, reset)
	fmt.Printf("  %-18s %s$%.0f%s\n", "Starting capital:", bold, capital, reset)
	fmt.Printf("  %-18s %sEMA Crossover, RSI, Bollinger Bands, Buy & Hold%s\n", "Strategies:", grey, reset)
	divider()

	confirm := prompt("\nPress enter to run, or type 'quit' to exit")
	if strings.ToLower(confirm) == "quit" {
		fmt.Println("\n  Exiting. Run 'go run main.go' to start again.")
		os.Exit(0)
	}

	limit := 1000
	fmt.Printf("\n%s  Fetching data...%s\n", grey, reset)

	var candles []binance.Candle
	switch mode {
	case "crypto":
		candles, err = binance.NewClient().FetchKlines(symbol, interval, limit)
	case "stocks":
		candles, err = yahoo.NewClient().FetchCandles(symbol, interval, limit)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n  %sError fetching data: %v%s\n", yellow, err, reset)
		fmt.Fprintf(os.Stderr, "  Double-check the symbol and try again.\n")
		os.Exit(1)
	}

	success(fmt.Sprintf("Fetched %d candles  (%s → %s)",
		len(candles),
		candles[0].OpenTime.Format("Jan 02 2006"),
		candles[len(candles)-1].OpenTime.Format("Jan 02 2006"),
	))

	fmt.Printf("\n%s  Running strategies...%s\n\n", grey, reset)

	strats := []struct {
		name     string
		strategy backtester.Strategy
	}{
		{"EMA Crossover (9/21)",    strategies.NewEMACrossover(9, 21)},
		{"RSI (14, 30/70)",         strategies.NewRSI(14, 30, 70)},
		{"Bollinger Bands (20, 2)", strategies.NewBollingerBands(20, 2.0)},
		{"Buy & Hold",              strategies.NewBuyAndHold()},
	}

	var results []dashboard.StrategyResult
	for _, s := range strats {
		bt := backtester.New(capital, s.strategy)
		result := bt.Run(candles)
		results = append(results, dashboard.StrategyResult{Name: s.name, Result: result})

		retColor := green
		if result.TotalReturnPct < 0 {
			retColor = yellow
		}
		fmt.Printf("  %s%-26s%s  return %s%+.2f%%%s   drawdown %.2f%%   trades %d\n",
			bold, s.name, reset,
			retColor, result.TotalReturnPct, reset,
			result.MaxDrawdownPct,
			result.TotalTrades,
		)
	}

	cfg := dashboard.Config{
		Symbol:         symbol,
		Interval:       interval,
		InitialCapital: capital,
		Candles:        candles,
		Mode:           mode,
	}

	divider()
	if err := dashboard.Render(results, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "\n  Dashboard error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s%s  Done! Your dashboard has opened in the browser.%s\n\n", bold, green, reset)
}