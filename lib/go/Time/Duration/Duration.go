package Duration

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grdvsng/stock-analysis/Regex"
	"github.com/grdvsng/stock-analysis/Result"
)

var durationRegex = Regex.Compile(`^(?P<value>\d+).(?P<type>Hour|Minute|Second)s?$`)

type Type string

type Duration struct {
	type_   Type
	value   int
	seconds time.Duration
}

const (
	HourDuration   Type = "Hour"
	MinuteDuration Type = "Minute"
	SecondDuration Type = "Second"
)

func calculateSeconds(value int, type_ Type) (time.Duration, error) {
	var e error

	factor := 1

	switch Type(type_) {
	case HourDuration:
		factor = 60 * 60
	case MinuteDuration:
		factor = 60
	case SecondDuration:
		factor = 1
	default:
		e = fmt.Errorf("Invalid duration type %s", type_)
	}

	return time.Duration(value*factor) * time.Second, e
}

func (t *Duration) Duration() time.Duration {
	return t.seconds
}

func (t *Duration) UnmarshalText(b []byte) (err error) {
	raw := strings.Trim(string(b), " ")
	groups := durationRegex.NamedGroups(raw)
	value, err := groups.Get("value")

	if err != nil {
		return fmt.Errorf("Invalid duration %q", raw)
	}

	_type, _ := groups.Get("type")
	interval, err := strconv.Atoi(value)

	if err != nil {
		return err
	}

	t.type_ = Type(_type)
	t.seconds, err = calculateSeconds(interval, t.type_)
	t.value = interval

	return err
}

func (d Duration) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%d %s", d.value, d.type_), nil
}

func Create(value int, type_ Type) *Duration {
	return &Duration{
		type_:   type_,
		value:   value,
		seconds: Result.Ok(calculateSeconds(value, type_)),
	}
}
