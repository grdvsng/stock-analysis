module github.com/grdvsng/stock-analysis/Server

go 1.26

replace (
	github.com/grdvsng/stock-analysis/Config v0.0.1 => ../../lib/go/Config
	github.com/grdvsng/stock-analysis/Data v0.0.1 => ../../lib/go/Data
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1 => ../../lib/go/Exceptions
	github.com/grdvsng/stock-analysis/File v0.0.1 => ../../lib/go/File
	github.com/grdvsng/stock-analysis/Logger v0.0.1 => ../../lib/go/Logger
	github.com/grdvsng/stock-analysis/Regex v0.0.1 => ../../lib/go/Regex
	github.com/grdvsng/stock-analysis/Result v0.0.1 => ../../lib/go/Result
	github.com/grdvsng/stock-analysis/Sources v0.0.1 => ../../lib/go/Sources
	github.com/grdvsng/stock-analysis/Storage v0.0.1 => ../../lib/go/Storage
	github.com/grdvsng/stock-analysis/Time v0.0.1 => ../../lib/go/Time
)

require (
	github.com/coder/websocket v1.8.15
	github.com/grdvsng/stock-analysis/Config v0.0.1
	github.com/grdvsng/stock-analysis/Data v0.0.1
	github.com/grdvsng/stock-analysis/Exceptions v0.0.1
	github.com/grdvsng/stock-analysis/Logger v0.0.1
	github.com/grdvsng/stock-analysis/Result v0.0.1
	github.com/grdvsng/stock-analysis/Sources v0.0.1
	github.com/grdvsng/stock-analysis/Storage v0.0.1
	github.com/grdvsng/stock-analysis/Time v0.0.1
)

require (
	github.com/ThomasMarcelis/ibkr-go/v2 v2.1.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/grdvsng/stock-analysis/Control v0.0.1
	github.com/grdvsng/stock-analysis/Indicators v0.0.1
	github.com/grdvsng/stock-analysis/Process v0.0.1
	github.com/grdvsng/stock-analysis/Regex v0.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/shopspring/decimal v1.5.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/grdvsng/stock-analysis/Control v0.0.1 => ../../lib/go/Control

replace github.com/grdvsng/stock-analysis/Indicators v0.0.1 => ../../lib/go/Indicators

replace github.com/grdvsng/stock-analysis/Process v0.0.1 => ../../lib/go/Process
