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
        console.warn("Failed to send log to server", response.status);
      }
    }
  } catch (err) {
    console.warn("Failed to send logs to server", err);
    logBuffer = [...logsToSend, ...logBuffer];
  }
  flushScheduled = false;
}

function scheduleFlush(): void {
  if (flushScheduled) return;
  flushScheduled = true;
  setTimeout(flushLogs, 1000);
}

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
    const entry = createLogEntry(level, message, metadata, component);
    logBuffer.push(entry);
    console.log(`[${level}] ${message}`, metadata || "");
    scheduleFlush();
  }, []);

  const info = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
    log("INFO", message, metadata, component);
  }, [log]);

  const warn = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
    log("WARN", message, metadata, component);
  }, [log]);

  const error = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
    log("ERROR", message, metadata, component);
  }, [log]);

  const debug = useCallback((message: string, metadata?: Record<string, unknown>, component?: string) => {
    log("DEBUG", message, metadata, component);
  }, [log]);

  const logNavigation = useCallback((page: string) => {
    log("INFO", "Navigation", { page, type: "navigation" }, "router");
  }, [log]);

  const logAction = useCallback((action: string, metadata?: Record<string, unknown>) => {
    log("INFO", `User Action: ${action}`, { ...metadata, type: "user_action" }, "action");
  }, [log]);

  const logAPICall = useCallback((method: string, path: string, status: number, durationMs: number, error?: string) => {
    const meta = { method, path, status, duration_ms: durationMs, type: "api_call" };
    if (error) {
      log("ERROR", `API Call Error: ${method} ${path}`, meta, "fetch");
    } else {
      log("INFO", `API Call: ${method} ${path} ${status}`, meta, "fetch");
    }
  }, [log]);

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
  if (typeof window === "undefined") return;

  window.onerror = (message, source, lineno, colno, error) => {
    const entry = createLogEntry("ERROR", `Uncaught error: ${message}`, {
      source,
      lineno,
      colno,
      error: error?.stack || String(error),
    });
    logBuffer.push(entry);
    scheduleFlush();
  };

  window.onunhandledrejection = (event) => {
    const entry = createLogEntry("ERROR", "Unhandled promise rejection", {
      reason: String(event.reason),
    });
    logBuffer.push(entry);
    scheduleFlush();
  };
}

if (typeof window !== "undefined") {
  const originalFetch = window.fetch;
  window.fetch = async function (input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
    const startTime = Date.now();
    const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
    const method = init?.method || "GET";

    try {
      const response = await originalFetch(input, init);
      const duration = Date.now() - startTime;
      const status = response.status;

      const meta = { method, path: url, status, duration_ms: duration, type: "api_call" };
      const entry = createLogEntry(
        status >= 500 ? "ERROR" : status >= 400 ? "WARN" : "INFO",
        `API Call: ${method} ${url} ${status}`,
        meta,
        "fetch"
      );
      logBuffer.push(entry);
      scheduleFlush();

      return response;
    } catch (err) {
      const duration = Date.now() - startTime;
      const entry = createLogEntry("ERROR", `API Call Error: ${method} ${url}`, {
        method,
        path: url,
        error: String(err),
        duration_ms: duration,
        type: "api_call",
      }, "fetch");
      logBuffer.push(entry);
      scheduleFlush();
      throw err;
    }
  };

  let navigationLogged = false;
  const originalPushState = history.pushState;
  history.pushState = function (data: unknown, title: string, url?: string | URL): void {
    if (!navigationLogged && url) {
      const entry = createLogEntry("INFO", "Navigation", { url: String(url), type: "navigation" }, "router");
      logBuffer.push(entry);
      scheduleFlush();
    }
    navigationLogged = false;
    return originalPushState.apply(history, [data, title, url]);
  };
}

export function logPageView(page: string) {
  if (typeof window === "undefined") return;
  const entry = createLogEntry("INFO", `Page View: ${page}`, { page, type: "page_view" }, "router");
  logBuffer.push(entry);
  scheduleFlush();
}

export function logUserAction(action: string, metadata?: Record<string, unknown>) {
  if (typeof window === "undefined") return;
  const entry = createLogEntry("INFO", `User Action: ${action}`, { ...metadata, type: "user_action" }, "action");
  logBuffer.push(entry);
  scheduleFlush();
}

export function logComponentMount(component: string) {
  if (typeof window === "undefined") return;
  const entry = createLogEntry("DEBUG", `Component mounted: ${component}`, { component, type: "component_mount" }, "react");
  logBuffer.push(entry);
  scheduleFlush();
}