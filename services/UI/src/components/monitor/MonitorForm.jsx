import { useState } from "react";
import { useI18n } from "../../i18n/I18n.jsx";
import NotificationPicker from "./NotificationPicker.jsx";
import { INTERVALS, LIFETIME_UNITS } from "./intervals.js";
import DurationInput from "../DurationInput.jsx";

export default function MonitorForm({ contracts, catalog, initial, onSubmit, onCancel }) {
  const { t } = useI18n();
  const [name, setName] = useState(initial?.name || "");
  const [stocks, setStocks] = useState(initial?.stocks || []);
  const [interval, setTimeframe] = useState(initial?.interval || "1m");
  const [notifications, setNotifications] = useState(initial?.notifications || []);
  const [lifetime, setLifetime] = useState(initial?.event_lifetime || null);
  const [status, setStatus] = useState(null);

  const toggle = (id) => setStocks(stocks.includes(id) ? stocks.filter((stock) => stock !== id) : [...stocks, id]);

  const submit = (event) => {
    event.preventDefault();
    setStatus({ state: "saving" });
    Promise.resolve(onSubmit({ name, stocks, interval, notifications, event_lifetime: lifetime })).catch((error) => setStatus({ state: "error", message: error.message }));
  };

  return (
    <form className="monitor-form" onSubmit={submit}>
      <fieldset disabled={status?.state === "saving"}>
        <label className="field">
          <span>{t("monitor.name")}</span>
          <input type="text" value={name} required onChange={(event) => setName(event.target.value)} />
        </label>
        <label className="field">
          <span>{t("monitor.interval")}</span>
          <select value={interval} onChange={(event) => setTimeframe(event.target.value)}>
            {INTERVALS.map((value) => (
              <option key={value} value={value}>
                {t(`interval.${value}`)}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>{t("monitor.lifetime")}</span>
          <DurationInput value={lifetime} units={LIFETIME_UNITS} placeholder={t("monitor.lifetimeDefault")} onChange={setLifetime} />
        </label>
      </fieldset>
      <h4>{t("monitor.contracts")}</h4>
      {contracts.length === 0 ? (
        <p className="hint">{t("contracts.none")}</p>
      ) : (
        <div className="choices">
          {contracts.map(({ metric, contract }) => (
            <label key={metric.id} className="picker-item">
              <input type="checkbox" checked={stocks.includes(metric.id)} onChange={() => toggle(metric.id)} />
              <span>
                <strong>{metric.id}</strong> {contract?.name}
              </span>
            </label>
          ))}
        </div>
      )}
      <h4>{t("monitor.notifications")}</h4>
      <NotificationPicker catalog={catalog} selected={notifications} onChange={setNotifications} />
      <div className="actions">
        <button type="submit" disabled={status?.state === "saving"}>
          {t("common.save")}
        </button>
        <button type="button" onClick={onCancel}>
          {t("common.cancel")}
        </button>
      </div>
      {status?.state === "error" && <p className="notice error">{status.message}</p>}
    </form>
  );
}
