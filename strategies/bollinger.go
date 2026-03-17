package strategies

import (
	"fmt"
	"math"

	"github.com/thesarfo/backtester/backtester"
	"github.com/thesarfo/backtester/binance"
)

// BollingerBands strategy:
// Buy when price touches or crosses below the lower band (mean reversion).
// Sell when price touches or crosses above the upper band.
type BollingerBands struct {
	period float64
	stdMul float64
}

func NewBollingerBands(period int, stdMultiplier float64) *BollingerBands {
	return &BollingerBands{period: float64(period), stdMul: stdMultiplier}
}

func (b *BollingerBands) Name() string {
	return fmt.Sprintf("Bollinger Bands (%.0f, %.1f)", b.period, b.stdMul)
}

func (b *BollingerBands) Compute(candles []binance.Candle, i int) backtester.Signal {
	p := int(b.period)
	if i < p {
		return backtester.Hold
	}

	closes := closePrices(candles[i-p+1 : i+1])
	mid := sma(closes)
	std := stdDevSlice(closes, mid)

	upper := mid + b.stdMul*std
	lower := mid - b.stdMul*std
	price := candles[i].Close

	if price <= lower {
		return backtester.Buy
	}
	if price >= upper {
		return backtester.Sell
	}
	return backtester.Hold
}

func stdDevSlice(prices []float64, mean float64) float64 {
	sum := 0.0
	for _, p := range prices {
		diff := p - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(prices)))
}
