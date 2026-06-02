import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import { authApi } from './api'
import type { User } from './types'

interface AuthState {
  user: User | null
  /** True while the initial /auth/me check is in flight */
  loading: boolean
}

interface AuthContextValue extends AuthState {
  login: (username: string, password: string) => Promise<void>
  passkeyLogin: () => Promise<void>
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
      setState({ user: null, loading: false })
    }
  }, [])

  // Check session on mount
  useEffect(() => {
    refresh()
  }, [refresh])

  const login = useCallback(
    async (username: string, password: string) => {
      await authApi.login({ username, password })
      await refresh()
    },
    [refresh],
  )

  const passkeyLogin = useCallback(async () => {
    await authApi.passkeyLogin()
    await refresh()
  }, [refresh])

  const logout = useCallback(async () => {
    await authApi.logout()
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
