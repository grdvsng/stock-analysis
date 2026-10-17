module github.com/grdvsng/stock-analysis/Sources

go 1.26

replace (
	github.com/grdvsng/stock-analysis/Data v0.0.1 => ../Data
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1 => ../Exceptions
	github.com/grdvsng/stock-analysis/File v0.0.1 => ../File
	github.com/grdvsng/stock-analysis/Logger v0.0.1 => ../Logger
	github.com/grdvsng/stock-analysis/Regex v0.0.1 => ../Regex
	github.com/grdvsng/stock-analysis/Result v0.0.1 => ../Result
	github.com/grdvsng/stock-analysis/Time v0.0.1 => ../Time
)

require (
	github.com/ThomasMarcelis/ibkr-go/v2 v2.1.0
	github.com/grdvsng/stock-analysis/Data v0.0.1
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1
	github.com/grdvsng/stock-analysis/Logger v0.0.1
	github.com/grdvsng/stock-analysis/Result v0.0.1
	github.com/grdvsng/stock-analysis/Time v0.0.1
)

require (
	github.com/shopspring/decimal v1.5.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
