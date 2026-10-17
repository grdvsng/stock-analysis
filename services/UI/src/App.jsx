import { useState } from "react";
import { dictionaries, useI18n } from "./i18n/I18n.jsx";
import ConfigTab from "./components/ConfigTab.jsx";
import ContractsTab from "./components/ContractsTab.jsx";
import SearchTab from "./components/SearchTab.jsx";
import MonitorsTab from "./components/monitor/MonitorsTab.jsx";
import ForecastTab from "./components/forecast/ForecastTab.jsx";

const TABS = {
  monitors: MonitorsTab,
  forecast: ForecastTab,
  contracts: ContractsTab,
  search: SearchTab,
  config: ConfigTab,
};

export default function App() {
  const { t, language, setLanguage } = useI18n();
  const [tab, setTab] = useState("monitors");
  const Current = TABS[tab];

  return (
    <div className="app">
      <header>
        <h1>{t("app.title")}</h1>
        <nav>
          {Object.keys(TABS).map((key) => (
            <button key={key} type="button" className={key === tab ? "active" : ""} onClick={() => setTab(key)}>
              {t(`tab.${key}`)}
            </button>
          ))}
        </nav>
        <select value={language} onChange={(event) => setLanguage(event.target.value)}>
          {Object.entries(dictionaries).map(([code, dictionary]) => (
            <option key={code} value={code}>
              {dictionary["language.name"]}
            </option>
          ))}
        </select>
      </header>
      <main>
        <Current key={tab} />
      </main>
    </div>
  );
}
