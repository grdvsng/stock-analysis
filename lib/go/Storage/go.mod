module github.com/grdvsng/stock-analysis/Storage

go 1.26

replace (
	github.com/grdvsng/stock-analysis/Data v0.0.1 => ../Data
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1 => ../Exceptions
	github.com/grdvsng/stock-analysis/Logger v0.0.1 => ../Logger
	github.com/grdvsng/stock-analysis/Regex v0.0.1 => ../Regex
	github.com/grdvsng/stock-analysis/Result v0.0.1 => ../Result
	github.com/grdvsng/stock-analysis/Time v0.0.1 => ../Time
)

require (
	github.com/grdvsng/stock-analysis/Data v0.0.1
	github.com/grdvsng/stock-analysis/Time v0.0.1
	github.com/shopspring/decimal v1.5.0
)
