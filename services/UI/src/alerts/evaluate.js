export const CONDITIONS = ["any", "buy", "sell"];

export function evaluateAlert(alert, signals, moment) {
  const latest = new Map();

  for (const signal of signals) {
    if (alert.metrics.includes(signal.metric) && signal.time <= moment && (!latest.has(signal.metric) || signal.time > latest.get(signal.metric).time)) {
      latest.set(signal.metric, signal);
    }
  }

  const matched = alert.metrics.map((metric) => latest.get(metric));

  if (matched.some((signal) => !signal || moment >= signal.expires)) {
    return null;
  }

  const direction = matched[0].action;

  if (matched.some((signal) => signal.action !== direction)) {
    return null;
  }

  if (alert.condition !== "any" && alert.condition !== direction) {
    return null;
  }

  return {
    direction,
    signals: matched,
    episode: matched.map((signal) => `${signal.metric}@${signal.time}`).sort().join("|"),
  };
}
