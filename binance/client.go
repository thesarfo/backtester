package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.binance.com"

// Candle represents a single OHLCV candlestick
type Candle struct {
	OpenTime  time.Time
	CloseTime time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}


type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}


func (c *Client) FetchKlines(symbol, interval string, limit int) ([]Candle, error) {
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		baseURL, symbol, interval, limit)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Binance API error %d: %s", resp.StatusCode, string(body))
	}

	var raw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	candles := make([]Candle, 0, len(raw))
	for _, row := range raw {
		c, err := parseKline(row)
		if err != nil {
			return nil, err
		}
		candles = append(candles, c)
	}

	return candles, nil
}


func (c *Client) FetchKlinesRange(symbol, interval string, start, end time.Time) ([]Candle, error) {
	var all []Candle
	current := start

	for current.Before(end) {
		url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=1000&startTime=%d&endTime=%d",
			baseURL, symbol, interval,
			current.UnixMilli(),
			end.UnixMilli(),
		)

		resp, err := c.httpClient.Get(url)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var raw [][]interface{}
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, err
		}
		if len(raw) == 0 {
			break
		}

		for _, row := range raw {
			candle, err := parseKline(row)
			if err != nil {
				return nil, err
			}
			all = append(all, candle)
		}

		last := all[len(all)-1]
		current = last.CloseTime.Add(time.Millisecond)

		time.Sleep(100 * time.Millisecond)
	}

	return all, nil
}

func parseKline(row []interface{}) (Candle, error) {
	parse := func(v interface{}) (float64, error) {
		s, ok := v.(string)
		if !ok {
			return 0, fmt.Errorf("expected string, got %T", v)
		}
		return strconv.ParseFloat(s, 64)
	}

	openMs, _ := row[0].(float64)
	closeMs, _ := row[6].(float64)

	open, err := parse(row[1])
	if err != nil {
		return Candle{}, err
	}
	high, _ := parse(row[2])
	low, _ := parse(row[3])
	close_, _ := parse(row[4])
	volume, _ := parse(row[5])

	return Candle{
		OpenTime:  time.UnixMilli(int64(openMs)),
		CloseTime: time.UnixMilli(int64(closeMs)),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close_,
		Volume:    volume,
	}, nil
}
