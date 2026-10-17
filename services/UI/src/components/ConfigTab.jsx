import { useEffect, useState } from "react";
import { api } from "../api.js";
import { useI18n } from "../i18n/I18n.jsx";
import { CONFIG_SECTIONS, KNOWN_KEYS, getAt, setAt } from "../config/schema.js";
import Field from "./Field.jsx";

export default function ConfigTab() {
  const { t } = useI18n();
  const [loaded, setLoaded] = useState(null);
  const [draft, setDraft] = useState(null);
  const [status, setStatus] = useState({ state: "loading" });

  const load = () => {
    setStatus({ state: "loading" });
    api
      .getConfig()
      .then((config) => {
        setLoaded(config);
        setDraft(config);
        setStatus({ state: "idle" });
      })
      .catch((error) => setStatus({ state: "error", message: error.message }));
  };

  useEffect(load, []);

  const save = (event) => {
    event.preventDefault();
    setStatus({ state: "saving" });
    api
      .saveConfig(draft)
      .then((config) => {
        setLoaded(config);
        setDraft(config);
        setStatus({ state: "saved" });
      })
      .catch((error) => setStatus({ state: "error", message: t("config.failed", { error: error.message }) }));
  };

  if (!draft) {
    return <p className={status.state === "error" ? "notice error" : "notice"}>{status.message || t("common.loading")}</p>;
  }

  const extra = Object.keys(draft).filter((key) => !KNOWN_KEYS.has(key));
  const busy = status.state === "saving";

  return (
    <form className="config" onSubmit={save}>
      {CONFIG_SECTIONS.map((section) => (
        <fieldset key={section.key} disabled={busy}>
          <legend>{t(`config.section.${section.key}`)}</legend>
          {section.fields.map((field) => (
            <Field
              key={field.path.join(".")}
              label={t(field.label)}
              field={field}
              value={getAt(draft, field.path)}
              onChange={(value) => setDraft(setAt(draft, field.path, value))}
            />
          ))}
        </fieldset>
      ))}
      {extra.length > 0 && <p className="hint">{t("config.extra", { fields: extra.join(", ") })}</p>}
      <div className="actions">
        <button type="submit" disabled={busy}>
          {t("common.save")}
        </button>
        <button type="button" disabled={busy} onClick={() => setDraft(loaded)}>
          {t("common.reset")}
        </button>
      </div>
      {status.state === "saving" && <p className="notice">{t("config.applying")}</p>}
      {status.state === "saved" && <p className="notice success">{t("config.saved")}</p>}
      {status.state === "error" && <p className="notice error">{status.message}</p>}
    </form>
  );
}
