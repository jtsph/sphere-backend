export const API_BASE = '/api/v1';

export type ApiResponse<T> = T | { message: string };

export async function fetchJson<T>(path: string, options: RequestInit = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    ...options
  });

  if (!response.ok) {
    const errorBody = await response.json().catch(() => ({ message: response.statusText }));
    throw new Error(errorBody.message || response.statusText);
  }

  return response.json() as Promise<T>;
}
