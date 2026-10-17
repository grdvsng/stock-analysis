package Indicators

const FIBONACCI_LOOKBACK = 55

type swing struct {
	high, low float64
	rising    bool
}

func swingAt(high, low []float64, i int) (swing, bool) {
	start := i - FIBONACCI_LOOKBACK

	if start < 0 {
		return swing{}, false
	}

	top, bottom := start, start

	for j := start; j < i; j++ {
		if high[j] >= high[top] {
			top = j
		}

		if low[j] <= low[bottom] {
			bottom = j
		}
	}

	if high[top] == low[bottom] {
		return swing{}, false
	}

	return swing{high: high[top], low: low[bottom], rising: bottom < top}, true
}

func (s swing) retracement(ratio float64) float64 {
	if s.rising {
		return s.high - ratio*(s.high-s.low)
	}

	return s.low + ratio*(s.high-s.low)
}

func (s swing) extension(ratio float64) float64 {
	if s.rising {
		return s.low + ratio*(s.high-s.low)
	}

	return s.high - ratio*(s.high-s.low)
}

func fibonacciZone(candles []Candle, from, to float64) []point {
	p := split(candles)
	points := make([]point, len(candles))

	for i := range candles {
		s, ok := swingAt(p.high, p.low, i)

		if !ok {
			continue
		}

		near, far := s.retracement(from), s.retracement(to)

		if s.rising && p.low[i] <= near && p.low[i] >= far*0.995 && p.close[i] > near && p.close[i] > p.open[i] {
			points[i] = point{BUY, near}
		} else if !s.rising && p.high[i] >= near && p.high[i] <= far*1.005 && p.close[i] < near && p.close[i] < p.open[i] {
			points[i] = point{SELL, near}
		}
	}

	return points
}

func fibonacciLevel(candles []Candle, ratio float64) []point {
	return fibonacciZone(candles, ratio, ratio+0.05)
}

func fibonacciExtension(candles []Candle, ratio float64) []point {
	p := split(candles)
	points := make([]point, len(candles))

	for i := FIBONACCI_LOOKBACK + 1; i < len(candles); i++ {
		s, ok := swingAt(p.high, p.low, i-FIBONACCI_LOOKBACK)

		if !ok {
			continue
		}

		level := s.extension(ratio)

		if s.rising && p.close[i-1] <= level && p.close[i] > level {
			points[i] = point{SELL, level}
		} else if !s.rising && p.close[i-1] >= level && p.close[i] < level {
			points[i] = point{BUY, level}
		}
	}

	return points
}
