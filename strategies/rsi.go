package strategies

import (
	"fmt"

	"github.com/thesarfo/backtester/backtester"
	"github.com/thesarfo/backtester/binance"
)

// RSI strategy: Buy when RSI crosses above oversold level,
// Sell when RSI crosses above overbought level.
type RSIStrategy struct {
	period     int
	oversold   float64
	overbought float64
}

func NewRSI(period int, oversold, overbought float64) *RSIStrategy {
	return &RSIStrategy{period: period, oversold: oversold, overbought: overbought}
}

func (r *RSIStrategy) Name() string {
	return fmt.Sprintf("RSI(%d, %.0f/%.0f)", r.period, r.oversold, r.overbought)
}

func (r *RSIStrategy) Compute(candles []binance.Candle, i int) backtester.Signal {
	if i < r.period+1 {
		return backtester.Hold
	}

	closes := closePrices(candles[:i+1])
	rsiNow := rsi(closes, r.period)

	prevCloses := closePrices(candles[:i])
	rsiPrev := rsi(prevCloses, r.period)

	if rsiPrev <= r.oversold && rsiNow > r.oversold {
		return backtester.Buy
	}

	if rsiPrev <= r.overbought && rsiNow > r.overbought {
		return backtester.Sell
	}

	return backtester.Hold
}

func rsi(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50 
	}

	changes := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		changes[i-1] = prices[i] - prices[i-1]
	}

	window := changes[len(changes)-period:]

	var gains, losses float64
	for _, c := range window {
		if c > 0 {
			gains += c
		} else {
			losses += -c
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}
