"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ChartBarIcon,
  WalletIcon,
  GlobeAltIcon,
  CpuChipIcon,
  BuildingLibraryIcon,
} from "@heroicons/react/24/outline";

const NAV_ITEMS = [
  { href: "/trade", label: "Trade", icon: ChartBarIcon },
  { href: "/portfolio", label: "Portfolio", icon: WalletIcon },
  { href: "/markets", label: "Markets", icon: GlobeAltIcon },
  { href: "/compute", label: "Compute", icon: CpuChipIcon },
  { href: "/governance", label: "Gov", icon: BuildingLibraryIcon },
];

export default function MobileNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 md:hidden bg-card-bg border-t border-border">
      <div className="flex items-center justify-around h-14">
        {NAV_ITEMS.map((item) => {
          const isActive = pathname === item.href;
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex flex-col items-center gap-0.5 px-2 py-1 text-xs transition-colors ${
                isActive ? "text-accent" : "text-gray-400"
              }`}
            >
              <Icon className="w-5 h-5" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
