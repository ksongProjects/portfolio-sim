import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1 text-[10px] font-semibold uppercase tracking-[0.08em] px-1.5 py-0.5",
  {
    variants: {
      variant: {
        default: "bg-primary text-on-primary",
        secondary: "bg-surface-container text-on-surface-variant",
        outline: "border border-outline text-on-surface-variant",
        success: "bg-primary/15 text-primary border border-primary/20",
        warning: "bg-warning/15 text-warning border border-warning/20",
        error: "bg-error/15 text-error border border-error/20",
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
        "h-1.5 w-1.5 shrink-0 rounded-full",
        active ? "bg-primary" : "bg-outline",
        className
      )}
      {...props}
    />
  );
}

export { Badge, badgeVariants, StatusIndicator };
