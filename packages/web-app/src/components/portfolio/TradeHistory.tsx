"use client";

import { useState, useMemo } from "react";

interface Trade {
  id: string;
  time: string;
  pair: string;
  side: "buy" | "sell";
  price: number;
  quantity: number;
  fee: number;
  realizedPnl: number;
}

const MOCK_TRADES: Trade[] = [
  { id: "t1", time: "2024-01-15 14:22:10", pair: "CU-PERP/USDT", side: "sell", price: 45100, quantity: 1.0, fee: 6.77, realizedPnl: 320.5 },
  { id: "t2", time: "2024-01-15 12:10:45", pair: "CU-PERP/USDT", side: "buy", price: 44780, quantity: 1.0, fee: 4.48, realizedPnl: 0 },
  { id: "t3", time: "2024-01-15 10:05:33", pair: "GPU-SPOT/USDT", side: "buy", price: 12450, quantity: 2.0, fee: 3.74, realizedPnl: 0 },
  { id: "t4", time: "2024-01-14 22:30:15", pair: "TPU-PERP/USDT", side: "sell", price: 9200, quantity: 3.0, fee: 4.14, realizedPnl: -150.0 },
  { id: "t5", time: "2024-01-14 18:45:02", pair: "CU-PERP/USDT", side: "buy", price: 44600, quantity: 2.5, fee: 5.01, realizedPnl: 0 },
  { id: "t6", time: "2024-01-14 15:20:48", pair: "GPU-SPOT/USDT", side: "sell", price: 12680, quantity: 1.5, fee: 2.85, realizedPnl: 285.0 },
];

type SortField = "time" | "pair" | "price" | "quantity" | "realizedPnl";

export default function TradeHistory() {
  const [filterPair, setFilterPair] = useState<string>("all");
  const [sortField, setSortField] = useState<SortField>("time");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");

  const pairs = useMemo(() => {
    const uniquePairs = [...new Set(MOCK_TRADES.map((t) => t.pair))];
    return ["all", ...uniquePairs];
  }, []);

  const filteredTrades = useMemo(() => {
    let trades = [...MOCK_TRADES];
    if (filterPair !== "all") {
      trades = trades.filter((t) => t.pair === filterPair);
    }
    trades.sort((a, b) => {
      const aVal = a[sortField];
      const bVal = b[sortField];
      if (typeof aVal === "string" && typeof bVal === "string") {
        return sortDir === "asc" ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
      }
      return sortDir === "asc"
        ? (aVal as number) - (bVal as number)
        : (bVal as number) - (aVal as number);
    });
    return trades;
  }, [filterPair, sortField, sortDir]);

  const handleSort = (field: SortField) => {
    if (field === sortField) {
      setSortDir(sortDir === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortDir("desc");
    }
  };

  const handleExportCSV = () => {
    const headers = "Time,Pair,Side,Price,Quantity,Fee,Realized P&L\n";
    const rows = filteredTrades
      .map(
        (t) =>
          `${t.time},${t.pair},${t.side},${t.price},${t.quantity},${t.fee},${t.realizedPnl}`
      )
      .join("\n");
    const csv = headers + rows;
    const blob = new Blob([csv], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "trade_history.csv";
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <select
            value={filterPair}
            onChange={(e) => setFilterPair(e.target.value)}
            className="text-xs bg-background border border-border rounded px-2 py-1 text-foreground"
          >
            {pairs.map((p) => (
              <option key={p} value={p}>
                {p === "all" ? "All Pairs" : p}
              </option>
            ))}
          </select>
        </div>
        <button
          onClick={handleExportCSV}
          className="text-xs px-3 py-1 border border-border rounded hover:bg-card-bg-hover transition-colors text-gray-400 hover:text-foreground"
        >
          Export CSV
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-xs text-gray-400">
              <th
                className="text-left py-2 px-3 font-medium cursor-pointer hover:text-foreground"
                onClick={() => handleSort("time")}
              >
                Time {sortField === "time" && (sortDir === "asc" ? "\u2191" : "\u2193")}
              </th>
              <th
                className="text-left py-2 px-3 font-medium cursor-pointer hover:text-foreground"
                onClick={() => handleSort("pair")}
              >
                Pair {sortField === "pair" && (sortDir === "asc" ? "\u2191" : "\u2193")}
              </th>
              <th className="text-left py-2 px-3 font-medium">Side</th>
              <th
                className="text-right py-2 px-3 font-medium cursor-pointer hover:text-foreground"
                onClick={() => handleSort("price")}
              >
                Price {sortField === "price" && (sortDir === "asc" ? "\u2191" : "\u2193")}
              </th>
              <th
                className="text-right py-2 px-3 font-medium cursor-pointer hover:text-foreground"
                onClick={() => handleSort("quantity")}
              >
                Quantity {sortField === "quantity" && (sortDir === "asc" ? "\u2191" : "\u2193")}
              </th>
              <th className="text-right py-2 px-3 font-medium">Fee</th>
              <th
                className="text-right py-2 px-3 font-medium cursor-pointer hover:text-foreground"
                onClick={() => handleSort("realizedPnl")}
              >
                P&L {sortField === "realizedPnl" && (sortDir === "asc" ? "\u2191" : "\u2193")}
              </th>
            </tr>
          </thead>
          <tbody>
            {filteredTrades.map((trade) => (
              <tr
                key={trade.id}
                className="border-b border-border/50 hover:bg-card-bg-hover transition-colors"
              >
                <td className="py-2 px-3 text-xs font-mono text-gray-400">
                  {trade.time}
                </td>
                <td className="py-2 px-3 font-medium">{trade.pair}</td>
                <td className="py-2 px-3">
                  <span
                    className={`text-xs font-medium uppercase ${
                      trade.side === "buy" ? "text-positive" : "text-negative"
                    }`}
                  >
                    {trade.side}
                  </span>
                </td>
                <td className="py-2 px-3 text-right font-mono">
                  {trade.price.toLocaleString(undefined, { minimumFractionDigits: 2 })}
                </td>
                <td className="py-2 px-3 text-right font-mono">
                  {trade.quantity.toFixed(4)}
                </td>
                <td className="py-2 px-3 text-right font-mono text-gray-400">
                  {trade.fee.toFixed(2)}
                </td>
                <td
                  className={`py-2 px-3 text-right font-mono ${
                    trade.realizedPnl > 0
                      ? "text-positive"
                      : trade.realizedPnl < 0
                      ? "text-negative"
                      : "text-gray-400"
                  }`}
                >
                  {trade.realizedPnl !== 0
                    ? `${trade.realizedPnl > 0 ? "+" : ""}${trade.realizedPnl.toFixed(2)}`
                    : "-"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
