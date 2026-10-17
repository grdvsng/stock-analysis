package Location

import (
	"time"
)

type Location struct {
	time.Location

	Name string
}

func (location Location) MarshalText() ([]byte, error) {
	return []byte(location.Name), nil
}

func (location *Location) UnmarshalText(text []byte) error {
	parsed, e := Parse(string(text))

	if e != nil {
		return e
	}

	*location = *parsed

	return nil
}

func (location *Location) ParseIn(format, raw string) (time.Time, error) {
	return time.ParseInLocation(format, raw, &location.Location)
}

func (location *Location) Now() time.Time {
	now := time.Now()

	return now.In(&location.Location)
}

func Parse(id string) (*Location, error) {
	loc, e := time.LoadLocation(id)

	if e == nil {
		return &Location{
			Location: *loc,
			Name:     id,
		}, nil
	}

	return nil, e
}
