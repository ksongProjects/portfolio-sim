"use client";

import { useState } from "react";

export type SimulationStatus = "idle" | "running" | "completed" | "error";

export function useSimulation() {
  const [status, setStatus] = useState<SimulationStatus>("idle");
  const [progress, setProgress] = useState(0);
  const [result, setResult] = useState<string | null>(null);

  const run = async () => {
    setStatus("running");
    setProgress(0);
    setResult(null);
    for (let i = 0; i <= 100; i += 10) {
      await new Promise((r) => setTimeout(r, 150));
      setProgress(i);
    }
    setStatus("completed");
    setResult("Simulation completed. 847 trades executed, +2.34% return.");
  };

  const reset = () => {
    setStatus("idle");
    setProgress(0);
    setResult(null);
  };

  return { status, progress, result, run, reset };
}

export type ToastVariant = "success" | "error" | "info";

interface Toast {
  id: number;
  message: string;
  variant: ToastVariant;
}

let toastCounter = 0;

export function useToasts() {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const addToast = (message: string, variant: ToastVariant = "info") => {
    const id = ++toastCounter;
    setToasts((prev) => [...prev, { id, message, variant }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 3500);
  };

  return { toasts, addToast };
}
