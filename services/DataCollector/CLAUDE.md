# DataCollector

Collects market data from IBKR into QuestDB for every enabled subscription in the `metrics` table. Project rules are in `/CLAUDE.md`, and the libraries it uses in `/lib/go/CLAUDE.md`.

## Run

```bash
go run . --config-path Examples/config.json [--ram-limit 0.5]
```

- The config file is described in `lib/go/Config`. Since subscriptions moved to the DB, it holds only infrastructure. `Examples/config.json` is the user's file: it may still contain a legacy `stocks` key (ignored) or stale values, so do not edit it without asking.
- Logs go to `<logging.path without ext>.<YYYY-MM-DD><ext>`; the directory is created if missing.

## Process structure (`main.go`)

- `main` loads the config once, starts the **control WebSocket server** on `ws.host:ws.port` (it lives for the whole process), and then loops `run()`.
- **Startup:** `run()` loads the config, configures the logger and calls `start()`:
  - connect to QuestDB;
  - `Tables.Create` and `Config.CreateMetrics`;
  - open `workers` IBK sources with client ids `client_id + i`;
  - build the `Collector`.
- After a successful start: `Sync` with the subscriptions from the DB, then `Cleanup`.
- **Loop events:**
  - `Process.Context()` cancelled (SIGINT, SIGTERM, SIGHUP or the parent's death): stop the process; a forced exit follows after 15 s if shutdown hangs. The IBK poll pause checks the stop flags every 250 ms, so subscriptions stop quickly;
  - local midnight: restart with a new log file, a fresh config and a cleanup;
  - every `update_interval`: reload subscriptions and `Sync`;
  - every minute: `Refresh`, which retries failed subscriptions and refreshes stale contracts;
  - control messages, see below.
- **Startup failures:** a failure without a pending config update is fatal (`Logger.Throw`). With a pending update, the archived config is restored, the error is returned to the caller, and the process restarts with the old config.
- A change to `ws.*` takes effect only after a full process restart.

## Collector (`Collector/`)

- **`Sync(metrics)`:** unsubscribes stocks that are gone or whose `update_interval` changed, and subscribes new ones. A stock whose contract was not found is not retried on every Sync, only in `Refresh` (once a minute).
- **On subscribe:**
  1. `GetContract`;
  2. save `stocks` and `contracts` snapshots;
  3. check whether the stock is **new** (no `metrics_<id>` table, or the table is empty);
  4. create the table;
  5. `Subscribe` on a round-robin source;
  6. if the stock is new and `reprocess > 0`, backfill `reprocess` trading days in the background (`History`) while live polling runs; dedup on `time` merges overlaps.
- **Live metrics:** `onPost` inserts `Tables.NewMetric`. Source errors are logged.
- `Resubscribe(id, metrics)` forces an unsubscribe and a new subscribe (signal `metric_updated`). `Remove(id, metrics)` unsubscribes and drops `metrics_<id>` (signal `metric_removed`).
- **`Cleanup`** runs at every start, so daily:
  - drops `metrics_<id>` partitions older than the subscription TTL;
  - drops `metrics_*` tables of stocks that are no longer subscribed;
  - writes fresh `stocks` / `contracts` snapshots for active stocks, then drops their partitions older than today.
- **`Close`** cancels backfills, waits for them, and unsubscribes everything.

## Control API (WebSocket, `ws://127.0.0.1:9100/control`)

The protocol is in `lib/go/Control`. Every request gets `{"id", "type": "response", "ok", "error"?}`.

| type | payload | behaviour |
|---|---|---|
| `update_config` | `config` (raw JSON) | Validate → archive the current file as `<name>.<YYYY-MM-DD_HH-MM-SS>.json` → write the new file → restart. Reply `ok` after a successful restart. If start fails, restore the archive, restart with it, and reply `ok: false` with the error. Invalid JSON or config gets `ok: false` without a restart. |
| `metric_updated` | `metric` (stock id) | Reload subscriptions, then unsubscribe and subscribe that stock again. |
| `metric_removed` | `metric` | Reload subscriptions, unsubscribe, and drop `metrics_<id>` (unless the stock is still enabled). |

Requests queue behind the run loop; while a restart is in progress they wait.

## Known limitations

- Live `size` is the size of the last trade, repeated on every poll, so summing it over a candle overstates volume. Backfilled minute bars have the real volume.
- A backfill interrupted by a restart is not resumed: the stock is no longer "new".
- Only SMART/USD stocks can be subscribed (see IBK in `lib/go/CLAUDE.md`).
