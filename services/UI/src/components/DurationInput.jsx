import { useI18n } from "../i18n/I18n.jsx";

const PATTERN = /^\s*(\d+)\s*([A-Za-z]+?)s?\s*$/;

export function parseDuration(value, units) {
  const match = typeof value === "string" ? value.match(PATTERN) : null;

  if (!match || !units.includes(match[2])) {
    return { amount: "", unit: units[0] };
  }

  return { amount: match[1], unit: match[2] };
}

export function formatDuration(amount, unit) {
  if (amount === "" || amount == null) {
    return null;
  }

  return `${amount} ${unit}${Number(amount) === 1 ? "" : "s"}`;
}

export default function DurationInput({ value, units, onChange, placeholder }) {
  const { t } = useI18n();
  const { amount, unit } = parseDuration(value, units);

  return (
    <div className="duration">
      <input
        type="number"
        min="1"
        value={amount}
        placeholder={placeholder}
        onChange={(event) => onChange(formatDuration(event.target.value, unit))}
      />
      <select value={unit} onChange={(event) => onChange(formatDuration(amount, event.target.value))}>
        {units.map((option) => (
          <option key={option} value={option}>
            {t(`unit.${option}`)}
          </option>
        ))}
      </select>
    </div>
  );
}
