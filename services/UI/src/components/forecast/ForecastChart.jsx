import { useEffect, useRef } from "react";
import { CandlestickSeries, createChart, CrosshairMode, LineStyle } from "lightweight-charts";
import { CANDLE, CHROME, STATUS } from "../../monitor/palette.js";
import useDarkMode from "../../monitor/useDarkMode.js";
import { useI18n } from "../../i18n/I18n.jsx";

export default function ForecastChart({ candles, forecast }) {
  const { t, language } = useI18n();
  const dark = useDarkMode();
  const container = useRef(null);

  useEffect(() => {
    const chrome = dark ? CHROME.dark : CHROME.light;
    const full = new Intl.DateTimeFormat(language, { dateStyle: "medium", timeStyle: "short" });
    const clock = new Intl.DateTimeFormat(language, { hour: "2-digit", minute: "2-digit" });
    const day = new Intl.DateTimeFormat(language, { month: "short", day: "numeric" });
    const chart = createChart(container.current, {
      autoSize: true,
      layout: { background: { color: chrome.surface }, textColor: chrome.muted, attributionLogo: true },
      grid: { vertLines: { visible: false }, horzLines: { color: chrome.grid } },
      rightPriceScale: { borderColor: chrome.axis },
      timeScale: { borderColor: chrome.axis, timeVisible: true, tickMarkFormatter: (time, type) => (type <= 2 ? day : clock).format(new Date(time * 1000)) },
      crosshair: { mode: CrosshairMode.Magnet },
      localization: { timeFormatter: (time) => full.format(new Date(time * 1000)) },
    });
    const series = chart.addSeries(CandlestickSeries, {
      upColor: "rgba(0, 0, 0, 0)",
      borderUpColor: CANDLE.up,
      wickUpColor: CANDLE.up,
      downColor: CANDLE.down,
      borderDownColor: CANDLE.down,
      wickDownColor: CANDLE.down,
      priceLineVisible: false,
    });

    series.setData(candles.map((candle) => ({ time: Math.floor(Date.parse(candle.time) / 1000), open: candle.open, high: candle.high, low: candle.low, close: candle.close })));

    const line = (price, color, style, title) => series.createPriceLine({ price, color, lineWidth: style === LineStyle.Solid ? 2 : 1, lineStyle: style, axisLabelVisible: true, title });

    line(forecast.buy.price, STATUS.buy, LineStyle.Solid, t("forecast.line.buy"));
    line(forecast.buy.stop, STATUS.buy, LineStyle.Dashed, t("forecast.line.buyStop"));
    forecast.buy.targets.slice(0, 1).forEach((price) => line(price, STATUS.buy, LineStyle.Dotted, t("forecast.line.target")));
    line(forecast.sell.price, STATUS.sell, LineStyle.Solid, t("forecast.line.sell"));
    line(forecast.sell.stop, STATUS.sell, LineStyle.Dashed, t("forecast.line.sellStop"));
    forecast.sell.targets.slice(0, 1).forEach((price) => line(price, STATUS.sell, LineStyle.Dotted, t("forecast.line.target")));
    chart.timeScale().fitContent();

    return () => chart.remove();
  }, [candles, forecast, dark, language]);

  return <div className="forecast-chart" ref={container} />;
}
