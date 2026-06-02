import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '@/features/auth/AuthContext'

interface Props {
  children: ReactNode
  /** Minimum role required. If omitted, any authenticated user is allowed. */
  requireRole?: 'admin' | 'teacher' | 'student'
}

const roleRank: Record<string, number> = { admin: 3, teacher: 2, student: 1 }

export default function ProtectedRoute({ children, requireRole }: Props) {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) return null

  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  if (requireRole && (roleRank[user.role] ?? 0) < (roleRank[requireRole] ?? 0)) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}
