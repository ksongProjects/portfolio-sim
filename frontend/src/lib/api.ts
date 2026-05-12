export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export function buildApiUrl(path: string) {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }

  return `${API_BASE}${path}`;
}

export async function fetchJson<T>(
  path: string,
  init?: RequestInit,
  errorMessage = "Request failed"
): Promise<T> {
  const res = await fetch(buildApiUrl(path), init);

  if (!res.ok) {
    throw new Error(errorMessage);
  }

  return res.json() as Promise<T>;
}

export function getErrorMessage(error: unknown, fallback = "Unknown error") {
  return error instanceof Error ? error.message : fallback;
}
