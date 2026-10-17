export const LOG_LEVELS = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL", "DISABLED"];
export const SOURCE_TYPES = ["ibk"];
export const INTERVAL_UNITS = ["Second", "Minute", "Hour"];
export const TTL_UNITS = ["Minute", "Hour", "Day"];

export const CONFIG_SECTIONS = [
  {
    key: "general",
    fields: [
      { path: ["workers"], label: "config.workers", type: "number", min: 1 },
      { path: ["update_interval"], label: "config.update_interval", type: "duration", units: INTERVAL_UNITS },
    ],
  },
  {
    key: "source",
    fields: [
      { path: ["source", "type"], label: "config.source.type", type: "select", options: SOURCE_TYPES },
      { path: ["source", "host"], label: "config.source.host", type: "text" },
      { path: ["source", "port"], label: "config.source.port", type: "number", min: 1 },
      { path: ["source", "timeout"], label: "config.source.timeout", type: "number", min: 1 },
      { path: ["source", "client_id"], label: "config.source.client_id", type: "number", min: 0 },
    ],
  },
  {
    key: "storage",
    fields: [
      { path: ["storage", "host"], label: "config.storage.host", type: "text" },
      { path: ["storage", "port"], label: "config.storage.port", type: "number", min: 1 },
      { path: ["storage", "timeout"], label: "config.storage.timeout", type: "number", min: 1 },
    ],
  },
  {
    key: "logging",
    fields: [
      { path: ["logging", "level"], label: "config.logging.level", type: "select", options: LOG_LEVELS },
      { path: ["logging", "path"], label: "config.logging.path", type: "text" },
    ],
  },
  {
    key: "ws",
    fields: [
      { path: ["ws", "host"], label: "config.ws.host", type: "text" },
      { path: ["ws", "port"], label: "config.ws.port", type: "number", min: 1 },
    ],
  },
];

export const KNOWN_KEYS = new Set(CONFIG_SECTIONS.flatMap((section) => section.fields.map((field) => field.path[0])));

export function getAt(object, path) {
  return path.reduce((value, key) => (value == null ? undefined : value[key]), object);
}

export function setAt(object, path, value) {
  const [key, ...rest] = path;
  const copy = { ...(object || {}) };

  copy[key] = rest.length === 0 ? value : setAt(copy[key], rest, value);

  return copy;
}
