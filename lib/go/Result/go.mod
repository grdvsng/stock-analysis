module github.com/grdvsng/stock-analysis/Result

go 1.25.0

replace github.com/grdvsng/stock-analysis/Logger v0.0.1 => ../Logger

replace github.com/grdvsng/stock-analysis/Exceptions v0.0.1 => ../Exceptions

require github.com/grdvsng/stock-analysis/Logger v0.0.1

require github.com/grdvsng/stock-analysis/Exceptions v0.0.1 // indirect
