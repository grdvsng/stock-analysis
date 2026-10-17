import { useMemo, useState } from "react";
import { useI18n } from "../../i18n/I18n.jsx";
import { STATUS } from "../../monitor/palette.js";
import { isActive } from "./intervals.js";

const PAGE = 100;

export default function EventsLog({ signals, stocks, notifications, names, reference, onOpen }) {
  const { t, language } = useI18n();
  const [stock, setStock] = useState("");
  const [metric, setMetric] = useState("");
  const [action, setAction] = useState("");
  const [limit, setLimit] = useState(PAGE);
  const time = new Intl.DateTimeFormat(language, { dateStyle: "short", timeStyle: "short" });

  const latest = useMemo(() => {
    const result = new Map();

    for (const signal of signals) {
      const key = `${signal.stock}:${signal.metric}`;

      if (signal.time <= reference && (!result.has(key) || signal.time > result.get(key).time)) {
        result.set(key, signal);
      }
    }

    return result;
  }, [signals, reference]);

  const filtered = useMemo(
    () => signals.filter((signal) => (!stock || signal.stock === stock) && (!metric || signal.metric === metric) && (!action || signal.action === action)),
    [signals, stock, metric, action],
  );

  const status = (signal) => {
    if (latest.get(`${signal.stock}:${signal.metric}`) === signal && isActive(signal, reference)) {
      return <span className="badge active">{t("events.activeUntil", { time: time.format(new Date(signal.expires)) })}</span>;
    }

    if (signal.time > reference) {
      return <span className="muted">{t("common.empty")}</span>;
    }

    return <span className="muted">{latest.get(`${signal.stock}:${signal.metric}`) === signal ? t("events.expired") : t("events.superseded")}</span>;
  };

  return (
    <section className="events">
      <div className="toolbar">
        <h3>{t("events.title", { count: filtered.length })}</h3>
        <select value={stock} onChange={(event) => setStock(event.target.value)} aria-label={t("contracts.stock")}>
          <option value="">{t("events.allStocks")}</option>
          {stocks.map((item) => (
            <option key={item} value={item}>
              {item}
            </option>
          ))}
        </select>
        <select value={metric} onChange={(event) => setMetric(event.target.value)} aria-label={t("monitor.metric")}>
          <option value="">{t("events.allMetrics")}</option>
          {notifications.map((key) => (
            <option key={key} value={key}>
              {names.get(key) || key}
            </option>
          ))}
        </select>
        <select value={action} onChange={(event) => setAction(event.target.value)} aria-label={t("events.signal")}>
          <option value="">{t("events.allSignals")}</option>
          <option value="buy">{t("action.buy")}</option>
          <option value="sell">{t("action.sell")}</option>
        </select>
      </div>
      {filtered.length === 0 ? (
        <p className="hint">{t("events.none")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("events.time")}</th>
              <th>{t("contracts.stock")}</th>
              <th>{t("monitor.metric")}</th>
              <th>{t("events.signal")}</th>
              <th>{t("events.value")}</th>
              <th>{t("events.status")}</th>
            </tr>
          </thead>
          <tbody>
            {filtered.slice(0, limit).map((signal) => (
              <tr key={`${signal.stock}:${signal.metric}:${signal.time}`} className="row" onClick={() => onOpen(signal)} title={t("events.open")}>
                <td>{time.format(new Date(signal.time))}</td>
                <td>{signal.stock}</td>
                <td>{names.get(signal.metric) || signal.metric}</td>
                <td>
                  <span className="action" style={{ color: STATUS[signal.action] }}>
                    {signal.action === "buy" ? "▲" : "▼"} {t(`action.${signal.action}`)}
                  </span>
                </td>
                <td className="number">{signal.value.toFixed(2)}</td>
                <td>{status(signal)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {filtered.length > limit && (
        <button type="button" onClick={() => setLimit(limit + PAGE)}>
          {t("events.more", { count: Math.min(PAGE, filtered.length - limit) })}
        </button>
      )}
    </section>
  );
}
