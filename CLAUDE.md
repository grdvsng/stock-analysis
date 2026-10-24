# stock-analysis

Personal stock market data platform: collects market data from Interactive Brokers into QuestDB, computes trading signals, and shows them in a web UI.

Per-area docs (read the one for the area you touch):

- `lib/go/CLAUDE.md` — shared Go libraries (ORM, config, sources, indicators, logger…)
- `services/DataCollector/CLAUDE.md` — collector service and its control socket
- `services/Server/CLAUDE.md` — REST + WebSocket API, monitors
- `services/UI/CLAUDE.md` — React UI
- `services/ReprocessWatcher/CLAUDE.md` — history backfill (past days) and its progress
- `services/Builder/CLAUDE.md` and `lib/python/CLAUDE.md` — Python builder, runner and shared Python package

## Data flow

```
IB Gateway ──► DataCollector ──► QuestDB ◄── Server ◄── UI (browser)
                  ▲   control WS (127.0.0.1:9100)  │
                  └────────────────────────────────┘
```

- DataCollector writes market data for the current day (`stocks`, `contracts`, `metrics_<stock>`). It reads subscriptions from the `metrics` table.
- ReprocessWatcher loads past trading days into `metrics_<stock>`, keeps the backfill bookkeeping (`history_heads`, `history_days`), lowers `metrics.reprocess` to what IBKR can provide, and writes backfill progress (`reprocesses`).
- Server writes subscriptions (`metrics`) and monitors (`monitors`), and reads everything. It notifies DataCollector through the control WebSocket.
- UI talks only to Server.

## Environment

- WSL2 Ubuntu on Windows. Go 1.26 (some lib modules declare `go 1.25.0`). Node 22 with npm in `/usr/bin`; the Windows npm under `/mnt/c/Program Files/nodejs` may also be on PATH, so do not use it.
- Run `go vet` / `go test` with `CGO_ENABLED=0`: cgo fails here because libc headers are missing, so `-race` is unavailable.
- QuestDB is run **by the user**: SQL over HTTP at `localhost:9000/exec`, GET only (POST returns 405). Port 9003 is only the health check (`Status: Healthy`), not the API.
- IB Gateway: `172.31.48.1:4001` (the Windows host, as seen from WSL).
- To look at the UI, use the Windows Chrome (`/mnt/c/Program Files/Google/Chrome/Application/chrome.exe`) in headless mode with a temporary `--user-data-dir`. WSL cannot reach its debugging port, so drive it over CDP from the Windows `node.exe`. Plain `--screenshot` fires before WebSocket data arrives.

## Hard rules

- **No comments in code.** Explain decisions in replies, not in source.
- **Never start your own QuestDB.** Use the user's instance. If it is unreachable, ask.
- **No destructive or polluting SQL on the user's tables** (`stocks`, `contracts`, `metrics`, `metrics_*`, `monitors`) without asking. Reads are fine. To test writes, use tables named `tmp_test_*` and drop them afterwards.
- Verification so far is done with temporary `tmp_test.go` files that are deleted afterwards. There are no committed tests. Ask before adding permanent ones.
- Do not commit or push unless asked. `lib/` was ignored by an old `.gitignore`; that is fixed now.
- Do not edit the user's spec files (`services/UI/monitor.md`, example configs the user changed) without asking. Point out the needed change instead.
- The user writes in Russian; reply in Russian. **Everything in the project is English**: code, identifiers, log and error messages, API texts, configs, docs and readmes. The only Russian text allowed is the UI dictionary `services/UI/src/i18n/ru.js` (and the user's own spec files).
- In shell removals, never use `rm -rf $VAR/...`. Use a literal path or `"${VAR:?}"`.
- **No orphan processes or open ports.** Every service process gets its shutdown context from `Process.Context()` (`lib/go/Process`). When you stop any HTTP/WS server you started (a preview binary, `go run`, headless Chrome), also stop all of its children, then confirm the ports are free with `ss -ltnp` (WSL) or `Get-NetTCPConnection` (Windows). `go run` leaves the built binary holding the port if only the parent dies. Find processes by exact pid or name: `pgrep -f '<pattern>'` also matches your own shell. Never kill the user's running services.

## Go conventions

- Directory, package and module names are PascalCase (`Storage`, `Tables`, `DataCollector`). Every `lib/go/<Name>` is its own module, `github.com/grdvsng/stock-analysis/<Name>`, wired through `replace … => ../<Name>` (services use `../../lib/go/<Name>`) with version `v0.0.1`. A new dependency on a lib needs both `replace` and `require` in every module whose graph contains it; then run `go mod tidy`.
- Exported constants are `UPPER_SNAKE` (`METRICS_TABLE`, `DEFAULT_INTERVAL`).
- The error variable is `e`; receivers are full words (`collector`, `logger`, `db`).
- Style: gofmt, plus a blank line between a declaration or statement group and the following `if`/`for`/`return`. Early returns.
- JSON tags are snake_case. Optional values are pointers (nil means JSON `null`, and SQL `NULL` in the ORM).
- Money and sizes use `shopspring/decimal`. Indicators use `float64`.
- Times are stored and transported in UTC and converted only for display.
- Errors:
  - **Message templates:** `Exceptions.CreateError("… %s")` / `CreateWarning`, kept in a package-level `Error.go`.
  - **Logging:** `Logger.Error(template, args...)` / `Logger.Warning(...)`.
  - **Fatal:** `Logger.Throw` and `Result.Ok` / `Result.Check` log and panic. Use them only at startup. Runtime failures (a request, a tick, a subscription) are logged or returned, never thrown.
- Use the global logger (`Logger.Info`, `Logger.Error`…), configured once in `main`.

## QuestDB rules (all storage goes through `lib/go/Storage`)

- **No `DELETE`.** Mutable entities (`metrics` subscriptions, `monitors`) are **journal tables**: every change appends a row with `enabled` and `updated_at`. The current state is `SELECT * FROM (SELECT * FROM t LATEST ON updated_at PARTITION BY id) WHERE enabled`. Removing an entity appends a row with `enabled = false`.
- **Retention:** `ALTER TABLE … DROP PARTITION WHERE ts < timestamp_floor(unit, X)`, which only drops whole partitions. Dropping a whole table is used for unsubscribed stocks.
- All tables are WAL with a designated timestamp. Use `DEDUP UPSERT KEYS` (the designated column plus `dedup` fields) for idempotent inserts. WAL applies asynchronously, so reads right after writes can lag by about a second.
- **Schema migration is automatic:** `db.Create(model)` runs `CREATE TABLE IF NOT EXISTS` and then `ALTER TABLE … ADD COLUMN IF NOT EXISTS` for new struct fields. Never rename or drop columns this way.
- Values are inlined as escaped literals (`Storage.Text`, `Storage.Timestamp`); the HTTP API has no parameters. Decimals are written as quoted strings into `DECIMAL(24,8)`.

## Tables

| Table | Writer | Columns (designated ts last) | Notes |
|---|---|---|---|
| `stocks` | DataCollector | id, type, min_tick, updated_at | snapshots, dedup (updated_at, id); only the current day kept |
| `contracts` | DataCollector | id (market name), stock_id, name, exchange, currency, industry, category, sub_category, primary_exchange, country, time_zone, trading_hours, liquid_hours, updated_at | snapshots, dedup (updated_at, stock_id); hours are JSON `{"time_zone","sessions":[{"start","end"}]}` in UTC |
| `metrics_<stock>` | DataCollector | stock_id, contract_id, time, price, size, bid, ask, bid_size, ask_size, collected_at | one table per stock, name `metrics_` + lowercased id with non-alphanumerics replaced by `_`; dedup (time); TTL per subscription |
| `metrics` | Server (ReprocessWatcher lowers `reprocess`) | id, ttl, update_interval, contract_update_interval, reprocess, enabled, created_at, updated_at | subscriptions journal; durations stored as text (`"30 Day"`, `"10 Second"`) |
| `reprocesses` | ReprocessWatcher | stock_id, required, collected, percent, state (`running` / `completed` / `limited`), started_at, finished_at, available_from, active, updated_at | journal; a row is written only when the percent step, the state or the requirement changes; `active = false` once the stock is no longer subscribed with `reprocess > 0` |
| `history_heads` | ReprocessWatcher | stock_id, head, updated_at | earliest TRADES data IBKR has for the stock (`HeadTimestamp`) |
| `history_days` | ReprocessWatcher | stock_id, date, rows, updated_at | a past trading day fully loaded from IBKR bars; only counts while that day still has data; rows 0 marks a holiday |
| `monitors` | Server | id, name, stocks (JSON), notifications (JSON), interval, event_lifetime, alerts (JSON), enabled, created_at, updated_at | monitors journal; event_lifetime is duration text or NULL (= one candle); alerts is `[{"id","name","stock","metrics","condition","enabled"}]` |

## Build and run

```bash
python3 -m services.Builder.build [--skip-ui] [--service NAME] [--out DIR]
build/run.py [--config-dir DIR]
```

- **Builder** (`services/Builder`, run from the project root as a module) builds the three Go services and the UI into `build/`. It creates any missing config from the examples, copies `lib/python/stock_analysis` to `build/lib/`, creates `build/.venv` from `lib/python/requirements.txt`, and writes `build/run.py` with that venv's interpreter in the shebang.
- **`build/run.py`** (`stock_analysis.runner`) starts DataCollector → Server → ReprocessWatcher one by one, each with its own config: `config/dc.json`, `config/server.json`, `config/reprocess.json`. It waits for the control port, then for `GET /settings`, and forwards every service's output with a coloured name prefix.
  - It refuses to start if a configured port is busy, and names the owner.
  - It sets `PROCESS_BIND_PARENT=1`, so the services die even if the runner is SIGKILLed.
  - On Ctrl+C, SIGTERM, SIGHUP or the exit of any service, it terminates every service and its children with psutil (kill after 20 s), then checks the ports.
- Stop the running services before rebuilding into the same folder.

## Python conventions

- Strict typing everywhere, including `-> None`; `lib/python/.venv/bin/mypy --strict --explicit-package-bases lib/python/stock_analysis services/Builder/build.py` must pass.
- Dependencies live in `lib/python/.venv`. After any change, freeze: `lib/python/.venv/bin/python -m pip freeze --exclude pip > lib/python/requirements.txt`. Python 3.14 here has no ensurepip, so venvs are created with `--without-pip` and bootstrapped with the system pip (`python3 -m pip --python <venv>/bin/python install pip`).
- No `sys.path` manipulation. Modules are run from the project root (`python3 -m services.Builder.build`) and import `lib.python.stock_analysis`; inside `build/` the runner imports `lib.stock_analysis`. All paths use `pathlib.Path`.
- Reusable code goes into submodules of `lib/python/stock_analysis`.

## Verify a change

```bash
cd lib/go && for m in */; do (cd $m && CGO_ENABLED=0 go vet ./...) ; done
cd services/DataCollector && CGO_ENABLED=0 go vet ./...
cd services/Server && CGO_ENABLED=0 go vet ./...
cd services/ReprocessWatcher && CGO_ENABLED=0 go vet ./...
cd services/UI && npm run build
lib/python/.venv/bin/mypy --strict --explicit-package-bases lib/python/stock_analysis services/Builder/build.py
```
