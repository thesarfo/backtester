package strategies

import (
	"github.com/thesarfo/backtester/backtester"
	"github.com/thesarfo/backtester/binance"
)

// BuyAndHold simply buys at the first candle and never sells.
// Used as a baseline to beat.
type BuyAndHold struct {
	bought bool
}

func NewBuyAndHold() *BuyAndHold {
	return &BuyAndHold{}
}

func (b *BuyAndHold) Name() string {
	return "Buy & Hold"
}

func (b *BuyAndHold) Compute(candles []binance.Candle, i int) backtester.Signal {
	if i == 0 {
		return backtester.Buy
	}
	return backtester.Hold
}
