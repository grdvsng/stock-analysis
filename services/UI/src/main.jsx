import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { I18nProvider } from "./i18n/I18n.jsx";
import { AlertsProvider } from "./alerts/Alerts.jsx";
import App from "./App.jsx";
import "./styles.css";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <I18nProvider>
      <AlertsProvider>
        <App />
      </AlertsProvider>
    </I18nProvider>
  </StrictMode>,
);
