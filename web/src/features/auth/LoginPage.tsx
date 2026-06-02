import { type FormEvent, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ApiError } from '@/api/client'
import { useAuth } from './AuthContext'

export default function LoginPage() {
  const { login, passkeyLogin } = useAuth()
  const navigate = useNavigate()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [passkeyLoading, setPasskeyLoading] = useState(false)

  async function handlePasswordLogin(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await login(username, password)
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  async function handlePasskeyLogin() {
    setError(null)
    setPasskeyLoading(true)
    try {
      await passkeyLogin()
      navigate('/', { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Passkey authentication failed')
    } finally {
      setPasskeyLoading(false)
    }
  }

  return (
    <main style={{ fontFamily: 'sans-serif', maxWidth: 400, margin: '4rem auto', padding: '0 1rem' }}>
      <h1>Sign in to Librarie</h1>

      <form onSubmit={handlePasswordLogin} style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        <label>
          Username
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            autoComplete="username"
            required
            style={{ display: 'block', width: '100%', marginTop: '0.25rem', padding: '0.5rem' }}
          />
        </label>

        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
            style={{ display: 'block', width: '100%', marginTop: '0.25rem', padding: '0.5rem' }}
          />
        </label>

        {error && <p style={{ color: 'crimson', margin: 0 }}>{error}</p>}

        <button type="submit" disabled={loading} style={{ padding: '0.6rem', cursor: 'pointer' }}>
          {loading ? 'Signing in…' : 'Sign in'}
        </button>
      </form>

      <hr style={{ margin: '1.5rem 0' }} />

      <button
        onClick={handlePasskeyLogin}
        disabled={passkeyLoading}
        style={{ width: '100%', padding: '0.6rem', cursor: 'pointer' }}
      >
        {passkeyLoading ? 'Authenticating…' : '🔑 Sign in with passkey'}
      </button>
    </main>
  )
}
