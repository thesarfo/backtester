package backtester

import (
	"math"

	"github.com/thesarfo/backtester/binance"
)

type Signal int

const (
	Hold Signal = iota
	Buy
	Sell
)

type Strategy interface {
	Name() string
	Compute(candles []binance.Candle, i int) Signal
}

type Trade struct {
	EntryPrice float64
	ExitPrice  float64
	PnL        float64
	PnLPct     float64
}

type Result struct {
	FinalCapital   float64
	TotalReturnPct float64
	MaxDrawdownPct float64
	SharpeRatio    float64
	TotalTrades    int
	WinRate        float64
	ProfitFactor   float64
	Trades         []Trade
	EquityCurve    []float64
}

type Backtester struct {
	initialCapital float64
	strategy       Strategy
	commission     float64 
}

func New(initialCapital float64, strategy Strategy) *Backtester {
	return &Backtester{
		initialCapital: initialCapital,
		strategy:       strategy,
		commission:     0.001, 
	}
}

func (bt *Backtester) Run(candles []binance.Candle) Result {
	capital := bt.initialCapital
	var position float64 
	var entryPrice float64
	inPosition := false

	var trades []Trade
	equityCurve := make([]float64, 0, len(candles))
	peakEquity := capital

	for i := range candles {
		price := candles[i].Close
		signal := bt.strategy.Compute(candles, i)

		switch signal {
		case Buy:
			if !inPosition {
				units := (capital * (1 - bt.commission)) / price
				position = units
				entryPrice = price
				capital = 0
				inPosition = true
			}

		case Sell:
			if inPosition {
				proceeds := position * price * (1 - bt.commission)
				pnl := proceeds - (position * entryPrice)
				pnlPct := (price/entryPrice - 1) * 100

				trades = append(trades, Trade{
					EntryPrice: entryPrice,
					ExitPrice:  price,
					PnL:        pnl,
					PnLPct:     pnlPct,
				})

				capital = proceeds
				position = 0
				inPosition = false
			}
		}

		equity := capital
		if inPosition {
			equity = position * price
		}
		equityCurve = append(equityCurve, equity)

		if equity > peakEquity {
			peakEquity = equity
		}
	}

	if inPosition {
		lastPrice := candles[len(candles)-1].Close
		proceeds := position * lastPrice * (1 - bt.commission)
		pnl := proceeds - (position * entryPrice)
		pnlPct := (lastPrice/entryPrice - 1) * 100
		trades = append(trades, Trade{
			EntryPrice: entryPrice,
			ExitPrice:  lastPrice,
			PnL:        pnl,
			PnLPct:     pnlPct,
		})
		capital = proceeds
	}

	return Result{
		FinalCapital:   capital,
		TotalReturnPct: (capital/bt.initialCapital - 1) * 100,
		MaxDrawdownPct: maxDrawdown(equityCurve),
		SharpeRatio:    sharpeRatio(equityCurve),
		TotalTrades:    len(trades),
		WinRate:        winRate(trades),
		ProfitFactor:   profitFactor(trades),
		Trades:         trades,
		EquityCurve:    equityCurve,
	}
}

func maxDrawdown(equity []float64) float64 {
	peak := equity[0]
	maxDD := 0.0
	for _, e := range equity {
		if e > peak {
			peak = e
		}
		dd := (peak - e) / peak * 100
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

func sharpeRatio(equity []float64) float64 {
	if len(equity) < 2 {
		return 0
	}
	returns := make([]float64, len(equity)-1)
	for i := 1; i < len(equity); i++ {
		if equity[i-1] != 0 {
			returns[i-1] = (equity[i] - equity[i-1]) / equity[i-1]
		}
	}
	mean := avg(returns)
	std := stdDev(returns, mean)
	if std == 0 {
		return 0
	}
	return (mean / std) * math.Sqrt(8760)
}

func winRate(trades []Trade) float64 {
	if len(trades) == 0 {
		return 0
	}
	wins := 0
	for _, t := range trades {
		if t.PnL > 0 {
			wins++
		}
	}
	return float64(wins) / float64(len(trades)) * 100
}

func profitFactor(trades []Trade) float64 {
	var gross, loss float64
	for _, t := range trades {
		if t.PnL > 0 {
			gross += t.PnL
		} else {
			loss += math.Abs(t.PnL)
		}
	}
	if loss == 0 {
		return math.Inf(1)
	}
	return gross / loss
}

func avg(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func stdDev(data []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range data {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(data)))
}
