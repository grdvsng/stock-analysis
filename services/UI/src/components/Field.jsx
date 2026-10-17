import DurationInput from "./DurationInput.jsx";

export default function Field({ label, field, value, onChange }) {
  let input;

  if (field.type === "select") {
    input = (
      <select value={value ?? ""} onChange={(event) => onChange(event.target.value)}>
        {field.options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    );
  } else if (field.type === "duration") {
    input = <DurationInput value={value} units={field.units} onChange={onChange} />;
  } else if (field.type === "number") {
    input = (
      <input
        type="number"
        min={field.min}
        value={value ?? ""}
        onChange={(event) => onChange(event.target.value === "" ? undefined : Number(event.target.value))}
      />
    );
  } else {
    input = <input type="text" value={value ?? ""} onChange={(event) => onChange(event.target.value)} />;
  }

  return (
    <label className="field">
      <span>{label}</span>
      {input}
    </label>
  );
}
