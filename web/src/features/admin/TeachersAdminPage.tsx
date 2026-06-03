import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi } from './api'
import type { Teacher } from './types'

function parseSubjectIds(raw: string): string[] {
  return raw
    .split(/[\n,]/g)
    .map((v) => v.trim())
    .filter((v) => v.length > 0)
}

export default function TeachersAdminPage() {
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteStatus, setInviteStatus] = useState<string | null>(null)
  const [subjectDraftByTeacher, setSubjectDraftByTeacher] = useState<Record<string, string>>({})
  const [savingTeacherId, setSavingTeacherId] = useState<string | null>(null)

  async function loadTeachers() {
    setLoading(true)
    setError(null)
    try {
      const data = await adminApi.listTeachers()
      setTeachers(data.teachers)
      setSubjectDraftByTeacher(
        Object.fromEntries(
          data.teachers.map((t) => [
            t.id,
            t.subjects.map((s) => s.id).join('\n'),
          ]),
        ),
      )
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load teachers'
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadTeachers()
  }, [])

  async function onInvite(e: FormEvent) {
    e.preventDefault()
    setInviteStatus(null)
    setError(null)
    try {
      const created = await adminApi.inviteTeacher({ email: inviteEmail.trim() })
      setInviteEmail('')
      setInviteStatus(`Invitation created for ${created.email}`)
      await loadTeachers()
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to invite teacher'
      setError(message)
    }
  }

  async function onSaveSubjects(teacherId: string) {
    setSavingTeacherId(teacherId)
    setError(null)
    try {
      const subject_ids = parseSubjectIds(subjectDraftByTeacher[teacherId] ?? '')
      const updated = await adminApi.replaceTeacherSubjects(teacherId, { subject_ids })
      setTeachers((current) =>
        current.map((t) =>
          t.id === teacherId
            ? {
                ...t,
                subjects: updated.subjects,
              }
            : t,
        ),
      )
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to save assignments'
      setError(message)
    } finally {
      setSavingTeacherId(null)
    }
  }

  const sortedTeachers = useMemo(
    () => [...teachers].sort((a, b) => a.username.localeCompare(b.username)),
    [teachers],
  )

  return (
    <main style={{ fontFamily: 'sans-serif', padding: '2rem', maxWidth: 960, margin: '0 auto' }}>
      <h1>Teacher Management</h1>
      <p>Invite teachers and assign subjects by ID.</p>

      <form onSubmit={onInvite} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem' }}>
        <input
          type="email"
          required
          placeholder="teacher@school.edu"
          value={inviteEmail}
          onChange={(e) => setInviteEmail(e.target.value)}
          style={{ flex: 1, padding: '0.6rem' }}
        />
        <button type="submit">Invite Teacher</button>
      </form>

      {inviteStatus && <p style={{ color: 'green' }}>{inviteStatus}</p>}
      {error && <p style={{ color: 'crimson' }}>{error}</p>}
      {loading && <p>Loading teachers...</p>}

      {!loading && sortedTeachers.length === 0 && <p>No teachers found yet.</p>}

      {!loading &&
        sortedTeachers.map((teacher) => (
          <section
            key={teacher.id}
            style={{
              border: '1px solid #ddd',
              borderRadius: 8,
              padding: '1rem',
              marginBottom: '1rem',
            }}
          >
            <h2 style={{ marginTop: 0 }}>{teacher.username}</h2>
            <p style={{ margin: '0.2rem 0' }}>{teacher.email}</p>
            <p style={{ margin: '0.2rem 0' }}>
              Assigned subjects:{' '}
              {teacher.subjects.length > 0
                ? teacher.subjects.map((s) => `${s.name} (${s.id})`).join(', ')
                : 'none'}
            </p>

            <label htmlFor={`subject-ids-${teacher.id}`} style={{ display: 'block', marginTop: '0.8rem' }}>
              Subject IDs (one per line or comma-separated)
            </label>
            <textarea
              id={`subject-ids-${teacher.id}`}
              rows={4}
              value={subjectDraftByTeacher[teacher.id] ?? ''}
              onChange={(e) =>
                setSubjectDraftByTeacher((current) => ({
                  ...current,
                  [teacher.id]: e.target.value,
                }))
              }
              style={{ width: '100%', marginTop: '0.4rem', padding: '0.6rem' }}
            />

            <button
              type="button"
              onClick={() => onSaveSubjects(teacher.id)}
              disabled={savingTeacherId === teacher.id}
              style={{ marginTop: '0.6rem' }}
            >
              {savingTeacherId === teacher.id ? 'Saving...' : 'Save Assignments'}
            </button>
          </section>
        ))}
    </main>
  )
}
