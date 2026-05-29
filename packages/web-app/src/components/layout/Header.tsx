"use client";

import { useState } from "react";
import {
  ChevronDownIcon,
  BellIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";

const MARKETS = [
  "CU-PERP/USDT",
  "GPU-SPOT/USDT",
  "TPU-PERP/USDT",
  "FPGA-SPOT/USDT",
];

export default function Header() {
  const [selectedMarket, setSelectedMarket] = useState(MARKETS[0]);
  const [dropdownOpen, setDropdownOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 h-14 flex items-center justify-between px-4 bg-card-bg border-b border-border">
      <div className="flex items-center gap-4">
        <span className="text-lg font-bold text-accent">ComputeCoin</span>

        <div className="relative hidden sm:block">
          <button
            onClick={() => setDropdownOpen(!dropdownOpen)}
            className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-background border border-border text-sm font-medium hover:bg-card-bg-hover transition-colors"
          >
            <span className="font-mono">{selectedMarket}</span>
            <ChevronDownIcon className="w-4 h-4 text-gray-400" />
          </button>

          {dropdownOpen && (
            <div className="absolute top-full mt-1 left-0 w-48 rounded-md bg-card-bg border border-border shadow-lg py-1 z-50">
              {MARKETS.map((market) => (
                <button
                  key={market}
                  onClick={() => {
                    setSelectedMarket(market);
                    setDropdownOpen(false);
                  }}
                  className={`w-full text-left px-3 py-2 text-sm font-mono hover:bg-card-bg-hover transition-colors ${
                    market === selectedMarket ? "text-accent" : "text-foreground"
                  }`}
                >
                  {market}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-3">
        <button className="p-2 rounded-md hover:bg-card-bg-hover transition-colors relative">
          <BellIcon className="w-5 h-5 text-gray-400" />
          <span className="absolute top-1 right-1 w-2 h-2 bg-accent rounded-full" />
        </button>
        <button className="p-2 rounded-md hover:bg-card-bg-hover transition-colors">
          <UserCircleIcon className="w-5 h-5 text-gray-400" />
        </button>
      </div>
    </header>
  );
}
