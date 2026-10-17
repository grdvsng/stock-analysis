package Indicators

import "math"

func crossedAbove(a, b []float64, i int) bool {
	return i > 0 && defined(a[i-1], b[i-1], a[i], b[i]) && a[i-1] <= b[i-1] && a[i] > b[i]
}

func crossedBelow(a, b []float64, i int) bool {
	return i > 0 && defined(a[i-1], b[i-1], a[i], b[i]) && a[i-1] >= b[i-1] && a[i] < b[i]
}

func cross(a, b []float64) []point {
	points := make([]point, len(a))

	for i := range a {
		if crossedAbove(a, b, i) {
			points[i] = point{BUY, a[i]}
		} else if crossedBelow(a, b, i) {
			points[i] = point{SELL, a[i]}
		}
	}

	return points
}

func levels(values []float64, low, high float64) []point {
	n := len(values)
	bottom, top := constant(n, low), constant(n, high)
	points := make([]point, n)

	for i := range values {
		if crossedAbove(values, bottom, i) {
			points[i] = point{BUY, values[i]}
		} else if crossedBelow(values, top, i) {
			points[i] = point{SELL, values[i]}
		}
	}

	return points
}

func reversal(close, upper, lower []float64) []point {
	points := make([]point, len(close))

	for i := range close {
		if crossedAbove(close, lower, i) {
			points[i] = point{BUY, close[i]}
		} else if crossedBelow(close, upper, i) {
			points[i] = point{SELL, close[i]}
		}
	}

	return points
}

func breakout(close, upper, lower []float64) []point {
	points := make([]point, len(close))

	for i := range close {
		if crossedAbove(close, upper, i) {
			points[i] = point{BUY, close[i]}
		} else if crossedBelow(close, lower, i) {
			points[i] = point{SELL, close[i]}
		}
	}

	return points
}

func zero(values []float64) []point {
	return cross(values, constant(len(values), 0))
}

func turning(values []float64) []point {
	points := make([]point, len(values))

	for i := 2; i < len(values); i++ {
		if !defined(values[i-2], values[i-1], values[i]) {
			continue
		}

		if values[i-1] < 0 && values[i-2] > values[i-1] && values[i] > values[i-1] {
			points[i] = point{BUY, values[i]}
		} else if values[i-1] > 0 && values[i-2] < values[i-1] && values[i] < values[i-1] {
			points[i] = point{SELL, values[i]}
		}
	}

	return points
}

func gated(points []point, strength []float64, minimum float64) []point {
	for i := range points {
		if points[i].action != "" && !(defined(strength[i]) && strength[i] >= minimum) {
			points[i] = point{}
		}
	}

	return points
}

func squeeze(close []float64) []point {
	_, upper, lower := bollinger(close, 20, 2)
	widths := combine(upper, lower, func(x, y float64) float64 { return x - y })
	narrowest := lowest(widths, 120)
	points := make([]point, len(close))

	for i := 1; i < len(close); i++ {
		if !defined(narrowest[i-1], widths[i-1]) || widths[i-1] > narrowest[i-1]*1.05 {
			continue
		}

		if crossedAbove(close, upper, i) {
			points[i] = point{BUY, close[i]}
		} else if crossedBelow(close, lower, i) {
			points[i] = point{SELL, close[i]}
		}
	}

	return points
}

func saucer(values []float64) []point {
	points := make([]point, len(values))

	for i := 2; i < len(values); i++ {
		if !defined(values[i-2], values[i-1], values[i]) {
			continue
		}

		falling := values[i-2] > values[i-1]

		if values[i] > 0 && values[i-1] > 0 && values[i-2] > 0 && falling && values[i] > values[i-1] {
			points[i] = point{BUY, values[i]}
		} else if values[i] < 0 && values[i-1] < 0 && values[i-2] < 0 && !falling && values[i] < values[i-1] {
			points[i] = point{SELL, values[i]}
		}
	}

	return points
}

func atrBreakout(high, low, close []float64, period int, width float64) []point {
	ranges := atr(high, low, close, period)
	points := make([]point, len(close))

	for i := 1; i < len(close); i++ {
		if !defined(ranges[i-1]) {
			continue
		}

		if close[i] > close[i-1]+width*ranges[i-1] {
			points[i] = point{BUY, close[i]}
		} else if close[i] < close[i-1]-width*ranges[i-1] {
			points[i] = point{SELL, close[i]}
		}
	}

	return points
}

func elder(high, low, close []float64, period int) []point {
	average := ema(close, period)
	bull := subtract(high, average)
	bear := subtract(low, average)
	points := make([]point, len(close))

	for i := 1; i < len(close); i++ {
		if !defined(average[i], average[i-1], bull[i], bear[i], bear[i-1], bull[i-1]) {
			continue
		}

		rising := average[i] > average[i-1]

		if rising && bear[i] < 0 && bear[i] > bear[i-1] {
			points[i] = point{BUY, bear[i]}
		} else if !rising && bull[i] > 0 && bull[i] < bull[i-1] {
			points[i] = point{SELL, bull[i]}
		}
	}

	return points
}

func psarFlip(high, low []float64) []point {
	sar, rising := parabolic(high, low, 0.02, 0.2)
	points := make([]point, len(high))

	for i := 2; i < len(high); i++ {
		if rising[i] && !rising[i-1] {
			points[i] = point{BUY, sar[i]}
		} else if !rising[i] && rising[i-1] {
			points[i] = point{SELL, sar[i]}
		}
	}

	return points
}

func cloud(close []float64, high, low []float64) []point {
	_, _, spanA, spanB := ichimoku(high, low)
	top := combine(spanA, spanB, math.Max)
	bottom := combine(spanA, spanB, math.Min)

	return breakout(close, top, bottom)
}
