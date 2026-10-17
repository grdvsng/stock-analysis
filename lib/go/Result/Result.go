package Result

import "github.com/grdvsng/stock-analysis/Logger"

func Ok[T any](result T, err error) T {
	Check(err)

	return result
}

func Check(err error) {
	if err != nil {
		Logger.Throw(RunTimeError, err)
	}
}
