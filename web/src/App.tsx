import { BrowserRouter, Link, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from '@/features/auth/AuthContext'
import LoginPage from '@/features/auth/LoginPage'
import InvitePage from '@/features/auth/InvitePage'
import SetupPage from '@/features/auth/SetupPage'
import ProtectedRoute from '@/components/ProtectedRoute'
import TeachersAdminPage from '@/features/admin/TeachersAdminPage'
import AIProvidersAdminPage from '@/features/admin/AIProvidersAdminPage'
import ContentPage from '@/features/content/ContentPage'

function HomePage() {
  const { user, logout } = useAuth()
  return (
    <main style={{ fontFamily: 'sans-serif', padding: '2rem' }}>
      <h1>Librarie</h1>
      <p>Welcome, <strong>{user?.username}</strong> ({user?.role})</p>
      {(user?.role === 'teacher' || user?.role === 'admin') && (
        <Link to="/content">Content</Link>
      )}
      {user?.role === 'admin' && (
        <nav style={{ display: 'flex', gap: '0.75rem', marginBottom: '1rem' }}>
          <Link to="/admin/teachers">Manage Teachers</Link>
          <Link to="/admin/ai-providers">AI Providers</Link>
        </nav>
      )}
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
      <Route path="/setup" element={<SetupPage />} />
      <Route path="/invite/:token" element={<InvitePage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <HomePage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/teachers"
        element={
          <ProtectedRoute requireRole="admin">
            <TeachersAdminPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/ai-providers"
        element={
          <ProtectedRoute requireRole="admin">
            <AIProvidersAdminPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/content"
        element={
          <ProtectedRoute requireRole="teacher">
            <ContentPage />
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
