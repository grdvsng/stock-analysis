import { useEffect, useRef, useState } from "react";
import { CandlestickSeries, createChart, createSeriesMarkers, CrosshairMode, LineSeries } from "lightweight-charts";
import { useI18n } from "../../i18n/I18n.jsx";
import { CANDLE, CHROME, STATUS } from "../../monitor/palette.js";
import { isActive } from "./intervals.js";

const MAX_EVENTS = 6;
const INTERVAL_SECONDS = { "1m": 60, "5m": 300, "15m": 900, "30m": 1800, "1h": 3600, "1d": 86400 };
const FOLLOW_TOLERANCE = 2;

const seconds = (time) => Math.floor(time / 1000);
const percent = (value) => `${value >= 0 ? "+" : ""}${value.toFixed(2)}%`;
const price = (value) => value.toFixed(2);

function formatter(language, options) {
  return new Intl.DateTimeFormat(language, options);
}

function chartOptions(dark, language) {
  const chrome = dark ? CHROME.dark : CHROME.light;
  const full = formatter(language, { dateStyle: "medium", timeStyle: "short" });
  const day = formatter(language, { month: "short", day: "numeric" });
  const clock = formatter(language, { hour: "2-digit", minute: "2-digit" });

  return {
    autoSize: true,
    layout: {
      background: { color: chrome.surface },
      textColor: chrome.muted,
      fontFamily: 'system-ui, -apple-system, "Segoe UI", sans-serif',
      attributionLogo: true,
    },
    grid: { vertLines: { visible: false }, horzLines: { color: chrome.grid } },
    rightPriceScale: { borderColor: chrome.axis },
    timeScale: {
      borderColor: chrome.axis,
      timeVisible: true,
      tickMarkFormatter: (time, type) => (type <= 2 ? day : clock).format(new Date(time * 1000)),
    },
    crosshair: { mode: CrosshairMode.Magnet, vertLine: { labelVisible: false }, horzLine: { labelVisible: false } },
    localization: { timeFormatter: (time) => full.format(new Date(time * 1000)) },
  };
}

function addSeries(instance, item) {
  if (item.type === "candle") {
    const format = item.format === "percent" ? percent : price;

    return instance.addSeries(CandlestickSeries, {
      upColor: "rgba(0, 0, 0, 0)",
      borderUpColor: CANDLE.up,
      wickUpColor: CANDLE.up,
      downColor: CANDLE.down,
      borderDownColor: CANDLE.down,
      wickDownColor: CANDLE.down,
      priceLineVisible: false,
      priceFormat: { type: "custom", formatter: format, minMove: 0.01 },
    });
  }

  return instance.addSeries(LineSeries, {
    lineWidth: 2,
    priceLineVisible: false,
    lastValueVisible: true,
    crosshairMarkerRadius: 4,
    priceFormat: { type: "custom", formatter: percent, minMove: 0.01 },
  });
}

function activeEvents(signals, moment, names) {
  const latest = new Map();

  for (const signal of signals) {
    if (signal.time <= moment && (!latest.has(signal.metric) || signal.time > latest.get(signal.metric).time)) {
      latest.set(signal.metric, signal);
    }
  }

  return [...latest.values()]
    .filter((signal) => isActive(signal, moment))
    .sort((a, b) => b.time - a.time)
    .map((signal) => ({ key: signal.metric, action: signal.action, name: names.get(signal.metric) || signal.metric }));
}

function Events({ events }) {
  const { t } = useI18n();

  if (events.length === 0) {
    return null;
  }

  return (
    <ul className="tooltip-events">
      {events.slice(0, MAX_EVENTS).map((event) => (
        <li key={event.key}>
          <span style={{ color: STATUS[event.action] }}>{event.action === "buy" ? "▲" : "▼"}</span> {event.name}
        </li>
      ))}
      {events.length > MAX_EVENTS && <li className="muted">{t("monitor.moreEvents", { count: events.length - MAX_EVENTS })}</li>}
    </ul>
  );
}

function Row({ row }) {
  const { t } = useI18n();

  if (row.type === "candle") {
    const up = row.close >= row.open;

    return (
      <div className="tooltip-row">
        <span className="candle-key" style={{ borderColor: up ? CANDLE.up : CANDLE.down, background: up ? "transparent" : CANDLE.down }} />
        <strong>{price(row.close)}</strong>
        <span style={{ color: up ? CANDLE.up : CANDLE.down }}>
          {up ? "▲" : "▼"} {percent((row.close / row.open - 1) * 100)}
        </span>
        <span>{row.label}</span>
        <span className="ohlc muted">
          {t("candle.open")} {price(row.open)} · {t("candle.high")} {price(row.high)} · {t("candle.low")} {price(row.low)}
        </span>
        <Events events={row.events} />
      </div>
    );
  }

  return (
    <div className="tooltip-row">
      <i style={{ background: row.color }} />
      <strong>{percent(row.change)}</strong>
      <span>{row.label}</span>
      {row.price !== undefined && <span className="muted">{price(row.price)}</span>}
      <Events events={row.events} />
    </div>
  );
}

export default function MonitorChart({ series: items, markers, dark, names, focusTime, interval }) {
  const { t, language } = useI18n();
  const container = useRef(null);
  const chart = useRef(null);
  const apis = useRef(new Map());
  const markerPlugin = useRef(null);
  const fitted = useRef(false);
  const itemsRef = useRef(items);
  const namesRef = useRef(names);
  const [tooltip, setTooltip] = useState(null);

  itemsRef.current = items;
  namesRef.current = names;

  useEffect(() => {
    const instance = createChart(container.current, chartOptions(dark, language));
    const registry = apis.current;

    chart.current = instance;

    const onMove = (param) => {
      if (!param.time || !param.point || param.point.x < 0) {
        setTooltip(null);

        return;
      }

      const moment = param.time * 1000;
      const rows = [];

      for (const item of itemsRef.current) {
        const api = registry.get(item.id);
        const value = api && param.seriesData.get(api);

        if (!value) {
          continue;
        }

        const events = activeEvents(item.signals, moment, namesRef.current);

        if (item.type === "candle" && value.close !== undefined) {
          const candle = item.raw.get(param.time) || value;

          rows.push({ id: item.id, type: "candle", label: item.label, open: candle.open, high: candle.high, low: candle.low, close: candle.close, events });
        } else if (value.value !== undefined) {
          rows.push({ id: item.id, type: "line", label: item.label, color: item.color, change: value.value, price: item.prices.get(param.time), events });
        }
      }

      rows.sort((a, b) => (b.change ?? 0) - (a.change ?? 0));
      setTooltip(rows.length ? { x: param.point.x, time: param.time, rows } : null);
    };

    instance.subscribeCrosshairMove(onMove);

    return () => {
      instance.unsubscribeCrosshairMove(onMove);
      instance.remove();
      registry.clear();
      markerPlugin.current = null;
      chart.current = null;
      fitted.current = false;
    };
  }, []);

  useEffect(() => {
    chart.current && chart.current.applyOptions(chartOptions(dark, language));
  }, [dark, language]);

  useEffect(() => {
    const instance = chart.current;

    if (!instance) {
      return;
    }

    const wanted = new Set(items.map((item) => item.id));
    const following = fitted.current && instance.timeScale().scrollPosition() >= -FOLLOW_TOLERANCE;

    for (const [id, api] of apis.current) {
      if (!wanted.has(id)) {
        instance.removeSeries(api);
        apis.current.delete(id);
      }
    }

    for (const item of items) {
      let api = apis.current.get(item.id);

      if (!api) {
        api = addSeries(instance, item);
        apis.current.set(item.id, api);
      }

      if (item.type === "line") {
        api.applyOptions({ color: item.color, title: item.label, lineWidth: item.thin ? 1 : 2 });
      } else {
        api.applyOptions({ title: item.overlay ? "" : item.label, lastValueVisible: !item.overlay });
      }

      api.setData(item.points);
    }

    const focused = markers.series && apis.current.get(markers.series);

    if (markerPlugin.current) {
      markerPlugin.current.detach();
      markerPlugin.current = null;
    }

    if (focused && markers.items.length) {
      markerPlugin.current = createSeriesMarkers(
        focused,
        markers.items.map((item) => ({
          time: seconds(item.time),
          position: item.action === "buy" ? "belowBar" : "aboveBar",
          shape: item.action === "buy" ? "arrowUp" : "arrowDown",
          color: STATUS[item.action],
          size: 1,
        })),
      );
    }

    if (!fitted.current && items.some((item) => item.points.length)) {
      instance.timeScale().fitContent();
      fitted.current = true;
    } else if (following) {
      instance.timeScale().scrollToRealTime();
    }
  }, [items, markers]);

  useEffect(() => {
    const instance = chart.current;

    if (!instance || !focusTime) {
      return;
    }

    const center = seconds(focusTime.time);
    const span = Math.max(30, Math.round((INTERVAL_SECONDS[interval] || 60) * 40));

    instance.timeScale().setVisibleRange({ from: center - span, to: center + span });
  }, [focusTime]);

  const time = tooltip && formatter(language, { dateStyle: "medium", timeStyle: "short" }).format(new Date(tooltip.time * 1000));
  const candles = items.some((item) => item.type === "candle");

  return (
    <div className="monitor-chart">
      <div className="legend" aria-label={t("monitor.legend")}>
        {items.map((item) => (
          <span key={item.id}>
            {item.type === "line" && <i style={{ background: item.color }} />}
            {item.label}
          </span>
        ))}
        {candles && (
          <>
            <span>
              <span className="candle-key" style={{ borderColor: CANDLE.up }} /> {t("candle.up")}
            </span>
            <span>
              <span className="candle-key" style={{ borderColor: CANDLE.down, background: CANDLE.down }} /> {t("candle.down")}
            </span>
          </>
        )}
      </div>
      <div className="canvas" ref={container}>
        {tooltip && (
          <div className="tooltip" style={{ left: Math.min(tooltip.x + 16, (container.current?.clientWidth || 0) - 250) }}>
            <div className="tooltip-time">{time}</div>
            {tooltip.rows.map((row) => (
              <Row key={row.id} row={row} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
