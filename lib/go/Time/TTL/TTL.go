package TTL

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grdvsng/stock-analysis/Regex"
)

type Type string

const (
	DayTTL    Type = "Day"
	HourTTL   Type = "Hour"
	MinuteTTL Type = "Minute"
)

var ttlRegex = Regex.Compile(`^(?P<value>\d+).(?P<type>Day|Hour|Minute)s?$`)

type TTL struct {
	type_   Type
	value   int
	seconds time.Duration
}

func (t *TTL) UnmarshalText(b []byte) (err error) {
	raw := strings.Trim(string(b), " ")
	groups := ttlRegex.NamedGroups(raw)
	value, err := groups.Get("value")

	if err != nil {
		return fmt.Errorf("Invalid ttl %q", raw)
	}

	_type, _ := groups.Get("type")
	interval, err := strconv.Atoi(value)

	if err != nil {
		return err
	}

	seconds := interval

	switch Type(_type) {
	case DayTTL:
		seconds = (interval * 60 * 60 * 24)
	case HourTTL:
		seconds = (interval * 60 * 60)
	case MinuteTTL:
		seconds = (interval * 60)
	default:
		return fmt.Errorf("Invalid ttl type %s", _type)
	}

	t.type_ = Type(_type)
	t.seconds = time.Duration(seconds) * time.Second
	t.value = interval

	return nil
}

func (t *TTL) Duration() time.Duration {
	return t.seconds
}

func (t TTL) MarshalText() ([]byte, error) {
	return fmt.Appendf(nil, "%d %s", t.value, t.type_), nil
}
