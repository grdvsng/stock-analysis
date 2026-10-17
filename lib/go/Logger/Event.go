package Logger

import "time"

type Event struct {
	Time    time.Time
	Level   Level
	Message string
	Host    string
	File    string
	Trace   string
}
