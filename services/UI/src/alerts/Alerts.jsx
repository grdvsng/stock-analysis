import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useI18n } from "../i18n/I18n.jsx";
import { STATUS } from "../monitor/palette.js";
import { onSoundChange, ring, soundReady, unlockSound } from "./bell.js";

const TOASTS_KEY = "alertToasts";
const EPISODES_KEY = "alertEpisodes";
const AlertsContext = createContext(null);

function load(key, fallback) {
  try {
    return JSON.parse(localStorage.getItem(key)) ?? fallback;
  } catch {
    return fallback;
  }
}

function save(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    return;
  }
}

function Bell() {
  return (
    <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
      <path
        fill="currentColor"
        d="M12 22a2.5 2.5 0 0 0 2.45-2h-4.9A2.5 2.5 0 0 0 12 22Zm7-6V11a7 7 0 0 0-5.5-6.84V3.5a1.5 1.5 0 0 0-3 0v.66A7 7 0 0 0 5 11v5l-2 2v1h18v-1l-2-2Z"
      />
    </svg>
  );
}

function Toasts({ toasts, onClose, sound }) {
  const { t, language } = useI18n();
  const time = new Intl.DateTimeFormat(language, { dateStyle: "short", timeStyle: "medium" });

  if (toasts.length === 0) {
    return null;
  }

  return (
    <div className="toasts" role="region" aria-label={t("alerts.popups")}>
      {!sound && (
        <button type="button" className="toast-sound" onClick={unlockSound}>
          {t("alerts.soundBlocked")}
        </button>
      )}
      {toasts.map((toast) => (
        <div key={toast.id} className="toast" role="alertdialog" aria-labelledby={`toast-${toast.id}`}>
          <span className="toast-bell">
            <Bell />
          </span>
          <div className="toast-body">
            <strong id={`toast-${toast.id}`}>{toast.alert}</strong>
            <div>
              {toast.stock}
              {toast.contract ? ` — ${toast.contract}` : ""}
            </div>
            <div className="muted">
              {time.format(new Date(toast.at))} · {toast.monitor}
            </div>
            <ul>
              {toast.signals.map((signal) => (
                <li key={signal.metric}>
                  <span style={{ color: STATUS[signal.action] }}>{signal.action === "buy" ? "▲" : "▼"}</span> {signal.name}
                </li>
              ))}
            </ul>
          </div>
          <button type="button" className="icon" onClick={() => onClose(toast.id)} aria-label={t("common.close")}>
            ×
          </button>
        </div>
      ))}
    </div>
  );
}

export function AlertsProvider({ children }) {
  const [toasts, setToasts] = useState(() => load(TOASTS_KEY, []));
  const [sound, setSound] = useState(soundReady);

  useEffect(() => onSoundChange(setSound), []);

  useEffect(() => save(TOASTS_KEY, toasts), [toasts]);

  const fire = useCallback((key, episode, toast) => {
    const episodes = load(EPISODES_KEY, {});

    if (episodes[key] === episode) {
      return;
    }

    episodes[key] = episode;
    save(EPISODES_KEY, episodes);
    setToasts((current) => [{ ...toast, id: `${key}:${episode}`, at: Date.now() }, ...current.filter((item) => item.id !== `${key}:${episode}`)]);
    ring();
  }, []);

  const close = useCallback((id) => setToasts((current) => current.filter((toast) => toast.id !== id)), []);

  const value = useMemo(() => ({ fire }), [fire]);

  return (
    <AlertsContext.Provider value={value}>
      {children}
      <Toasts toasts={toasts} onClose={close} sound={sound} />
    </AlertsContext.Provider>
  );
}

export function useAlerts() {
  return useContext(AlertsContext);
}
