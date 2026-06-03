import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi } from './api'
import type { CapabilityConfig } from './types'

const defaultProviderKeys = ['openai', 'anthropic']

export default function AIProvidersAdminPage() {
  const [providerKey, setProviderKey] = useState('openai')
  const [capabilities, setCapabilities] = useState<CapabilityConfig[]>([])
  const [credentialsDraft, setCredentialsDraft] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [savingCapability, setSavingCapability] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [status, setStatus] = useState<string | null>(null)

  async function loadCapabilities(nextProviderKey = providerKey) {
    setLoading(true)
    setError(null)
    setStatus(null)
    try {
      const data = await adminApi.listProviderCapabilities(nextProviderKey)
      setCapabilities(data.capabilities)
      setCredentialsDraft(Object.fromEntries(data.capabilities.map((c) => [c.capability, '{}'])))
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load provider capabilities'
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCapabilities()
  }, [])

  async function onLoadProvider(e: FormEvent) {
    e.preventDefault()
    await loadCapabilities(providerKey.trim())
  }

  async function onSaveCapability(capability: CapabilityConfig) {
    setSavingCapability(capability.capability)
    setError(null)
    setStatus(null)

    const rawCredentials = credentialsDraft[capability.capability] ?? '{}'
    let credentials: Record<string, unknown> = {}
    try {
      const parsed = JSON.parse(rawCredentials)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        credentials = parsed as Record<string, unknown>
      } else {
        throw new Error('Credentials must be a JSON object')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Invalid credentials JSON'
      setError(message)
      setSavingCapability(null)
      return
    }

    try {
      const updated = await adminApi.upsertProviderCapability(providerKey, capability.capability, {
        model: capability.model,
        is_enabled: capability.is_enabled,
        credentials,
      })
      setCapabilities((current) =>
        current.map((item) => (item.capability === updated.capability ? updated : item)),
      )
      setStatus(`Saved ${updated.capability} for ${updated.provider_key}`)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to save capability'
      setError(message)
    } finally {
      setSavingCapability(null)
    }
  }

  return (
    <main style={{ fontFamily: 'sans-serif', padding: '2rem', maxWidth: 960, margin: '0 auto' }}>
      <h1>AI Provider Configuration</h1>
      <p>Configure model, enabled state, and encrypted credentials per capability.</p>

      <form onSubmit={onLoadProvider} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        <input
          type="text"
          list="provider-key-options"
          value={providerKey}
          onChange={(e) => setProviderKey(e.target.value)}
          placeholder="provider key"
          required
          style={{ padding: '0.6rem', flex: 1 }}
        />
        <datalist id="provider-key-options">
          {defaultProviderKeys.map((k) => (
            <option key={k} value={k} />
          ))}
        </datalist>
        <button type="submit">Load Provider</button>
      </form>

      {status && <p style={{ color: 'green' }}>{status}</p>}
      {error && <p style={{ color: 'crimson' }}>{error}</p>}
      {loading && <p>Loading capabilities...</p>}

      {!loading &&
        capabilities.map((capability) => (
          <section
            key={capability.capability}
            style={{ border: '1px solid #ddd', borderRadius: 8, padding: '1rem', marginBottom: '1rem' }}
          >
            <h2 style={{ marginTop: 0 }}>{capability.capability}</h2>

            <label style={{ display: 'block', marginBottom: '0.4rem' }}>
              Model
              <input
                type="text"
                value={capability.model}
                onChange={(e) =>
                  setCapabilities((current) =>
                    current.map((item) =>
                      item.capability === capability.capability
                        ? { ...item, model: e.target.value }
                        : item,
                    ),
                  )
                }
                style={{ width: '100%', marginTop: '0.3rem', padding: '0.6rem' }}
              />
            </label>

            <label style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.8rem' }}>
              <input
                type="checkbox"
                checked={capability.is_enabled}
                onChange={(e) =>
                  setCapabilities((current) =>
                    current.map((item) =>
                      item.capability === capability.capability
                        ? { ...item, is_enabled: e.target.checked }
                        : item,
                    ),
                  )
                }
              />
              Enabled
            </label>

            <label style={{ display: 'block' }}>
              Credentials JSON
              <textarea
                rows={5}
                value={credentialsDraft[capability.capability] ?? '{}'}
                onChange={(e) =>
                  setCredentialsDraft((current) => ({
                    ...current,
                    [capability.capability]: e.target.value,
                  }))
                }
                style={{ width: '100%', marginTop: '0.3rem', padding: '0.6rem' }}
              />
            </label>

            <button
              type="button"
              onClick={() => onSaveCapability(capability)}
              disabled={savingCapability === capability.capability}
              style={{ marginTop: '0.8rem' }}
            >
              {savingCapability === capability.capability ? 'Saving...' : 'Save Capability'}
            </button>
          </section>
        ))}
    </main>
  )
}
