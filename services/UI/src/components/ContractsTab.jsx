import { Fragment, useEffect, useState } from "react";
import { api } from "../api.js";
import { useI18n } from "../i18n/I18n.jsx";
import ContractDetails from "./ContractDetails.jsx";
import SubscriptionForm from "./SubscriptionForm.jsx";

function formatDate(value, language) {
  if (!value) {
    return "";
  }

  return new Intl.DateTimeFormat(language, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}

export default function ContractsTab() {
  const { t, language } = useI18n();
  const [items, setItems] = useState(null);
  const [open, setOpen] = useState(null);
  const [status, setStatus] = useState(null);

  const load = () => {
    api
      .getContracts()
      .then((contracts) => setItems(contracts))
      .catch((error) => setStatus({ state: "error", message: error.message }));
  };

  useEffect(load, []);

  const update = (id) => (metric) =>
    api.subscribe(id, metric).then(() => {
      setStatus({ state: "success", message: t("contracts.updated") });
      load();
    });

  const unsubscribe = (id) => {
    if (!window.confirm(t("contracts.unsubscribeConfirm", { id }))) {
      return;
    }

    api
      .unsubscribe(id)
      .then(() => {
        setOpen(null);
        load();
      })
      .catch((error) => setStatus({ state: "error", message: error.message }));
  };

  if (!items) {
    return <p className={status ? "notice error" : "notice"}>{status ? status.message : t("common.loading")}</p>;
  }

  return (
    <div className="contracts">
      <div className="toolbar">
        <button type="button" onClick={load}>
          {t("common.refresh")}
        </button>
      </div>
      {status && <p className={`notice ${status.state}`}>{status.message}</p>}
      {items.length === 0 ? (
        <p className="hint">{t("contracts.none")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("contracts.added")}</th>
              <th>{t("contracts.name")}</th>
              <th>{t("contracts.stock")}</th>
              <th>{t("contracts.country")}</th>
            </tr>
          </thead>
          <tbody>
            {items.map(({ metric, contract }) => (
              <Fragment key={metric.id}>
                <tr className="row" onClick={() => setOpen(open === metric.id ? null : metric.id)}>
                  <td>{formatDate(metric.created_at, language)}</td>
                  <td>{contract ? contract.name : <em>{t("contracts.pending")}</em>}</td>
                  <td>{metric.id}</td>
                  <td>{(contract && contract.country) || t("common.empty")}</td>
                </tr>
                {open === metric.id && (
                  <tr className="card">
                    <td colSpan={4}>
                      {contract && <ContractDetails contract={contract} />}
                      <SubscriptionForm initial={metric} submitLabel={t("common.save")} onSubmit={update(metric.id)} />
                      <button type="button" className="danger" onClick={() => unsubscribe(metric.id)}>
                        {t("contracts.unsubscribe")}
                      </button>
                    </td>
                  </tr>
                )}
              </Fragment>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
