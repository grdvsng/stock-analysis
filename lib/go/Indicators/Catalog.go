package Indicators

import (
	"fmt"
	"time"
)

const (
	TREND      = "trend"
	MOMENTUM   = "momentum"
	VOLATILITY = "volatility"
	VOLUME     = "volume"
	PATTERN    = "pattern"
	FIBONACCI  = "fibonacci"
)

type evaluator func(candles []Candle, p prices) []point

type Definition struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Warmup      int    `json:"warmup"`
	evaluate    evaluator
}

func (definition Definition) Evaluate(candles []Candle, from time.Time) []Signal {
	if len(candles) == 0 {
		return nil
	}

	points := definition.evaluate(candles, split(candles))
	signals := []Signal{}

	for i, current := range points {
		if current.action == "" || candles[i].Time.Before(from) || !defined(current.value) {
			continue
		}

		signals = append(signals, Signal{Metric: definition.Key, Time: candles[i].Time, Action: current.action, Value: current.value})
	}

	return signals
}

var (
	catalog = build()
	byKey   = index(catalog)
)

func Catalog() []Definition {
	return catalog
}

func Lookup(key string) (Definition, bool) {
	definition, exists := byKey[key]

	return definition, exists
}

func index(definitions []Definition) map[string]Definition {
	result := make(map[string]Definition, len(definitions))

	for _, definition := range definitions {
		if _, exists := result[definition.Key]; exists {
			panic("duplicate indicator " + definition.Key)
		}

		result[definition.Key] = definition
	}

	return result
}

func priceCross(key, name string, period int, average func([]float64, int) []float64, warmup int) Definition {
	return Definition{
		Key:         key,
		Group:       TREND,
		Name:        name,
		Description: fmt.Sprintf("Buy when the close crosses above the %s, sell when it crosses below.", name),
		Warmup:      warmup,
		evaluate: func(_ []Candle, p prices) []point {
			return cross(p.close, average(p.close, period))
		},
	}
}

func averageCross(key, name string, fast, slow int, average func([]float64, int) []float64, warmup int) Definition {
	return Definition{
		Key:         key,
		Group:       TREND,
		Name:        name,
		Description: fmt.Sprintf("Buy when the %d-period average crosses above the %d-period one (golden cross), sell on the opposite (death cross).", fast, slow),
		Warmup:      warmup,
		evaluate: func(_ []Candle, p prices) []point {
			return cross(average(p.close, fast), average(p.close, slow))
		},
	}
}

func rsiLevel(key, name string, period int, low, high float64) Definition {
	return Definition{
		Key:         key,
		Group:       MOMENTUM,
		Name:        name,
		Description: fmt.Sprintf("Buy when RSI(%d) rises back above %.0f, sell when it falls back below %.0f.", period, low, high),
		Warmup:      period*3 + 1,
		evaluate: func(_ []Candle, p prices) []point {
			return levels(rsi(p.close, period), low, high)
		},
	}
}

func candlePattern(key, name, description string, size int, bullish, bearish pattern) Definition {
	return Definition{
		Key:         key,
		Group:       PATTERN,
		Name:        name,
		Description: description,
		Warmup:      size + TREND_PERIOD + 11,
		evaluate: func(candles []Candle, _ prices) []point {
			return detect(candles, size+TREND_PERIOD, bullish, bearish)
		},
	}
}

func fibonacci(key, name, description string, evaluate func([]Candle) []point) Definition {
	return Definition{
		Key:         key,
		Group:       FIBONACCI,
		Name:        name,
		Description: description,
		Warmup:      FIBONACCI_LOOKBACK*2 + 1,
		evaluate: func(candles []Candle, _ prices) []point {
			return evaluate(candles)
		},
	}
}

func build() []Definition {
	definitions := []Definition{}

	for _, period := range []int{5, 10, 20, 50, 100, 200} {
		definitions = append(definitions, priceCross(fmt.Sprintf("sma_%d_price", period), fmt.Sprintf("SMA %d", period), period, sma, period+1))
	}

	for _, period := range []int{5, 10, 20, 50, 100, 200} {
		definitions = append(definitions, priceCross(fmt.Sprintf("ema_%d_price", period), fmt.Sprintf("EMA %d", period), period, ema, period*2+1))
	}

	for _, pair := range [][2]int{{5, 20}, {10, 50}, {20, 50}, {50, 200}} {
		definitions = append(definitions, averageCross(fmt.Sprintf("sma_%d_%d_cross", pair[0], pair[1]), fmt.Sprintf("SMA %d/%d cross", pair[0], pair[1]), pair[0], pair[1], sma, pair[1]+1))
	}

	for _, pair := range [][2]int{{9, 21}, {12, 26}, {20, 50}, {50, 200}} {
		definitions = append(definitions, averageCross(fmt.Sprintf("ema_%d_%d_cross", pair[0], pair[1]), fmt.Sprintf("EMA %d/%d cross", pair[0], pair[1]), pair[0], pair[1], ema, pair[1]*2+1))
	}

	definitions = append(definitions,
		priceCross("wma_20_price", "WMA 20", 20, wma, 21),
		priceCross("dema_20_price", "DEMA 20", 20, dema, 61),
		priceCross("tema_20_price", "TEMA 20", 20, tema, 81),
		priceCross("hma_20_price", "HMA 20", 20, hma, 31),
		averageCross("hma_9_21_cross", "HMA 9/21 cross", 9, 21, hma, 31),
		Definition{Key: "macd_signal_cross", Group: TREND, Name: "MACD signal cross", Warmup: 70,
			Description: "Buy when MACD(12,26) crosses above its 9-period signal line, sell when it crosses below.",
			evaluate: func(_ []Candle, p prices) []point {
				line, signal, _ := macd(p.close, 12, 26, 9)

				return cross(line, signal)
			}},
		Definition{Key: "macd_zero_cross", Group: TREND, Name: "MACD zero cross", Warmup: 60,
			Description: "Buy when MACD(12,26) crosses above zero, sell when it crosses below.",
			evaluate: func(_ []Candle, p prices) []point {
				line, _, _ := macd(p.close, 12, 26, 9)

				return zero(line)
			}},
		Definition{Key: "macd_histogram_turn", Group: TREND, Name: "MACD histogram turn", Warmup: 70,
			Description: "Buy when a negative MACD histogram turns up, sell when a positive one turns down.",
			evaluate: func(_ []Candle, p prices) []point {
				_, _, histogram := macd(p.close, 12, 26, 9)

				return turning(histogram)
			}},
		Definition{Key: "macd_5_35_5_cross", Group: TREND, Name: "MACD 5/35/5 signal cross", Warmup: 80,
			Description: "Fast MACD(5,35) crossing its 5-period signal line.",
			evaluate: func(_ []Candle, p prices) []point {
				line, signal, _ := macd(p.close, 5, 35, 5)

				return cross(line, signal)
			}},
		Definition{Key: "ppo_signal_cross", Group: TREND, Name: "PPO signal cross", Warmup: 70,
			Description: "Percentage price oscillator (12,26) crossing its 9-period signal line.",
			evaluate: func(_ []Candle, p prices) []point {
				line, signal := ppo(p.close, 12, 26, 9)

				return cross(line, signal)
			}},
		Definition{Key: "adx_di_cross", Group: TREND, Name: "ADX trend DI cross", Warmup: 45,
			Description: "+DI/-DI cross confirmed by ADX(14) of at least 20.",
			evaluate: func(_ []Candle, p prices) []point {
				strength, plus, minus := directional(p.high, p.low, p.close, 14)

				return gated(cross(plus, minus), strength, 20)
			}},
		Definition{Key: "di_cross", Group: TREND, Name: "DI cross", Warmup: 30,
			Description: "Buy when +DI(14) crosses above -DI, sell when it crosses below.",
			evaluate: func(_ []Candle, p prices) []point {
				_, plus, minus := directional(p.high, p.low, p.close, 14)

				return cross(plus, minus)
			}},
		Definition{Key: "aroon_25_cross", Group: TREND, Name: "Aroon 25 cross", Warmup: 27,
			Description: "Buy when Aroon up(25) crosses above Aroon down, sell when it crosses below.",
			evaluate: func(_ []Candle, p prices) []point {
				up, down := aroon(p.high, p.low, 25)

				return cross(up, down)
			}},
		Definition{Key: "aroon_14_cross", Group: TREND, Name: "Aroon 14 cross", Warmup: 16,
			Description: "Buy when Aroon up(14) crosses above Aroon down, sell when it crosses below.",
			evaluate: func(_ []Candle, p prices) []point {
				up, down := aroon(p.high, p.low, 14)

				return cross(up, down)
			}},
		Definition{Key: "psar_flip", Group: TREND, Name: "Parabolic SAR flip", Warmup: 10,
			Description: "Parabolic SAR (0.02, 0.2) switching sides of the price.",
			evaluate: func(_ []Candle, p prices) []point {
				return psarFlip(p.high, p.low)
			}},
		Definition{Key: "ichimoku_tk_cross", Group: TREND, Name: "Ichimoku TK cross", Warmup: 30,
			Description: "Tenkan-sen (9) crossing Kijun-sen (26).",
			evaluate: func(_ []Candle, p prices) []point {
				tenkan, kijun, _, _ := ichimoku(p.high, p.low)

				return cross(tenkan, kijun)
			}},
		Definition{Key: "ichimoku_cloud_breakout", Group: TREND, Name: "Ichimoku cloud breakout", Warmup: 80,
			Description: "Close breaking above or below the Ichimoku cloud.",
			evaluate: func(_ []Candle, p prices) []point {
				return cloud(p.close, p.high, p.low)
			}},
		Definition{Key: "trix_signal_cross", Group: TREND, Name: "TRIX signal cross", Warmup: 60,
			Description: "TRIX(15) crossing its 9-period signal line.",
			evaluate: func(_ []Candle, p prices) []point {
				line := trix(p.close, 15)

				return cross(line, ema(line, 9))
			}},
		Definition{Key: "trix_zero_cross", Group: TREND, Name: "TRIX zero cross", Warmup: 50,
			Description: "TRIX(15) crossing zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return zero(trix(p.close, 15))
			}},
		rsiLevel("rsi_7", "RSI 7", 7, 30, 70),
		rsiLevel("rsi_14", "RSI 14", 14, 30, 70),
		rsiLevel("rsi_21", "RSI 21", 21, 30, 70),
		rsiLevel("rsi_14_extreme", "RSI 14 extreme", 14, 20, 80),
		Definition{Key: "rsi_14_midline", Group: MOMENTUM, Name: "RSI 14 midline", Warmup: 43,
			Description: "RSI(14) crossing 50.",
			evaluate: func(_ []Candle, p prices) []point {
				values := rsi(p.close, 14)

				return cross(values, constant(len(values), 50))
			}},
		Definition{Key: "stoch_rsi_14", Group: MOMENTUM, Name: "Stochastic RSI 14", Warmup: 50,
			Description: "Stochastic RSI(14,3,3) %K leaving the 20/80 zones.",
			evaluate: func(_ []Candle, p prices) []point {
				k, _ := stochasticRsi(p.close, 14, 3, 3)

				return levels(k, 20, 80)
			}},
		Definition{Key: "stoch_14_3", Group: MOMENTUM, Name: "Stochastic 14,3", Warmup: 18,
			Description: "Stochastic %K(14) leaving the 20/80 zones.",
			evaluate: func(_ []Candle, p prices) []point {
				k, _ := stochastic(p.high, p.low, p.close, 14, 3)

				return levels(k, 20, 80)
			}},
		Definition{Key: "stoch_5_3", Group: MOMENTUM, Name: "Stochastic 5,3", Warmup: 9,
			Description: "Fast stochastic %K(5) leaving the 20/80 zones.",
			evaluate: func(_ []Candle, p prices) []point {
				k, _ := stochastic(p.high, p.low, p.close, 5, 3)

				return levels(k, 20, 80)
			}},
		Definition{Key: "stoch_14_3_cross", Group: MOMENTUM, Name: "Stochastic %K/%D cross", Warmup: 18,
			Description: "Stochastic %K(14) crossing %D(3).",
			evaluate: func(_ []Candle, p prices) []point {
				k, d := stochastic(p.high, p.low, p.close, 14, 3)

				return cross(k, d)
			}},
		Definition{Key: "cci_14", Group: MOMENTUM, Name: "CCI 14", Warmup: 16,
			Description: "CCI(14) coming back from beyond ±100.",
			evaluate: func(_ []Candle, p prices) []point {
				return levels(cci(p.typical, 14), -100, 100)
			}},
		Definition{Key: "cci_20", Group: MOMENTUM, Name: "CCI 20", Warmup: 22,
			Description: "CCI(20) coming back from beyond ±100.",
			evaluate: func(_ []Candle, p prices) []point {
				return levels(cci(p.typical, 20), -100, 100)
			}},
		Definition{Key: "williams_r_14", Group: MOMENTUM, Name: "Williams %R 14", Warmup: 16,
			Description: "Williams %R(14) leaving the -80/-20 zones.",
			evaluate: func(_ []Candle, p prices) []point {
				return levels(williams(p.high, p.low, p.close, 14), -80, -20)
			}},
		Definition{Key: "roc_9_zero", Group: MOMENTUM, Name: "ROC 9 zero cross", Warmup: 11,
			Description: "Rate of change (9) crossing zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return zero(roc(p.close, 9))
			}},
		Definition{Key: "roc_12_zero", Group: MOMENTUM, Name: "ROC 12 zero cross", Warmup: 14,
			Description: "Rate of change (12) crossing zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return zero(roc(p.close, 12))
			}},
		Definition{Key: "momentum_10_zero", Group: MOMENTUM, Name: "Momentum 10 zero cross", Warmup: 12,
			Description: "Momentum (10) crossing zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return zero(change(p.close, 10))
			}},
		Definition{Key: "cmo_14", Group: MOMENTUM, Name: "CMO 14", Warmup: 16,
			Description: "Chande momentum oscillator (14) coming back from beyond ±50.",
			evaluate: func(_ []Candle, p prices) []point {
				return levels(cmo(p.close, 14), -50, 50)
			}},
		Definition{Key: "ultimate_oscillator", Group: MOMENTUM, Name: "Ultimate oscillator", Warmup: 30,
			Description: "Ultimate oscillator (7,14,28) leaving the 30/70 zones.",
			evaluate: func(_ []Candle, p prices) []point {
				return levels(ultimate(p.high, p.low, p.close), 30, 70)
			}},
		Definition{Key: "awesome_zero", Group: MOMENTUM, Name: "Awesome oscillator zero cross", Warmup: 36,
			Description: "Awesome oscillator (5,34) crossing zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return zero(awesome(p.median))
			}},
		Definition{Key: "awesome_saucer", Group: MOMENTUM, Name: "Awesome oscillator saucer", Warmup: 38,
			Description: "Awesome oscillator saucer: a dip and turn on the same side of zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return saucer(awesome(p.median))
			}},
		Definition{Key: "elder_ray_13", Group: MOMENTUM, Name: "Elder ray 13", Warmup: 30,
			Description: "Bear power rising under a rising EMA(13) (buy) or bull power falling under a falling one (sell).",
			evaluate: func(_ []Candle, p prices) []point {
				return elder(p.high, p.low, p.close, 13)
			}},
		Definition{Key: "bollinger_20_reversal", Group: VOLATILITY, Name: "Bollinger 20,2 reversal", Warmup: 22,
			Description: "Close returning inside the Bollinger bands (20, 2σ).",
			evaluate: func(_ []Candle, p prices) []point {
				_, upper, lower := bollinger(p.close, 20, 2)

				return reversal(p.close, upper, lower)
			}},
		Definition{Key: "bollinger_20_breakout", Group: VOLATILITY, Name: "Bollinger 20,2 breakout", Warmup: 22,
			Description: "Close breaking out of the Bollinger bands (20, 2σ).",
			evaluate: func(_ []Candle, p prices) []point {
				_, upper, lower := bollinger(p.close, 20, 2)

				return breakout(p.close, upper, lower)
			}},
		Definition{Key: "bollinger_10_reversal", Group: VOLATILITY, Name: "Bollinger 10,1.5 reversal", Warmup: 12,
			Description: "Close returning inside the fast Bollinger bands (10, 1.5σ).",
			evaluate: func(_ []Candle, p prices) []point {
				_, upper, lower := bollinger(p.close, 10, 1.5)

				return reversal(p.close, upper, lower)
			}},
		Definition{Key: "bollinger_squeeze", Group: VOLATILITY, Name: "Bollinger squeeze breakout", Warmup: 142,
			Description: "Breakout of the Bollinger bands right after their narrowest width in 120 candles.",
			evaluate: func(_ []Candle, p prices) []point {
				return squeeze(p.close)
			}},
		Definition{Key: "keltner_20_breakout", Group: VOLATILITY, Name: "Keltner 20,2 breakout", Warmup: 42,
			Description: "Close breaking out of the Keltner channel (EMA 20 ± 2 ATR).",
			evaluate: func(_ []Candle, p prices) []point {
				upper, lower := keltner(p.high, p.low, p.close, 20, 2)

				return breakout(p.close, upper, lower)
			}},
		Definition{Key: "keltner_20_reversal", Group: VOLATILITY, Name: "Keltner 20,2 reversal", Warmup: 42,
			Description: "Close returning inside the Keltner channel (EMA 20 ± 2 ATR).",
			evaluate: func(_ []Candle, p prices) []point {
				upper, lower := keltner(p.high, p.low, p.close, 20, 2)

				return reversal(p.close, upper, lower)
			}},
		Definition{Key: "donchian_20_breakout", Group: VOLATILITY, Name: "Donchian 20 breakout", Warmup: 22,
			Description: "Close above the previous 20-candle high (buy) or below the 20-candle low (sell).",
			evaluate: func(_ []Candle, p prices) []point {
				return breakout(p.close, shift(highest(p.high, 20), 1), shift(lowest(p.low, 20), 1))
			}},
		Definition{Key: "donchian_55_breakout", Group: VOLATILITY, Name: "Donchian 55 breakout", Warmup: 57,
			Description: "Close above the previous 55-candle high (buy) or below the 55-candle low (sell).",
			evaluate: func(_ []Candle, p prices) []point {
				return breakout(p.close, shift(highest(p.high, 55), 1), shift(lowest(p.low, 55), 1))
			}},
		Definition{Key: "atr_14_breakout", Group: VOLATILITY, Name: "ATR 14 breakout", Warmup: 16,
			Description: "Close moving more than 2 ATR(14) from the previous close.",
			evaluate: func(_ []Candle, p prices) []point {
				return atrBreakout(p.high, p.low, p.close, 14, 2)
			}},
		Definition{Key: "mfi_14", Group: VOLUME, Name: "MFI 14", Warmup: 16,
			Description: "Money flow index (14) leaving the 20/80 zones.",
			evaluate: func(_ []Candle, p prices) []point {
				return levels(mfi(p.typical, p.volume, 14), 20, 80)
			}},
		Definition{Key: "obv_sma_20_cross", Group: VOLUME, Name: "OBV / SMA 20 cross", Warmup: 22,
			Description: "On-balance volume crossing its 20-period average.",
			evaluate: func(_ []Candle, p prices) []point {
				values := obv(p.close, p.volume)

				return cross(values, sma(values, 20))
			}},
		Definition{Key: "cmf_20_zero", Group: VOLUME, Name: "CMF 20 zero cross", Warmup: 22,
			Description: "Chaikin money flow (20) crossing zero.",
			evaluate: func(_ []Candle, p prices) []point {
				return zero(cmf(p.high, p.low, p.close, p.volume, 20))
			}},
		Definition{Key: "vwap_20_cross", Group: VOLUME, Name: "VWAP 20 cross", Warmup: 22,
			Description: "Close crossing the rolling 20-candle volume weighted average price.",
			evaluate: func(_ []Candle, p prices) []point {
				return cross(p.close, vwap(p.typical, p.volume, 20))
			}},
		candlePattern("bullish_engulfing", "Bullish engulfing", "A bullish body engulfs the previous bearish one after a decline.", 2, engulfingUp, nil),
		candlePattern("bearish_engulfing", "Bearish engulfing", "A bearish body engulfs the previous bullish one after a rise.", 2, nil, engulfingDown),
		candlePattern("hammer", "Hammer", "Small body with a long lower shadow after a decline.", 1, hammer, nil),
		candlePattern("hanging_man", "Hanging man", "Small body with a long lower shadow after a rise.", 1, nil, hangingMan),
		candlePattern("inverted_hammer", "Inverted hammer", "Small body with a long upper shadow after a decline.", 1, invertedHammer, nil),
		candlePattern("shooting_star", "Shooting star", "Small body with a long upper shadow after a rise.", 1, nil, shootingStar),
		candlePattern("dragonfly_doji", "Dragonfly doji", "Doji with a long lower shadow after a decline.", 1, dragonfly, nil),
		candlePattern("gravestone_doji", "Gravestone doji", "Doji with a long upper shadow after a rise.", 1, nil, gravestone),
		candlePattern("morning_star", "Morning star", "Long bearish candle, small star, then a bullish close above the first midpoint.", 3, morningStar, nil),
		candlePattern("evening_star", "Evening star", "Long bullish candle, small star, then a bearish close below the first midpoint.", 3, nil, eveningStar),
		candlePattern("three_white_soldiers", "Three white soldiers", "Three long bullish candles closing progressively higher.", 3, soldiers, nil),
		candlePattern("three_black_crows", "Three black crows", "Three long bearish candles closing progressively lower.", 3, nil, crows),
		candlePattern("bullish_harami", "Bullish harami", "Small bullish body inside the previous long bearish one after a decline.", 2, haramiUp, nil),
		candlePattern("bearish_harami", "Bearish harami", "Small bearish body inside the previous long bullish one after a rise.", 2, nil, haramiDown),
		candlePattern("piercing_line", "Piercing line", "Opens below the previous low and closes above the middle of the previous bearish body.", 2, piercing, nil),
		candlePattern("dark_cloud_cover", "Dark cloud cover", "Opens above the previous high and closes below the middle of the previous bullish body.", 2, nil, darkCloud),
		candlePattern("tweezer_bottom", "Tweezer bottom", "Two candles with matching lows reversing a decline.", 2, tweezerBottom, nil),
		candlePattern("tweezer_top", "Tweezer top", "Two candles with matching highs reversing a rise.", 2, nil, tweezerTop),
		candlePattern("bullish_marubozu", "Bullish marubozu", "Long bullish candle almost without shadows.", 1, marubozuUp, nil),
		candlePattern("bearish_marubozu", "Bearish marubozu", "Long bearish candle almost without shadows.", 1, nil, marubozuDown),
		candlePattern("rising_three_methods", "Rising three methods", "Long bullish candle, three small ones inside it, then a higher bullish close.", 5, risingThree, nil),
		candlePattern("falling_three_methods", "Falling three methods", "Long bearish candle, three small ones inside it, then a lower bearish close.", 5, nil, fallingThree),
		candlePattern("bullish_kicker", "Bullish kicker", "Long bullish candle opening above the previous bearish open.", 2, kickerUp, nil),
		candlePattern("bearish_kicker", "Bearish kicker", "Long bearish candle opening below the previous bullish open.", 2, nil, kickerDown),
		candlePattern("three_inside_up", "Three inside up", "Bullish harami confirmed by a close above the first candle open.", 3, insideUp, nil),
		candlePattern("three_inside_down", "Three inside down", "Bearish harami confirmed by a close below the first candle open.", 3, nil, insideDown),
		candlePattern("three_outside_up", "Three outside up", "Bullish engulfing confirmed by a higher close.", 3, outsideUp, nil),
		candlePattern("three_outside_down", "Three outside down", "Bearish engulfing confirmed by a lower close.", 3, nil, outsideDown),
		fibonacci("fibonacci_236", "Fibonacci 23.6% bounce", "Price bouncing off the 23.6% retracement of the last 55-candle swing in its direction.", func(c []Candle) []point { return fibonacciLevel(c, 0.236) }),
		fibonacci("fibonacci_382", "Fibonacci 38.2% bounce", "Price bouncing off the 38.2% retracement of the last 55-candle swing in its direction.", func(c []Candle) []point { return fibonacciLevel(c, 0.382) }),
		fibonacci("fibonacci_500", "Fibonacci 50% bounce", "Price bouncing off the 50% retracement of the last 55-candle swing in its direction.", func(c []Candle) []point { return fibonacciLevel(c, 0.5) }),
		fibonacci("fibonacci_618", "Fibonacci 61.8% bounce", "Price bouncing off the 61.8% retracement of the last 55-candle swing in its direction.", func(c []Candle) []point { return fibonacciLevel(c, 0.618) }),
		fibonacci("fibonacci_786", "Fibonacci 78.6% bounce", "Price bouncing off the 78.6% retracement of the last 55-candle swing in its direction.", func(c []Candle) []point { return fibonacciLevel(c, 0.786) }),
		fibonacci("fibonacci_golden_pocket", "Fibonacci golden pocket", "Price bouncing out of the 61.8–65% retracement zone of the last 55-candle swing.", func(c []Candle) []point { return fibonacciZone(c, 0.618, 0.65) }),
		fibonacci("fibonacci_extension_1618", "Fibonacci 161.8% extension", "Close reaching the 161.8% extension target of the previous 55-candle swing: take profit (sell) after a rise, cover (buy) after a fall.", func(c []Candle) []point { return fibonacciExtension(c, 1.618) }),
	)

	return definitions
}
