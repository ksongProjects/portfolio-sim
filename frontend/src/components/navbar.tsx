"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
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
    <header className="sticky top-0 z-50 w-full border-b border-outline-variant bg-surface-container-low">
      <div className="flex h-14 items-center px-6">
        <div className="flex items-center gap-6">
          <Link href="/" className="flex items-center gap-2">
            <span className="font-semibold text-on-surface">Portfolio Sim</span>
          </Link>
          <nav className="flex items-center gap-1 text-sm font-medium">
            {routes.map((route) => {
              const Icon = route.icon;
              const isActive = pathname === route.href;
              return (
                <Link
                  key={route.href}
                  href={route.href}
                  className={cn(
                    "flex items-center gap-1.5 px-3 py-2 transition-colors hover:bg-surface-container-high",
                    isActive
                      ? "text-primary border-b-2 border-primary"
                      : "text-on-surface-variant"
                  )}
                >
                  <Icon className="h-4 w-4" />
                  {route.label}
                </Link>
              );
            })}
          </nav>
        </div>
      </div>
    </header>
  );
}
