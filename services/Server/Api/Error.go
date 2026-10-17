package Api

import "github.com/grdvsng/stock-analysis/Exceptions"

var RequestError = Exceptions.CreateError("%s %s failed: %s")

var SourceError = Exceptions.CreateError("Source unavailable: %s")

var SignalWarning = Exceptions.CreateWarning("Data collector signal %s for %s failed: %s")

var StreamError = Exceptions.CreateError("Stream error: %s")
