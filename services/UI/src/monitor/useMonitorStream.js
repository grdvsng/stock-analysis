import { useEffect, useState } from "react";
import { monitorSocketUrl } from "../api.js";

const toTime = (value) => Date.parse(value);

function normalize(update) {
  return {
    trading: update.trading,
    candles: (update.candles || []).map((candle) => ({ ...candle, time: toTime(candle.time) })),
    signals: (update.signals || []).map((signal) => ({ ...signal, time: toTime(signal.time), expires: toTime(signal.expires) })),
  };
}

function merge(previous, message) {
  const next = message.type === "snapshot" ? {} : { ...previous };

  for (const [stock, raw] of Object.entries(message.stocks || {})) {
    const update = normalize(raw);
    const since = toTime(raw.since);
    const current = next[stock];

    if (message.type === "snapshot" || !current) {
      next[stock] = update;
      continue;
    }

    next[stock] = {
      trading: update.trading,
      candles: current.candles.filter((candle) => candle.time < since).concat(update.candles),
      signals: current.signals.filter((signal) => signal.time < since).concat(update.signals),
    };
  }

  return next;
}

export default function useMonitorStream({ port, stocks, interval, notifications, lifetime, from, to }) {
  const [data, setData] = useState({});
  const [status, setStatus] = useState({ state: "idle" });
  const key = [port, stocks.join(","), interval, notifications.join(","), lifetime, from, to].join("|");

  useEffect(() => {
    setData({});

    if (!port || stocks.length === 0) {
      setStatus({ state: "idle" });

      return undefined;
    }

    const params = new URLSearchParams({ stocks: stocks.join(","), interval, notifications: notifications.join(","), from });

    if (to) {
      params.set("to", to);
    }

    if (lifetime) {
      params.set("lifetime", lifetime);
    }

    let disposed = false;
    const socket = new WebSocket(monitorSocketUrl(port, params));

    setStatus({ state: "connecting" });

    socket.onmessage = (event) => {
      if (disposed) {
        return;
      }

      const message = JSON.parse(event.data);

      if (message.type === "error") {
        setStatus({ state: "error", message: message.error });

        return;
      }

      setData((previous) => merge(previous, message));
      setStatus({ state: to ? "complete" : "live" });
    };

    socket.onerror = () => {
      if (!disposed) {
        setStatus({ state: "error" });
      }
    };

    socket.onclose = () => {
      if (!disposed) {
        setStatus((current) => (current.state === "error" ? current : { state: "closed" }));
      }
    };

    return () => {
      disposed = true;
      socket.onmessage = socket.onerror = socket.onclose = null;
      socket.close(1000, "monitor closed");
    };
  }, [key]);

  return { data, status };
}
