/**
 * Base API client.
 *
 * All requests include credentials (session cookie) automatically.
 * Throws an Error with the server's error message on non-2xx responses.
 */

const BASE = '/api/v1'

export class ApiError extends Error {
  readonly status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const init: RequestInit = {
    method,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  }
  if (body !== undefined) {
    init.body = JSON.stringify(body)
  }

  const res = await fetch(`${BASE}${path}`, init)

  if (!res.ok) {
    let message = res.statusText
    try {
      const json = (await res.json()) as { error?: string }
      if (json.error) message = json.error
    } catch {
      // ignore parse failures
    }
    throw new ApiError(res.status, message)
  }

  // 204 No Content — return undefined cast to T
  if (res.status === 204) return undefined as T

  return res.json() as Promise<T>
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  patch: <T>(path: string, body?: unknown) => request<T>('PATCH', path, body),
  delete: <T>(path: string) => request<T>('DELETE', path),
}
