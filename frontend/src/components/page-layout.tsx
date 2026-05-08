import { cn } from "@/lib/utils";

interface PageGridProps {
  children: React.ReactNode;
  className?: string;
}

export function PageGrid({ children, className }: PageGridProps) {
  return (
    <div className={cn("grid gap-[1px] bg-outline-variant p-[1px]", className)}>
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
    <div className={cn("bg-surface-container p-6", className)}>
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
    <div className="flex items-center justify-between py-4 border-b border-outline-variant">
      <div>
        <h1 className="text-[24px] font-medium tracking-[-0.01em] text-on-surface" style={{ fontFamily: "var(--font-work-sans)" }}>
          {title}
        </h1>
        {description && (
          <p className="mt-1 text-sm text-on-surface-variant">{description}</p>
        )}
      </div>
      {children && <div className="flex items-center gap-3">{children}</div>}
    </div>
  );
}

interface MetricCardProps {
  label: string;
  value: string;
  change?: string;
  positive?: boolean;
}

export function MetricCard({ label, value, change, positive }: MetricCardProps) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant">
        {label}
      </span>
      <span className="text-[13px] font-medium text-on-surface font-mono tracking-[0.02em]">
        {value}
      </span>
      {change && (
        <span
          className={cn(
            "text-[11px] font-semibold",
            positive ? "text-primary" : "text-error"
          )}
        >
          {change}
        </span>
      )}
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
    <div className="w-full overflow-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-outline-variant bg-surface-container-low">
            {columns.map((col) => (
              <th
                key={String(col.key)}
                className="h-10 px-4 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant"
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
                className="border-b border-outline-variant hover:bg-surface-container-low"
              >
                {columns.map((col) => (
                  <td
                    key={String(col.key)}
                    className="h-10 px-4 text-sm text-on-surface"
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
