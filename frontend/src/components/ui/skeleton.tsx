import * as React from "react";
import { cn } from "@/lib/utils";

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn("animate-pulse rounded bg-surface-container-high", className)}
      {...props}
    />
  );
}

function TableSkeleton({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="border border-outline-variant/30">
      <div className="h-10 bg-surface-container-low" />
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-4 px-4 py-3 border-t border-outline-variant/30">
          {Array.from({ length: cols }).map((_, j) => (
            <Skeleton key={j} className="h-4 flex-1" />
          ))}
        </div>
      ))}
    </div>
  );
}

function CardSkeleton() {
  return (
    <div className="border border-outline-variant/30 p-4 space-y-3">
      <Skeleton className="h-4 w-1/3" />
      <Skeleton className="h-8 w-1/2" />
      <Skeleton className="h-3 w-1/4" />
    </div>
  );
}

function MetricSkeleton() {
  return (
    <div className="space-y-2">
      <Skeleton className="h-3 w-20" />
      <Skeleton className="h-8 w-32" />
      <Skeleton className="h-3 w-24" />
    </div>
  );
}

function Spinner({ className }: { className?: string }) {
  return (
    <div
      className={cn("h-5 w-5 border-2 border-primary border-t-transparent rounded-full animate-spin", className)}
    />
  );
}

function LoadingScreen() {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="flex flex-col items-center gap-3">
        <Spinner className="h-8 w-8" />
        <span className="text-sm text-on-surface-variant">Loading...</span>
      </div>
    </div>
  );
}

function PageLoadingFallback() {
  return (
    <div className="flex flex-col h-full">
      <div className="px-6 pt-6 pb-4">
        <div className="flex items-center justify-between">
          <div className="space-y-2">
            <Skeleton className="h-7 w-32" />
            <Skeleton className="h-4 w-48" />
          </div>
          <div className="flex gap-3">
            <Skeleton className="h-9 w-28" />
            <Skeleton className="h-9 w-28" />
          </div>
        </div>
      </div>
      <div className="flex-1 px-6 pb-6 overflow-auto space-y-6">
        <div style={{ gridTemplateColumns: "repeat(4, 1fr)" }} className="grid gap-4 mb-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="border border-outline-variant/30 p-4">
              <MetricSkeleton />
            </div>
          ))}
        </div>
        <div style={{ gridTemplateColumns: "7fr 5fr" }} className="grid gap-4">
          <CardSkeleton />
          <CardSkeleton />
        </div>
      </div>
    </div>
  );
}

export { Skeleton, TableSkeleton, CardSkeleton, MetricSkeleton, Spinner, LoadingScreen, PageLoadingFallback };