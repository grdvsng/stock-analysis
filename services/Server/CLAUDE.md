# Server

The REST and WebSocket API used by the UI. It manages subscriptions and monitors, edits the DataCollector config through the collector, and computes monitor notifications. Project rules are in `/CLAUDE.md`, libraries in `/lib/go/CLAUDE.md`.

## Run

```bash
go run . --config-path Examples/config.json
```

Server config (`Settings/Settings.go`):

| field | default | meaning |
|---|---|---|
| `http_port` | 8080 | REST API and the UI static files |
| `ws_port` | 8081 | WebSocket streams |
| `data_collector_config` | required | path to the DataCollector config, relative to this file; gives QuestDB, IBKR and the collector control socket |
| `stream_interval` | `"2 Seconds"` | stream tick |
| `update_timeout` | `"10 Minutes"` | how long `POST /config` waits for the collector restart |
| `ui_path` | — | built UI (`../../UI/dist`), served at `GET /` with SPA fallback |
| `logging` | INFO, stdout | `{level, path}` |

- **Startup:** `Tables.Create`, `Config.CreateMetrics` and `Monitors.Create` (creates tables and migrates columns). Then the server connects to the collector control socket and reconnects every 5 s.
- **Shutdown:** on `Process.Context()` cancellation (signals or the parent's death) both HTTP servers shut down within 10 s, with a forced exit at 15 s.
- **IBKR lookups** (`GET /contract` fallback, search) use one lazily opened source with client id `client_id + workers`, so it does not collide with the collector. It is reset after a config update.

## Layout

- `Api/Api.go` — router, helpers `reply` / `fail` (errors as `{"error": "..."}`; status ≥ 500 is logged), lazy source, `signal()` to the collector.
- `Api/Config.go`, `Api/Contract.go`, `Api/Contracts.go`, `Api/Monitors.go` — REST handlers.
- `Api/Stream.go` (`/contracts` stream, origin check) and `Api/MonitorStream.go` (`/monitor`).
- `Api/Ui.go` — static UI.
- `Monitors/` — the `Monitor` model (table `monitors`) and candle loading (`SAMPLE BY`).

## REST API (`http_port`)

| method & path | request | response |
|---|---|---|
| `GET /config` | — | the DataCollector config file as-is (JSON) |
| `POST /config` | full config JSON | validated, then sent to the collector as `update_config`, waiting up to `update_timeout`. `200` + the config when applied; `400` when invalid or the collector failed with it (it already restored the old one); `503` when the collector is not connected; `504` on timeout |
| `GET /contract?id=SYM` | — | `{"contract": Market.Contract, "metric": Metric\|null}`. The contract comes from the DB; if absent it is qualified in IBKR. `404` not found, `502` IBKR unavailable |
| `POST /contract?id=SYM` | `{"ttl"?, "update_interval"?, "contract_update_interval"?, "reprocess"?}` | creates or updates the subscription (keeps `created_at`), signals `metric_updated`. `201` + the saved metric; `400` when invalid (for example `reprocess` > `ttl`, or a bad duration text) |
| `DELETE /contract?id=SYM` | — | appends a disabled row and signals `metric_removed` (the collector drops the data). `204`; `404` if not subscribed |
| `GET /contracts` | — | `[{"metric", "contract"\|null}]` for enabled subscriptions; `contract` is null until the collector has fetched it |
| `GET /contracts/history?from=&to=` | `from` required; `to` defaults to now; RFC3339 or `YYYY-MM-DD` | `{"from", "to", "metrics": {"SYM": [Tables.Metric]}}` across all `metrics_*` tables |
| `GET /contracts/search?query=` | text | `[Market.Match]` (stocks only) from IBKR; `502` when IBKR is unavailable |
| `GET /settings` | — | `{"ws_port"}` |
| `GET /indicators` | — | `[{"key", "group", "name", "description", "warmup"}]` (107 entries) |
| `GET /monitors` | — | `[Monitor]` ordered by `created_at` |
| `POST /monitors` | `{"name", "stocks", "notifications", "interval", "event_lifetime"?}` | `201` + Monitor. Validation: the name is non-empty; `interval` is one of `1m 5m 15m 30m 1h 1d` (default `1m`); every notification key is in the catalog; every stock is an enabled subscription; `event_lifetime` is a positive duration text (`"3 Minutes"`) or null, meaning one candle |
| `PUT /monitor?id=` | same fields (omitted fields keep their values) | `200` + Monitor; `404` |
| `DELETE /monitor?id=` | — | `204`; `404` |

Shapes:

- `Metric` (subscription): `{"id", "ttl", "update_interval", "contract_update_interval", "reprocess", "enabled", "created_at", "updated_at"}`.
- `Monitor`: `{"id", "name", "stocks", "notifications", "interval", "event_lifetime", "enabled", "created_at", "updated_at"}`.
- `Market.Contract` schedules are `{"time_zone", "sessions": [{"start", "end"}]}` in UTC.

## WebSocket API (`ws_port`)

Connections are accepted only from pages of the same host (any port).

### `GET /contracts`

Every `stream_interval` the server sends `{"metrics": {"SYM": [Tables.Metric]}}` with the rows whose `time` is newer than the last one sent for that table (rows since server start at first). Nothing is sent when there is nothing new.

### `GET /monitor?stocks=A,B&interval=5m&notifications=k1,k2&lifetime=3%20Minutes&from=&to=`

- **Parameters:** `from` defaults to the start of the current UTC day; `to` is optional, and without it the stream is live. At most 50 stocks. `lifetime` is the event lifetime as duration text; it defaults to one candle of `interval`.
- **Signals** are `{"metric", "time", "action", "value", "expires"}`, where `expires = time + lifetime`. An event is active while `time <= t < expires`; after that it has expired. Only the latest signal of a metric counts: a newer one supersedes an older one even if the older one has not expired. The client applies that rule, and the server sends every signal.
- **First message:** `{"type": "snapshot", "stocks": {"SYM": {"since", "trading", "candles": [Candle], "signals": [Signal]}}}`.
  - Candles cover `from`–`to`; a warmup of up to 1000 earlier candles is loaded but not sent.
  - Signals are computed for each requested notification over warmup plus range, and only those at or after `from` are sent.
- **Updates:** while live, every tick `{"type": "update", "stocks": {...}}` is sent only for stocks whose tail candles or `trading` flag changed. The client replaces its candles and signals with `time >= since`, because the last candle is still forming.
- **Other fields and messages:**
  - `trading` comes from the contract schedule and is refreshed every minute;
  - bad parameters → `{"type": "error", "error"}` and the socket closes;
  - when `to` is in the past, no updates follow the snapshot.
- New parameters mean a new connection.
- Candles are built from `metrics_<sym>` with `SAMPLE BY <interval> ALIGN TO CALENDAR` (first, max, min and last of price; sum of size). Rows without a price are ignored.
