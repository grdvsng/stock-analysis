export const SERIES = {
  light: ["#2a78d6", "#eb6834", "#1baf7a", "#eda100", "#e87ba4", "#008300", "#6250d6", "#e34948"],
  dark: ["#3987e5", "#d95926", "#199e70", "#c98500", "#d55181", "#008300", "#9085e9", "#e66767"],
};

export const STATUS = { buy: "#0ca30c", sell: "#d03b3b" };

export const CANDLE = { up: STATUS.buy, down: STATUS.sell };

export const CHROME = {
  light: { surface: "#fcfcfb", text: "#52514e", muted: "#898781", grid: "#e1e0d9", axis: "#c3c2b7" },
  dark: { surface: "#1a1a19", text: "#c3c2b7", muted: "#898781", grid: "#2c2c2a", axis: "#383835" },
};

export const MAX_LINES = SERIES.light.length;

export function assignSlots(previous, stocks) {
  const slots = {};
  const used = new Set();

  for (const stock of stocks) {
    if (previous[stock] !== undefined && !used.has(previous[stock])) {
      slots[stock] = previous[stock];
      used.add(previous[stock]);
    }
  }

  for (const stock of stocks) {
    if (slots[stock] !== undefined) {
      continue;
    }

    for (let slot = 0; slot < MAX_LINES; slot++) {
      if (!used.has(slot)) {
        slots[stock] = slot;
        used.add(slot);
        break;
      }
    }
  }

  return slots;
}
