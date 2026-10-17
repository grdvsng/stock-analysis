# UI

React 19 + Vite 7 single-page app. It talks only to Server (`services/Server/CLAUDE.md`). Project rules are in `/CLAUDE.md`. The user's product spec for monitors is `monitor.md` (do not edit it without asking).

## Run

```bash
npm install
npm run build   # → dist/, served by Server via ui_path on http_port
npm run dev     # Vite dev server; REST is proxied to SERVER_URL (default http://localhost:8080)
```

- `dist/` and `node_modules/` are git-ignored.
- WebSockets are not proxied: the UI asks `GET /settings` for `ws_port` and connects to `ws(s)://<page hostname>:<ws_port>`.
- To check rendering, take screenshots (see Environment in `/CLAUDE.md`). Build success alone does not prove the UI is fine.

## Structure (`src/`)

- `main.jsx`, `App.jsx` — tabs in order: Monitors (default), Contracts, Search, Configuration; plus a language switch.
- `api.js` — `request(method, path, body)` throws `Error(body.error)` on non-2xx; the `api.*` calls; `monitorSocketUrl(port, params)`.
- `i18n/` — `I18nProvider`, `useI18n()` → `{t, language, setLanguage}`. Dictionaries are `en.js` (default and fallback) and `ru.js`. `t(key, {vars})` fills `{name}` placeholders. Every visible string goes through `t`; to add a language, create a file with the same keys and register it in `dictionaries`. Indicator names and descriptions come from the server in English.
- `config/schema.js` — Configuration tab fields (path, type, units). Unknown config keys are preserved on save.
- `components/` — `ConfigTab`, `ContractsTab`, `SearchTab`, `ContractDetails` (contract card; schedules shown in the **user's** time zone with its name), `SubscriptionForm`, `DurationInput` (text `"<n> <Unit>s"`), `Field`, `Modal`.
- `components/monitor/` — `MonitorsTab` (list, create, delete; the selected id is kept in localStorage), `MonitorForm`, `MonitorView`, `MonitorChart`, `ContractPanel`, `NotificationPicker`, `intervals.js`.
- `monitor/` — `useMonitorStream` (WebSocket lifecycle and message merge), `palette.js`, `useDarkMode`.

## Behaviour rules

- **Configuration:** saving shows "waiting for the data collector to restart", then success or the server's error. Invalid configs are rolled back by the collector.
- **Subscriptions:** unsubscribing asks for confirmation, because the collector drops that stock's data.
- **Monitor stream lifecycle:** one WebSocket per parameter set. Any change (stocks, interval, notifications, from or to) closes the socket and opens a new one, and unmount closes it. Handlers are detached before `close()`, so late messages are ignored. Merge rule: a snapshot replaces everything; an update replaces candles and signals with `time >= since` for the stocks it contains.
- **Monitor chart** (`lightweight-charts` v5, with the attribution logo on). It has two independent layers, Lines and Candles (at least one is on; the choice is stored in localStorage `chartLayers`), and the chart remounts when the layers change:
  - candles are drawn for the focused stock: rising candles hollow green, falling candles filled red, wicks in the same colour. With Lines on they are scaled to % of the same base as that stock's line (one axis); that line is drawn thin and the candle series has no own last-value label. With Candles only they are in price. The tooltip always shows the real O/H/L/C, the direction ▲/▼ and the change from the open;
  - one line per stock as % change from the first candle in range, on a single axis;
  - at most 8 lines (`MAX_LINES`);
  - the colour slot is pinned to the stock (`assignSlots`), so survivors are not repainted;
  - crosshair tooltip lists every visible series (value first, then name and price);
  - legend above the chart;
  - signal markers only for the metric chosen in the panel's hits table, or for "show all"; buy is ▲ green and sell is ▼ red;
  - the hover tooltip also lists, per stock, the events **active at the hovered moment** (the latest signal per metric with `time <= t < expires`), at most 6 plus "+N more";
  - live mode: data is re-set on every update; when the view is at the right edge (`scrollPosition() >= -2`) it scrolls to real time, otherwise the user's scroll is kept.
- **Event history** (`EventsLog` under the chart): every signal in the range, newest first, 100 per page. It filters by stock, metric and action. The status column shows "active until" (the latest signal of its metric, not yet expired), "expired" or "superseded". Clicking a row focuses that stock and metric markers and zooms the time scale to about ±40 candles around the event.
- **Panel:** a non-trading contract is greyed and labelled "Sleeping". `N/M active` counts metrics whose latest signal is active at the reference time: now when live, otherwise the range end. The hits table shows the total hits, the active signal with "until" and the last signal. Contracts, notifications and the event lifetime can be edited live (`PUT /monitor`, which triggers a reconnect); the lifetime input is debounced by 800 ms.
- **Colours** follow the dataviz reference palette: categorical series in a fixed order, separate light and dark steps; status good `#0ca30c` and critical `#d03b3b` only for buy and sell, always with an arrow and a label. Chart chrome comes from `CHROME`.
- **Styling:** plain CSS in `styles.css` with tokens on `:root` and a `prefers-color-scheme: dark` override. No CSS framework.
