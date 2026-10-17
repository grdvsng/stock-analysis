package Schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/grdvsng/stock-analysis/Time/Location"
)

const (
	YMD_FORMAT               = "2006-01-02"
	RAW_TIME_DELIMETER       = `;`
	RAW_TIME_PAIR_DELIMETER  = `-`
	RAW_TIME_DATETIME_FORMAT = "20060102:1504"
	RAW_TIME_CLOSED          = ":CLOSED"
)

type Schedule struct {
	End      map[string]int64   `json:"end"`
	Start    map[string]int64   `json:"start"`
	Location *Location.Location `json:"location"`
}

func (s *Schedule) Available() bool {
	if s.Location == nil {
		return false
	}

	now := s.Location.Now()

	return s.availableAt(now, now) || s.availableAt(now, now.AddDate(0, 0, -1))
}

func (s *Schedule) Actual() bool {
	if s.Location == nil {
		return false
	}

	now := time.Now().Unix()

	for _, end := range s.End {
		if end > now {
			return true
		}
	}

	return false
}

func (s *Schedule) availableAt(now time.Time, day time.Time) bool {
	ymd := day.Format(YMD_FORMAT)

	if start, exists := s.Start[ymd]; exists {
		if end, exists := s.End[ymd]; exists {
			nowSeconds := now.Unix()

			return nowSeconds >= start && end > nowSeconds
		}
	}

	return false
}

func parseError(raw string, msg string, args ...any) error {
	return fmt.Errorf("Can't parse '%s', error: '%s'", raw, fmt.Sprintf(msg, args...))
}

func Parse(raw string, timeZoneId string) (*Schedule, error) {
	pairs := strings.Split(raw, RAW_TIME_DELIMETER)
	start := map[string]int64{}
	end := map[string]int64{}
	loc, e := Location.Parse(timeZoneId)

	if e != nil {
		return nil, e
	}

	if strings.TrimSpace(raw) == "" {
		return nil, parseError(raw, "Raw string is empty")
	}

	for _, pair := range pairs {
		if pair == "" || strings.HasSuffix(pair, RAW_TIME_CLOSED) {
			continue
		}

		splited := strings.Split(pair, RAW_TIME_PAIR_DELIMETER)

		if len(splited) != 2 {
			return nil, parseError(raw, "Cant split line '%s' to end and start time ", pair)
		}

		ymd := ""

		for i, rawTime := range splited {
			parsed, e := loc.ParseIn(RAW_TIME_DATETIME_FORMAT, rawTime)

			if e == nil {
				if i == 0 {
					ymd = parsed.Format(YMD_FORMAT)
					start[ymd] = parsed.Unix()
				} else {
					end[ymd] = parsed.Unix()
				}
			} else {
				return nil, parseError(raw, "Cant parse time '%s' to end and start time: '%s'", rawTime, e)
			}
		}
	}

	return &Schedule{
		Start:    start,
		End:      end,
		Location: loc,
	}, nil
}
