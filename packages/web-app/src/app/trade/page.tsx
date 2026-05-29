"use client";

import { useState } from "react";
import TradingChart from "@/components/chart/TradingChart";
import OrderBook from "@/components/orderbook/OrderBook";
import OrderForm from "@/components/order-form/OrderForm";
import PositionsTable from "@/components/portfolio/PositionsTable";
import OrdersTable from "@/components/portfolio/OrdersTable";

export default function TradePage() {
  const [activeTab, setActiveTab] = useState<"positions" | "orders" | "history">(
    "positions"
  );
  const [orderPrice, setOrderPrice] = useState<number | undefined>(undefined);

  const handlePriceClick = (price: number) => {
    setOrderPrice(price);
  };

  return (
    <div className="p-4 h-[calc(100vh-3.5rem)]">
      <div className="grid grid-cols-12 gap-3 h-full">
        {/* Chart Section */}
        <div className="col-span-12 lg:col-span-8 row-span-2 rounded-lg bg-card-bg border border-border p-3 min-h-[400px]">
          <TradingChart />
        </div>

        {/* Order Book */}
        <div className="col-span-12 sm:col-span-6 lg:col-span-4 rounded-lg bg-card-bg border border-border p-3 min-h-[300px] lg:min-h-0">
          <OrderBook onPriceClick={handlePriceClick} />
        </div>

        {/* Order Entry */}
        <div className="col-span-12 sm:col-span-6 lg:col-span-4 rounded-lg bg-card-bg border border-border p-3 min-h-[300px] lg:min-h-0">
          <OrderForm initialPrice={orderPrice} />
        </div>

        {/* Positions / Orders */}
        <div className="col-span-12 rounded-lg bg-card-bg border border-border p-3">
          <div className="flex gap-4 mb-3 border-b border-border pb-2">
            <button
              onClick={() => setActiveTab("positions")}
              className={`text-sm font-medium transition-colors ${
                activeTab === "positions"
                  ? "text-accent"
                  : "text-gray-400 hover:text-foreground"
              }`}
            >
              Open Positions
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
          </div>
          {activeTab === "positions" && <PositionsTable />}
          {activeTab === "orders" && <OrdersTable />}
        </div>
      </div>
    </div>
  );
}
