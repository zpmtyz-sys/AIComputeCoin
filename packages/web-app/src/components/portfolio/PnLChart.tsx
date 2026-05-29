"use client";

import { useEffect, useRef, useState } from "react";
import {
  createChart,
  type IChartApi,
  type LineData,
  type Time,
  ColorType,
} from "lightweight-charts";

const PERIODS = ["1D", "1W", "1M", "3M"] as const;
type Period = (typeof PERIODS)[number];

function generateEquityData(days: number): LineData<Time>[] {
  const data: LineData<Time>[] = [];
  let time = Math.floor(Date.now() / 1000) - days * 86400;
  let equity = 100000;

  for (let i = 0; i < days; i++) {
    equity += (Math.random() - 0.45) * 2000;
    data.push({
      time: time as Time,
      value: Math.round(equity * 100) / 100,
    });
    time += 86400;
  }
  return data;
}

function getDaysForPeriod(period: Period): number {
  switch (period) {
    case "1D":
      return 1;
    case "1W":
      return 7;
    case "1M":
      return 30;
    case "3M":
      return 90;
  }
}

export default function PnLChart() {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const [period, setPeriod] = useState<Period>("1M");

  useEffect(() => {
    if (!containerRef.current) return;

    if (chartRef.current) {
      chartRef.current.remove();
      chartRef.current = null;
    }

    const chart = createChart(containerRef.current, {
      width: containerRef.current.clientWidth,
      height: containerRef.current.clientHeight,
      layout: {
        background: { type: ColorType.Solid, color: "transparent" },
        textColor: "#e6edf3",
      },
      grid: {
        vertLines: { color: "#30363d" },
        horzLines: { color: "#30363d" },
      },
      rightPriceScale: { borderColor: "#30363d" },
      timeScale: { borderColor: "#30363d" },
    });

    chartRef.current = chart;

    const series = chart.addLineSeries({
      color: "#58a6ff",
      lineWidth: 2,
      priceLineVisible: false,
      lastValueVisible: true,
    });

    const days = getDaysForPeriod(period);
    series.setData(generateEquityData(days));
    chart.timeScale().fitContent();

    const resizeObserver = new ResizeObserver(() => {
      if (chartRef.current && containerRef.current) {
        chartRef.current.applyOptions({
          width: containerRef.current.clientWidth,
          height: containerRef.current.clientHeight,
        });
      }
    });
    resizeObserver.observe(containerRef.current);

    return () => {
      resizeObserver.disconnect();
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
    };
  }, [period]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-semibold">Equity Curve</h3>
        <div className="flex gap-1">
          {PERIODS.map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-2 py-1 text-xs rounded transition-colors ${
                p === period
                  ? "bg-accent/20 text-accent"
                  : "text-gray-400 hover:text-foreground hover:bg-card-bg-hover"
              }`}
            >
              {p}
            </button>
          ))}
        </div>
      </div>
      <div ref={containerRef} className="flex-1 min-h-[200px]" />
    </div>
  );
}
