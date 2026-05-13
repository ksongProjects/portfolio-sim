"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Notifications } from "@/components/notifications";
import { LiveMarketIndicator } from "@/components/live-market-indicator";
import {
  LayoutDashboard,
  Briefcase,
  Newspaper,
  TrendingUp,
  Eye,
  Settings,
} from "lucide-react";

const routes = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/portfolio", label: "Portfolio", icon: Briefcase },
  { href: "/news-feed", label: "News Feed", icon: Newspaper },
  { href: "/strategy", label: "Strategy", icon: TrendingUp },
  { href: "/observability", label: "Observability", icon: Eye },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Navbar() {
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-50 w-full border-b border-outline-variant/50 bg-surface">
      <div className="flex h-12 items-center px-6 justify-between">
        <div className="flex items-center gap-6">
          <Link href="/" className="flex items-center gap-2 mr-4">
            <span className="text-sm font-semibold text-on-surface tracking-tight">Portfolio Sim</span>
          </Link>
          <nav className="flex items-center gap-0.5 text-sm font-medium">
            {routes.map((route) => {
              const Icon = route.icon;
              const isActive = pathname === route.href;
              return (
                <Link
                  key={route.href}
                  href={route.href}
                  className={cn(
                    "flex items-center gap-1.5 px-3 py-2 text-[13px] transition-colors",
                    isActive
                      ? "text-primary bg-primary/5 border-b border-b-primary"
                      : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container"
                  )}
                >
                  <Icon className="h-3.5 w-3.5" />
                  {route.label}
                </Link>
              );
            })}
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <LiveMarketIndicator />
          <Notifications />
        </div>
      </div>
    </header>
  );
}
