package yahoo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/thesarfo/backtester/binance"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

var intervalMap = map[string]string{
	"1m":  "1m",
	"5m":  "5m",
	"15m": "15m",
	"30m": "30m",
	"1h":  "60m",
	"4h":  "60m",
	"1d":  "1d",
	"1wk": "1wk",
	"1mo": "1mo",
}

func (c *Client) FetchCandles(symbol, interval string, limit int) ([]binance.Candle, error) {
	yahooInterval, ok := intervalMap[interval]
	if !ok {
		return nil, fmt.Errorf("unsupported interval %q; supported: 1m 5m 15m 30m 1h 1d 1wk 1mo", interval)
	}

	rangeDuration := estimateRange(interval, limit)
	end := time.Now()
	start := end.Add(-rangeDuration)

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=%s&period1=%d&period2=%d&includePrePost=false",
		symbol, yahooInterval, start.Unix(), end.Unix(),
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; backtester/1.0)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yahoo Finance API error %d: %s", resp.StatusCode, string(body))
	}

	candles, err := parseYahooResponse(body)
	if err != nil {
		return nil, err
	}

	if len(candles) > limit {
		candles = candles[len(candles)-limit:]
	}

	return candles, nil
}

func (c *Client) FetchCandlesRange(symbol, interval string, start, end time.Time) ([]binance.Candle, error) {
	yahooInterval, ok := intervalMap[interval]
	if !ok {
		return nil, fmt.Errorf("unsupported interval %q", interval)
	}

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=%s&period1=%d&period2=%d&includePrePost=false",
		symbol, yahooInterval, start.Unix(), end.Unix(),
	)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; backtester/1.0)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return parseYahooResponse(body)
}

type yahooResponse struct {
	Chart struct {
		Result []struct {
			Timestamps []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func parseYahooResponse(body []byte) ([]binance.Candle, error) {
	var yr yahooResponse
	if err := json.Unmarshal(body, &yr); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	if yr.Chart.Error != nil {
		return nil, fmt.Errorf("Yahoo Finance error: %s, %s",
			yr.Chart.Error.Code, yr.Chart.Error.Description)
	}

	if len(yr.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned, check the symbol is valid")
	}

	result := yr.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data in response")
	}

	timestamps := result.Timestamps
	q := result.Indicators.Quote[0]
	n := len(timestamps)

	candles := make([]binance.Candle, 0, n)
	for i := 0; i < n; i++ {
		if i >= len(q.Close) || q.Close[i] == 0 {
			continue
		}
		open := safeGet(q.Open, i)
		high := safeGet(q.High, i)
		low := safeGet(q.Low, i)
		close_ := safeGet(q.Close, i)
		vol := safeGet(q.Volume, i)

		if close_ == 0 || high == 0 || low == 0 {
			continue
		}

		t := time.Unix(timestamps[i], 0)
		candles = append(candles, binance.Candle{
			OpenTime:  t,
			CloseTime: t,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close_,
			Volume:    vol,
		})
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("no valid candles after parsing, symbol may be delisted or have no data for this range")
	}

	return candles, nil
}

func safeGet(slice []float64, i int) float64 {
	if i < len(slice) {
		return slice[i]
	}
	return 0
}

func estimateRange(interval string, limit int) time.Duration {
	perCandle := map[string]time.Duration{
		"1m":  time.Minute,
		"5m":  5 * time.Minute,
		"15m": 15 * time.Minute,
		"30m": 30 * time.Minute,
		"1h":  time.Hour,
		"4h":  4 * time.Hour,
		"1d":  24 * time.Hour,
		"1wk": 7 * 24 * time.Hour,
		"1mo": 30 * 24 * time.Hour,
	}
	d, ok := perCandle[interval]
	if !ok {
		d = 24 * time.Hour
	}
	return time.Duration(float64(d) * float64(limit) * 1.4)
}
