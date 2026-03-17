package strategies

import (
	"fmt"

	"github.com/thesarfo/backtester/backtester"
	"github.com/thesarfo/backtester/binance"
)

// EMACrossover generates a Buy when fast EMA crosses above slow EMA,
// and a Sell when fast EMA crosses below slow EMA.
type EMACrossover struct {
	fast int
	slow int
}

func NewEMACrossover(fast, slow int) *EMACrossover {
	return &EMACrossover{fast: fast, slow: slow}
}

func (e *EMACrossover) Name() string {
	return fmt.Sprintf("EMA Crossover (%d/%d)", e.fast, e.slow)
}

func (e *EMACrossover) Compute(candles []binance.Candle, i int) backtester.Signal {
	if i < e.slow {
		return backtester.Hold
	}

	closes := closePrices(candles[:i+1])

	fastNow := ema(closes, e.fast)
	slowNow := ema(closes, e.slow)

	// Previous bar
	if i < e.slow+1 {
		return backtester.Hold
	}
	prevCloses := closePrices(candles[:i])
	fastPrev := ema(prevCloses, e.fast)
	slowPrev := ema(prevCloses, e.slow)

	// Crossover detection
	crossedAbove := fastPrev <= slowPrev && fastNow > slowNow
	crossedBelow := fastPrev >= slowPrev && fastNow < slowNow

	if crossedAbove {
		return backtester.Buy
	}
	if crossedBelow {
		return backtester.Sell
	}
	return backtester.Hold
}

// ema computes the Exponential Moving Average for the last `period` bars
func ema(prices []float64, period int) float64 {
	if len(prices) < period {
		return sma(prices)
	}

	k := 2.0 / float64(period+1)

	// Seed with SMA of first `period` values
	seed := sma(prices[:period])
	result := seed

	for _, p := range prices[period:] {
		result = p*k + result*(1-k)
	}
	return result
}

func sma(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	return sum / float64(len(prices))
}

func closePrices(candles []binance.Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Close
	}
	return out
}
