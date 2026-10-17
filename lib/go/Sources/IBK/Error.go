package IBK

import "github.com/grdvsng/stock-analysis/Exceptions"

var IBKApiError = Exceptions.CreateError("IBK API error: %s")

var ScheduleParseError = Exceptions.CreateError("Can't parse %s schedule for %s: %s")

var ContractNotFoundError = Exceptions.CreateError("Contract for %s not found")

var TicksError = Exceptions.CreateError("Can't get %s ticks for %s: %s")
