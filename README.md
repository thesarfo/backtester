A backtester that real historical price data from, runs four trading strategies side by side for comparison.


### Features

- **Two markets**: crypto via Binance, stocks via Yahoo Finance
- **Visual dashboard**: equity curves, comparison table, drawdown bars, trade log, in your browser
- **Four strategies** included out of the box, EMA Crossover, RSI, Bollinger Bands, Buy & Hold
- **Five performance metrics**: Total Return, Max Drawdown, Sharpe Ratio, Win Rate, Profit Factor


### The Four Strategies

All four strategies work identically across crypto and stocks.

### EMA Crossover
Watches a fast and a slow exponential moving average. Buys when the fast crosses above the slow (upward momentum), sells when it crosses back below. Works well in trending markets, struggles in sideways chop.

### RSI (Relative Strength Index)
Buys when RSI drops below 30 (oversold, the market may have fallen too far) and sells when it rises above 70 (overbought, may be due for a pullback). A contrarian, mean-reversion approach.

### Bollinger Bands
Draws a volatility-adjusted band around the price. Buys at the lower band (statistically cheap), sells at the upper band (statistically expensive). Band width adapts automatically to market conditions.

### Buy & Hold
Buys on the first candle and never sells. Included as a baseline, every other strategy must beat it to justify the extra effort.



<!-- ## Notes

- No API keys are required for either data source, both Binance market data and Yahoo Finance are publicly accessible
- Commission is set to **0.1% per trade** (Binance spot taker fee). Adjust the `commission` field in `backtester/backtester.go` to match your actual broker
- The Sharpe Ratio is annualised assuming hourly bars (`sqrt(8760)`). If you switch to daily bars, change the multiplier to `sqrt(365)`
- Yahoo Finance hourly data is limited to the last 60 days, use daily bars for longer stock backtests
- Past performance does not predict future results. Always validate on out-of-sample data before trading live -->