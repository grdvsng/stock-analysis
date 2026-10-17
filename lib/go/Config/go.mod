module github.com/grdvsng/stock-analysis/Config

go 1.26

replace (
	github.com/grdvsng/stock-analysis/Data v0.0.1 => ../Data
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1 => ../Exceptions
	github.com/grdvsng/stock-analysis/File v0.0.1 => ../File
	github.com/grdvsng/stock-analysis/Logger v0.0.1 => ../Logger
	github.com/grdvsng/stock-analysis/Regex v0.0.1 => ../Regex
	github.com/grdvsng/stock-analysis/Result v0.0.1 => ../Result
	github.com/grdvsng/stock-analysis/Sources v0.0.1 => ../Sources
	github.com/grdvsng/stock-analysis/Time v0.0.1 => ../Time
)

require (
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/grdvsng/stock-analysis/Logger v0.0.1
	github.com/grdvsng/stock-analysis/Sources v0.0.1
	github.com/grdvsng/stock-analysis/Time v0.0.1
	github.com/spf13/viper v1.21.0
)

require (
	github.com/ThomasMarcelis/ibkr-go/v2 v2.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/grdvsng/stock-analysis/Data v0.0.1 // indirect
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1 // indirect
	github.com/grdvsng/stock-analysis/Regex v0.0.1 // indirect
	github.com/grdvsng/stock-analysis/Result v0.0.1 // indirect
	github.com/grdvsng/stock-analysis/Storage v0.0.1
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/shopspring/decimal v1.5.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/grdvsng/stock-analysis/Storage v0.0.1 => ../Storage
