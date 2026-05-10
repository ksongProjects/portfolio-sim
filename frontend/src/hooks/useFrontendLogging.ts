"use client";

import { useCallback, useRef } from "react";

export type LogLevel = "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL";

export interface FrontendLogEntry {
  id: string;
  timestamp: string;
  level: LogLevel;
  service: string;
  component: string | null;
  message: string;
  metadata: Record<string, unknown> | null;
  trace_id: string | null;
  span_id: string | null;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

let logBuffer: FrontendLogEntry[] = [];
let flushScheduled = false;

function generateId(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

async function flushLogs(): Promise<void> {
  if (logBuffer.length === 0) return;

  const logsToSend = [...logBuffer];
  logBuffer = [];

  try {
    for (const entry of logsToSend) {
      const response = await fetch(`${API_BASE}/api/logs`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(entry),
      });
      if (!response.ok) {
      }
    }
  } catch {
    logBuffer = [...logsToSend, ...logBuffer];
  }
  flushScheduled = false;
}

function scheduleFlush(): void {
  if (flushScheduled) return;
  flushScheduled = true;
  setTimeout(flushLogs, 1000);
}

const LOGGING_ENDPOINT = "/api/logs";

function createLogEntry(level: LogLevel, message: string, metadata?: Record<string, unknown>, component?: string): FrontendLogEntry {
  return {
    id: generateId(),
    timestamp: new Date().toISOString(),
    level,
    service: "frontend",
    component: component || null,
    message,
    metadata: metadata || null,
    trace_id: null,
    span_id: null,
  };
}

export function useFrontendLogging() {
  const initialisedRef = useRef(false);

  const log = useCallback((level: LogLevel, message: string, metadata?: Record<string, unknown>, component?: string) => {
  }, []);

  const info = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
  }, []);

  const warn = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
  }, []);

  const error = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
  }, []);

  const debug = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
  }, []);

  const logNavigation = useCallback((page: string) => {
  }, []);

  const logAction = useCallback((action: string, metadata?: Record<string, unknown>) => {
  }, []);

  const logAPICall = useCallback((method: string, path: string, status: number, durationMs: number, error?: string) => {
  }, []);

  return {
    log,
    info,
    warn,
    error,
    debug,
    logNavigation,
    logAction,
    logAPICall,
    flush: flushLogs,
  };
}

export function setupGlobalErrorHandling() {
}

export function logPageView(page: string) {
}

export function logUserAction(action: string, metadata?: Record<string, unknown>) {
}

export function logComponentMount(component: string) {
}