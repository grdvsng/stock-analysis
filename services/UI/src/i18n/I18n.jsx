import { createContext, useCallback, useContext, useMemo, useState } from "react";
import en from "./en.js";
import ru from "./ru.js";

export const dictionaries = { en, ru };

const STORAGE_KEY = "language";
const I18nContext = createContext(null);

function initialLanguage() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);

    if (stored && dictionaries[stored]) {
      return stored;
    }
  } catch {
    return "en";
  }

  return "en";
}

export function I18nProvider({ children }) {
  const [language, setLanguage] = useState(initialLanguage);

  const choose = useCallback((next) => {
    setLanguage(next);

    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {
      return;
    }
  }, []);

  const t = useCallback(
    (key, values = {}) => {
      const template = dictionaries[language][key] ?? dictionaries.en[key] ?? key;

      return template.replace(/\{(\w+)\}/g, (_, name) => values[name] ?? "");
    },
    [language],
  );

  const value = useMemo(() => ({ language, setLanguage: choose, t }), [language, choose, t]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  return useContext(I18nContext);
}
