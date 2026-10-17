import { useState } from "react";
import { useI18n } from "../../i18n/I18n.jsx";
import { CONDITIONS } from "../../alerts/evaluate.js";

const EMPTY = { name: "", stock: "", metrics: [], condition: "any", enabled: true };

function AlertForm({ initial, stocks, notifications, names, onSave, onCancel }) {
  const { t } = useI18n();
  const [alert, setAlert] = useState({ ...EMPTY, stock: stocks[0] || "", ...initial });
  const [error, setError] = useState(null);
  const change = (key) => (value) => setAlert({ ...alert, [key]: value });

  const toggle = (metric) =>
    change("metrics")(alert.metrics.includes(metric) ? alert.metrics.filter((item) => item !== metric) : notifications.filter((item) => item === metric || alert.metrics.includes(item)));

  const submit = (event) => {
    event.preventDefault();

    if (!alert.metrics.length) {
      setError(t("alerts.needMetric"));

      return;
    }

    setError(null);
    Promise.resolve(onSave(alert)).catch((reason) => setError(reason.message));
  };

  return (
    <form className="alert-form" onSubmit={submit}>
      <fieldset>
        <label className="field">
          <span>{t("alerts.name")}</span>
          <input type="text" required value={alert.name} onChange={(event) => change("name")(event.target.value)} />
        </label>
        <label className="field">
          <span>{t("alerts.contract")}</span>
          <select value={alert.stock} onChange={(event) => change("stock")(event.target.value)}>
            {stocks.map((stock) => (
              <option key={stock} value={stock}>
                {stock}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>{t("alerts.condition")}</span>
          <select value={alert.condition} onChange={(event) => change("condition")(event.target.value)}>
            {CONDITIONS.map((condition) => (
              <option key={condition} value={condition}>
                {t(`alerts.condition.${condition}`)}
              </option>
            ))}
          </select>
        </label>
        <label className="picker-item">
          <input type="checkbox" checked={alert.enabled} onChange={(event) => change("enabled")(event.target.checked)} />
          <span>{t("alerts.enabled")}</span>
        </label>
      </fieldset>
      <h4>{t("alerts.metrics")}</h4>
      <p className="hint">{t("alerts.metricsHint")}</p>
      <div className="choices">
        {notifications.map((metric) => (
          <label key={metric} className="picker-item">
            <input type="checkbox" checked={alert.metrics.includes(metric)} onChange={() => toggle(metric)} />
            <span>{names.get(metric) || metric}</span>
          </label>
        ))}
      </div>
      <div className="actions">
        <button type="submit">{t("common.save")}</button>
        <button type="button" onClick={onCancel}>
          {t("common.cancel")}
        </button>
      </div>
      {error && <p className="notice error">{error}</p>}
    </form>
  );
}

export default function AlertsManager({ monitor, names, onSave }) {
  const { t } = useI18n();
  const alerts = monitor.alerts || [];
  const [editing, setEditing] = useState(null);

  const persist = (next) => onSave(next).then(() => setEditing(null));

  if (monitor.stocks.length === 0 || monitor.notifications.length === 0) {
    return <p className="hint">{t("alerts.unavailable")}</p>;
  }

  if (editing) {
    return (
      <AlertForm
        initial={editing === "new" ? {} : editing}
        stocks={monitor.stocks}
        notifications={monitor.notifications}
        names={names}
        onCancel={() => setEditing(null)}
        onSave={(alert) => persist(editing === "new" ? [...alerts, alert] : alerts.map((item) => (item.id === alert.id ? alert : item)))}
      />
    );
  }

  return (
    <div className="alerts-list">
      {alerts.length === 0 ? (
        <p className="hint">{t("alerts.none")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("alerts.name")}</th>
              <th>{t("alerts.contract")}</th>
              <th>{t("alerts.condition")}</th>
              <th>{t("alerts.metrics")}</th>
              <th>{t("alerts.enabled")}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {alerts.map((alert) => (
              <tr key={alert.id}>
                <td>{alert.name}</td>
                <td>{alert.stock}</td>
                <td>{t(`alerts.condition.${alert.condition}`)}</td>
                <td>{alert.metrics.map((metric) => names.get(metric) || metric).join(" + ")}</td>
                <td>
                  <input
                    type="checkbox"
                    checked={alert.enabled}
                    aria-label={t("alerts.enabled")}
                    onChange={(event) => onSave(alerts.map((item) => (item.id === alert.id ? { ...item, enabled: event.target.checked } : item)))}
                  />
                </td>
                <td className="row-actions">
                  <button type="button" className="link" onClick={() => setEditing(alert)}>
                    {t("alerts.edit")}
                  </button>
                  <button
                    type="button"
                    className="link danger-text"
                    onClick={() => window.confirm(t("alerts.deleteConfirm", { name: alert.name })) && onSave(alerts.filter((item) => item.id !== alert.id))}
                  >
                    {t("alerts.delete")}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <div className="actions">
        <button type="button" className="primary" onClick={() => setEditing("new")}>
          {t("alerts.add")}
        </button>
      </div>
    </div>
  );
}
