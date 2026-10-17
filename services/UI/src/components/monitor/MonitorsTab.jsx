import { useEffect, useState } from "react";
import { api } from "../../api.js";
import { useI18n } from "../../i18n/I18n.jsx";
import MonitorForm from "./MonitorForm.jsx";
import MonitorView from "./MonitorView.jsx";

const STORAGE_KEY = "monitor";

function remembered() {
  try {
    return localStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

function remember(id) {
  try {
    id ? localStorage.setItem(STORAGE_KEY, id) : localStorage.removeItem(STORAGE_KEY);
  } catch {
    return;
  }
}

export default function MonitorsTab() {
  const { t, language } = useI18n();
  const [state, setState] = useState({ loading: true });
  const [selected, setSelected] = useState(remembered);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    let active = true;

    Promise.all([api.getMonitors(), api.getContracts(), api.getIndicators(), api.getSettings()])
      .then(([monitors, contracts, catalog, settings]) => active && setState({ monitors, contracts, catalog, wsPort: settings.ws_port }))
      .catch((reason) => active && setState({ error: reason.message }));

    return () => {
      active = false;
    };
  }, []);

  if (state.loading) {
    return <p className="notice">{t("common.loading")}</p>;
  }

  if (state.error) {
    return <p className="notice error">{state.error}</p>;
  }

  const { monitors, contracts, catalog, wsPort } = state;
  const current = monitors.find((monitor) => monitor.id === selected) || monitors[0] || null;
  const date = new Intl.DateTimeFormat(language, { dateStyle: "medium", timeStyle: "short" });

  const select = (id) => {
    setSelected(id);
    remember(id);
  };

  const replace = (monitor) =>
    setState((previous) => ({ ...previous, monitors: previous.monitors.map((item) => (item.id === monitor.id ? monitor : item)) }));

  const create = (body) =>
    api.createMonitor(body).then((monitor) => {
      setState((previous) => ({ ...previous, monitors: [...previous.monitors, monitor] }));
      select(monitor.id);
      setCreating(false);
    });

  const remove = () => {
    if (!current || !window.confirm(t("monitor.deleteConfirm", { name: current.name }))) {
      return;
    }

    api
      .deleteMonitor(current.id)
      .then(() => {
        setState((previous) => ({ ...previous, monitors: previous.monitors.filter((item) => item.id !== current.id) }));
        select(null);
      })
      .catch((reason) => setError(reason.message));
  };

  if (creating || monitors.length === 0) {
    return (
      <div>
        {monitors.length === 0 && <p className="hint">{t("monitor.none")}</p>}
        <MonitorForm contracts={contracts} catalog={catalog} onSubmit={create} onCancel={() => setCreating(false)} />
      </div>
    );
  }

  return (
    <div>
      <div className="toolbar">
        <select value={current.id} onChange={(event) => select(event.target.value)} aria-label={t("monitor.select")}>
          {monitors.map((monitor) => (
            <option key={monitor.id} value={monitor.id}>
              {monitor.name}
            </option>
          ))}
        </select>
        <button type="button" onClick={() => setCreating(true)}>
          {t("monitor.add")}
        </button>
        <button type="button" className="danger" onClick={remove}>
          {t("monitor.delete")}
        </button>
        <span className="hint">
          {t("monitor.meta", { id: current.id, created: date.format(new Date(current.created_at)), updated: date.format(new Date(current.updated_at)) })}
        </span>
      </div>
      {error && <p className="notice error">{error}</p>}
      <MonitorView key={current.id} monitor={current} contracts={contracts} catalog={catalog} wsPort={wsPort} onChange={replace} />
    </div>
  );
}
