"use client";

interface Position {
  id: string;
  pair: string;
  side: "long" | "short";
  size: number;
  entryPrice: number;
  markPrice: number;
  unrealizedPnl: number;
  margin: number;
}

const MOCK_POSITIONS: Position[] = [
  {
    id: "1",
    pair: "CU-PERP/USDT",
    side: "long",
    size: 2.5,
    entryPrice: 44800,
    markPrice: 45050,
    unrealizedPnl: 625,
    margin: 11200,
  },
  {
    id: "2",
    pair: "GPU-SPOT/USDT",
    side: "short",
    size: 1.2,
    entryPrice: 12500,
    markPrice: 12650,
    unrealizedPnl: -180,
    margin: 3000,
  },
  {
    id: "3",
    pair: "TPU-PERP/USDT",
    side: "long",
    size: 5.0,
    entryPrice: 8900,
    markPrice: 9100,
    unrealizedPnl: 1000,
    margin: 8900,
  },
];

export default function PositionsTable() {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-xs text-gray-400">
            <th className="text-left py-2 px-3 font-medium">Pair</th>
            <th className="text-left py-2 px-3 font-medium">Side</th>
            <th className="text-right py-2 px-3 font-medium">Size</th>
            <th className="text-right py-2 px-3 font-medium">Entry Price</th>
            <th className="text-right py-2 px-3 font-medium">Mark Price</th>
            <th className="text-right py-2 px-3 font-medium">Unrealized P&L</th>
            <th className="text-right py-2 px-3 font-medium">Margin</th>
          </tr>
        </thead>
        <tbody>
          {MOCK_POSITIONS.map((pos) => (
            <tr
              key={pos.id}
              className="border-b border-border/50 hover:bg-card-bg-hover transition-colors"
            >
              <td className="py-2 px-3 font-medium">{pos.pair}</td>
              <td className="py-2 px-3">
                <span
                  className={`text-xs font-medium uppercase ${
                    pos.side === "long" ? "text-positive" : "text-negative"
                  }`}
                >
                  {pos.side}
                </span>
              </td>
              <td className="py-2 px-3 text-right font-mono">{pos.size.toFixed(4)}</td>
              <td className="py-2 px-3 text-right font-mono">
                {pos.entryPrice.toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </td>
              <td className="py-2 px-3 text-right font-mono">
                {pos.markPrice.toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </td>
              <td
                className={`py-2 px-3 text-right font-mono ${
                  pos.unrealizedPnl >= 0 ? "text-positive" : "text-negative"
                }`}
              >
                {pos.unrealizedPnl >= 0 ? "+" : ""}
                {pos.unrealizedPnl.toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </td>
              <td className="py-2 px-3 text-right font-mono">
                {pos.margin.toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {MOCK_POSITIONS.length === 0 && (
        <div className="text-center py-8 text-gray-400 text-sm">
          No open positions
        </div>
      )}
    </div>
  );
}
