import { useCallback, useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { useAuth } from '@/features/auth/AuthContext'
import { contentApi } from './api'
import type { Subject, ContentItem, Page, BlockInput, BlockType } from './types'
import { BLOCK_TYPES } from './types'
import BlockEditor from './BlockEditor'

// ── Default data templates ────────────────────────────────────────────────────

const defaultData: Record<BlockType, Record<string, unknown>> = {
  text: { content: '' },
  article: { content: '' },
  speech: { source_type: 'tts', text: '' },
  flash_card: { front: '', back: '', interaction_mode: 'flip' },
  image: { source_type: 'url', external_url: '' },
  diagram: { diagram_json: {} },
  video: { source_type: 'embed', embed_url: '' },
  translation: { source_language: 'en', source: '', translations: [] },
}

// ── Layout ────────────────────────────────────────────────────────────────────

const navColumnStyle: React.CSSProperties = {
  width: 240,
  borderRight: '1px solid #e0e0e0',
  display: 'flex',
  flexDirection: 'column',
  flexShrink: 0,
  overflowY: 'auto',
}

const navHeaderStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: '0.4rem',
  padding: '0.6rem 1rem 0.4rem',
  borderBottom: '1px solid #e8e8e8',
  flexShrink: 0,
}

const navBodyStyle: React.CSSProperties = {
  padding: '0.75rem 1rem',
  flex: 1,
  overflowY: 'auto',
}

const editorStyle: React.CSSProperties = {
  flex: 1,
  padding: '1.5rem 2rem',
  overflowY: 'auto',
  minWidth: 0,
}

// ── Sidebar: subject list ─────────────────────────────────────────────────────

interface SubjectListProps {
  subjects: Subject[]
  selectedId: string | null
  onSelect: (id: string) => void
  onRefresh: () => void
  isAdmin: boolean
}

function SubjectList({ subjects, selectedId, onSelect, onRefresh, isAdmin }: SubjectListProps) {
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState<string | null>(null)

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    setError(null)
    try {
      const form = new FormData()
      form.append('name', name.trim())
      form.append('description', description.trim())
      await contentApi.createSubject(form)
      setName('')
      setDescription('')
      setCreating(false)
      onRefresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create subject')
    }
  }

  return (
    <div>
      {subjects.map((s) => (
        <div
          key={s.id}
          onClick={() => onSelect(s.id)}
          style={{
            padding: '6px 8px',
            cursor: 'pointer',
            borderRadius: 4,
            background: selectedId === s.id ? '#e8f0fe' : 'transparent',
            marginBottom: 2,
            fontSize: 14,
          }}
        >
          {s.name}
        </div>
      ))}
      {isAdmin && !creating && (
        <button
          onClick={() => setCreating(true)}
          style={{ marginTop: '0.5rem', fontSize: 13, width: '100%' }}
        >
          + New subject
        </button>
      )}
      {creating && (
        <form onSubmit={handleCreate} style={{ marginTop: '0.5rem' }}>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Subject name"
            style={{ width: '100%', boxSizing: 'border-box', marginBottom: 4 }}
          />
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description (optional)"
            style={{ width: '100%', boxSizing: 'border-box', marginBottom: 4 }}
          />
          {error && <p style={{ color: '#c00', fontSize: 12 }}>{error}</p>}
          <div style={{ display: 'flex', gap: 4 }}>
            <button type="submit" style={{ flex: 1 }}>
              Save
            </button>
            <button type="button" onClick={() => setCreating(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}
    </div>
  )
}

// ── Panel: content list ───────────────────────────────────────────────────────

interface ContentListProps {
  subjectId: string
  contents: ContentItem[]
  selectedId: string | null
  onSelect: (id: string) => void
  onRefresh: () => void
}

function ContentList({ subjectId, contents, selectedId, onSelect, onRefresh }: ContentListProps) {
  const [creating, setCreating] = useState(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState<string | null>(null)

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    setError(null)
    try {
      await contentApi.createContent({
        subject_id: subjectId,
        title: title.trim(),
        description: description.trim(),
      })
      setTitle('')
      setDescription('')
      setCreating(false)
      onRefresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create content')
    }
  }

  return (
    <div>
      {contents.map((c) => (
        <div
          key={c.id}
          onClick={() => onSelect(c.id)}
          style={{
            padding: '6px 8px',
            cursor: 'pointer',
            borderRadius: 4,
            background: selectedId === c.id ? '#e8f0fe' : 'transparent',
            marginBottom: 2,
            fontSize: 14,
          }}
        >
          {c.title}
        </div>
      ))}
      {!creating && (
        <button
          onClick={() => setCreating(true)}
          style={{ marginTop: '0.5rem', fontSize: 13, width: '100%' }}
        >
          + New content
        </button>
      )}
      {creating && (
        <form onSubmit={handleCreate} style={{ marginTop: '0.5rem' }}>
          <input
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Content title"
            style={{ width: '100%', boxSizing: 'border-box', marginBottom: 4 }}
          />
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description (optional)"
            style={{ width: '100%', boxSizing: 'border-box', marginBottom: 4 }}
          />
          {error && <p style={{ color: '#c00', fontSize: 12 }}>{error}</p>}
          <div style={{ display: 'flex', gap: 4 }}>
            <button type="submit" style={{ flex: 1 }}>
              Save
            </button>
            <button type="button" onClick={() => setCreating(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}
    </div>
  )
}

// ── Panel: page list ──────────────────────────────────────────────────────────

interface PageListProps {
  contentId: string
  pages: Page[]
  selectedId: string | null
  onSelect: (id: string) => void
  onRefresh: () => void
}

function PageList({ contentId, pages, selectedId, onSelect, onRefresh }: PageListProps) {
  const [creating, setCreating] = useState(false)
  const [pageName, setPageName] = useState('')
  const [error, setError] = useState<string | null>(null)

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    setError(null)
    try {
      await contentApi.createPage(contentId, pageName.trim())
      setPageName('')
      setCreating(false)
      onRefresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create page')
    }
  }

  return (
    <div>
      {pages.map((p) => (
        <div
          key={p.id}
          onClick={() => onSelect(p.id)}
          style={{
            padding: '6px 8px',
            cursor: 'pointer',
            borderRadius: 4,
            background: selectedId === p.id ? '#e8f0fe' : 'transparent',
            marginBottom: 2,
            fontSize: 14,
          }}
        >
          {p.name}
        </div>
      ))}
      {!creating && (
        <button
          onClick={() => setCreating(true)}
          style={{ marginTop: '0.5rem', fontSize: 13, width: '100%' }}
        >
          + New page
        </button>
      )}
      {creating && (
        <form onSubmit={handleCreate} style={{ marginTop: '0.5rem' }}>
          <input
            required
            value={pageName}
            onChange={(e) => setPageName(e.target.value)}
            placeholder="Page name"
            style={{ width: '100%', boxSizing: 'border-box', marginBottom: 4 }}
          />
          {error && <p style={{ color: '#c00', fontSize: 12 }}>{error}</p>}
          <div style={{ display: 'flex', gap: 4 }}>
            <button type="submit" style={{ flex: 1 }}>
              Save
            </button>
            <button type="button" onClick={() => setCreating(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}
    </div>
  )
}

// ── Page editor ───────────────────────────────────────────────────────────────

interface PageEditorProps {
  pageId: string
}

function PageEditor({ pageId }: PageEditorProps) {
  const [blocks, setBlocks] = useState<BlockInput[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    setBlocks([])
    setError(null)
    setSaved(false)
    contentApi
      .getPage(pageId)
      .then(({ page }) => {
        if (page.blocks && page.blocks.length > 0) {
          setBlocks(
            page.blocks.map((b) => ({
              type: b.type,
              position: b.position,
              data: b.data,
            })),
          )
        }
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Failed to load page')
      })
  }, [pageId])

  function addBlock(type: BlockType) {
    const position = blocks.length > 0 ? Math.max(...blocks.map((b) => b.position)) + 1 : 0
    setBlocks((prev) => [...prev, { type, position, data: { ...defaultData[type] } }])
    setSaved(false)
  }

  function updateBlock(index: number, data: Record<string, unknown>) {
    setBlocks((prev) => prev.map((b, i) => (i === index ? { ...b, data } : b)))
    setSaved(false)
  }

  function deleteBlock(index: number) {
    setBlocks((prev) => {
      const next = prev.filter((_, i) => i !== index)
      return next.map((b, i) => ({ ...b, position: i }))
    })
    setSaved(false)
  }

  function moveBlock(index: number, direction: 'up' | 'down') {
    const target = direction === 'up' ? index - 1 : index + 1
    if (target < 0 || target >= blocks.length) return
    setBlocks((prev) => {
      const next = [...prev]
      ;[next[index], next[target]] = [next[target], next[index]]
      return next.map((b, i) => ({ ...b, position: i }))
    })
    setSaved(false)
  }

  async function handleSave() {
    setSaving(true)
    setError(null)
    setSaved(false)
    try {
      const normalised = blocks.map((b, i) => ({ ...b, position: i }))
      await contentApi.replaceBlocks(pageId, normalised)
      setBlocks(normalised)
      setSaved(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save blocks')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '1rem',
        }}
      >
        <h2 style={{ margin: 0, fontSize: 16 }}>Page editor</h2>
        <button onClick={handleSave} disabled={saving} style={{ fontWeight: 600 }}>
          {saving ? 'Saving…' : 'Save'}
        </button>
      </div>

      {error && <p style={{ color: '#c00', marginBottom: '0.5rem' }}>{error}</p>}
      {saved && <p style={{ color: 'green', marginBottom: '0.5rem' }}>Saved.</p>}

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1rem' }}>
        {blocks.map((block, i) => (
          <BlockEditor
            key={i}
            block={block}
            index={i}
            onChange={updateBlock}
            onDelete={deleteBlock}
            onMoveUp={(idx) => moveBlock(idx, 'up')}
            onMoveDown={(idx) => moveBlock(idx, 'down')}
            isFirst={i === 0}
            isLast={i === blocks.length - 1}
          />
        ))}
      </div>

      <div>
        <p style={{ fontSize: 13, fontWeight: 600, marginBottom: '0.25rem' }}>Add block:</p>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {BLOCK_TYPES.map(({ type, label }) => (
            <button key={type} onClick={() => addBlock(type)} style={{ fontSize: 12 }}>
              {label}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

// ── Nav level ─────────────────────────────────────────────────────────────────

type NavLevel = 'subjects' | 'contents' | 'pages'

// ── Main page ─────────────────────────────────────────────────────────────────

export default function ContentPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'

  const [subjects, setSubjects] = useState<Subject[]>([])
  const [contents, setContents] = useState<ContentItem[]>([])
  const [pages, setPages] = useState<Page[]>([])

  const [selectedSubjectId, setSelectedSubjectId] = useState<string | null>(null)
  const [selectedSubjectName, setSelectedSubjectName] = useState<string>('')
  const [selectedContentId, setSelectedContentId] = useState<string | null>(null)
  const [selectedContentTitle, setSelectedContentTitle] = useState<string>('')
  const [selectedPageId, setSelectedPageId] = useState<string | null>(null)

  const [navLevel, setNavLevel] = useState<NavLevel>('subjects')
  const [loadError, setLoadError] = useState<string | null>(null)

  const loadSubjects = useCallback(async () => {
    try {
      const data = await contentApi.listSubjects()
      setSubjects(data.subjects)
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : 'Failed to load subjects')
    }
  }, [])

  const loadContents = useCallback(async (subjectId: string) => {
    try {
      const data = await contentApi.listContents(subjectId)
      setContents(data.contents)
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : 'Failed to load contents')
    }
  }, [])

  const loadPages = useCallback(async (contentId: string) => {
    try {
      const data = await contentApi.listPages(contentId)
      setPages(data.pages)
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : 'Failed to load pages')
    }
  }, [])

  useEffect(() => {
    loadSubjects()
  }, [loadSubjects])

  function selectSubject(subject: Subject) {
    setSelectedSubjectId(subject.id)
    setSelectedSubjectName(subject.name)
    setSelectedContentId(null)
    setSelectedContentTitle('')
    setSelectedPageId(null)
    setContents([])
    setPages([])
    loadContents(subject.id)
    setNavLevel('contents')
  }

  function selectContent(content: ContentItem) {
    setSelectedContentId(content.id)
    setSelectedContentTitle(content.title)
    setSelectedPageId(null)
    setPages([])
    loadPages(content.id)
    setNavLevel('pages')
  }

  function selectPage(id: string) {
    setSelectedPageId(id)
  }

  function goBack() {
    if (navLevel === 'pages') {
      setNavLevel('contents')
      setSelectedPageId(null)
    } else if (navLevel === 'contents') {
      setNavLevel('subjects')
      setSelectedSubjectId(null)
      setSelectedSubjectName('')
      setSelectedContentId(null)
      setSelectedContentTitle('')
      setSelectedPageId(null)
      setContents([])
      setPages([])
    }
  }

  // ── Nav column header ──────────────────────────────────────────────────────

  function NavHeader() {
    if (navLevel === 'subjects') {
      return (
        <div style={navHeaderStyle}>
          <span style={{ fontWeight: 600, fontSize: 14 }}>Subjects</span>
        </div>
      )
    }
    if (navLevel === 'contents') {
      return (
        <div style={navHeaderStyle}>
          <button
            onClick={goBack}
            title="Back to subjects"
            style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontSize: 16, lineHeight: 1 }}
          >
            ←
          </button>
          <span style={{ fontWeight: 600, fontSize: 14, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {selectedSubjectName}
          </span>
        </div>
      )
    }
    // pages
    return (
      <div style={navHeaderStyle}>
        <button
          onClick={goBack}
          title="Back to contents"
          style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, fontSize: 16, lineHeight: 1 }}
        >
          ←
        </button>
        <span style={{ fontWeight: 600, fontSize: 14, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {selectedContentTitle}
        </span>
      </div>
    )
  }

  return (
    <div style={{ fontFamily: 'sans-serif', display: 'flex', height: '100vh', overflow: 'hidden' }}>
      {/* Single-column drill-down nav */}
      <div style={navColumnStyle}>
        <NavHeader />
        <div style={navBodyStyle}>
          {loadError && <p style={{ color: '#c00', fontSize: 12 }}>{loadError}</p>}

          {navLevel === 'subjects' && (
            <SubjectList
              subjects={subjects}
              selectedId={selectedSubjectId}
              onSelect={(id) => {
                const s = subjects.find((s) => s.id === id)
                if (s) selectSubject(s)
              }}
              onRefresh={loadSubjects}
              isAdmin={isAdmin}
            />
          )}

          {navLevel === 'contents' && selectedSubjectId && (
            <ContentList
              subjectId={selectedSubjectId}
              contents={contents}
              selectedId={selectedContentId}
              onSelect={(id) => {
                const c = contents.find((c) => c.id === id)
                if (c) selectContent(c)
              }}
              onRefresh={() => loadContents(selectedSubjectId)}
            />
          )}

          {navLevel === 'pages' && selectedContentId && (
            <PageList
              contentId={selectedContentId}
              pages={pages}
              selectedId={selectedPageId}
              onSelect={selectPage}
              onRefresh={() => loadPages(selectedContentId)}
            />
          )}
        </div>
      </div>

      {/* Page editor — fills remaining width */}
      <div style={editorStyle}>
        {selectedPageId ? (
          <PageEditor pageId={selectedPageId} />
        ) : (
          <div style={{ color: '#888', marginTop: '4rem', textAlign: 'center', fontSize: 15 }}>
            {navLevel === 'pages'
              ? 'Select or create a page to start editing.'
              : navLevel === 'contents'
                ? 'Select or create content, then open a page.'
                : 'Select a subject to get started.'}
          </div>
        )}
      </div>
    </div>
  )
}
