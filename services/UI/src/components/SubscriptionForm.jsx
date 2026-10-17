import { useState } from "react";
import { useI18n } from "../i18n/I18n.jsx";
import { INTERVAL_UNITS, TTL_UNITS } from "../config/schema.js";
import DurationInput from "./DurationInput.jsx";

const EMPTY = { ttl: null, update_interval: null, contract_update_interval: null, reprocess: 0 };

export default function SubscriptionForm({ initial, submitLabel, onSubmit, onCancel }) {
  const { t } = useI18n();
  const [metric, setMetric] = useState({ ...EMPTY, ...(initial || {}) });
  const [status, setStatus] = useState(null);
  const change = (key) => (value) => setMetric({ ...metric, [key]: value });

  const submit = (event) => {
    event.preventDefault();
    setStatus({ state: "saving" });

    const body = {
      ttl: metric.ttl,
      update_interval: metric.update_interval,
      contract_update_interval: metric.contract_update_interval,
      reprocess: Number(metric.reprocess) || 0,
    };

    Promise.resolve(onSubmit(body))
      .then(() => setStatus(null))
      .catch((error) => setStatus({ state: "error", message: error.message }));
  };

  const busy = status && status.state === "saving";

  return (
    <form className="subscription" onSubmit={submit}>
      <h4>{t("subscription.title")}</h4>
      <fieldset disabled={busy}>
        <label className="field">
          <span>{t("subscription.ttl")}</span>
          <DurationInput value={metric.ttl} units={TTL_UNITS} placeholder={t("common.default")} onChange={change("ttl")} />
        </label>
        <label className="field">
          <span>{t("subscription.update_interval")}</span>
          <DurationInput
            value={metric.update_interval}
            units={INTERVAL_UNITS}
            placeholder={t("common.default")}
            onChange={change("update_interval")}
          />
        </label>
        <label className="field">
          <span>{t("subscription.contract_update_interval")}</span>
          <DurationInput
            value={metric.contract_update_interval}
            units={INTERVAL_UNITS}
            placeholder={t("common.default")}
            onChange={change("contract_update_interval")}
          />
        </label>
        <label className="field">
          <span>{t("subscription.reprocess")}</span>
          <input type="number" min="0" value={metric.reprocess} onChange={(event) => change("reprocess")(event.target.value)} />
        </label>
      </fieldset>
      <div className="actions">
        <button type="submit" disabled={busy}>
          {submitLabel}
        </button>
        {onCancel && (
          <button type="button" disabled={busy} onClick={onCancel}>
            {t("common.cancel")}
          </button>
        )}
      </div>
      {status && status.state === "error" && <p className="notice error">{status.message}</p>}
    </form>
  );
}
