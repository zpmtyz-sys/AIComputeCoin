const MOCK_MARKETS = [
  { pair: "CU-PERP/USDT", lastPrice: 45023.5, change24h: 523.5, changePct: 1.18, volume: 125430000, high: 45500, low: 44200 },
  { pair: "GPU-SPOT/USDT", lastPrice: 12450.0, change24h: -180.0, changePct: -1.43, volume: 45200000, high: 12800, low: 12300 },
  { pair: "TPU-PERP/USDT", lastPrice: 9150.0, change24h: 92.0, changePct: 1.02, volume: 23100000, high: 9300, low: 8950 },
  { pair: "FPGA-SPOT/USDT", lastPrice: 3420.0, change24h: -45.0, changePct: -1.3, volume: 8700000, high: 3500, low: 3380 },
  { pair: "ASIC-PERP/USDT", lastPrice: 67200.0, change24h: 1200.0, changePct: 1.82, volume: 210000000, high: 67800, low: 65500 },
  { pair: "NPU-SPOT/USDT", lastPrice: 5890.0, change24h: -120.0, changePct: -2.0, volume: 15600000, high: 6100, low: 5800 },
];

export default function MarketsPage() {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Markets</h1>

      <div className="rounded-lg bg-card-bg border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-xs text-gray-400 bg-background/50">
                <th className="text-left py-3 px-4 font-medium">Pair</th>
                <th className="text-right py-3 px-4 font-medium">Price</th>
                <th className="text-right py-3 px-4 font-medium">24h Change</th>
                <th className="text-right py-3 px-4 font-medium">24h High</th>
                <th className="text-right py-3 px-4 font-medium">24h Low</th>
                <th className="text-right py-3 px-4 font-medium">24h Volume</th>
              </tr>
            </thead>
            <tbody>
              {MOCK_MARKETS.map((market) => (
                <tr
                  key={market.pair}
                  className="border-b border-border/50 hover:bg-card-bg-hover transition-colors cursor-pointer"
                >
                  <td className="py-3 px-4 font-medium">{market.pair}</td>
                  <td className="py-3 px-4 text-right font-mono">
                    {market.lastPrice.toLocaleString(undefined, {
                      minimumFractionDigits: 2,
                    })}
                  </td>
                  <td
                    className={`py-3 px-4 text-right font-mono ${
                      market.changePct >= 0 ? "text-positive" : "text-negative"
                    }`}
                  >
                    {market.changePct >= 0 ? "+" : ""}
                    {market.changePct.toFixed(2)}%
                  </td>
                  <td className="py-3 px-4 text-right font-mono text-gray-400">
                    {market.high.toLocaleString(undefined, {
                      minimumFractionDigits: 2,
                    })}
                  </td>
                  <td className="py-3 px-4 text-right font-mono text-gray-400">
                    {market.low.toLocaleString(undefined, {
                      minimumFractionDigits: 2,
                    })}
                  </td>
                  <td className="py-3 px-4 text-right font-mono text-gray-400">
                    ${(market.volume / 1000000).toFixed(1)}M
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
