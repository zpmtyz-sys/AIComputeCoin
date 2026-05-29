"use client";

import { useState, useCallback } from "react";
import { z } from "zod";
import ConfirmationModal from "./ConfirmationModal";

type OrderType = "limit" | "market" | "stop-limit";
type OrderSide = "buy" | "sell";

const limitSchema = z.object({
  price: z.number().positive("Price must be positive"),
  quantity: z.number().positive("Quantity must be positive"),
});

const marketSchema = z.object({
  quantity: z.number().positive("Quantity must be positive"),
});

const stopLimitSchema = z.object({
  stopPrice: z.number().positive("Stop price must be positive"),
  limitPrice: z.number().positive("Limit price must be positive"),
  quantity: z.number().positive("Quantity must be positive"),
});

interface OrderFormProps {
  initialPrice?: number;
}

const AVAILABLE_BALANCE = 50000;
const MAKER_FEE = 0.001;
const TAKER_FEE = 0.0015;

export default function OrderForm({ initialPrice }: OrderFormProps) {
  const [orderType, setOrderType] = useState<OrderType>("limit");
  const [side, setSide] = useState<OrderSide>("buy");
  const [price, setPrice] = useState(initialPrice?.toString() ?? "45000");
  const [quantity, setQuantity] = useState("");
  const [stopPrice, setStopPrice] = useState("");
  const [limitPrice, setLimitPrice] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [showConfirmation, setShowConfirmation] = useState(false);

  const total = parseFloat(price || "0") * parseFloat(quantity || "0");
  const fee = total * (orderType === "market" ? TAKER_FEE : MAKER_FEE);

  const handlePercentage = useCallback(
    (pct: number) => {
      const priceVal = parseFloat(price || "0");
      if (priceVal > 0) {
        const maxQty = (AVAILABLE_BALANCE * pct) / 100 / priceVal;
        setQuantity(maxQty.toFixed(4));
      }
    },
    [price]
  );

  const validateAndSubmit = () => {
    setErrors({});
    let result;

    if (orderType === "limit") {
      result = limitSchema.safeParse({
        price: parseFloat(price),
        quantity: parseFloat(quantity),
      });
    } else if (orderType === "market") {
      result = marketSchema.safeParse({
        quantity: parseFloat(quantity),
      });
    } else {
      result = stopLimitSchema.safeParse({
        stopPrice: parseFloat(stopPrice),
        limitPrice: parseFloat(limitPrice),
        quantity: parseFloat(quantity),
      });
    }

    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.errors.forEach((err) => {
        const field = err.path[0] as string;
        fieldErrors[field] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }

    // Show confirmation for large orders (> 10000 USDT)
    if (total > 10000) {
      setShowConfirmation(true);
      return;
    }

    submitOrder();
  };

  const submitOrder = () => {
    setShowConfirmation(false);
    // In production this would call the API
    setQuantity("");
  };

  return (
    <div className="flex flex-col h-full">
      {/* Order Type Tabs */}
      <div className="flex border-b border-border mb-3">
        {(["limit", "market", "stop-limit"] as const).map((type) => (
          <button
            key={type}
            onClick={() => setOrderType(type)}
            className={`flex-1 py-2 text-xs font-medium capitalize transition-colors ${
              orderType === type
                ? "text-foreground border-b-2 border-accent"
                : "text-gray-400 hover:text-foreground"
            }`}
          >
            {type}
          </button>
        ))}
      </div>

      {/* Buy/Sell Toggle */}
      <div className="grid grid-cols-2 gap-1 mb-3 p-1 bg-background rounded-md">
        <button
          onClick={() => setSide("buy")}
          className={`py-1.5 text-sm font-medium rounded transition-colors ${
            side === "buy"
              ? "bg-positive text-white"
              : "text-gray-400 hover:text-foreground"
          }`}
        >
          Buy
        </button>
        <button
          onClick={() => setSide("sell")}
          className={`py-1.5 text-sm font-medium rounded transition-colors ${
            side === "sell"
              ? "bg-negative text-white"
              : "text-gray-400 hover:text-foreground"
          }`}
        >
          Sell
        </button>
      </div>

      {/* Form Fields */}
      <div className="space-y-2 flex-1">
        {(orderType === "limit" || orderType === "stop-limit") && (
          <div>
            {orderType === "stop-limit" && (
              <div className="mb-2">
                <label className="text-xs text-gray-400 mb-1 block">Stop Price</label>
                <input
                  type="number"
                  value={stopPrice}
                  onChange={(e) => setStopPrice(e.target.value)}
                  placeholder="0.00"
                  className="w-full px-3 py-2 bg-background border border-border rounded-md text-sm font-mono focus:outline-none focus:border-accent"
                />
                {errors.stopPrice && (
                  <span className="text-xs text-negative">{errors.stopPrice}</span>
                )}
              </div>
            )}
            <label className="text-xs text-gray-400 mb-1 block">
              {orderType === "stop-limit" ? "Limit Price" : "Price"}
            </label>
            <input
              type="number"
              value={orderType === "stop-limit" ? limitPrice : price}
              onChange={(e) =>
                orderType === "stop-limit"
                  ? setLimitPrice(e.target.value)
                  : setPrice(e.target.value)
              }
              placeholder="0.00"
              className="w-full px-3 py-2 bg-background border border-border rounded-md text-sm font-mono focus:outline-none focus:border-accent"
            />
            {errors.price && (
              <span className="text-xs text-negative">{errors.price}</span>
            )}
            {errors.limitPrice && (
              <span className="text-xs text-negative">{errors.limitPrice}</span>
            )}
          </div>
        )}

        {orderType === "market" && (
          <div className="text-xs text-gray-400 mb-1">
            Estimated fill: <span className="font-mono text-foreground">~45,025.00</span>
          </div>
        )}

        <div>
          <label className="text-xs text-gray-400 mb-1 block">Quantity</label>
          <input
            type="number"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            placeholder="0.0000"
            className="w-full px-3 py-2 bg-background border border-border rounded-md text-sm font-mono focus:outline-none focus:border-accent"
          />
          {errors.quantity && (
            <span className="text-xs text-negative">{errors.quantity}</span>
          )}
        </div>

        {/* Percentage buttons */}
        <div className="flex gap-1">
          {[25, 50, 75, 100].map((pct) => (
            <button
              key={pct}
              onClick={() => handlePercentage(pct)}
              className="flex-1 py-1 text-xs text-gray-400 border border-border rounded hover:bg-card-bg-hover transition-colors"
            >
              {pct}%
            </button>
          ))}
        </div>

        {/* Total and fee */}
        <div className="space-y-1 pt-2 border-t border-border">
          <div className="flex justify-between text-xs">
            <span className="text-gray-400">Total</span>
            <span className="font-mono">{isNaN(total) ? "0.00" : total.toFixed(2)} USDT</span>
          </div>
          <div className="flex justify-between text-xs">
            <span className="text-gray-400">
              Fee ({orderType === "market" ? "0.15%" : "0.10%"})
            </span>
            <span className="font-mono text-gray-400">
              ~{isNaN(fee) ? "0.00" : fee.toFixed(2)} USDT
            </span>
          </div>
          <div className="flex justify-between text-xs">
            <span className="text-gray-400">Available</span>
            <span className="font-mono">{AVAILABLE_BALANCE.toLocaleString()} USDT</span>
          </div>
        </div>
      </div>

      {/* Submit */}
      <button
        onClick={validateAndSubmit}
        className={`w-full py-3 mt-3 rounded-md text-sm font-semibold transition-colors ${
          side === "buy"
            ? "bg-positive hover:bg-positive/90 text-white"
            : "bg-negative hover:bg-negative/90 text-white"
        }`}
      >
        {side === "buy" ? "Buy" : "Sell"} CU-PERP
      </button>

      {showConfirmation && (
        <ConfirmationModal
          side={side}
          type={orderType}
          price={orderType === "market" ? 45025 : parseFloat(price)}
          quantity={parseFloat(quantity)}
          total={total}
          onConfirm={submitOrder}
          onCancel={() => setShowConfirmation(false)}
        />
      )}
    </div>
  );
}
