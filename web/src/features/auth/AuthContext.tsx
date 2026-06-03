import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import { clearAuthTokens, setAuthTokens } from '@/api/client'
import { authApi } from './api'
import type { User } from './types'

interface AuthState {
  user: User | null
  /** True while the initial /auth/me check is in flight */
  loading: boolean
}

interface AuthContextValue extends AuthState {
  login: (username: string, password: string) => Promise<User>
  passkeyLogin: () => Promise<User>
  logout: () => Promise<void>
  /** Call after successful registration to refresh auth state */
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ user: null, loading: true })

  const refresh = useCallback(async () => {
    try {
      const user = await authApi.me()
      setState({ user, loading: false })
    } catch {
      clearAuthTokens()
      setState({ user: null, loading: false })
    }
  }, [])

  // Check session on mount
  useEffect(() => {
    refresh()
  }, [refresh])

  const login = useCallback(
    async (username: string, password: string) => {
      const auth = await authApi.login({ username, password })
      setAuthTokens(auth.access_token, auth.refresh_token)
      setState({ user: auth.user, loading: false })
      return auth.user
    },
    [],
  )

  const passkeyLogin = useCallback(async () => {
    const auth = await authApi.passkeyLogin()
    setAuthTokens(auth.access_token, auth.refresh_token)
    setState({ user: auth.user, loading: false })
    return auth.user
  }, [])

  const logout = useCallback(async () => {
    await authApi.logout()
    clearAuthTokens()
    setState({ user: null, loading: false })
  }, [])

  return (
    <AuthContext.Provider value={{ ...state, login, passkeyLogin, logout, refresh }}>
      {children}
    </AuthContext.Provider>
  )
}

/** Hook to access the auth context. Must be used inside <AuthProvider>. */
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within <AuthProvider>')
  return ctx
}
