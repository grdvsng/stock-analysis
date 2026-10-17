package Schedule

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/grdvsng/stock-analysis/Time/Location"
)

type Session struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type document struct {
	TimeZone string    `json:"time_zone"`
	Sessions []Session `json:"sessions"`
}

type legacy struct {
	Start    map[string]int64   `json:"start"`
	End      map[string]int64   `json:"end"`
	Location *Location.Location `json:"location"`
}

func (s Schedule) Sessions() []Session {
	sessions := make([]Session, 0, len(s.Start))

	for day, start := range s.Start {
		if end, exists := s.End[day]; exists {
			sessions = append(sessions, Session{Start: time.Unix(start, 0).UTC(), End: time.Unix(end, 0).UTC()})
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Start.Before(sessions[j].Start)
	})

	return sessions
}

func (s Schedule) MarshalJSON() ([]byte, error) {
	result := document{Sessions: s.Sessions()}

	if s.Location != nil {
		result.TimeZone = s.Location.Name
	}

	return json.Marshal(result)
}

func (s *Schedule) UnmarshalJSON(raw []byte) error {
	current := document{}

	if e := json.Unmarshal(raw, &current); e != nil {
		return e
	}

	if current.Sessions == nil && current.TimeZone == "" {
		old := legacy{}

		if e := json.Unmarshal(raw, &old); e != nil {
			return e
		}

		*s = Schedule{Start: old.Start, End: old.End, Location: old.Location}

		return nil
	}

	result := Schedule{Start: map[string]int64{}, End: map[string]int64{}}

	if current.TimeZone != "" {
		location, e := Location.Parse(current.TimeZone)

		if e != nil {
			return e
		}

		result.Location = location
	}

	for _, session := range current.Sessions {
		start := session.Start

		if result.Location != nil {
			start = start.In(&result.Location.Location)
		}

		day := start.Format(YMD_FORMAT)
		result.Start[day] = session.Start.Unix()
		result.End[day] = session.End.Unix()
	}

	*s = result

	return nil
}
