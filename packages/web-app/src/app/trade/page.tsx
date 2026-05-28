export default function TradePage() {
  return (
    <main className="min-h-screen p-4">
      <div className="grid grid-cols-12 gap-4 h-[calc(100vh-2rem)]">
        {/* Chart Section */}
        <div className="col-span-8 row-span-2 rounded-lg bg-[var(--card-bg)] border border-[var(--border)] p-4">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">CU-PERP/USDT</h2>
            <div className="flex gap-2 text-sm text-gray-400">
              <span>1m</span>
              <span>5m</span>
              <span>15m</span>
              <span>1h</span>
              <span>4h</span>
              <span>1D</span>
            </div>
          </div>
          <div className="flex items-center justify-center h-[calc(100%-3rem)] text-gray-500">
            Chart placeholder - TradingView Lightweight Charts
          </div>
        </div>

        {/* Order Book */}
        <div className="col-span-4 rounded-lg bg-[var(--card-bg)] border border-[var(--border)] p-4">
          <h2 className="text-lg font-semibold mb-4">Order Book</h2>
          <div className="flex items-center justify-center h-[calc(100%-3rem)] text-gray-500">
            Order book placeholder
          </div>
        </div>

        {/* Order Entry */}
        <div className="col-span-4 rounded-lg bg-[var(--card-bg)] border border-[var(--border)] p-4">
          <h2 className="text-lg font-semibold mb-4">Place Order</h2>
          <div className="flex items-center justify-center h-[calc(100%-3rem)] text-gray-500">
            Order entry form placeholder
          </div>
        </div>

        {/* Positions / Orders */}
        <div className="col-span-12 rounded-lg bg-[var(--card-bg)] border border-[var(--border)] p-4">
          <div className="flex gap-4 mb-4">
            <button className="text-sm font-semibold text-compute-blue-400">
              Open Positions
            </button>
            <button className="text-sm text-gray-400">Open Orders</button>
            <button className="text-sm text-gray-400">Trade History</button>
          </div>
          <div className="flex items-center justify-center h-32 text-gray-500">
            Positions table placeholder
          </div>
        </div>
      </div>
    </main>
  );
}
