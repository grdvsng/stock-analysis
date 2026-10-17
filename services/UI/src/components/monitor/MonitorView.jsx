import { useEffect, useMemo, useRef, useState } from "react";
import { api } from "../../api.js";
import { useI18n } from "../../i18n/I18n.jsx";
import useMonitorStream from "../../monitor/useMonitorStream.js";
import useDarkMode from "../../monitor/useDarkMode.js";
import { MAX_LINES, SERIES, assignSlots } from "../../monitor/palette.js";
import ContractDetails from "../ContractDetails.jsx";
import Modal from "../Modal.jsx";
import MonitorChart from "./MonitorChart.jsx";
import ContractPanel from "./ContractPanel.jsx";
import NotificationPicker from "./NotificationPicker.jsx";
import EventsLog from "./EventsLog.jsx";
import AlertsManager from "./AlertsManager.jsx";
import { useAlerts } from "../../alerts/Alerts.jsx";
import { evaluateAlert } from "../../alerts/evaluate.js";
import { INTERVALS, LIFETIME_UNITS, RANGES } from "./intervals.js";
import DurationInput from "../DurationInput.jsx";

const LIFETIME_DEBOUNCE = 800;
const LAYERS = ["lines", "candles"];
const LAYERS_KEY = "chartLayers";

function rememberedLayers() {
  try {
    const stored = JSON.parse(localStorage.getItem(LAYERS_KEY));

    if (Array.isArray(stored) && stored.length && stored.every((layer) => LAYERS.includes(layer))) {
      return stored;
    }
  } catch {
    return ["lines"];
  }

  return ["lines"];
}

function rememberLayers(layers) {
  try {
    localStorage.setItem(LAYERS_KEY, JSON.stringify(layers));
  } catch {
    return;
  }
}

function startOfRange(days) {
  const date = new Date();

  date.setHours(0, 0, 0, 0);
  date.setDate(date.getDate() - days + 1);

  return date.toISOString();
}

function toLocalInput(iso) {
  if (!iso) {
    return "";
  }

  const date = new Date(iso);

  return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
}

function fromLocalInput(value) {
  return value ? new Date(value).toISOString() : "";
}

function Details({ stock, onClose }) {
  const { t } = useI18n();
  const [state, setState] = useState({ loading: true });

  useEffect(() => {
    const controller = { active: true };

    api
      .getContract(stock)
      .then((data) => controller.active && setState({ data }))
      .catch((error) => controller.active && setState({ error: error.message }));

    return () => {
      controller.active = false;
    };
  }, [stock]);

  return (
    <Modal title={stock} onClose={onClose}>
      {state.loading && <p className="notice">{t("common.loading")}</p>}
      {state.error && <p className="notice error">{state.error}</p>}
      {state.data && <ContractDetails contract={state.data.contract} />}
    </Modal>
  );
}

export default function MonitorView({ monitor, contracts, catalog, wsPort, onChange }) {
  const { t } = useI18n();
  const dark = useDarkMode();
  const [range, setRange] = useState({ preset: "5d", from: startOfRange(5), to: "" });
  const [visible, setVisible] = useState(() => monitor.stocks.slice(0, MAX_LINES));
  const [focused, setFocused] = useState(monitor.stocks[0] || null);
  const [highlight, setHighlight] = useState({ stock: null, metric: null });
  const [details, setDetails] = useState(null);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(monitor.notifications);
  const [error, setError] = useState(null);
  const slotsRef = useRef({});
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000);

    return () => clearInterval(timer);
  }, []);

  const [lifetime, setLifetime] = useState(monitor.event_lifetime || null);

  useEffect(() => {
    if ((lifetime || null) === (monitor.event_lifetime || null)) {
      return undefined;
    }

    const timer = setTimeout(() => update({ event_lifetime: lifetime }).catch(() => undefined), LIFETIME_DEBOUNCE);

    return () => clearTimeout(timer);
  }, [lifetime]);

  const { fire } = useAlerts();
  const [alertsOpen, setAlertsOpen] = useState(false);
  const alerts = monitor.alerts || [];
  const live = !range.to;

  const names = useMemo(() => new Map(catalog.map((item) => [item.key, item.name])), [catalog]);

  useEffect(() => {
    setVisible((current) => {
      const kept = current.filter((stock) => monitor.stocks.includes(stock));
      const added = monitor.stocks.filter((stock) => !current.includes(stock));

      return kept.concat(added).slice(0, MAX_LINES);
    });

    if (!monitor.stocks.includes(focused)) {
      setFocused(monitor.stocks[0] || null);
    }
  }, [monitor.stocks]);

  const { data, status } = useMonitorStream({
    port: wsPort,
    stocks: monitor.stocks,
    interval: monitor.interval,
    notifications: monitor.notifications,
    lifetime: monitor.event_lifetime,
    from: range.from,
    to: range.to,
  });

  useEffect(() => {
    if (!live || status.state !== "live") {
      return;
    }

    for (const alert of alerts) {
      if (!alert.enabled) {
        continue;
      }

      const result = evaluateAlert(alert, data[alert.stock]?.signals || [], now);

      if (result) {
        const contract = contracts.find((item) => item.metric.id === alert.stock)?.contract;

        fire(`${monitor.id}:${alert.id}`, result.episode, {
          alert: alert.name,
          stock: alert.stock,
          contract: contract?.name || "",
          monitor: monitor.name,
          signals: result.signals.map((signal) => ({ metric: signal.metric, action: signal.action, name: names.get(signal.metric) || signal.metric })),
        });
      }
    }
  }, [now, data, alerts, live, status.state]);

  const slots = useMemo(() => {
    slotsRef.current = assignSlots(slotsRef.current, visible);

    return slotsRef.current;
  }, [visible]);

  const colors = dark ? SERIES.dark : SERIES.light;

  const [layers, setLayers] = useState(rememberedLayers);
  const [focusTime, setFocusTime] = useState(null);
  const showLines = layers.includes("lines");
  const showCandles = layers.includes("candles");

  const toggleLayer = (layer) => {
    const next = layers.includes(layer) ? layers.filter((item) => item !== layer) : LAYERS.filter((item) => item === layer || layers.includes(item));

    if (next.length) {
      setLayers(next);
      rememberLayers(next);
    }
  };

  const lines = useMemo(
    () =>
      visible.map((stock) => {
        const candles = data[stock]?.candles || [];
        const base = candles.length ? candles[0].close : 1;
        const prices = new Map();
        const points = candles.map((candle) => {
          const time = Math.floor(candle.time / 1000);

          prices.set(time, candle.close);

          return { time, value: (candle.close / base - 1) * 100 };
        });

        return { id: stock, type: "line", label: stock, color: colors[slots[stock]], points, prices, signals: data[stock]?.signals || [] };
      }),
    [visible, data, slots, colors],
  );

  const candleStock = monitor.stocks.includes(focused) ? focused : monitor.stocks[0];

  const series = useMemo(() => {
    const result = showLines ? lines.map((line) => (showCandles && line.id === candleStock ? { ...line, thin: true } : line)) : [];

    if (showCandles && candleStock) {
      const candles = data[candleStock]?.candles || [];
      const base = showLines && candles.length ? candles[0].close : null;
      const scale = (value) => (base ? (value / base - 1) * 100 : value);
      const raw = new Map();

      result.push({
        id: `candle:${candleStock}`,
        stock: candleStock,
        type: "candle",
        format: base ? "percent" : "price",
        label: showLines ? t("chart.candlesOf", { stock: candleStock }) : candleStock,
        overlay: Boolean(base),
        points: candles.map((candle) => {
          const time = Math.floor(candle.time / 1000);

          raw.set(time, candle);

          return { time, open: scale(candle.open), high: scale(candle.high), low: scale(candle.low), close: scale(candle.close) };
        }),
        raw,
        prices: new Map(),
        signals: data[candleStock]?.signals || [],
      });
    }

    return result;
  }, [showLines, showCandles, lines, data, candleStock, t]);

  const markers = useMemo(() => {
    const target = highlight.stock && series.find((item) => item.type === "candle" && item.stock === highlight.stock) || series.find((item) => item.type === "line" && item.id === highlight.stock);
    const items = target
      ? (data[highlight.stock]?.signals || []).filter((signal) => highlight.metric === "all" || signal.metric === highlight.metric).sort((a, b) => a.time - b.time)
      : [];

    return { series: target ? target.id : null, items };
  }, [data, highlight, series]);

  const allSignals = useMemo(
    () => monitor.stocks.flatMap((stock) => (data[stock]?.signals || []).map((signal) => ({ ...signal, stock }))).sort((a, b) => b.time - a.time),
    [data, monitor.stocks],
  );

  const openEvent = (event) => {
    setHighlight({ stock: event.stock, metric: event.metric });
    setFocused(event.stock);
    setVisible((current) => (current.includes(event.stock) || current.length >= MAX_LINES ? current : [...current, event.stock]));
    setFocusTime({ time: event.time, at: Date.now() });
  };

  const update = (changes) => {
    setError(null);

    return api
      .updateMonitor(monitor.id, { ...monitor, ...changes })
      .then(onChange)
      .catch((reason) => {
        setError(reason.message);
        throw reason;
      });
  };

  const choosePreset = (preset) => {
    const option = RANGES.find((item) => item.key === preset);

    setRange({ preset, from: startOfRange(option.days), to: "" });
  };

  const toggle = (stock) =>
    setVisible((current) => (current.includes(stock) ? current.filter((item) => item !== stock) : current.length < MAX_LINES ? [...current, stock] : current));

  return (
    <div className="monitor">
      <div className="toolbar filters">
        <div className="segmented" role="group" aria-label={t("chart.mode")}>
          {LAYERS.map((layer) => (
            <button key={layer} type="button" className={layers.includes(layer) ? "active" : ""} aria-pressed={layers.includes(layer)} onClick={() => toggleLayer(layer)}>
              {t(`chart.${layer}`)}
            </button>
          ))}
        </div>
        {RANGES.map((option) => (
          <button key={option.key} type="button" className={range.preset === option.key ? "active" : ""} onClick={() => choosePreset(option.key)}>
            {t(`range.${option.key}`)}
          </button>
        ))}
        <label className="inline">
          {t("monitor.from")}
          <input
            type="datetime-local"
            value={toLocalInput(range.from)}
            onChange={(event) => event.target.value && setRange({ ...range, preset: null, from: fromLocalInput(event.target.value) })}
          />
        </label>
        <label className="inline">
          {t("monitor.to")}
          <input type="datetime-local" value={toLocalInput(range.to)} onChange={(event) => setRange({ ...range, preset: null, to: fromLocalInput(event.target.value) })} />
        </label>
        {range.to && (
          <button type="button" onClick={() => setRange({ ...range, to: "" })}>
            {t("monitor.live")}
          </button>
        )}
        <select value={monitor.interval} onChange={(event) => update({ interval: event.target.value })} aria-label={t("monitor.interval")}>
          {INTERVALS.map((value) => (
            <option key={value} value={value}>
              {t(`interval.${value}`)}
            </option>
          ))}
        </select>
        <label className="inline">
          {t("monitor.lifetime")}
          <DurationInput value={lifetime} units={LIFETIME_UNITS} placeholder={t("monitor.lifetimeDefault")} onChange={setLifetime} />
        </label>
        <button
          type="button"
          onClick={() => {
            setDraft(monitor.notifications);
            setEditing(true);
          }}
        >
          {t("monitor.notificationsCount", { count: monitor.notifications.length })}
        </button>
        <button type="button" onClick={() => setAlertsOpen(true)}>
          🔔 {t("alerts.button", { count: alerts.filter((alert) => alert.enabled).length })}
        </button>
        <span className={`stream ${status.state}`}>{t(`stream.${status.state}`)}</span>
      </div>
      {(error || status.message) && <p className="notice error">{error || status.message}</p>}
      <div className={`monitor-body${status.state === "connecting" ? " refreshing" : ""}`}>
        {monitor.stocks.length === 0 ? (
          <p className="hint">{t("monitor.empty")}</p>
        ) : (
          <MonitorChart key={layers.join("+")} series={series} markers={markers} dark={dark} names={names} focusTime={focusTime} interval={monitor.interval} />
        )}
        <ContractPanel
          stocks={monitor.stocks}
          contracts={contracts}
          data={data}
          slots={slots}
          colors={colors}
          visible={visible}
          canShow={visible.length < MAX_LINES}
          notifications={monitor.notifications}
          catalog={catalog}
          reference={range.to ? Date.parse(range.to) : now}
          focused={focused}
          highlight={highlight}
          onHighlight={(next) => {
            setHighlight(next);
            next.stock && setFocused(next.stock);
          }}
          onToggle={toggle}
          onFocus={setFocused}
          onDetails={setDetails}
          onRemove={(stock) => {
            const affected = alerts.filter((alert) => alert.stock === stock).length;

            if (!affected || window.confirm(t("alerts.removeContractConfirm", { id: stock, count: affected }))) {
              update({ stocks: monitor.stocks.filter((item) => item !== stock) });
            }
          }}
          onAdd={(stock) => update({ stocks: [...monitor.stocks, stock] })}
        />
      </div>
      <EventsLog
        signals={allSignals}
        stocks={monitor.stocks}
        notifications={monitor.notifications}
        names={names}
        reference={range.to ? Date.parse(range.to) : now}
        onOpen={openEvent}
      />
      {details && <Details stock={details} onClose={() => setDetails(null)} />}
      {alertsOpen && (
        <Modal title={t("alerts.title", { name: monitor.name })} onClose={() => setAlertsOpen(false)}>
          <AlertsManager monitor={monitor} names={names} onSave={(next) => update({ alerts: next })} />
        </Modal>
      )}
      {editing && (
        <Modal title={t("monitor.notifications")} onClose={() => setEditing(false)}>
          <NotificationPicker catalog={catalog} selected={draft} onChange={setDraft} />
          <div className="actions">
            <button type="button" className="primary" onClick={() => update({ notifications: draft }).then(() => setEditing(false))}>
              {t("common.save")}
            </button>
            <button type="button" onClick={() => setEditing(false)}>
              {t("common.cancel")}
            </button>
          </div>
        </Modal>
      )}
    </div>
  );
}
