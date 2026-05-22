import { useEffect } from 'react'
import { Route, Routes } from 'react-router-dom'
import { configureAuth } from './api/client'
import { AuthProvider, useAuth } from './auth/AuthContext'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { Dashboard } from './pages/Dashboard'
import { LedgerDetail } from './pages/LedgerDetail'
import { Ledgers } from './pages/Ledgers'
import { Login } from './pages/Login'

function AuthBridge({ children }: { children: React.ReactNode }) {
  const { getAccessToken, refresh } = useAuth()
  useEffect(() => {
    configureAuth(getAccessToken, refresh)
  }, [getAccessToken, refresh])
  return <>{children}</>
}

export default function App() {
  return (
    <AuthProvider>
      <AuthBridge>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route element={<ProtectedRoute />}>
            <Route element={<Layout />}>
              <Route index element={<Dashboard />} />
              <Route path="ledgers" element={<Ledgers />} />
              <Route path="ledgers/:id" element={<LedgerDetail />} />
            </Route>
          </Route>
        </Routes>
      </AuthBridge>
    </AuthProvider>
  )
}
