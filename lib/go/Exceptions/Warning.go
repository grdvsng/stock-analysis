package Exceptions

type Warning interface {
	Exception
}

type warning struct {
	exception
}

func CreateWarning(template string) Warning {
	return &ierror{
		exception: exception{Template: template},
	}
}
