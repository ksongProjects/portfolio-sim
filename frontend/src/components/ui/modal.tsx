"use client";

import { useRef } from "react";
import { Dialog } from "@base-ui/react/dialog";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
  hideCloseButton?: boolean;
  preventCloseOnBackdropClick?: boolean;
}

export function Modal({
  open,
  onClose,
  title,
  description,
  children,
  className,
  hideCloseButton = false,
  preventCloseOnBackdropClick = false,
}: ModalProps) {
  return (
    <Dialog.Root open={open} onOpenChange={(open) => !open && onClose()}>
      <Dialog.Backdrop
        className={cn(
          "fixed inset-0 z-50 bg-black/50",
          "data-[state=open]:animate-in data-[state=closed]:animate-out",
          "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
        )}
        onClick={preventCloseOnBackdropClick ? undefined : onClose}
      />
      <Dialog.Portal>
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <Dialog.Popup
            className={cn(
              "bg-surface border border-outline-variant w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col rounded-lg shadow-xl",
              "data-[state=open]:animate-in data-[state=closed]:animate-out",
              "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
              "data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
              "data-[state=closed]:slide-out-to-anchor-center data-[state=open]:slide-in-from-anchor-center",
              className
            )}
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-outline-variant">
              <div>
                <Dialog.Title className="text-base font-semibold text-on-surface">
                  {title}
                </Dialog.Title>
                {description && (
                  <p className="text-xs text-on-surface-variant mt-0.5">
                    {description}
                  </p>
                )}
              </div>
              {!hideCloseButton && (
                <button
                  onClick={onClose}
                  className="text-on-surface-variant hover:text-on-surface transition-colors"
                  aria-label="Close"
                >
                  <X className="h-5 w-5" />
                </button>
              )}
            </div>
            <div className="flex-1 overflow-y-auto p-5">{children}</div>
          </Dialog.Popup>
        </div>
      </Dialog.Portal>
    </Dialog.Root>
  );
}