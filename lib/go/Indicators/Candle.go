package Indicators

import "time"

type Candle struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

type Action string

const (
	BUY  Action = "buy"
	SELL Action = "sell"
)

type Signal struct {
	Metric  string    `json:"metric"`
	Time    time.Time `json:"time"`
	Action  Action    `json:"action"`
	Value   float64   `json:"value"`
	Expires time.Time `json:"expires,omitzero"`
}

type point struct {
	action Action
	value  float64
}

type prices struct {
	open, high, low, close, volume, median, typical []float64
}

func split(candles []Candle) prices {
	result := prices{
		open:    make([]float64, len(candles)),
		high:    make([]float64, len(candles)),
		low:     make([]float64, len(candles)),
		close:   make([]float64, len(candles)),
		volume:  make([]float64, len(candles)),
		median:  make([]float64, len(candles)),
		typical: make([]float64, len(candles)),
	}

	for i, candle := range candles {
		result.open[i] = candle.Open
		result.high[i] = candle.High
		result.low[i] = candle.Low
		result.close[i] = candle.Close
		result.volume[i] = candle.Volume
		result.median[i] = (candle.High + candle.Low) / 2
		result.typical[i] = (candle.High + candle.Low + candle.Close) / 3
	}

	return result
}
