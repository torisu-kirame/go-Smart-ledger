import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { api, ApiError } from '../api/client'
import type { UserInfo } from '../types/auth'

interface AuthState {
  user: UserInfo | null
  accessToken: string | null
  loading: boolean
}

interface AuthContextValue extends AuthState {
  login: (accessToken: string, user: UserInfo, expiresIn: number) => void
  logout: () => Promise<void>
  refresh: () => Promise<boolean>
  getAccessToken: () => string | null
}

const AuthContext = createContext<AuthContextValue | null>(null)

/** 短期 access token 仅存内存，避免 localStorage 被 XSS 读取 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const tokenRef = useRef<string | null>(null)
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)

  const login = useCallback((accessToken: string, u: UserInfo, _expiresIn: number) => {
    tokenRef.current = accessToken
    setUser(u)
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } catch {
      /* ignore */
    }
    tokenRef.current = null
    setUser(null)
  }, [])

  const refresh = useCallback(async () => {
    try {
      const res = await api.refresh()
      tokenRef.current = res.accessToken
      setUser(res.user)
      return true
    } catch {
      tokenRef.current = null
      setUser(null)
      return false
    }
  }, [])

  const getAccessToken = useCallback(() => tokenRef.current, [])

  useEffect(() => {
    void (async () => {
      setLoading(true)
      await refresh()
      setLoading(false)
    })()
  }, [refresh])

  const value = useMemo(
    () => ({
      user,
      accessToken: tokenRef.current,
      loading,
      login,
      logout,
      refresh,
      getAccessToken,
    }),
    [user, loading, login, logout, refresh, getAccessToken],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth outside AuthProvider')
  return ctx
}

export function useAuthGuard() {
  const { user, loading, refresh } = useAuth()
  return { isAuthenticated: !!user && loading === false, loading, refresh }
}

export { ApiError }
