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
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        <div className="mr-4 hidden md:flex">
          <Link href="/" className="mr-6 flex items-center space-x-2">
            <span className="hidden font-bold sm:inline-block">Portfolio Sim</span>
          </Link>
          <nav className="flex items-center space-x-6 text-sm font-medium">
            {routes.map((route) => {
              const Icon = route.icon;
              const isActive = pathname === route.href;
              return (
                <Link
                  key={route.href}
                  href={route.href}
                  className={cn(
                    "flex items-center gap-1.5 text-sm transition-colors hover:text-foreground/80",
                    isActive
                      ? "text-foreground font-semibold"
                      : "text-foreground/60"
                  )}
                >
                  <Icon className="h-4 w-4" />
                  {route.label}
                </Link>
              );
            })}
          </nav>
        </div>
        <div className="flex md:hidden">
          <Link href="/" className="mr-4 flex items-center space-x-2">
            <span className="font-bold">Portfolio Sim</span>
          </Link>
        </div>
      </div>
    </header>
  );
}