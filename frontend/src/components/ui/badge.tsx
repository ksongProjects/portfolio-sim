import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1 text-[11px] font-semibold uppercase tracking-[0.08em]",
  {
    variants: {
      variant: {
        default: "bg-primary text-on-primary",
        secondary: "bg-surface-container text-on-surface-variant",
        outline: "border border-outline text-on-surface-variant",
        success: "bg-primary text-on-primary",
        warning: "bg-error-container text-on-error-container",
        error: "bg-error text-on-error",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

function StatusIndicator({
  className,
  active = false,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { active?: boolean }) {
  return (
    <div
      className={cn(
        "h-2 w-2 shrink-0",
        active ? "bg-primary" : "bg-outline",
        className
      )}
      {...props}
    />
  );
}

export { Badge, badgeVariants, StatusIndicator };
