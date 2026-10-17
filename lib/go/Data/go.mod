module github.com/grdvsng/stock-analysis/Data

go 1.26

replace github.com/grdvsng/stock-analysis/Time v0.0.1 => ../Time

require (
	github.com/grdvsng/stock-analysis/Time v0.0.1
	github.com/shopspring/decimal v1.5.0
)
