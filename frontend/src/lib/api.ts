export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const API_REFERRER_POLICY: ReferrerPolicy = "unsafe-url";

export class ApiError extends Error {
  status?: number;
  retryable: boolean;

  constructor(
    message: string,
    options?: { status?: number; retryable?: boolean; cause?: unknown }
  ) {
    super(message);
    this.name = "ApiError";
    this.status = options?.status;
    this.retryable = options?.retryable ?? false;
    if (options?.cause !== undefined) {
      this.cause = options.cause;
    }
  }
}

export function buildApiUrl(path: string) {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }

  return `${API_BASE}${path}`;
}

export function apiFetch(path: string, init?: RequestInit) {
  return fetch(buildApiUrl(path), {
    ...init,
    referrerPolicy: init?.referrerPolicy ?? API_REFERRER_POLICY,
  });
}

export async function fetchJson<T>(
  path: string,
  init?: RequestInit,
  errorMessage = "Request failed"
): Promise<T> {
  let res: Response;
  try {
    res = await apiFetch(path, init);
  } catch (error) {
    throw new ApiError(errorMessage, { retryable: false, cause: error });
  }

  if (!res.ok) {
    throw new ApiError(errorMessage, {
      status: res.status,
      retryable: res.status >= 500 || res.status === 429,
    });
  }

  const text = await res.text();
  if (!text.trim()) {
    throw new ApiError(`${errorMessage}: empty response`, {
      status: res.status,
      retryable: false,
    });
  }

  try {
    return JSON.parse(text) as T;
  } catch (error) {
    throw new ApiError(`${errorMessage}: invalid JSON response`, {
      status: res.status,
      retryable: false,
      cause: error,
    });
  }
}

export function getErrorMessage(error: unknown, fallback = "Unknown error") {
  return error instanceof Error ? error.message : fallback;
}
