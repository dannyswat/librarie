/**
 * Base API client.
 *
 * All requests include credentials (session cookie) automatically.
 * Throws an Error with the server's error message on non-2xx responses.
 */

const BASE = '/api/v1'
const ACCESS_TOKEN_KEY = 'librarie_access_token'
const REFRESH_TOKEN_KEY = 'librarie_refresh_token'

type RefreshResponse = {
  access_token: string
  refresh_token: string
}

let refreshPromise: Promise<string | null> | null = null

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setAuthTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

export function clearAuthTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export class ApiError extends Error {
  readonly status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) {
    return refreshPromise
  }

  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    return null
  }

  refreshPromise = (async () => {
    try {
      const res = await fetch(`${BASE}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!res.ok) {
        clearAuthTokens()
        return null
      }

      const data = (await res.json()) as RefreshResponse
      setAuthTokens(data.access_token, data.refresh_token)
      return data.access_token
    } catch {
      clearAuthTokens()
      return null
    } finally {
      refreshPromise = null
    }
  })()

  return refreshPromise
}

async function request<T>(method: string, path: string, body?: unknown, hasRetried = false): Promise<T> {
  const init: RequestInit = {
    method,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  }

  const accessToken = getAccessToken()
  if (accessToken) {
    ;(init.headers as Record<string, string>).Authorization = `Bearer ${accessToken}`
  }

  if (body !== undefined) {
    init.body = JSON.stringify(body)
  }

  let res = await fetch(`${BASE}${path}`, init)

  if (res.status === 401 && !hasRetried && path !== '/auth/login' && path !== '/auth/refresh') {
    const refreshedAccessToken = await refreshAccessToken()
    if (refreshedAccessToken) {
      ;(init.headers as Record<string, string>).Authorization = `Bearer ${refreshedAccessToken}`
      res = await fetch(`${BASE}${path}`, init)
    }
  }

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
