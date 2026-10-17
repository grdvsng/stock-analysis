import { useMemo, useState } from "react";
import { useI18n } from "../../i18n/I18n.jsx";

const GROUPS = ["trend", "momentum", "volatility", "volume", "pattern", "fibonacci"];

export default function NotificationPicker({ catalog, selected, onChange }) {
  const { t } = useI18n();
  const [filter, setFilter] = useState("");
  const chosen = new Set(selected);
  const needle = filter.trim().toLowerCase();

  const groups = useMemo(
    () =>
      GROUPS.map((group) => ({
        group,
        items: catalog.filter(
          (item) =>
            item.group === group &&
            (!needle || item.name.toLowerCase().includes(needle) || item.key.includes(needle) || item.description.toLowerCase().includes(needle)),
        ),
      })).filter((group) => group.items.length > 0),
    [catalog, needle],
  );

  const toggle = (key) => {
    const next = new Set(chosen);

    next.has(key) ? next.delete(key) : next.add(key);
    onChange(catalog.map((item) => item.key).filter((item) => next.has(item)));
  };

  const setGroup = (items, on) => {
    const next = new Set(chosen);

    items.forEach((item) => (on ? next.add(item.key) : next.delete(item.key)));
    onChange(catalog.map((item) => item.key).filter((item) => next.has(item)));
  };

  return (
    <div className="picker">
      <div className="toolbar">
        <input type="search" value={filter} placeholder={t("notifications.filter")} onChange={(event) => setFilter(event.target.value)} />
        <span className="hint">{t("notifications.selected", { count: selected.length, total: catalog.length })}</span>
      </div>
      {groups.map(({ group, items }) => {
        const all = items.every((item) => chosen.has(item.key));

        return (
          <fieldset key={group} className="picker-group">
            <legend>
              <label>
                <input type="checkbox" checked={all} onChange={() => setGroup(items, !all)} /> {t(`group.${group}`)} ({items.length})
              </label>
            </legend>
            {items.map((item) => (
              <label key={item.key} className="picker-item" title={item.description}>
                <input type="checkbox" checked={chosen.has(item.key)} onChange={() => toggle(item.key)} />
                <span>{item.name}</span>
              </label>
            ))}
          </fieldset>
        );
      })}
    </div>
  );
}
