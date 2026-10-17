import { Fragment, useEffect, useState } from "react";
import { api } from "../api.js";
import { useI18n } from "../i18n/I18n.jsx";
import ContractDetails from "./ContractDetails.jsx";
import SubscriptionForm from "./SubscriptionForm.jsx";

function Card({ symbol }) {
  const { t } = useI18n();
  const [state, setState] = useState({ loading: true });
  const [subscribing, setSubscribing] = useState(false);

  useEffect(() => {
    let active = true;

    api
      .getContract(symbol)
      .then((data) => active && setState({ data }))
      .catch((error) => active && setState({ error: error.message }));

    return () => {
      active = false;
    };
  }, [symbol]);

  if (state.loading) {
    return <p className="notice">{t("common.loading")}</p>;
  }

  if (state.error) {
    return <p className="notice error">{state.error}</p>;
  }

  const { contract, metric } = state.data;

  const subscribe = (settings) =>
    api.subscribe(symbol, settings).then((saved) => {
      setSubscribing(false);
      setState({ data: { contract, metric: saved } });
    });

  return (
    <div>
      <ContractDetails contract={contract} />
      {metric ? (
        <p className="notice success">{t("subscription.already")}</p>
      ) : subscribing ? (
        <SubscriptionForm submitLabel={t("subscription.subscribe")} onSubmit={subscribe} onCancel={() => setSubscribing(false)} />
      ) : (
        <button type="button" onClick={() => setSubscribing(true)}>
          {t("subscription.subscribe")}
        </button>
      )}
    </div>
  );
}

export default function SearchTab() {
  const { t } = useI18n();
  const [text, setText] = useState("");
  const [results, setResults] = useState(null);
  const [status, setStatus] = useState(null);
  const [open, setOpen] = useState(null);

  const search = (event) => {
    event.preventDefault();

    if (!text.trim()) {
      return;
    }

    setStatus({ state: "loading" });
    setOpen(null);
    api
      .search(text.trim())
      .then((matches) => {
        setResults(matches);
        setStatus(null);
      })
      .catch((error) => setStatus({ state: "error", message: error.message }));
  };

  return (
    <div className="search">
      <form className="toolbar" onSubmit={search}>
        <input type="search" value={text} placeholder={t("search.placeholder")} onChange={(event) => setText(event.target.value)} />
        <button type="submit">{t("search.submit")}</button>
      </form>
      {status && status.state === "loading" && <p className="notice">{t("common.loading")}</p>}
      {status && status.state === "error" && <p className="notice error">{status.message}</p>}
      {!results && !status && <p className="hint">{t("search.hint")}</p>}
      {results && results.length === 0 && <p className="hint">{t("search.none")}</p>}
      {results && results.length > 0 && (
        <table>
          <thead>
            <tr>
              <th>{t("contracts.name")}</th>
              <th>{t("contracts.stock")}</th>
              <th>{t("contracts.country")}</th>
              <th>{t("contracts.exchange")}</th>
              <th>{t("contracts.currency")}</th>
            </tr>
          </thead>
          <tbody>
            {results.map((match) => {
              const key = `${match.symbol}:${match.primary_exchange}:${match.currency}`;

              return (
                <Fragment key={key}>
                  <tr className="row" onClick={() => setOpen(open === key ? null : key)}>
                    <td>{match.name}</td>
                    <td>{match.symbol}</td>
                    <td>{match.country || t("common.empty")}</td>
                    <td>{match.primary_exchange}</td>
                    <td>{match.currency}</td>
                  </tr>
                  {open === key && (
                    <tr className="card">
                      <td colSpan={5}>
                        <Card symbol={match.symbol} />
                      </td>
                    </tr>
                  )}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}
