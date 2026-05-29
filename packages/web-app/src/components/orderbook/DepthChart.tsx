"use client";

import { useEffect, useRef } from "react";

interface DepthLevel {
  price: number;
  cumulative: number;
}

function generateMockDepth(): { bids: DepthLevel[]; asks: DepthLevel[] } {
  const bids: DepthLevel[] = [];
  const asks: DepthLevel[] = [];
  let bidPrice = 45000;
  let askPrice = 45050;
  let bidCum = 0;
  let askCum = 0;

  for (let i = 0; i < 30; i++) {
    bidCum += Math.random() * 3 + 0.5;
    bids.push({ price: bidPrice, cumulative: bidCum });
    bidPrice -= Math.random() * 30 + 10;

    askCum += Math.random() * 3 + 0.5;
    asks.push({ price: askPrice, cumulative: askCum });
    askPrice += Math.random() * 30 + 10;
  }

  return { bids, asks };
}

export default function DepthChart() {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const resizeCanvas = () => {
      const rect = canvas.parentElement?.getBoundingClientRect();
      if (rect) {
        canvas.width = rect.width;
        canvas.height = rect.height;
      }
      draw();
    };

    const draw = () => {
      const { bids, asks } = generateMockDepth();
      const w = canvas.width;
      const h = canvas.height;

      ctx.clearRect(0, 0, w, h);

      const allPrices = [...bids.map((b) => b.price), ...asks.map((a) => a.price)];
      const minPrice = Math.min(...allPrices);
      const maxPrice = Math.max(...allPrices);
      const priceRange = maxPrice - minPrice;

      const maxCum = Math.max(
        bids[bids.length - 1]?.cumulative ?? 0,
        asks[asks.length - 1]?.cumulative ?? 0
      );

      const toX = (price: number) => ((price - minPrice) / priceRange) * w;
      const toY = (cum: number) => h - (cum / maxCum) * h;

      // Draw bids area
      ctx.beginPath();
      ctx.moveTo(toX(bids[0].price), h);
      for (const bid of bids) {
        ctx.lineTo(toX(bid.price), toY(bid.cumulative));
      }
      ctx.lineTo(toX(bids[bids.length - 1].price), h);
      ctx.closePath();
      ctx.fillStyle = "rgba(63,185,80,0.15)";
      ctx.fill();
      ctx.strokeStyle = "#3fb950";
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      for (let i = 0; i < bids.length; i++) {
        const x = toX(bids[i].price);
        const y = toY(bids[i].cumulative);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      }
      ctx.stroke();

      // Draw asks area
      ctx.beginPath();
      ctx.moveTo(toX(asks[0].price), h);
      for (const ask of asks) {
        ctx.lineTo(toX(ask.price), toY(ask.cumulative));
      }
      ctx.lineTo(toX(asks[asks.length - 1].price), h);
      ctx.closePath();
      ctx.fillStyle = "rgba(248,81,73,0.15)";
      ctx.fill();
      ctx.strokeStyle = "#f85149";
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      for (let i = 0; i < asks.length; i++) {
        const x = toX(asks[i].price);
        const y = toY(asks[i].cumulative);
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      }
      ctx.stroke();
    };

    resizeCanvas();
    const observer = new ResizeObserver(resizeCanvas);
    if (canvas.parentElement) {
      observer.observe(canvas.parentElement);
    }

    return () => {
      observer.disconnect();
    };
  }, []);

  return (
    <div className="relative w-full h-full min-h-[120px]">
      <canvas ref={canvasRef} className="w-full h-full" />
    </div>
  );
}
