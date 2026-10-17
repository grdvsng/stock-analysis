import { useI18n } from "../i18n/I18n.jsx";

export const USER_TIME_ZONE = Intl.DateTimeFormat().resolvedOptions().timeZone;

function formatSessions(schedule, language) {
  if (!schedule || !Array.isArray(schedule.sessions)) {
    return [];
  }

  const day = new Intl.DateTimeFormat(language, { weekday: "short", month: "short", day: "numeric" });
  const time = new Intl.DateTimeFormat(language, { hour: "2-digit", minute: "2-digit" });

  return schedule.sessions.map((session) => {
    const start = new Date(session.start);
    const end = new Date(session.end);
    const sameDay = start.toDateString() === end.toDateString();

    return `${day.format(start)}: ${time.format(start)} – ${sameDay ? "" : `${day.format(end)} `}${time.format(end)}`;
  });
}

function Sessions({ title, schedule }) {
  const { t, language } = useI18n();
  const sessions = formatSessions(schedule, language);

  return (
    <div className="sessions">
      <h4>{title}</h4>
      <p className="hint">{t("contract.inYourTimeZone", { zone: USER_TIME_ZONE })}</p>
      {sessions.length === 0 ? (
        <p className="hint">{t("contract.noSessions")}</p>
      ) : (
        <ul>
          {sessions.map((session) => (
            <li key={session}>{session}</li>
          ))}
        </ul>
      )}
    </div>
  );
}

export default function ContractDetails({ contract }) {
  const { t } = useI18n();
  const empty = t("common.empty");
  const rows = [
    ["contracts.name", contract.name],
    ["contract.symbol", contract.stock && contract.stock.id],
    ["contract.market", contract.id],
    ["contracts.exchange", contract.exchange],
    ["contract.primary_exchange", contract.primary_exchange],
    ["contracts.country", contract.country],
    ["contracts.currency", contract.currency],
    ["contract.industry", contract.industry],
    ["contract.category", contract.category],
    ["contract.sub_category", contract.sub_category],
    ["contract.type", contract.stock && contract.stock.type],
    ["contract.min_tick", contract.stock && contract.stock.min_tick],
    ["contract.time_zone", contract.trading_hours && contract.trading_hours.time_zone],
  ];

  return (
    <div className="details">
      <dl>
        {rows.map(([label, value]) => (
          <div key={label}>
            <dt>{t(label)}</dt>
            <dd>{value || empty}</dd>
          </div>
        ))}
      </dl>
      <div className="schedules">
        <Sessions title={t("contract.trading_hours")} schedule={contract.trading_hours} />
        <Sessions title={t("contract.liquid_hours")} schedule={contract.liquid_hours} />
      </div>
    </div>
  );
}
