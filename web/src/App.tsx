import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from '@/features/auth/AuthContext'
import LoginPage from '@/features/auth/LoginPage'
import InvitePage from '@/features/auth/InvitePage'
import ProtectedRoute from '@/components/ProtectedRoute'

function HomePage() {
  const { user, logout } = useAuth()
  return (
    <main style={{ fontFamily: 'sans-serif', padding: '2rem' }}>
      <h1>Librarie</h1>
      <p>Welcome, <strong>{user?.username}</strong> ({user?.role})</p>
      <button onClick={logout}>Sign out</button>
    </main>
  )
}

function NotFoundPage() {
  return (
    <main style={{ fontFamily: 'sans-serif', padding: '2rem' }}>
      <h1>404 — Page not found</h1>
    </main>
  )
}

function AppRoutes() {
  const { user, loading } = useAuth()

  if (loading) return null

  return (
    <Routes>
      <Route path="/login" element={user ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route path="/invite/:token" element={<InvitePage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <HomePage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  )
}
