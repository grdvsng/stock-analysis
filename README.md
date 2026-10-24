# stock-analysis

A self-hosted tool for short-term stock trading research. It collects market data from Interactive Brokers into QuestDB, computes technical signals, and shows them in a web UI.

> [!WARNING]
> **Built together with an AI assistant (Claude).** A large part of the code, the indicator logic and the forecasts were written with Claude and checked only partly. Expect bugs.
>
> - **Not financial advice.** Signals, forecasts and buy/sell levels come from technical indicators. They can be wrong, late or based on incomplete data.
> - **Verify before you act.** Check prices and levels in your broker before placing any order. Never trade on this tool's output alone.
> - **Use at your own risk.** The software is provided "as is", without warranty (MIT License). Test it with paper trading first.

## What it does

- **Data collection** — polls IBKR for the latest trade and quote of every subscribed stock, stores contracts and trading hours, refills the part of today missed during downtime, and cleans up data older than each subscription's TTL every day.
- **History backfill** — a separate service loads the requested past trading days and the past days missed during downtime. When IBKR has less history than requested, it lowers the subscription's `reprocess` to the days available (0 when none). It also tracks how many of the requested trading days are stored. The contract card shows the progress and the stored days, a popup announces completion, and stocks with less history in IBKR than requested are marked as "completed with a limit".
- **Subscriptions** — find a stock through IBKR search, subscribe or unsubscribe, and tune the polling interval, retention and history backfill.
- **Monitors** — live charts that compare stocks as lines (% change), as classic candlesticks, or both. Each chart has a contracts panel and marks stocks as "sleeping" when the exchange is closed.
- **Notifications** — 100 candle-based indicators (trend, momentum, volatility, volume, candle patterns) and 7 Fibonacci signals. Each one emits buy/sell events with a configurable lifetime. You can see the active events at any point of the chart and browse the full event history.
- **Alerts** — per-monitor rules such as "all of these metrics say buy for this stock". A match rings a loud bell and shows a popup that stays until you close it.
- **Forecast** — for a chosen contract and candle interval: a buy / hold / sell recommendation with entry, stop and target prices, supports and resistances, and the reasons behind it. It warns you when the company is not trading.
- **Configuration** — edit the collector config from the UI; a config that fails is rolled back automatically.
- **Languages** — English, Russian, German, French, Italian, Polish, Greek, Chinese (Simplified) and Japanese.

## Parts

| Path | What it is |
|---|---|
| `services/DataCollector` | Go service: IBKR → QuestDB, with a local control WebSocket |
| `services/Server` | Go service: REST + WebSocket API, monitors, forecasts; serves the UI |
| `services/UI` | React + Vite web app |
| `services/ReprocessWatcher` | Go service: loads past days of history and tracks backfill progress |
| `services/Builder` | Python builder: compiles everything into `build/` |
| `lib/python` | shared Python package: the runner and its helpers |
| `lib/go` | shared Go libraries (storage, indicators, config, sources…) |

## Requirements

Go 1.26, Node.js 22 with npm, a running QuestDB (HTTP API on port 9000), and IB Gateway or TWS with market data permissions.

## Quick start

```bash
python3 -m services.Builder.build
build/run.py
```

The builder compiles every service and the UI into `build/`, prepares configs in `build/config/` and a Python environment for the runner. `build/run.py` starts the DataCollector, the Server and the ReprocessWatcher one after another, shows their output in one console, prints a clickable link to the interface, and stops all of them on Ctrl+C, when any of them exits, or when the runner itself is killed. Python 3 with pip is needed for the builder. Developer notes and the full API are in the `CLAUDE.md` files.

## License

MIT, see [LICENSE](LICENSE).
