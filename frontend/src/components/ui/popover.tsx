"use client"

import * as React from "react"
import { Popover as PopoverPrimitive, PopoverAnchor } from "@base-ui-components/react"

import { cn } from "@/lib/utils"

function Popover({
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Root>) {
  return <PopoverPrimitive.Root data-slot="popover" {...props} />
}

function PopoverTrigger({
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Trigger>) {
  return <PopoverPrimitive.Trigger data-slot="popover-trigger" {...props} />
}

function PopoverAnchor({
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Anchor>) {
  return <PopoverPrimitive.Anchor data-slot="popover-anchor" {...props} />
}

function PopoverBackdrop({
  className,
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Backdrop>) {
  return (
    <PopoverPrimitive.Backdrop
      data-slot="popover-backdrop"
      className={cn(
        "fixed inset-0 z-50 bg-black/50 data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
        className
      )}
      {...props}
    />
  )
}

function PopoverContent({
  className,
  children,
  side = "right",
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Positioner> & {
  side?: "top" | "bottom" | "left" | "right"
}) {
  return (
    <PopoverPrimitive.Positioner side={side} sideOffset={0}>
      <PopoverPrimitive.Popup
        data-slot="popover-content"
        className={cn(
          "fixed z-50 flex h-full w-96 max-w-full flex-col bg-surface-container text-on-surface shadow-lg data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 data-[closed]:animate-out data-[closed]:fade-out-0 data-[closed]:zoom-out-95 data-[open]:animate-in data-[open]:fade-in-0 data-[open]:zoom-in-95",
          side === "right" && "inset-y-0 right-0 border-l",
          side === "left" && "inset-y-0 left-0 border-r",
          side === "bottom" && "inset-x-0 bottom-0 border-t",
          side === "top" && "inset-x-0 top-0 border-b",
          className
        )}
        {...props}
      >
        {children}
      </PopoverPrimitive.Popup>
    </PopoverPrimitive.Positioner>
  )
}

function PopoverArrow({
  ...props
}: React.ComponentProps<typeof PopoverPrimitive.Arrow>) {
  return <PopoverPrimitive.Arrow data-slot="popover-arrow" {...props} />
}

export {
  Popover,
  PopoverAnchor,
  PopoverArrow,
  PopoverBackdrop,
  PopoverContent,
  PopoverTrigger,
}