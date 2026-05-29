"use client";

import { useState, useMemo } from "react";

interface OrderLevel {
  price: number;
  quantity: number;
  total: number;
}

interface OrderBookProps {
  onPriceClick?: (price: number) => void;
}

const GROUPING_OPTIONS = [0.01, 0.1, 1, 10];

function generateMockAsks(): OrderLevel[] {
  const levels: OrderLevel[] = [];
  let price = 45050;
  let cumTotal = 0;
  for (let i = 0; i < 15; i++) {
    const quantity = Math.round((Math.random() * 5 + 0.1) * 1000) / 1000;
    cumTotal += quantity;
    levels.push({ price, quantity, total: Math.round(cumTotal * 1000) / 1000 });
    price += Math.round(Math.random() * 20 + 5);
  }
  return levels;
}

function generateMockBids(): OrderLevel[] {
  const levels: OrderLevel[] = [];
  let price = 45000;
  let cumTotal = 0;
  for (let i = 0; i < 15; i++) {
    const quantity = Math.round((Math.random() * 5 + 0.1) * 1000) / 1000;
    cumTotal += quantity;
    levels.push({ price, quantity, total: Math.round(cumTotal * 1000) / 1000 });
    price -= Math.round(Math.random() * 20 + 5);
  }
  return levels;
}

export default function OrderBook({ onPriceClick }: OrderBookProps) {
  const [grouping, setGrouping] = useState(1);

  const asks = useMemo(() => generateMockAsks(), []);
  const bids = useMemo(() => generateMockBids(), []);

  const maxTotal = Math.max(
    asks[asks.length - 1]?.total ?? 0,
    bids[bids.length - 1]?.total ?? 0
  );

  const spread = asks[0] ? asks[0].price - bids[0].price : 0;
  const spreadPct = bids[0] ? ((spread / bids[0].price) * 100).toFixed(3) : "0";

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-semibold">Order Book</h3>
        <select
          value={grouping}
          onChange={(e) => setGrouping(Number(e.target.value))}
          className="text-xs bg-background border border-border rounded px-2 py-1 text-foreground"
        >
          {GROUPING_OPTIONS.map((g) => (
            <option key={g} value={g}>
              {g}
            </option>
          ))}
        </select>
      </div>

      {/* Header */}
      <div className="grid grid-cols-3 text-xs text-gray-400 px-2 py-1 border-b border-border">
        <span>Price</span>
        <span className="text-right">Qty</span>
        <span className="text-right">Total</span>
      </div>

      {/* Asks - reversed so lowest ask is at bottom */}
      <div className="flex-1 overflow-hidden flex flex-col justify-end">
        {[...asks].reverse().map((level, i) => (
          <button
            key={`ask-${i}`}
            onClick={() => onPriceClick?.(level.price)}
            className="grid grid-cols-3 text-xs px-2 py-0.5 relative hover:bg-card-bg-hover transition-colors w-full text-left"
          >
            <div
              className="absolute inset-0 bg-negative/10"
              style={{ width: `${(level.total / maxTotal) * 100}%`, right: 0, left: "auto" }}
            />
            <span className="font-mono text-negative relative z-10">
              {level.price.toFixed(2)}
            </span>
            <span className="font-mono text-right relative z-10">
              {level.quantity.toFixed(3)}
            </span>
            <span className="font-mono text-right text-gray-400 relative z-10">
              {level.total.toFixed(3)}
            </span>
          </button>
        ))}
      </div>

      {/* Spread */}
      <div className="flex items-center justify-center py-1.5 border-y border-border text-xs">
        <span className="font-mono text-foreground">{spread.toFixed(2)}</span>
        <span className="text-gray-400 ml-2">({spreadPct}%)</span>
      </div>

      {/* Bids */}
      <div className="flex-1 overflow-hidden">
        {bids.map((level, i) => (
          <button
            key={`bid-${i}`}
            onClick={() => onPriceClick?.(level.price)}
            className="grid grid-cols-3 text-xs px-2 py-0.5 relative hover:bg-card-bg-hover transition-colors w-full text-left"
          >
            <div
              className="absolute inset-0 bg-positive/10"
              style={{ width: `${(level.total / maxTotal) * 100}%`, right: 0, left: "auto" }}
            />
            <span className="font-mono text-positive relative z-10">
              {level.price.toFixed(2)}
            </span>
            <span className="font-mono text-right relative z-10">
              {level.quantity.toFixed(3)}
            </span>
            <span className="font-mono text-right text-gray-400 relative z-10">
              {level.total.toFixed(3)}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
