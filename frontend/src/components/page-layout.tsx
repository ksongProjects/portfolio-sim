import { cn } from "@/lib/utils";

interface PageGridProps {
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

export function PageGrid({ children, className, style }: PageGridProps) {
  return (
    <div className={cn("grid gap-px bg-outline-variant", className)} style={style}>
      {children}
    </div>
  );
}

interface PageCellProps {
  children: React.ReactNode;
  className?: string;
}

export function PageCell({ children, className }: PageCellProps) {
  return (
    <div className={cn("bg-surface-container p-5", className)}>
      {children}
    </div>
  );
}

interface MetricLabelProps {
  children: React.ReactNode;
  className?: string;
}

export function MetricLabel({ children, className }: MetricLabelProps) {
  return (
    <div className={cn("text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant mb-2", className)}>
      {children}
    </div>
  );
}

interface MetricValueProps {
  children: React.ReactNode;
  className?: string;
  highlight?: boolean;
  style?: React.CSSProperties;
}

export function MetricValue({ children, className, highlight, style }: MetricValueProps) {
  return (
    <div
      className={cn(
        "text-[22px] font-medium tracking-[-0.01em] text-on-surface",
        highlight && "text-primary",
        className
      )}
      style={{ fontFamily: "var(--font-display, 'Work Sans', sans-serif)", ...style }}
    >
      {children}
    </div>
  );
}

interface MetricSubValueProps {
  children: React.ReactNode;
  className?: string;
  positive?: boolean;
}

export function MetricSubValue({ children, className, positive }: MetricSubValueProps) {
  return (
    <div
      className={cn(
        "text-[12px] font-mono mt-1",
        positive === true ? "text-primary" : positive === false ? "text-error" : "text-on-surface-variant",
        className
      )}
    >
      {children}
    </div>
  );
}

interface PageHeaderProps {
  title: string;
  description?: string;
  children?: React.ReactNode;
}

export function PageHeader({ title, description, children }: PageHeaderProps) {
  return (
    <div className="flex items-center justify-between pb-4 border-b border-outline-variant/50">
      <div>
        <h1 className="text-xl font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-display, 'Work Sans', sans-serif)" }}>
          {title}
        </h1>
        {description && (
          <p className="mt-0.5 text-sm text-on-surface-variant">{description}</p>
        )}
      </div>
      {children && <div className="flex items-center gap-3">{children}</div>}
    </div>
  );
}

interface DataTableProps<T extends { id: string | number }> {
  columns: { key: keyof T | string; label: string; width?: string }[];
  data: T[];
  renderRow?: (row: T) => React.ReactNode;
}

export function DataTable<T extends { id: string | number }>({
  columns,
  data,
  renderRow,
}: DataTableProps<T>) {
  return (
    <div className="w-full overflow-auto border border-outline-variant/30">
      <table className="w-full">
        <thead>
          <tr className="border-b border-outline-variant/30 bg-surface-container-low">
            {columns.map((col) => (
              <th
                key={String(col.key)}
                className="h-9 px-4 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant"
                style={{ width: col.width }}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row) =>
            renderRow ? (
              renderRow(row)
            ) : (
              <tr
                key={row.id}
                className="border-b border-outline-variant/30 hover:bg-surface-container-low/50"
              >
                {columns.map((col) => (
                  <td
                    key={String(col.key)}
                    className="h-9 px-4 text-sm text-on-surface"
                  >
                    {String((row as Record<string, unknown>)[col.key as string] ?? "")}
                  </td>
                ))}
              </tr>
            )
          )}
        </tbody>
      </table>
    </div>
  );
}
