package Indicators

import "math"

const TREND_PERIOD = 5

type bar struct {
	Candle
}

func (b bar) body() float64        { return math.Abs(b.Close - b.Open) }
func (b bar) span() float64        { return b.High - b.Low }
func (b bar) top() float64         { return math.Max(b.Open, b.Close) }
func (b bar) bottom() float64      { return math.Min(b.Open, b.Close) }
func (b bar) upperShadow() float64 { return b.High - b.top() }
func (b bar) lowerShadow() float64 { return b.bottom() - b.Low }
func (b bar) bullish() bool        { return b.Close > b.Open }
func (b bar) bearish() bool        { return b.Close < b.Open }
func (b bar) middle() float64      { return (b.Open + b.Close) / 2 }

func (b bar) long(average float64) bool {
	return b.body() >= average
}

func (b bar) small(average float64) bool {
	return b.body() <= average*0.5
}

func (b bar) doji() bool {
	return b.span() > 0 && b.body() <= b.span()*0.1
}

type context struct {
	bars    []bar
	average []float64
}

func newContext(candles []Candle) context {
	bars := make([]bar, len(candles))
	bodies := make([]float64, len(candles))

	for i, candle := range candles {
		bars[i] = bar{candle}
		bodies[i] = bars[i].body()
	}

	return context{bars: bars, average: shift(sma(bodies, 10), 1)}
}

func (c context) down(i int) bool {
	return i-TREND_PERIOD >= 0 && c.bars[i].Close < c.bars[i-TREND_PERIOD].Close
}

func (c context) up(i int) bool {
	return i-TREND_PERIOD >= 0 && c.bars[i].Close > c.bars[i-TREND_PERIOD].Close
}

type pattern func(c context, i int) bool

func detect(candles []Candle, size int, bullish, bearish pattern) []point {
	c := newContext(candles)
	points := make([]point, len(candles))

	for i := size - 1; i < len(candles); i++ {
		if !defined(c.average[i]) || c.average[i] == 0 {
			continue
		}

		if bullish != nil && bullish(c, i) {
			points[i] = point{BUY, candles[i].Close}
		} else if bearish != nil && bearish(c, i) {
			points[i] = point{SELL, candles[i].Close}
		}
	}

	return points
}

func engulfingUp(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.down(i-1) && a.bearish() && b.bullish() && b.Open <= a.Close && b.Close >= a.Open && b.body() > a.body()
}

func engulfingDown(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.up(i-1) && a.bullish() && b.bearish() && b.Open >= a.Close && b.Close <= a.Open && b.body() > a.body()
}

func hammerShape(b bar) bool {
	return b.span() > 0 && b.body() > 0 && b.lowerShadow() >= 2*b.body() && b.upperShadow() <= b.body()*0.5
}

func invertedShape(b bar) bool {
	return b.span() > 0 && b.body() > 0 && b.upperShadow() >= 2*b.body() && b.lowerShadow() <= b.body()*0.5
}

func hammer(c context, i int) bool     { return c.down(i-1) && hammerShape(c.bars[i]) }
func hangingMan(c context, i int) bool { return c.up(i-1) && hammerShape(c.bars[i]) }
func invertedHammer(c context, i int) bool {
	return c.down(i-1) && invertedShape(c.bars[i])
}
func shootingStar(c context, i int) bool { return c.up(i-1) && invertedShape(c.bars[i]) }

func dragonfly(c context, i int) bool {
	b := c.bars[i]

	return c.down(i-1) && b.doji() && b.upperShadow() <= b.span()*0.1 && b.lowerShadow() >= b.span()*0.6
}

func gravestone(c context, i int) bool {
	b := c.bars[i]

	return c.up(i-1) && b.doji() && b.lowerShadow() <= b.span()*0.1 && b.upperShadow() >= b.span()*0.6
}

func morningStar(c context, i int) bool {
	a, b, d := c.bars[i-2], c.bars[i-1], c.bars[i]

	return c.down(i-2) && a.bearish() && a.long(c.average[i]) && b.small(c.average[i]) && b.top() <= a.Close && d.bullish() && d.Close > a.middle()
}

func eveningStar(c context, i int) bool {
	a, b, d := c.bars[i-2], c.bars[i-1], c.bars[i]

	return c.up(i-2) && a.bullish() && a.long(c.average[i]) && b.small(c.average[i]) && b.bottom() >= a.Close && d.bearish() && d.Close < a.middle()
}

func soldiers(c context, i int) bool {
	for j := i - 2; j <= i; j++ {
		b := c.bars[j]

		if !b.bullish() || !b.long(c.average[i]*0.8) || b.upperShadow() > b.body()*0.3 {
			return false
		}

		if j > i-2 && (b.Close <= c.bars[j-1].Close || b.Open < c.bars[j-1].Open || b.Open > c.bars[j-1].Close) {
			return false
		}
	}

	return true
}

func crows(c context, i int) bool {
	for j := i - 2; j <= i; j++ {
		b := c.bars[j]

		if !b.bearish() || !b.long(c.average[i]*0.8) || b.lowerShadow() > b.body()*0.3 {
			return false
		}

		if j > i-2 && (b.Close >= c.bars[j-1].Close || b.Open > c.bars[j-1].Open || b.Open < c.bars[j-1].Close) {
			return false
		}
	}

	return true
}

func haramiUp(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.down(i-1) && a.bearish() && a.long(c.average[i]) && b.bullish() && b.small(c.average[i]) && b.top() <= a.Open && b.bottom() >= a.Close
}

func haramiDown(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.up(i-1) && a.bullish() && a.long(c.average[i]) && b.bearish() && b.small(c.average[i]) && b.top() <= a.Close && b.bottom() >= a.Open
}

func piercing(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.down(i-1) && a.bearish() && a.long(c.average[i]) && b.bullish() && b.Open < a.Low && b.Close > a.middle() && b.Close < a.Open
}

func darkCloud(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.up(i-1) && a.bullish() && a.long(c.average[i]) && b.bearish() && b.Open > a.High && b.Close < a.middle() && b.Close > a.Open
}

func tweezerBottom(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.down(i-1) && a.bearish() && b.bullish() && math.Abs(a.Low-b.Low) <= math.Max(a.span(), b.span())*0.05
}

func tweezerTop(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return c.up(i-1) && a.bullish() && b.bearish() && math.Abs(a.High-b.High) <= math.Max(a.span(), b.span())*0.05
}

func marubozuUp(c context, i int) bool {
	b := c.bars[i]

	return b.bullish() && b.long(c.average[i]) && b.body() >= b.span()*0.95
}

func marubozuDown(c context, i int) bool {
	b := c.bars[i]

	return b.bearish() && b.long(c.average[i]) && b.body() >= b.span()*0.95
}

func risingThree(c context, i int) bool {
	first, last := c.bars[i-4], c.bars[i]

	if !first.bullish() || !first.long(c.average[i]) || !last.bullish() || last.Close <= first.Close {
		return false
	}

	for j := i - 3; j < i; j++ {
		if c.bars[j].top() > first.High || c.bars[j].bottom() < first.Low || !c.bars[j].small(c.average[i]*1.5) {
			return false
		}
	}

	return true
}

func fallingThree(c context, i int) bool {
	first, last := c.bars[i-4], c.bars[i]

	if !first.bearish() || !first.long(c.average[i]) || !last.bearish() || last.Close >= first.Close {
		return false
	}

	for j := i - 3; j < i; j++ {
		if c.bars[j].top() > first.High || c.bars[j].bottom() < first.Low || !c.bars[j].small(c.average[i]*1.5) {
			return false
		}
	}

	return true
}

func kickerUp(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return a.bearish() && b.bullish() && b.Open > a.Open && b.long(c.average[i])
}

func kickerDown(c context, i int) bool {
	a, b := c.bars[i-1], c.bars[i]

	return a.bullish() && b.bearish() && b.Open < a.Open && b.long(c.average[i])
}

func insideUp(c context, i int) bool {
	return haramiUp(c, i-1) && c.bars[i].bullish() && c.bars[i].Close > c.bars[i-2].Open
}

func insideDown(c context, i int) bool {
	return haramiDown(c, i-1) && c.bars[i].bearish() && c.bars[i].Close < c.bars[i-2].Open
}

func outsideUp(c context, i int) bool {
	return engulfingUp(c, i-1) && c.bars[i].bullish() && c.bars[i].Close > c.bars[i-1].Close
}

func outsideDown(c context, i int) bool {
	return engulfingDown(c, i-1) && c.bars[i].bearish() && c.bars[i].Close < c.bars[i-1].Close
}
