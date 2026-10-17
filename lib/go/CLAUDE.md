# lib/go — shared libraries

Each directory is a separate Go module, `github.com/grdvsng/stock-analysis/<Name>`. Project-wide rules are in `/CLAUDE.md`.

## Dependency order (no cycles)

`Exceptions` → `Logger` → `Result` → `Regex` → `Time` → `Data` → `Storage` → `Sources` → `Config`. `Control`, `Indicators` and `Process` are standalone (`Control` needs `coder/websocket`). `File` is a small file helper.

## Process

`Process.Context()` returns the shutdown context for a service `main`. The context is cancelled on:

- SIGINT, SIGTERM or SIGHUP;
- the parent's death: Linux `PR_SET_PDEATHSIG` sends SIGTERM, and a watchdog also cancels when the parent pid changes (checked every second).

After cancellation the process force-exits within 15 s, so ports never stay held. Every service must use it instead of `signal.NotifyContext`.

## Exceptions / Logger / Result

- `Exceptions.CreateError(template)` and `CreateWarning(template)` return message templates. `Render(wantStack, args...)` formats them; `Error.Panic(args...)` panics.
- Logger levels: `DEBUG_LEVEL` … `CRITICAL_LEVEL`, `DISABLED_LEVEL`. In JSON they are written as `"DEBUG"`, `"INFO"`, `"WARNING"`, `"ERROR"`, `"CRITICAL"`, `"DISABLED"`.
- Global functions: `Logger.Configure(Options{Level, Formatter, FilePath})`, `Debug` / `Info(format, args...)`, `Warning(warning, args...)`, `Error` / `Critical` / `Exception(error, args...)`. `Exception` includes a stack trace. `Throw(error, args...)` logs at CRITICAL, closes the log file and panics.
- Formatters: `TextFormatter{Fields, TimeLayout, Separator}` and `JSONFormatter`.
- `Result.Ok(value, err)` and `Result.Check(err)` throw `RunTimeError` on error. Use them only at startup.

## Time

- `Duration`: text `"<n> <Hour|Minute|Second>[s]"`. `Duration()` returns a `time.Duration`; `Create(n, unit)`.
- `TTL`: text `"<n> <Day|Hour|Minute>[s]"`. `Duration()` returns a `time.Duration`.
- `Location`: a wrapper over `time.Location` that keeps the IANA `Name` and is (un)marshalled as text. `Parse(id)`, `Now()`, `ParseIn(format, raw)`.
- `Schedule`:
  - **Parsing:** `Parse(raw, timeZone)` reads the IBKR format `20261019:0400-20261019:2000;20261020:CLOSED`. `CLOSED` days and empty segments are skipped.
  - **Internal form:** `Start` and `End` are maps from the exchange-local start date (`YYYY-MM-DD`) to unix seconds.
  - **Queries:** `Available()` checks today's and yesterday's session against now in exchange time. `Actual()` is true while any session has not ended yet.
  - **JSON** is UTC sessions: `{"time_zone": "US/Eastern", "sessions": [{"start": "…Z", "end": "…Z"}]}`. The legacy map format `{"start", "end", "location"}` is still accepted when reading.

## Data/Market

- `Stock{Id, Type, MinTick}`.
- `Contract{Id (market name), Name, Exchange, Currency, Industry, Category, Subcategory, PrimaryExchange, Country, TradingHours, LiquidHours, Stock}` with `IsTradingHour()`, `IsLiquidHours()`, `IsActual()`.
- `TradeMetric{Contract, Time, Price, Size, Bid, Ask, BidSize, AskSize}`, all decimals.
- `Match{Symbol, Name, PrimaryExchange, Currency, Country}` is a search hit.
- `CountryOf(primaryExchange, isin)` returns the ISIN prefix if present, otherwise looks up the exchange table.

## Storage (QuestDB ORM)

- `Connect(host, port, timeout)` uses HTTP `/exec` (GET). `db.Exec(sql)` returns `Result{Columns, Dataset [][]json.RawMessage, Error}`.
- **Model:** a struct with `Table() string` and, optionally, `Partition() Partition` (`HOUR` / `DAY` / `MONTH` / `YEAR`; DAY by default).
- **Field tags:**
  - `db:"column[,designated][,dedup]"` — exactly one `designated` field, and it must be a TIMESTAMP;
  - `dbtype:"SYMBOL"` overrides the inferred type;
  - `json:"…"` is required and is what reading uses.
- **Inferred types:** `string` → VARCHAR, `time.Time` → TIMESTAMP (zero → NULL), `decimal.Decimal` → `DECIMAL(24,8)`, ints → LONG/INT/SHORT/BYTE, floats → DOUBLE/FLOAT, `bool` → BOOLEAN, `encoding.TextMarshaler` → VARCHAR holding the text. Any other struct, slice or map → VARCHAR holding JSON. A nil pointer → NULL.
- **Operations:**
  - `db.Create(model)` creates the table and migrates new columns;
  - `db.Insert(models...)` groups rows by table, 100 rows per statement;
  - `Storage.Select[T](db, sql)` reads rows through JSON into `T`;
  - `db.Exists(model)`, `db.Count(model)`, `db.Tables(prefix)`, `db.Drop(table)`;
  - `db.DropPartitionsBefore(model, t)` drops whole partitions strictly older than `t`'s partition.
- `Storage/Tables`: the `Stock`, `Contract`, `Metric` models; `NewStock`, `NewContract`, `NewMetric` (zero decimals become nil) and `.Market()` back-conversions; `MetricsTable(stockId)`; `Tables.Create(db)` creates `stocks` and `contracts`.

## Sources

- The `Source` interface:
  - `GetContract(id) *Market.Contract` — nil if not found or on error; errors are logged;
  - `Subscribe(contract, onPost, onError, updateIntervalSeconds)`, `UnSubscribe(contract)`;
  - `History(ctx, contract, days, onBatch) error`, `Search(pattern) ([]Market.Match, error)`;
  - `Close()`, `Run()`.
- `Subscription` carries the `OnPost(TradeMetric)` and `OnError(error)` handlers. A nil `OnError` logs. Subscriptions are keyed by `Stock.Id`.
- `IBK` is the Interactive Brokers implementation (`ThomasMarcelis/ibkr-go/v2`):
  - **Connecting:** `Connect(host, port, timeout, clientId)` returns an error; `New` throws. Each connection needs a unique client id. The connection context lives until `Close`; the timeout applies per request.
  - **Contracts:** qualified as `ibkr.Stock(symbol)`, i.e. **SMART / USD only**. Non-US listings found by search cannot be qualified yet.
  - **Subscription loop:** every interval, if the contract is in trading hours, it requests the last TRADES tick and the last BID_ASK tick (`NumberOfTicks = 1`, `EndTime = now`). It re-qualifies the contract when `IsActual()` is false.
  - **History:** walks back with 1-minute TRADES and BID_ASK bars, one trading day per request, until N trading days are collected. Price and size come from TRADES (close and volume), bid and ask from BID_ASK (open and close); there are no bid or ask sizes. ibkr-go handles IBKR pacing.
  - **Search:** `Contracts().Search`, keeping only stocks.

## Config (DataCollector config; Server reads and edits it)

- `Load(path)` reads the JSON through viper with `json` tags; `Validate(raw, near)` validates raw JSON through a temporary file.
- File operations: `Archive(path, now)` copies the file to `<name>.<2006-01-02_15-04-05><ext>`; `Replace(path, raw)` writes atomically; `Restore(path, archive)`.
- JSON fields and defaults:
  - `workers` (4);
  - `update_interval` (`"10 Seconds"`) — how often subscriptions are reloaded from the DB;
  - `source{type:"ibk", host, port, timeout:300, client_id:1}`;
  - `storage{host:"localhost", port:9000, timeout:30}`;
  - `logging{level:"INFO", path}`;
  - `ws{host:"127.0.0.1", port:9100}`.
- `source.Gateway(worker)` returns a source with client id `client_id + worker`.
- `Metric` is a subscription and the `metrics` table model. Fields:
  - `ID`;
  - `TTL`: retention, 30 days by default;
  - `UpdateInterval`: polling, 10 s by default;
  - `ContractUpdateInterval`: 24 h by default;
  - `Reprocess`: trading days of history to backfill for a stock new to the DB; it must not exceed the TTL;
  - `Enabled`, `CreatedAt`, `UpdatedAt`.
- Metric functions: `CreateMetrics(db)`, `LoadMetrics(db)`, `LoadMetric(db, id)`, `SaveMetric(db, m)` (keeps `CreatedAt`), `RemoveMetric(db, id)`. All of them tolerate a missing table.

## Control (Server ⇄ DataCollector protocol)

- WebSocket at `ws://<ws.host>:<ws.port>/control`. A message is `{"id", "type", "metric"?, "config"?, "ok", "error"?}`. Every request gets one reply with the same `id` and `type: "response"`.
- Types: `update_config` (with `config`), `metric_updated` and `metric_removed` (with `metric`).
- `Serve(handler)` is the server side; handlers run concurrently and writes are serialized.
- `NewClient(host, port)` with `Run(ctx, onError)` reconnects every 5 s. `Request(ctx, message)` returns `ErrDisconnected` when there is no connection.

## Indicators

- `Candle{Time, Open, High, Low, Close, Volume}` is float64. `Signal{Metric, Time, Action: "buy"|"sell", Value}`.
- `Catalog()` is ordered: 100 candle metrics plus 7 Fibonacci. Groups are `trend`, `momentum`, `volatility`, `volume`, `pattern`, `fibonacci`. `Lookup(key)` finds one definition.
- `Definition{Key, Group, Name, Description, Warmup}`. `Evaluate(candles, from)` returns the signals at candles with `time >= from`. `Warmup` is how many earlier candles it needs.
- **Rules:**
  - crosses: price vs average, fast vs slow average, line vs signal line, zero;
  - leaving overbought or oversold zones;
  - band reversal or breakout;
  - candle patterns, where trend context is the close vs the close 5 candles back;
  - Fibonacci: bounces off retracements of the last 55-candle swing, and the 161.8% extension of the previous swing as a profit target.
- **To add a metric:** append a `Definition` in `build()` in `Catalog.go`. Keys must be unique (checked at init), and the candle-metric count is meant to stay at 100.
