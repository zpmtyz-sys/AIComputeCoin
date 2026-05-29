"use client";

interface ConfirmationModalProps {
  side: "buy" | "sell";
  type: string;
  price: number;
  quantity: number;
  total: number;
  onConfirm: () => void;
  onCancel: () => void;
}

export default function ConfirmationModal({
  side,
  type,
  price,
  quantity,
  total,
  onConfirm,
  onCancel,
}: ConfirmationModalProps) {
  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60">
      <div className="w-full max-w-sm mx-4 rounded-lg bg-card-bg border border-border p-6 animate-fade-in">
        <h3 className="text-lg font-semibold mb-4">Confirm Order</h3>

        <div className="space-y-2 mb-6">
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Side</span>
            <span
              className={`font-medium capitalize ${
                side === "buy" ? "text-positive" : "text-negative"
              }`}
            >
              {side}
            </span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Type</span>
            <span className="capitalize">{type}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Price</span>
            <span className="font-mono">{price.toFixed(2)}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Quantity</span>
            <span className="font-mono">{quantity.toFixed(4)}</span>
          </div>
          <div className="flex justify-between text-sm border-t border-border pt-2">
            <span className="text-gray-400">Total</span>
            <span className="font-mono font-semibold">{total.toFixed(2)} USDT</span>
          </div>
        </div>

        <p className="text-xs text-gray-400 mb-4">
          Are you sure you want to place this order? This action cannot be undone.
        </p>

        <div className="flex gap-3">
          <button
            onClick={onCancel}
            className="flex-1 py-2 rounded-md border border-border text-sm font-medium hover:bg-card-bg-hover transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className={`flex-1 py-2 rounded-md text-sm font-medium text-white transition-colors ${
              side === "buy"
                ? "bg-positive hover:bg-positive/90"
                : "bg-negative hover:bg-negative/90"
            }`}
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  );
}
