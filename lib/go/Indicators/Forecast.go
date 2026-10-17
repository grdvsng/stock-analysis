package Indicators

import (
	"errors"
	"math"
	"sort"
	"time"
)

const (
	FORECAST_MIN_CANDLES   = 60
	FORECAST_SWING_WINDOW  = 120
	FORECAST_PIVOT_SPAN    = 3
	FORECAST_MAX_LEVELS    = 3
	FORECAST_ZONE_ATR      = 0.25
	FORECAST_STOP_ATR      = 0.5
	FORECAST_SCORE_TRIGGER = 1.5
	HOLD                   = Action("hold")
)

var ErrNotEnoughCandles = errors.New("not enough candles for a forecast")

type Reason struct {
	Code   string  `json:"code"`
	Vote   float64 `json:"vote"`
	Value  float64 `json:"value"`
	Second float64 `json:"second,omitempty"`
}

type Trade struct {
	Price   float64   `json:"price"`
	Stop    float64   `json:"stop"`
	Targets []float64 `json:"targets"`
	Reward  float64   `json:"reward"`
}

type Forecast struct {
	AsOf        time.Time          `json:"as_of"`
	Price       float64            `json:"price"`
	Trend       string             `json:"trend"`
	Strength    string             `json:"strength"`
	Action      Action             `json:"action"`
	Score       float64            `json:"score"`
	Buy         Trade              `json:"buy"`
	Sell        Trade              `json:"sell"`
	HoldFrom    float64            `json:"hold_from"`
	HoldTo      float64            `json:"hold_to"`
	Supports    []float64          `json:"supports"`
	Resistances []float64          `json:"resistances"`
	Fibonacci   map[string]float64 `json:"fibonacci"`
	Indicators  map[string]float64 `json:"indicators"`
	Reasons     []Reason           `json:"reasons"`
}

func last(values []float64) float64 {
	for i := len(values) - 1; i >= 0; i-- {
		if defined(values[i]) {
			return values[i]
		}
	}

	return math.NaN()
}

func at(values []float64, back int) float64 {
	if len(values)-1-back < 0 {
		return math.NaN()
	}

	return values[len(values)-1-back]
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func pivots(high, low []float64, from int) ([]float64, []float64) {
	highs, lows := []float64{}, []float64{}

	for j := max(from, FORECAST_PIVOT_SPAN); j < len(high)-FORECAST_PIVOT_SPAN; j++ {
		top, bottom := true, true

		for k := j - FORECAST_PIVOT_SPAN; k <= j+FORECAST_PIVOT_SPAN; k++ {
			if high[k] > high[j] {
				top = false
			}

			if low[k] < low[j] {
				bottom = false
			}
		}

		if top {
			highs = append(highs, high[j])
		}

		if bottom {
			lows = append(lows, low[j])
		}
	}

	return highs, lows
}

func cluster(levels []float64, gap float64, ascending bool) []float64 {
	sort.Float64s(levels)

	if !ascending {
		sort.Sort(sort.Reverse(sort.Float64Slice(levels)))
	}

	result := []float64{}

	for _, level := range levels {
		if len(result) > 0 && math.Abs(result[len(result)-1]-level) < gap {
			continue
		}

		result = append(result, round(level))

		if len(result) == FORECAST_MAX_LEVELS {
			break
		}
	}

	return result
}

func Predict(candles []Candle) (Forecast, error) {
	if len(candles) < FORECAST_MIN_CANDLES {
		return Forecast{}, ErrNotEnoughCandles
	}

	p := split(candles)
	n := len(candles)
	price := p.close[n-1]
	fast, slow := ema(p.close, 20), ema(p.close, 50)
	strength, plus, minus := directional(p.high, p.low, p.close, 14)
	_, _, histogram := macd(p.close, 12, 26, 9)
	k, _ := stochastic(p.high, p.low, p.close, 14, 3)
	_, upper, lower := bollinger(p.close, 20, 2)
	ranges := atr(p.high, p.low, p.close, 14)
	strengthNow, rsiNow, atrNow := last(strength), last(rsi(p.close, 14)), last(ranges)
	fastNow, slowNow, vwapNow := last(fast), last(slow), last(vwap(p.typical, p.volume, 20))
	slope := (fastNow - at(fast, 5)) / at(fast, 5) * 100

	if !defined(atrNow) || atrNow <= 0 {
		atrNow = price * 0.005
	}

	forecast := Forecast{
		AsOf:  candles[n-1].Time,
		Price: round(price),
		Indicators: map[string]float64{
			"ema_20":     round(fastNow),
			"ema_50":     round(slowNow),
			"rsi_14":     round(rsiNow),
			"macd_hist":  round(last(histogram)),
			"stoch_k":    round(last(k)),
			"adx_14":     round(strengthNow),
			"plus_di":    round(last(plus)),
			"minus_di":   round(last(minus)),
			"atr_14":     round(atrNow),
			"bb_upper":   round(last(upper)),
			"bb_lower":   round(last(lower)),
			"vwap_20":    round(vwapNow),
			"ema_20_pct": round(slope),
		},
		Fibonacci: map[string]float64{},
	}

	forecast.Trend = "sideways"

	if fastNow > slowNow && slope > 0 && last(plus) > last(minus) {
		forecast.Trend = "up"
	} else if fastNow < slowNow && slope < 0 && last(plus) < last(minus) {
		forecast.Trend = "down"
	}

	switch {
	case strengthNow >= 40:
		forecast.Strength = "very_strong"
	case strengthNow >= 25:
		forecast.Strength = "strong"
	case strengthNow >= 20:
		forecast.Strength = "moderate"
	default:
		forecast.Strength = "weak"
	}

	highs, lows := pivots(p.high, p.low, max(0, n-FORECAST_SWING_WINDOW))
	below, above := []float64{}, []float64{}

	for _, level := range lows {
		if level < price {
			below = append(below, level)
		}
	}

	for _, level := range highs {
		if level > price {
			above = append(above, level)
		}
	}

	if len(below) == 0 {
		below = append(below, last(lowest(p.low, min(50, n))))
	}

	if len(above) == 0 {
		above = append(above, last(highest(p.high, min(50, n))))
	}

	forecast.Supports = cluster(below, atrNow*0.3, false)
	forecast.Resistances = cluster(above, atrNow*0.3, true)

	if s, ok := swingAt(p.high, p.low, n); ok {
		for name, ratio := range map[string]float64{"38.2": 0.382, "50": 0.5, "61.8": 0.618} {
			forecast.Fibonacci[name] = round(s.retracement(ratio))
		}
	}

	support, resistance := forecast.Supports[0], forecast.Resistances[0]
	buy, sell := support, resistance

	if forecast.Trend == "up" && fastNow < price && fastNow > support {
		buy = fastNow
	}

	if forecast.Trend == "down" && fastNow > price && fastNow < resistance {
		sell = fastNow
	}

	buy, sell = math.Min(buy, price), math.Max(sell, price)
	buyStop := support - FORECAST_STOP_ATR*atrNow
	sellStop := resistance + FORECAST_STOP_ATR*atrNow
	forecast.Buy = trade(buy, math.Min(buyStop, buy-atrNow), forecast.Resistances, 1)
	forecast.Sell = trade(sell, math.Max(sellStop, sell+atrNow), forecast.Supports, -1)
	forecast.HoldFrom, forecast.HoldTo = round(buy), round(sell)

	vote := func(code string, value, second, amount float64) {
		forecast.Reasons = append(forecast.Reasons, Reason{Code: code, Vote: amount, Value: round(value), Second: round(second)})
		forecast.Score += amount
	}

	trendWeight := 1.0

	if strengthNow >= 25 {
		trendWeight = 2
	}

	switch forecast.Trend {
	case "up":
		vote("trend_up", strengthNow, slope, trendWeight)
	case "down":
		vote("trend_down", strengthNow, slope, -trendWeight)
	default:
		vote("trend_sideways", strengthNow, slope, 0)
	}

	switch {
	case rsiNow < 30:
		vote("rsi_oversold", rsiNow, 0, 1)
	case rsiNow > 70:
		vote("rsi_overbought", rsiNow, 0, -1)
	case rsiNow >= 50:
		vote("rsi_bullish", rsiNow, 0, 0.5)
	default:
		vote("rsi_bearish", rsiNow, 0, -0.5)
	}

	if h, previous := last(histogram), at(histogram, 1); defined(h, previous) {
		switch {
		case h > 0 && h >= previous:
			vote("macd_rising", h, 0, 0.75)
		case h < 0 && h <= previous:
			vote("macd_falling", h, 0, -0.75)
		case h > 0:
			vote("macd_fading_up", h, 0, 0.25)
		default:
			vote("macd_fading_down", h, 0, -0.25)
		}
	}

	if stoch := last(k); defined(stoch) {
		if stoch < 20 {
			vote("stoch_oversold", stoch, 0, 0.5)
		} else if stoch > 80 {
			vote("stoch_overbought", stoch, 0, -0.5)
		}
	}

	if top, bottom := last(upper), last(lower); defined(top, bottom) {
		if price <= bottom {
			vote("below_bollinger", price, bottom, 0.5)
		} else if price >= top {
			vote("above_bollinger", price, top, -0.5)
		}
	}

	if defined(vwapNow) {
		if price > vwapNow {
			vote("above_vwap", price, vwapNow, 0.5)
		} else {
			vote("below_vwap", price, vwapNow, -0.5)
		}
	}

	forecast.Score = round(forecast.Score)
	forecast.Action = HOLD

	switch {
	case price <= buy+FORECAST_ZONE_ATR*atrNow && forecast.Score > -FORECAST_SCORE_TRIGGER:
		forecast.Action = BUY
		vote("near_support", price, buy, 0)
	case price >= sell-FORECAST_ZONE_ATR*atrNow && forecast.Score < FORECAST_SCORE_TRIGGER:
		forecast.Action = SELL
		vote("near_resistance", price, sell, 0)
	case forecast.Score >= FORECAST_SCORE_TRIGGER:
		forecast.Action = BUY
	case forecast.Score <= -FORECAST_SCORE_TRIGGER:
		forecast.Action = SELL
	}

	return forecast, nil
}

func trade(entry, stop float64, levels []float64, direction float64) Trade {
	risk := math.Abs(entry - stop)
	targets := []float64{}

	for _, level := range levels {
		if (level-entry)*direction > risk*0.5 {
			targets = append(targets, round(level))
		}
	}

	for _, multiple := range []float64{1.5, 2.5} {
		if len(targets) >= 2 {
			break
		}

		candidate := round(entry + direction*risk*multiple)

		if len(targets) == 0 || (candidate-targets[len(targets)-1])*direction > 0 {
			targets = append(targets, candidate)
		}
	}

	reward := 0.0

	if risk > 0 {
		reward = round(math.Abs(targets[0]-entry) / risk)
	}

	return Trade{Price: round(entry), Stop: round(stop), Targets: targets, Reward: reward}
}
