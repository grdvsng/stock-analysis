async function request(method, path, body) {
  const options = { method, headers: {} };

  if (body !== undefined) {
    options.headers["Content-Type"] = "application/json";
    options.body = typeof body === "string" ? body : JSON.stringify(body);
  }

  const response = await fetch(path, options);
  const text = await response.text();
  const data = text ? JSON.parse(text) : null;

  if (!response.ok) {
    throw new Error((data && data.error) || response.statusText);
  }

  return data;
}

const query = (id) => `?id=${encodeURIComponent(id)}`;

export const api = {
  getConfig: () => request("GET", "/config"),
  saveConfig: (config) => request("POST", "/config", config),
  getContracts: () => request("GET", "/contracts"),
  getContract: (id) => request("GET", "/contract" + query(id)),
  subscribe: (id, metric) => request("POST", "/contract" + query(id), metric),
  unsubscribe: (id) => request("DELETE", "/contract" + query(id)),
  search: (text) => request("GET", `/contracts/search?query=${encodeURIComponent(text)}`),
  getSettings: () => request("GET", "/settings"),
  getIndicators: () => request("GET", "/indicators"),
  getForecast: (id, interval) => request("GET", `/forecast?id=${encodeURIComponent(id)}&interval=${encodeURIComponent(interval)}`),
  getMonitors: () => request("GET", "/monitors"),
  createMonitor: (monitor) => request("POST", "/monitors", monitor),
  updateMonitor: (id, monitor) => request("PUT", "/monitor" + query(id), monitor),
  deleteMonitor: (id) => request("DELETE", "/monitor" + query(id)),
};

export function monitorSocketUrl(port, params) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";

  return `${protocol}//${window.location.hostname}:${port}/monitor?${params}`;
}
