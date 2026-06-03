import { type FormEvent, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ApiError } from '@/api/client'
import { authApi } from './api'

export default function SetupPage() {
  const navigate = useNavigate()

  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)
  const [alreadyDone, setAlreadyDone] = useState(false)

  // If an admin already exists, show a message rather than a form.
  useEffect(() => {
    authApi
      .setupStatus()
      .then(({ needs_setup }) => {
        if (!needs_setup) setAlreadyDone(true)
      })
      .catch(() => {
        // If the check fails, still render the form — the POST will reject with 409 if needed.
      })
      .finally(() => setChecking(false))
  }, [])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await authApi.registerAdmin({ username, email, password })
      navigate('/login', { replace: true })
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setAlreadyDone(true)
      } else {
        setError(err instanceof ApiError ? err.message : 'Registration failed')
      }
    } finally {
      setLoading(false)
    }
  }

  if (checking) return null

  if (alreadyDone) {
    return (
      <main style={{ fontFamily: 'sans-serif', maxWidth: 400, margin: '4rem auto', padding: '0 1rem' }}>
        <h1>Setup complete</h1>
        <p>An admin account already exists.</p>
        <a href="/login">Go to login</a>
      </main>
    )
  }

  return (
    <main style={{ fontFamily: 'sans-serif', maxWidth: 400, margin: '4rem auto', padding: '0 1rem' }}>
      <h1>Create admin account</h1>
      <p>This page is only available before the first admin is registered.</p>

      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
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
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
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
            autoComplete="new-password"
            required
            minLength={8}
            style={{ display: 'block', width: '100%', marginTop: '0.25rem', padding: '0.5rem' }}
          />
        </label>

        {error && <p style={{ color: 'crimson', margin: 0 }}>{error}</p>}

        <button type="submit" disabled={loading} style={{ padding: '0.6rem', cursor: 'pointer' }}>
          {loading ? 'Creating account…' : 'Create admin account'}
        </button>
      </form>
    </main>
  )
}
