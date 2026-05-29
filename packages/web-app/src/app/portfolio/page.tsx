"use client";

import { useState } from "react";
import PortfolioDashboard from "@/components/portfolio/PortfolioDashboard";
import PositionsTable from "@/components/portfolio/PositionsTable";
import OrdersTable from "@/components/portfolio/OrdersTable";
import TradeHistory from "@/components/portfolio/TradeHistory";
import PnLChart from "@/components/portfolio/PnLChart";

export default function PortfolioPage() {
  const [activeTab, setActiveTab] = useState<
    "positions" | "orders" | "history"
  >("positions");

  return (
    <div className="p-4 space-y-4">
      <h1 className="text-2xl font-bold">Portfolio</h1>

      <PortfolioDashboard />

      {/* Equity Chart */}
      <div className="rounded-lg bg-card-bg border border-border p-4 h-[300px]">
        <PnLChart />
      </div>

      {/* Tabs */}
      <div className="rounded-lg bg-card-bg border border-border p-4">
        <div className="flex gap-4 mb-4 border-b border-border pb-2">
          <button
            onClick={() => setActiveTab("positions")}
            className={`text-sm font-medium transition-colors ${
              activeTab === "positions"
                ? "text-accent"
                : "text-gray-400 hover:text-foreground"
            }`}
          >
            Positions
          </button>
          <button
            onClick={() => setActiveTab("orders")}
            className={`text-sm font-medium transition-colors ${
              activeTab === "orders"
                ? "text-accent"
                : "text-gray-400 hover:text-foreground"
            }`}
          >
            Open Orders
          </button>
          <button
            onClick={() => setActiveTab("history")}
            className={`text-sm font-medium transition-colors ${
              activeTab === "history"
                ? "text-accent"
                : "text-gray-400 hover:text-foreground"
            }`}
          >
            Trade History
          </button>
        </div>

        {activeTab === "positions" && <PositionsTable />}
        {activeTab === "orders" && <OrdersTable />}
        {activeTab === "history" && <TradeHistory />}
      </div>
    </div>
  );
}
