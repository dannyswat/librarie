import { type FormEvent, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ApiError } from '@/api/client'
import { authApi } from './api'
import { useAuth } from './AuthContext'

export default function InvitePage() {
  const { token } = useParams<{ token: string }>()
  const { refresh } = useAuth()
  const navigate = useNavigate()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    if (password !== confirm) {
      setError('Passwords do not match')
      return
    }
    if (!token) {
      setError('Invalid invitation link')
      return
    }

    setLoading(true)
    try {
      await authApi.acceptInvitation(token, { username, password })
      await refresh()
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main style={{ fontFamily: 'sans-serif', maxWidth: 400, margin: '4rem auto', padding: '0 1rem' }}>
      <h1>Create your account</h1>
      <p>You've been invited to Librarie. Choose a username and password to get started.</p>

      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <label>
          Username
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            required
            minLength={3}
            style={{ display: 'block', width: '100%', marginTop: '0.25rem', padding: '0.5rem' }}
          />
        </label>

        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="new-password"
            required
            minLength={8}
            style={{ display: 'block', width: '100%', marginTop: '0.25rem', padding: '0.5rem' }}
          />
        </label>

        <label>
          Confirm password
          <input
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            autoComplete="new-password"
            required
            style={{ display: 'block', width: '100%', marginTop: '0.25rem', padding: '0.5rem' }}
          />
        </label>

        {error && <p style={{ color: 'crimson', margin: 0 }}>{error}</p>}

        <button type="submit" disabled={loading} style={{ padding: '0.6rem', cursor: 'pointer' }}>
          {loading ? 'Creating account…' : 'Create account'}
        </button>
      </form>
    </main>
  )
}
