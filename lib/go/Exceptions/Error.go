package Exceptions

type Error interface {
	Exception
	Panic(args ...any)
}

type ierror struct {
	exception
}

func (e *ierror) Panic(args ...any) {
	panic(e.Render(true, args...))
}

func CreateError(template string) Error {
	return &ierror{
		exception: exception{Template: template},
	}
}
