package Exceptions

import (
	"fmt"
	"runtime/debug"
	"strings"
)

type Exception interface {
	Render(wantStack bool, args ...any) string
}

type exception struct {
	Template string
}

func trace() string {
	var builder strings.Builder

	builder.Write(debug.Stack())

	return strings.TrimRight(builder.String(), "\n")
}

func (e *exception) Render(wantStack bool, args ...any) string {
	message := e.Template

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	if wantStack {
		message = fmt.Sprintf("%s\n%s", trace(), message)
	}

	return message
}
