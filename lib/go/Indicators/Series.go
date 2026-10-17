package Indicators

import "math"

func empty(n int) []float64 {
	values := make([]float64, n)

	for i := range values {
		values[i] = math.NaN()
	}

	return values
}

func defined(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}

	return true
}

func constant(n int, value float64) []float64 {
	values := make([]float64, n)

	for i := range values {
		values[i] = value
	}

	return values
}

func combine(a, b []float64, operation func(x, y float64) float64) []float64 {
	result := empty(len(a))

	for i := range a {
		if defined(a[i], b[i]) {
			result[i] = operation(a[i], b[i])
		}
	}

	return result
}

func subtract(a, b []float64) []float64 {
	return combine(a, b, func(x, y float64) float64 { return x - y })
}

func shift(values []float64, by int) []float64 {
	result := empty(len(values))

	for i := range values {
		if i-by >= 0 && i-by < len(values) {
			result[i] = values[i-by]
		}
	}

	return result
}

func sma(values []float64, period int) []float64 {
	result := empty(len(values))
	sum, count := 0.0, 0

	for i, value := range values {
		if !defined(value) {
			sum, count = 0, 0

			continue
		}

		sum += value
		count++

		if count > period {
			sum -= values[i-period]
			count = period
		}

		if count == period {
			result[i] = sum / float64(period)
		}
	}

	return result
}

func smooth(values []float64, period int, alpha float64) []float64 {
	result := empty(len(values))
	seed, count := 0.0, 0
	previous := math.NaN()

	for i, value := range values {
		if !defined(value) {
			seed, count, previous = 0, 0, math.NaN()

			continue
		}

		if defined(previous) {
			previous = previous + alpha*(value-previous)
			result[i] = previous

			continue
		}

		seed += value
		count++

		if count == period {
			previous = seed / float64(period)
			result[i] = previous
		}
	}

	return result
}

func ema(values []float64, period int) []float64 {
	return smooth(values, period, 2/float64(period+1))
}

func wilder(values []float64, period int) []float64 {
	return smooth(values, period, 1/float64(period))
}

func wma(values []float64, period int) []float64 {
	result := empty(len(values))
	weights := float64(period*(period+1)) / 2

	for i := period - 1; i < len(values); i++ {
		sum := 0.0
		ok := true

		for j := 0; j < period; j++ {
			value := values[i-j]

			if !defined(value) {
				ok = false

				break
			}

			sum += value * float64(period-j)
		}

		if ok {
			result[i] = sum / weights
		}
	}

	return result
}

func dema(values []float64, period int) []float64 {
	first := ema(values, period)
	second := ema(first, period)

	return combine(first, second, func(x, y float64) float64 { return 2*x - y })
}

func tema(values []float64, period int) []float64 {
	first := ema(values, period)
	second := ema(first, period)
	third := ema(second, period)
	result := empty(len(values))

	for i := range values {
		if defined(first[i], second[i], third[i]) {
			result[i] = 3*first[i] - 3*second[i] + third[i]
		}
	}

	return result
}

func hma(values []float64, period int) []float64 {
	half := wma(values, period/2)
	full := wma(values, period)
	raw := combine(half, full, func(x, y float64) float64 { return 2*x - y })

	return wma(raw, int(math.Round(math.Sqrt(float64(period)))))
}

func window(values []float64, period, i int, pick func(a, b float64) float64) float64 {
	if i-period+1 < 0 {
		return math.NaN()
	}

	result := values[i]

	for j := i - period + 1; j <= i; j++ {
		if !defined(values[j]) {
			return math.NaN()
		}

		result = pick(result, values[j])
	}

	return result
}

func highest(values []float64, period int) []float64 {
	result := empty(len(values))

	for i := range values {
		result[i] = window(values, period, i, math.Max)
	}

	return result
}

func lowest(values []float64, period int) []float64 {
	result := empty(len(values))

	for i := range values {
		result[i] = window(values, period, i, math.Min)
	}

	return result
}

func stddev(values []float64, period int) []float64 {
	mean := sma(values, period)
	result := empty(len(values))

	for i := period - 1; i < len(values); i++ {
		if !defined(mean[i]) {
			continue
		}

		sum := 0.0

		for j := i - period + 1; j <= i; j++ {
			sum += (values[j] - mean[i]) * (values[j] - mean[i])
		}

		result[i] = math.Sqrt(sum / float64(period))
	}

	return result
}

func change(values []float64, period int) []float64 {
	return combine(values, shift(values, period), func(x, y float64) float64 { return x - y })
}

func roc(values []float64, period int) []float64 {
	return combine(values, shift(values, period), func(x, y float64) float64 {
		if y == 0 {
			return math.NaN()
		}

		return (x - y) / y * 100
	})
}

func rsi(values []float64, period int) []float64 {
	gains := empty(len(values))
	losses := empty(len(values))

	for i := 1; i < len(values); i++ {
		if defined(values[i], values[i-1]) {
			delta := values[i] - values[i-1]
			gains[i] = math.Max(delta, 0)
			losses[i] = math.Max(-delta, 0)
		}
	}

	up := wilder(gains, period)
	down := wilder(losses, period)

	return combine(up, down, func(x, y float64) float64 {
		if y == 0 {
			return 100
		}

		return 100 - 100/(1+x/y)
	})
}

func stochastic(high, low, close []float64, period, smoothing int) ([]float64, []float64) {
	top := highest(high, period)
	bottom := lowest(low, period)
	k := empty(len(close))

	for i := range close {
		if defined(top[i], bottom[i], close[i]) && top[i] != bottom[i] {
			k[i] = (close[i] - bottom[i]) / (top[i] - bottom[i]) * 100
		}
	}

	return k, sma(k, smoothing)
}

func stochasticRsi(close []float64, period, smoothK, smoothD int) ([]float64, []float64) {
	values := rsi(close, period)
	top := highest(values, period)
	bottom := lowest(values, period)
	raw := empty(len(close))

	for i := range close {
		if defined(top[i], bottom[i], values[i]) && top[i] != bottom[i] {
			raw[i] = (values[i] - bottom[i]) / (top[i] - bottom[i]) * 100
		}
	}

	k := sma(raw, smoothK)

	return k, sma(k, smoothD)
}

func macd(close []float64, fast, slow, signal int) ([]float64, []float64, []float64) {
	line := subtract(ema(close, fast), ema(close, slow))
	trigger := ema(line, signal)

	return line, trigger, subtract(line, trigger)
}

func ppo(close []float64, fast, slow, signal int) ([]float64, []float64) {
	line := combine(ema(close, fast), ema(close, slow), func(x, y float64) float64 {
		if y == 0 {
			return math.NaN()
		}

		return (x - y) / y * 100
	})

	return line, ema(line, signal)
}

func bollinger(close []float64, period int, width float64) ([]float64, []float64, []float64) {
	middle := sma(close, period)
	deviation := stddev(close, period)
	upper := combine(middle, deviation, func(x, y float64) float64 { return x + width*y })
	lower := combine(middle, deviation, func(x, y float64) float64 { return x - width*y })

	return middle, upper, lower
}

func trueRange(high, low, close []float64) []float64 {
	result := empty(len(close))

	for i := range close {
		if i == 0 {
			result[i] = high[i] - low[i]

			continue
		}

		result[i] = math.Max(high[i]-low[i], math.Max(math.Abs(high[i]-close[i-1]), math.Abs(low[i]-close[i-1])))
	}

	return result
}

func atr(high, low, close []float64, period int) []float64 {
	return wilder(trueRange(high, low, close), period)
}

func keltner(high, low, close []float64, period int, width float64) ([]float64, []float64) {
	middle := ema(close, period)
	ranges := atr(high, low, close, period)
	upper := combine(middle, ranges, func(x, y float64) float64 { return x + width*y })
	lower := combine(middle, ranges, func(x, y float64) float64 { return x - width*y })

	return upper, lower
}

func cci(typical []float64, period int) []float64 {
	mean := sma(typical, period)
	result := empty(len(typical))

	for i := period - 1; i < len(typical); i++ {
		if !defined(mean[i]) {
			continue
		}

		deviation := 0.0

		for j := i - period + 1; j <= i; j++ {
			deviation += math.Abs(typical[j] - mean[i])
		}

		deviation /= float64(period)

		if deviation != 0 {
			result[i] = (typical[i] - mean[i]) / (0.015 * deviation)
		}
	}

	return result
}

func williams(high, low, close []float64, period int) []float64 {
	top := highest(high, period)
	bottom := lowest(low, period)
	result := empty(len(close))

	for i := range close {
		if defined(top[i], bottom[i]) && top[i] != bottom[i] {
			result[i] = (top[i] - close[i]) / (top[i] - bottom[i]) * -100
		}
	}

	return result
}

func directional(high, low, close []float64, period int) ([]float64, []float64, []float64) {
	plus := empty(len(close))
	minus := empty(len(close))

	for i := 1; i < len(close); i++ {
		up := high[i] - high[i-1]
		down := low[i-1] - low[i]
		plus[i], minus[i] = 0, 0

		if up > down && up > 0 {
			plus[i] = up
		}

		if down > up && down > 0 {
			minus[i] = down
		}
	}

	ranges := wilder(trueRange(high, low, close), period)
	plusSmooth := wilder(plus, period)
	minusSmooth := wilder(minus, period)
	plusDi := combine(plusSmooth, ranges, func(x, y float64) float64 { return 100 * x / y })
	minusDi := combine(minusSmooth, ranges, func(x, y float64) float64 { return 100 * x / y })
	dx := combine(plusDi, minusDi, func(x, y float64) float64 {
		if x+y == 0 {
			return 0
		}

		return 100 * math.Abs(x-y) / (x + y)
	})

	return wilder(dx, period), plusDi, minusDi
}

func aroon(high, low []float64, period int) ([]float64, []float64) {
	up := empty(len(high))
	down := empty(len(high))

	for i := period; i < len(high); i++ {
		top, bottom := i-period, i-period

		for j := i - period; j <= i; j++ {
			if high[j] >= high[top] {
				top = j
			}

			if low[j] <= low[bottom] {
				bottom = j
			}
		}

		up[i] = float64(period-(i-top)) / float64(period) * 100
		down[i] = float64(period-(i-bottom)) / float64(period) * 100
	}

	return up, down
}

func parabolic(high, low []float64, step, limit float64) ([]float64, []bool) {
	sar := empty(len(high))
	rising := make([]bool, len(high))

	if len(high) < 2 {
		return sar, rising
	}

	up := high[1] >= high[0]
	acceleration := step
	extreme := high[0]
	current := low[0]

	if !up {
		extreme = low[0]
		current = high[0]
	}

	for i := 1; i < len(high); i++ {
		current = current + acceleration*(extreme-current)

		if up {
			current = math.Min(current, math.Min(low[i-1], low[max(i-2, 0)]))

			if low[i] < current {
				up, current, extreme, acceleration = false, extreme, low[i], step
			} else if high[i] > extreme {
				extreme = high[i]
				acceleration = math.Min(acceleration+step, limit)
			}
		} else {
			current = math.Max(current, math.Max(high[i-1], high[max(i-2, 0)]))

			if high[i] > current {
				up, current, extreme, acceleration = true, extreme, high[i], step
			} else if low[i] < extreme {
				extreme = low[i]
				acceleration = math.Min(acceleration+step, limit)
			}
		}

		sar[i] = current
		rising[i] = up
	}

	return sar, rising
}

func midpoint(high, low []float64, period int) []float64 {
	return combine(highest(high, period), lowest(low, period), func(x, y float64) float64 { return (x + y) / 2 })
}

func ichimoku(high, low []float64) ([]float64, []float64, []float64, []float64) {
	tenkan := midpoint(high, low, 9)
	kijun := midpoint(high, low, 26)
	spanA := shift(combine(tenkan, kijun, func(x, y float64) float64 { return (x + y) / 2 }), 26)
	spanB := shift(midpoint(high, low, 52), 26)

	return tenkan, kijun, spanA, spanB
}

func trix(close []float64, period int) []float64 {
	triple := ema(ema(ema(close, period), period), period)

	return roc(triple, 1)
}

func cmo(close []float64, period int) []float64 {
	result := empty(len(close))

	for i := period; i < len(close); i++ {
		up, down := 0.0, 0.0

		for j := i - period + 1; j <= i; j++ {
			delta := close[j] - close[j-1]

			if delta > 0 {
				up += delta
			} else {
				down -= delta
			}
		}

		if up+down != 0 {
			result[i] = (up - down) / (up + down) * 100
		}
	}

	return result
}

func ultimate(high, low, close []float64) []float64 {
	pressure := empty(len(close))
	ranges := empty(len(close))

	for i := 1; i < len(close); i++ {
		bottom := math.Min(low[i], close[i-1])
		top := math.Max(high[i], close[i-1])
		pressure[i] = close[i] - bottom
		ranges[i] = top - bottom
	}

	average := func(period int) []float64 {
		return combine(sma(pressure, period), sma(ranges, period), func(x, y float64) float64 {
			if y == 0 {
				return math.NaN()
			}

			return x / y
		})
	}

	short, middle, long := average(7), average(14), average(28)
	result := empty(len(close))

	for i := range close {
		if defined(short[i], middle[i], long[i]) {
			result[i] = 100 * (4*short[i] + 2*middle[i] + long[i]) / 7
		}
	}

	return result
}

func awesome(median []float64) []float64 {
	return subtract(sma(median, 5), sma(median, 34))
}

func mfi(typical, volume []float64, period int) []float64 {
	result := empty(len(typical))

	for i := period; i < len(typical); i++ {
		positive, negative := 0.0, 0.0

		for j := i - period + 1; j <= i; j++ {
			flow := typical[j] * volume[j]

			if typical[j] > typical[j-1] {
				positive += flow
			} else if typical[j] < typical[j-1] {
				negative += flow
			}
		}

		if negative == 0 {
			result[i] = 100
		} else {
			result[i] = 100 - 100/(1+positive/negative)
		}
	}

	return result
}

func obv(close, volume []float64) []float64 {
	result := make([]float64, len(close))

	for i := 1; i < len(close); i++ {
		result[i] = result[i-1]

		if close[i] > close[i-1] {
			result[i] += volume[i]
		} else if close[i] < close[i-1] {
			result[i] -= volume[i]
		}
	}

	return result
}

func cmf(high, low, close, volume []float64, period int) []float64 {
	flow := make([]float64, len(close))

	for i := range close {
		if high[i] != low[i] {
			flow[i] = ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i]) * volume[i]
		}
	}

	return combine(sma(flow, period), sma(volume, period), func(x, y float64) float64 {
		if y == 0 {
			return math.NaN()
		}

		return x / y
	})
}

func vwap(typical, volume []float64, period int) []float64 {
	weighted := make([]float64, len(typical))

	for i := range typical {
		weighted[i] = typical[i] * volume[i]
	}

	return combine(sma(weighted, period), sma(volume, period), func(x, y float64) float64 {
		if y == 0 {
			return math.NaN()
		}

		return x / y
	})
}
