"use client";

interface PortfolioDashboardProps {
  totalEquity?: number;
  change24h?: number;
  changePct24h?: number;
}

export default function PortfolioDashboard({
  totalEquity = 125430.52,
  change24h = 2340.18,
  changePct24h = 1.9,
}: PortfolioDashboardProps) {
  const isPositive = change24h >= 0;

  return (
    <div className="rounded-lg bg-card-bg border border-border p-6">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-gray-400 mb-1">Total Equity</p>
          <p className="text-3xl font-bold font-mono">
            ${totalEquity.toLocaleString(undefined, { minimumFractionDigits: 2 })}
          </p>
        </div>
        <div
          className={`flex items-center gap-1 px-2.5 py-1 rounded-full text-sm font-medium ${
            isPositive
              ? "bg-positive/10 text-positive"
              : "bg-negative/10 text-negative"
          }`}
        >
          <span>{isPositive ? "+" : ""}{changePct24h.toFixed(2)}%</span>
        </div>
      </div>
      <div className="mt-2">
        <span
          className={`text-sm font-mono ${
            isPositive ? "text-positive" : "text-negative"
          }`}
        >
          {isPositive ? "+" : ""}${change24h.toLocaleString(undefined, { minimumFractionDigits: 2 })}
        </span>
        <span className="text-xs text-gray-400 ml-2">24h Change</span>
      </div>
    </div>
  );
}
