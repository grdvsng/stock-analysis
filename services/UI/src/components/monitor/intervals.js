export const INTERVALS = ["1m", "5m", "15m", "30m", "1h", "1d"];

export const LIFETIME_UNITS = ["Second", "Minute", "Hour"];

export const INTERVAL_MS = { "1m": 60e3, "5m": 300e3, "15m": 900e3, "30m": 1800e3, "1h": 3600e3, "1d": 86400e3 };

export function isActive(signal, moment) {
  return signal.time <= moment && moment < signal.expires;
}

export const RANGES = [
  { key: "1d", days: 1 },
  { key: "5d", days: 5 },
  { key: "1mo", days: 30 },
  { key: "3mo", days: 90 },
];
