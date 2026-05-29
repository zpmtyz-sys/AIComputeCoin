"use client";

import { useState } from "react";

interface Order {
  id: string;
  pair: string;
  side: "buy" | "sell";
  type: "limit" | "market" | "stop-limit";
  price: number;
  quantity: number;
  filled: number;
  time: string;
}

const MOCK_ORDERS: Order[] = [
  {
    id: "ord-001",
    pair: "CU-PERP/USDT",
    side: "buy",
    type: "limit",
    price: 44500,
    quantity: 1.5,
    filled: 0,
    time: "2024-01-15 14:30:22",
  },
  {
    id: "ord-002",
    pair: "GPU-SPOT/USDT",
    side: "sell",
    type: "limit",
    price: 12800,
    quantity: 3.0,
    filled: 50,
    time: "2024-01-15 13:15:08",
  },
  {
    id: "ord-003",
    pair: "CU-PERP/USDT",
    side: "buy",
    type: "stop-limit",
    price: 46000,
    quantity: 0.5,
    filled: 0,
    time: "2024-01-15 10:45:33",
  },
];

interface OrdersTableProps {
  onCancel?: (orderId: string) => void;
}

export default function OrdersTable({ onCancel }: OrdersTableProps) {
  const [orders, setOrders] = useState(MOCK_ORDERS);

  const handleCancel = (orderId: string) => {
    setOrders((prev) => prev.filter((o) => o.id !== orderId));
    onCancel?.(orderId);
  };

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-border text-xs text-gray-400">
            <th className="text-left py-2 px-3 font-medium">Pair</th>
            <th className="text-left py-2 px-3 font-medium">Side</th>
            <th className="text-left py-2 px-3 font-medium">Type</th>
            <th className="text-right py-2 px-3 font-medium">Price</th>
            <th className="text-right py-2 px-3 font-medium">Quantity</th>
            <th className="text-right py-2 px-3 font-medium">Filled%</th>
            <th className="text-left py-2 px-3 font-medium">Time</th>
            <th className="text-right py-2 px-3 font-medium">Action</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((order) => (
            <tr
              key={order.id}
              className="border-b border-border/50 hover:bg-card-bg-hover transition-colors"
            >
              <td className="py-2 px-3 font-medium">{order.pair}</td>
              <td className="py-2 px-3">
                <span
                  className={`text-xs font-medium uppercase ${
                    order.side === "buy" ? "text-positive" : "text-negative"
                  }`}
                >
                  {order.side}
                </span>
              </td>
              <td className="py-2 px-3 text-xs capitalize text-gray-400">
                {order.type}
              </td>
              <td className="py-2 px-3 text-right font-mono">
                {order.price.toLocaleString(undefined, { minimumFractionDigits: 2 })}
              </td>
              <td className="py-2 px-3 text-right font-mono">
                {order.quantity.toFixed(4)}
              </td>
              <td className="py-2 px-3 text-right font-mono">
                {order.filled}%
              </td>
              <td className="py-2 px-3 text-xs text-gray-400 font-mono">
                {order.time}
              </td>
              <td className="py-2 px-3 text-right">
                <button
                  onClick={() => handleCancel(order.id)}
                  className="text-xs text-negative hover:text-negative/80 font-medium transition-colors"
                >
                  Cancel
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {orders.length === 0 && (
        <div className="text-center py-8 text-gray-400 text-sm">
          No open orders
        </div>
      )}
    </div>
  );
}
