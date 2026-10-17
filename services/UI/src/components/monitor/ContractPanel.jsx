import { useState } from "react";
import { useI18n } from "../../i18n/I18n.jsx";
import { STATUS } from "../../monitor/palette.js";
import { isActive } from "./intervals.js";

function summarize(signals, notifications, reference) {
  const byMetric = new Map(notifications.map((key) => [key, { hits: 0, last: null, current: null, active: null }]));

  for (const signal of signals) {
    const entry = byMetric.get(signal.metric);

    if (!entry) {
      continue;
    }

    entry.hits++;

    if (!entry.last || signal.time >= entry.last.time) {
      entry.last = signal;
    }

    if (signal.time <= reference && (!entry.current || signal.time >= entry.current.time)) {
      entry.current = signal;
    }
  }

  for (const entry of byMetric.values()) {
    entry.active = entry.current && isActive(entry.current, reference) ? entry.current : null;
  }

  return byMetric;
}

function Action({ action }) {
  const { t } = useI18n();

  return (
    <span className="action" style={{ color: STATUS[action] }}>
      {action === "buy" ? "▲" : "▼"} {t(`action.${action}`)}
    </span>
  );
}

function Row({ stock, contract, data, color, visible, canShow, notifications, catalog, reference, onToggle, onDetails, onRemove, onFocus, focused, highlight, onHighlight }) {
  const { t, language } = useI18n();
  const [open, setOpen] = useState(false);
  const candles = data?.candles || [];
  const last = candles[candles.length - 1];
  const first = candles[0];
  const change = first && last ? (last.close / first.close - 1) * 100 : null;
  const summary = summarize(data?.signals || [], notifications, reference);
  const notified = [...summary.values()].filter((entry) => entry.active).length;
  const sleeping = data ? !data.trading : false;
  const names = new Map(catalog.map((item) => [item.key, item.name]));
  const time = new Intl.DateTimeFormat(language, { dateStyle: "short", timeStyle: "short" });

  return (
    <li className={`panel-row${sleeping ? " sleeping" : ""}${focused ? " focused" : ""}`}>
      <div className="panel-head">
        <label className="panel-visible" title={t("monitor.showOnChart")}>
          <input type="checkbox" checked={visible} disabled={!visible && !canShow} onChange={onToggle} />
          <i style={{ background: visible ? color : "transparent", borderColor: color || "currentColor" }} />
        </label>
        <button type="button" className="link" onClick={() => onFocus(stock)}>
          <strong>{stock}</strong>
          <span className="muted"> {contract?.name}</span>
        </button>
        <button type="button" className="icon" onClick={() => onRemove(stock)} aria-label={t("monitor.removeContract", { id: stock })}>
          ×
        </button>
      </div>
      <div className="panel-stats">
        <span>{last ? last.close.toFixed(2) : t("common.empty")}</span>
        {change !== null && <span className={change >= 0 ? "up" : "down"}>{`${change >= 0 ? "+" : ""}${change.toFixed(2)}%`}</span>}
        {sleeping && <span className="badge">{t("monitor.sleeping")}</span>}
        <button type="button" className="link" onClick={() => setOpen(!open)}>
          {t("monitor.notified", { count: notified, total: notifications.length })}
        </button>
        <button type="button" className="link" onClick={() => onDetails(stock)}>
          {t("monitor.details")}
        </button>
      </div>
      {open && (
        <table className="hits">
          <thead>
            <tr>
              <th>{t("monitor.metric")}</th>
              <th>{t("monitor.hits")}</th>
              <th>{t("monitor.activeNow")}</th>
              <th>{t("monitor.lastSignal")}</th>
            </tr>
          </thead>
          <tbody>
            {[...summary.entries()].map(([key, entry]) => (
              <tr
                key={key}
                className={`${entry.hits ? "" : "muted"}${highlight.stock === stock && (highlight.metric === key || highlight.metric === "all") ? " highlighted" : ""}`}
                onClick={() => onHighlight(highlight.stock === stock && highlight.metric === key ? { stock: null, metric: null } : { stock, metric: key })}
              >
                <td>{names.get(key) || key}</td>
                <td>{entry.hits}</td>
                <td>
                  {entry.active ? (
                    <>
                      <Action action={entry.active.action} /> <span className="muted">{t("monitor.until", { time: time.format(new Date(entry.active.expires)) })}</span>
                    </>
                  ) : (
                    t("common.empty")
                  )}
                </td>
                <td>
                  {entry.last ? (
                    <>
                      <Action action={entry.last.action} /> <span className="muted">{time.format(new Date(entry.last.time))}</span>{" "}
                      {entry.last.value.toFixed(2)}
                    </>
                  ) : (
                    t("common.empty")
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {open && (
        <div className="hits-actions">
          <button type="button" className="link" onClick={() => onHighlight({ stock, metric: "all" })}>
            {t("monitor.markAll")}
          </button>{" "}
          {highlight.stock === stock && (
            <button type="button" className="link" onClick={() => onHighlight({ stock: null, metric: null })}>
              {t("monitor.markNone")}
            </button>
          )}
        </div>
      )}
    </li>
  );
}

export default function ContractPanel({ stocks, contracts, data, slots, colors, visible, canShow, notifications, catalog, reference, focused, highlight, onHighlight, onToggle, onFocus, onDetails, onRemove, onAdd }) {
  const { t } = useI18n();
  const byId = new Map(contracts.map((item) => [item.metric.id, item.contract]));
  const available = contracts.filter((item) => !stocks.includes(item.metric.id));

  return (
    <aside className="panel">
      <ul>
        {stocks.map((stock) => (
          <Row
            key={stock}
            stock={stock}
            contract={byId.get(stock)}
            data={data[stock]}
            color={slots[stock] !== undefined ? colors[slots[stock]] : null}
            visible={visible.includes(stock)}
            canShow={canShow}
            notifications={notifications}
            catalog={catalog}
            reference={reference}
            focused={focused === stock}
            highlight={highlight}
            onHighlight={onHighlight}
            onToggle={() => onToggle(stock)}
            onFocus={onFocus}
            onDetails={onDetails}
            onRemove={onRemove}
          />
        ))}
      </ul>
      {available.length > 0 && (
        <select value="" onChange={(event) => event.target.value && onAdd(event.target.value)} aria-label={t("monitor.addContract")}>
          <option value="">{t("monitor.addContract")}</option>
          {available.map(({ metric, contract }) => (
            <option key={metric.id} value={metric.id}>
              {metric.id} {contract?.name ? `— ${contract.name}` : ""}
            </option>
          ))}
        </select>
      )}
    </aside>
  );
}
