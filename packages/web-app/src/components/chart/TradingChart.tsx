"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import {
  createChart,
  type IChartApi,
  type ISeriesApi,
  type CandlestickData,
  type HistogramData,
  type LineData,
  type Time,
  CrosshairMode,
  ColorType,
} from "lightweight-charts";

const INTERVALS = ["1m", "5m", "15m", "1h", "4h", "1D", "1W"] as const;
type Interval = (typeof INTERVALS)[number];

function generateMockCandles(count: number): CandlestickData<Time>[] {
  const data: CandlestickData<Time>[] = [];
  let time = Math.floor(Date.now() / 1000) - count * 3600;
  let close = 45000;

  for (let i = 0; i < count; i++) {
    const open = close;
    const change = (Math.random() - 0.48) * 500;
    close = open + change;
    const high = Math.max(open, close) + Math.random() * 200;
    const low = Math.min(open, close) - Math.random() * 200;

    data.push({
      time: time as Time,
      open: Math.round(open * 100) / 100,
      high: Math.round(high * 100) / 100,
      low: Math.round(low * 100) / 100,
      close: Math.round(close * 100) / 100,
    });
    time += 3600;
  }
  return data;
}

function generateVolumeData(candles: CandlestickData<Time>[]): HistogramData<Time>[] {
  return candles.map((c) => ({
    time: c.time,
    value: Math.round(Math.random() * 1000 + 200),
    color: c.close >= c.open ? "rgba(63,185,80,0.3)" : "rgba(248,81,73,0.3)",
  }));
}

function calculateMA(candles: CandlestickData<Time>[], period: number): LineData<Time>[] {
  const result: LineData<Time>[] = [];
  for (let i = period - 1; i < candles.length; i++) {
    let sum = 0;
    for (let j = 0; j < period; j++) {
      sum += candles[i - j].close;
    }
    result.push({
      time: candles[i].time,
      value: Math.round((sum / period) * 100) / 100,
    });
  }
  return result;
}

interface TooltipData {
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export default function TradingChart() {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candleSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const [selectedInterval, setSelectedInterval] = useState<Interval>("1h");
  const [tooltip, setTooltip] = useState<TooltipData | null>(null);

  const initChart = useCallback(() => {
    if (!chartContainerRef.current) return;

    // Clean up existing chart
    if (chartRef.current) {
      chartRef.current.remove();
      chartRef.current = null;
    }

    const container = chartContainerRef.current;
    const chart = createChart(container, {
      width: container.clientWidth,
      height: container.clientHeight,
      layout: {
        background: { type: ColorType.Solid, color: "transparent" },
        textColor: "#e6edf3",
      },
      grid: {
        vertLines: { color: "#30363d" },
        horzLines: { color: "#30363d" },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
      },
      rightPriceScale: {
        borderColor: "#30363d",
      },
      timeScale: {
        borderColor: "#30363d",
        timeVisible: true,
      },
    });

    chartRef.current = chart;

    const candles = generateMockCandles(200);
    const volumeData = generateVolumeData(candles);

    const candleSeries = chart.addCandlestickSeries({
      upColor: "#3fb950",
      downColor: "#f85149",
      borderUpColor: "#3fb950",
      borderDownColor: "#f85149",
      wickUpColor: "#3fb950",
      wickDownColor: "#f85149",
    });
    candleSeries.setData(candles);
    candleSeriesRef.current = candleSeries;

    const volumeSeries = chart.addHistogramSeries({
      priceFormat: { type: "volume" },
      priceScaleId: "",
    });
    volumeSeries.priceScale().applyOptions({
      scaleMargins: { top: 0.8, bottom: 0 },
    });
    volumeSeries.setData(volumeData);

    // MA 7
    const ma7Series = chart.addLineSeries({
      color: "#f0b90b",
      lineWidth: 1,
      priceLineVisible: false,
      lastValueVisible: false,
    });
    ma7Series.setData(calculateMA(candles, 7));

    // MA 25
    const ma25Series = chart.addLineSeries({
      color: "#8b5cf6",
      lineWidth: 1,
      priceLineVisible: false,
      lastValueVisible: false,
    });
    ma25Series.setData(calculateMA(candles, 25));

    chart.subscribeCrosshairMove((param) => {
      if (!param.time || !param.seriesData) {
        setTooltip(null);
        return;
      }
      const candleData = param.seriesData.get(candleSeries) as CandlestickData<Time> | undefined;
      const volData = param.seriesData.get(volumeSeries) as HistogramData<Time> | undefined;
      if (candleData) {
        setTooltip({
          open: candleData.open,
          high: candleData.high,
          low: candleData.low,
          close: candleData.close,
          volume: volData?.value ?? 0,
        });
      }
    });

    chart.timeScale().fitContent();
  }, []);

  useEffect(() => {
    initChart();

    const resizeObserver = new ResizeObserver(() => {
      if (chartRef.current && chartContainerRef.current) {
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        });
      }
    });

    if (chartContainerRef.current) {
      resizeObserver.observe(chartContainerRef.current);
    }

    return () => {
      resizeObserver.disconnect();
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
      }
    };
  }, [initChart]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-2 px-1">
        <div className="flex items-center gap-2">
          {tooltip && (
            <div className="flex gap-3 text-xs font-mono">
              <span>O: <span className="text-foreground">{tooltip.open.toFixed(2)}</span></span>
              <span>H: <span className="text-positive">{tooltip.high.toFixed(2)}</span></span>
              <span>L: <span className="text-negative">{tooltip.low.toFixed(2)}</span></span>
              <span>C: <span className="text-foreground">{tooltip.close.toFixed(2)}</span></span>
              <span>V: <span className="text-gray-400">{tooltip.volume.toFixed(0)}</span></span>
            </div>
          )}
        </div>
        <div className="flex gap-1">
          {INTERVALS.map((interval) => (
            <button
              key={interval}
              onClick={() => setSelectedInterval(interval)}
              className={`px-2 py-1 text-xs rounded transition-colors ${
                interval === selectedInterval
                  ? "bg-accent/20 text-accent"
                  : "text-gray-400 hover:text-foreground hover:bg-card-bg-hover"
              }`}
            >
              {interval}
            </button>
          ))}
        </div>
      </div>
      <div ref={chartContainerRef} className="flex-1 min-h-0" />
    </div>
  );
}
