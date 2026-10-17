import { useEffect, useState } from "react";
import { api } from "../../api.js";
import { useI18n } from "../../i18n/I18n.jsx";
import { STATUS } from "../../monitor/palette.js";
import { INTERVALS } from "../monitor/intervals.js";
import ForecastChart from "./ForecastChart.jsx";

const ACTION_COLOR = { buy: STATUS.buy, sell: STATUS.sell, hold: "var(--muted)" };
const ACTION_ICON = { buy: "▲", sell: "▼", hold: "■" };
const INDICATORS = ["ema_20", "ema_50", "ema_20_pct", "rsi_14", "macd_hist", "stoch_k", "adx_14", "plus_di", "minus_di", "atr_14", "bb_upper", "bb_lower", "vwap_20"];

const money = (value) => value.toFixed(2);

function Plan({ kind, trade, t }) {
  return (
    <div className={`plan-card ${kind}`}>
      <h4 style={{ color: ACTION_COLOR[kind] }}>
        {ACTION_ICON[kind]} {t(`forecast.${kind}.title`)}
      </h4>
      <div className="plan-price">
        {kind === "buy" ? "≤ " : "≥ "}
        {money(trade.price)}
      </div>
      <dl>
        <div>
          <dt>{t("forecast.stop")}</dt>
          <dd>{money(trade.stop)}</dd>
        </div>
        <div>
          <dt>{t("forecast.targets")}</dt>
          <dd>{trade.targets.map(money).join(" → ")}</dd>
        </div>
        <div>
          <dt>{t("forecast.reward")}</dt>
          <dd>1 : {trade.reward.toFixed(2)}</dd>
        </div>
      </dl>
      <p className="hint">{t(`forecast.${kind}.hint`)}</p>
    </div>
  );
}

function Warning({ name, onClose }) {
  const { t } = useI18n();

  return (
    <div className="warning-popup" role="alert">
      <span aria-hidden="true">⚠</span>
      <div>
        <strong>{t("forecast.notTradingTitle")}</strong>
        <div>{t("forecast.notTrading", { name })}</div>
      </div>
      <button type="button" className="icon" onClick={onClose} aria-label={t("common.close")}>
        ×
      </button>
    </div>
  );
}

export default function ForecastTab() {
  const { t, language } = useI18n();
  const [contracts, setContracts] = useState(null);
  const [stock, setStock] = useState("");
  const [interval, setTimeframe] = useState("5m");
  const [state, setState] = useState({});
  const [warning, setWarning] = useState(false);

  useEffect(() => {
    let active = true;

    api
      .getContracts()
      .then((items) => {
        if (active) {
          setContracts(items);
          setStock((current) => current || items[0]?.metric.id || "");
        }
      })
      .catch((error) => active && setState({ error: error.message }));

    return () => {
      active = false;
    };
  }, []);

  const calculate = (event) => {
    event && event.preventDefault();

    if (!stock) {
      return;
    }

    setState((current) => ({ ...current, loading: true, error: null }));
    api
      .getForecast(stock, interval)
      .then((data) => {
        setState({ data });
        setWarning(!data.trading);
      })
      .catch((error) => setState({ error: error.message }));
  };

  if (!contracts) {
    return <p className={state.error ? "notice error" : "notice"}>{state.error || t("common.loading")}</p>;
  }

  const data = state.data;
  const forecast = data?.forecast;
  const time = new Intl.DateTimeFormat(language, { dateStyle: "medium", timeStyle: "short" });

  return (
    <div className="forecast">
      <form className="toolbar" onSubmit={calculate}>
        <select value={stock} onChange={(event) => setStock(event.target.value)} aria-label={t("alerts.contract")}>
          {contracts.map(({ metric, contract }) => (
            <option key={metric.id} value={metric.id}>
              {metric.id} {contract?.name ? `— ${contract.name}` : ""}
            </option>
          ))}
        </select>
        <select value={interval} onChange={(event) => setTimeframe(event.target.value)} aria-label={t("monitor.interval")}>
          {INTERVALS.map((value) => (
            <option key={value} value={value}>
              {t(`interval.${value}`)}
            </option>
          ))}
        </select>
        <button type="submit" disabled={!stock || state.loading}>
          {t("forecast.calculate")}
        </button>
      </form>
      {contracts.length === 0 && <p className="hint">{t("contracts.none")}</p>}
      {state.error && <p className="notice error">{state.error}</p>}
      {!data && !state.error && contracts.length > 0 && <p className="hint">{t("forecast.hint")}</p>}
      {forecast && (
        <div className={state.loading ? "refreshing" : ""}>
          <section className="forecast-summary">
            <div className="recommendation" style={{ borderColor: ACTION_COLOR[forecast.action] }}>
              <span className="muted">{t("forecast.now")}</span>
              <strong style={{ color: ACTION_COLOR[forecast.action] }}>
                {ACTION_ICON[forecast.action]} {t(`forecast.action.${forecast.action}`)}
              </strong>
              <span>
                {t("forecast.score")} {forecast.score > 0 ? "+" : ""}
                {forecast.score.toFixed(2)}
              </span>
            </div>
            <div>
              <div className="forecast-price">{money(forecast.price)}</div>
              <div className="muted">
                {data.stock}
                {data.name ? ` — ${data.name}` : ""} · {t("forecast.asOf", { time: time.format(new Date(forecast.as_of)) })} · {t(`interval.${data.interval}`)}
              </div>
              <div>
                {t(`forecast.trend.${forecast.trend}`)} · {t(`forecast.strength.${forecast.strength}`)}
              </div>
            </div>
          </section>
          <section className="plans">
            <Plan kind="buy" trade={forecast.buy} t={t} />
            <div className="plan-card hold">
              <h4 style={{ color: ACTION_COLOR.hold }}>
                {ACTION_ICON.hold} {t("forecast.hold.title")}
              </h4>
              <div className="plan-price">
                {money(forecast.hold_from)} – {money(forecast.hold_to)}
              </div>
              <p className="hint">{t("forecast.hold.hint")}</p>
            </div>
            <Plan kind="sell" trade={forecast.sell} t={t} />
          </section>
          <ForecastChart candles={data.candles} forecast={forecast} />
          <section className="forecast-details">
            <div>
              <h4>{t("forecast.reasons")}</h4>
              <ul className="reasons">
                {forecast.reasons.map((reason, index) => (
                  <li key={`${reason.code}${index}`}>
                    <span className={reason.vote > 0 ? "up" : reason.vote < 0 ? "down" : "muted"}>
                      {reason.vote > 0 ? "▲" : reason.vote < 0 ? "▼" : "•"}
                    </span>{" "}
                    {t(`reason.${reason.code}`, { value: reason.value, second: reason.second })}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h4>{t("forecast.levels")}</h4>
              <dl className="levels">
                <div>
                  <dt>{t("forecast.resistances")}</dt>
                  <dd>{forecast.resistances.map(money).join(", ")}</dd>
                </div>
                <div>
                  <dt>{t("forecast.supports")}</dt>
                  <dd>{forecast.supports.map(money).join(", ")}</dd>
                </div>
                {Object.entries(forecast.fibonacci)
                  .sort(([a], [b]) => Number(a) - Number(b))
                  .map(([level, price]) => (
                    <div key={level}>
                      <dt>{t("forecast.fibonacci", { level })}</dt>
                      <dd>{money(price)}</dd>
                    </div>
                  ))}
              </dl>
            </div>
            <div>
              <h4>{t("forecast.indicators")}</h4>
              <dl className="levels">
                {INDICATORS.filter((key) => forecast.indicators[key] !== undefined).map((key) => (
                  <div key={key}>
                    <dt>{t(`indicator.${key}`)}</dt>
                    <dd>{forecast.indicators[key].toFixed(2)}</dd>
                  </div>
                ))}
              </dl>
            </div>
          </section>
          <p className="hint disclaimer">{t("forecast.disclaimer")}</p>
        </div>
      )}
      {warning && data && <Warning name={data.name || data.stock} onClose={() => setWarning(false)} />}
    </div>
  );
}
