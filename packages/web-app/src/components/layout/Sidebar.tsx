"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ChartBarIcon,
  WalletIcon,
  GlobeAltIcon,
  CpuChipIcon,
  BuildingLibraryIcon,
  ChevronLeftIcon,
} from "@heroicons/react/24/outline";

const NAV_ITEMS = [
  { href: "/trade", label: "Trade", icon: ChartBarIcon },
  { href: "/portfolio", label: "Portfolio", icon: WalletIcon },
  { href: "/markets", label: "Markets", icon: GlobeAltIcon },
  { href: "/compute", label: "Compute Nodes", icon: CpuChipIcon },
  { href: "/governance", label: "Governance", icon: BuildingLibraryIcon },
];

export default function Sidebar() {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);

  return (
    <aside
      className={`hidden md:flex flex-col fixed top-14 left-0 bottom-0 z-40 bg-card-bg border-r border-border transition-all duration-200 ${
        collapsed ? "w-16" : "w-56"
      }`}
    >
      <nav className="flex-1 py-4 px-2 space-y-1">
        {NAV_ITEMS.map((item) => {
          const isActive = pathname === item.href;
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? "bg-accent/10 text-accent"
                  : "text-gray-400 hover:text-foreground hover:bg-card-bg-hover"
              }`}
            >
              <Icon className="w-5 h-5 shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      <button
        onClick={() => setCollapsed(!collapsed)}
        className="p-3 border-t border-border text-gray-400 hover:text-foreground transition-colors"
      >
        <ChevronLeftIcon
          className={`w-4 h-4 mx-auto transition-transform ${
            collapsed ? "rotate-180" : ""
          }`}
        />
      </button>
    </aside>
  );
}
