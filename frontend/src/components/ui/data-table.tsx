import * as React from "react";
import {
  ColumnDef,
  ColumnFiltersState,
  PaginationState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  Column,
} from "@tanstack/react-table";
import { ChevronDown, ChevronUp, ChevronsUpDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  loading?: boolean;
  emptyMessage?: string;
  searchPlaceholder?: string;
  searchColumn?: string;
  onRowClick?: (row: TData) => void;
  pageSizes?: number[];
  enablePagination?: boolean;
}

function SortIcon({ column }: { column: Column<any, any> }) {
  const sorted = column.getIsSorted();
  if (!sorted) return <ChevronsUpDown className="h-3 w-3 opacity-30" />;
  return sorted === "asc" ? <ChevronUp className="h-3 w-3 text-primary" /> : <ChevronDown className="h-3 w-3 text-primary" />;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  loading = false,
  emptyMessage = "No data available",
  searchPlaceholder,
  searchColumn,
  onRowClick,
  pageSizes = [10, 25, 50, 100],
  enablePagination = true,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [globalFilter, setGlobalFilter] = React.useState("");
  const [pagination, setPagination] = React.useState<PaginationState>({
    pageIndex: 0,
    pageSize: pageSizes[0] || 10,
  });

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      globalFilter,
      pagination,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onGlobalFilterChange: setGlobalFilter,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: enablePagination ? getPaginationRowModel() : undefined,
  });

  const cols = table.getAllColumns().filter((col) => col.getCanSort());
  const pageCount = table.getPageCount();
  const currentPage = pagination.pageIndex + 1;
  const totalRows = table.getFilteredRowModel().rows.length;

  return (
    <div className="space-y-4">
      {searchColumn && (
        <div className="flex items-center gap-2">
          <input
            type="text"
            placeholder={searchPlaceholder || `Search ${searchColumn}...`}
            value={globalFilter ?? ""}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className="h-9 w-[200px] rounded-md border border-outline bg-surface-container px-3 text-sm text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:ring-1 focus:ring-primary"
          />
        </div>
      )}

      <div className="border border-outline-variant/30 rounded-md overflow-hidden">
        <table className="w-full caption-bottom text-sm">
          <thead className="bg-surface-container-low">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  const canSort = header.column.getCanSort();
                  return (
                    <th
                      key={header.id}
                      className={cn(
                        "h-9 px-4 text-left align-middle text-[11px] font-semibold uppercase tracking-[0.08em] text-on-surface-variant",
                        canSort && "cursor-pointer select-none hover:text-on-surface"
                      )}
                      onClick={canSort ? header.column.getToggleSortingHandler() : undefined}
                    >
                      <div className="flex items-center gap-1">
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                        {canSort && <SortIcon column={header.column} />}
                      </div>
                    </th>
                  );
                })}
              </tr>
            ))}
          </thead>
          <tbody className="[&_tr:last-child]:border-0">
            {loading ? (
              Array.from({ length: pagination.pageSize }).map((_, i) => (
                <tr key={i} className="border-b border-outline-variant/30">
                  {columns.map((_, j) => (
                    <td key={j} className="h-9 px-4 align-middle">
                      <div className="h-4 bg-surface-container-high rounded animate-pulse" />
                    </td>
                  ))}
                </tr>
              ))
            ) : table.getRowModel().rows.length > 0 ? (
              table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  className={cn(
                    "border-b border-outline-variant/30 transition-colors hover:bg-surface-container-low/50",
                    onRowClick && "cursor-pointer"
                  )}
                  onClick={() => onRowClick?.(row.original)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="h-9 px-4 align-middle text-sm text-on-surface">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={columns.length} className="h-24 text-center text-on-surface-variant text-sm">
                  {emptyMessage}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {enablePagination && (
        <div className="flex items-center justify-between px-1">
          <div className="flex items-center gap-4">
            <span className="text-sm text-on-surface-variant">
              {totalRows} {totalRows === 1 ? "row" : "rows"}
            </span>
            <select
              value={pagination.pageSize}
              onChange={(e) => table.setPageSize(Number(e.target.value))}
              className="h-8 rounded-md border border-outline bg-surface-container px-2 text-sm text-on-surface focus:outline-none focus:ring-1 focus:ring-primary"
            >
              {pageSizes.map((size) => (
                <option key={size} value={size}>
                  {size} per page
                </option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-1">
            <span className="text-sm text-on-surface-variant mr-2">
              Page {currentPage} of {pageCount || 1}
            </span>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => table.setPageIndex(0)}
              disabled={!table.getCanPreviousPage()}
            >
              <ChevronsLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => table.setPageIndex(pageCount - 1)}
              disabled={!table.getCanNextPage()}
            >
              <ChevronsRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      {cols.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-on-surface-variant">Sort by:</span>
          {cols.map((col) => {
            const sorted = col.getIsSorted();
            return (
              <Button
                key={col.id}
                variant={sorted ? "default" : "ghost"}
                size="sm"
                className="h-7 text-[11px]"
                onClick={() => col.getToggleSortingHandler()?.(undefined)}
              >
                {col.id}
                {sorted === "asc" ? " ↑" : sorted === "desc" ? " ↓" : ""}
              </Button>
            );
          })}
        </div>
      )}
    </div>
  );
}

interface ColumnHelperMenuProps<T> {
  table: ReturnType<typeof useReactTable<T>>;
  className?: string;
}

export function ColumnHelperMenu<T>({ table, className }: ColumnHelperMenuProps<T>) {
  const [open, setOpen] = React.useState(false);

  return (
    <div className={cn("relative", className)}>
      <Button variant="ghost" size="sm" onClick={() => setOpen(!open)}>
        Columns
      </Button>
      {open && (
        <div className="absolute right-0 top-full mt-1 z-50 bg-surface-container-high border border-outline-variant rounded-md shadow-lg p-2 min-w-[150px]">
          <label className="flex items-center gap-2 px-2 py-1 text-sm hover:bg-surface-container-low cursor-pointer">
            <input
              type="checkbox"
              checked={table.getAllColumns().every((c) => c.getIsVisible())}
              onChange={(e) => table.toggleAllColumnsVisible(e.target.checked)}
              className="rounded border-outline"
            />
            Toggle All
          </label>
          <hr className="my-1 border-outline-variant/30" />
          {table.getAllColumns().filter((c) => c.getCanHide()).map((column) => (
            <label key={column.id} className="flex items-center gap-2 px-2 py-1 text-sm hover:bg-surface-container-low cursor-pointer">
              <input
                type="checkbox"
                checked={column.getIsVisible()}
                onChange={(e) => column.toggleVisibility(e.target.checked)}
                className="rounded border-outline"
              />
              {column.id}
            </label>
          ))}
        </div>
      )}
    </div>
  );
}