import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import type { BlockType, BlockInput } from './types'

interface BlockEditorProps {
  block: BlockInput
  index: number
  onChange: (index: number, data: Record<string, unknown>) => void
  onDelete: (index: number) => void
  onMoveUp: (index: number) => void
  onMoveDown: (index: number) => void
  isFirst: boolean
  isLast: boolean
}

// ── Rich text editor (text / article) ─────────────────────────────────────────

function RichTextEditor({
  value,
  onChange,
}: {
  value: string
  onChange: (v: string) => void
}) {
  const editor = useEditor({
    extensions: [StarterKit],
    content: value,
    onUpdate: ({ editor }) => {
      onChange(editor.getHTML())
    },
  })

  return (
    <div style={{ border: '1px solid #ccc', borderRadius: 4, minHeight: 80, padding: 4 }}>
      <EditorContent editor={editor} />
    </div>
  )
}

// ── Per-type editors ──────────────────────────────────────────────────────────

function TextEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  return (
    <RichTextEditor
      value={(data.content as string) ?? ''}
      onChange={(v) => onChange({ ...data, content: v })}
    />
  )
}

function SpeechEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  const sourceType = (data.source_type as string) ?? 'tts'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <label style={{ fontSize: 13 }}>Source type</label>
      <select
        value={sourceType}
        onChange={(e) => onChange({ ...data, source_type: e.target.value })}
      >
        <option value="tts">Text-to-Speech</option>
        <option value="stt">Speech-to-Text</option>
        <option value="upload">Upload</option>
        <option value="recorded">Recorded</option>
      </select>
      {sourceType === 'tts' && (
        <textarea
          placeholder="Text to synthesize…"
          value={(data.text as string) ?? ''}
          onChange={(e) => onChange({ ...data, text: e.target.value })}
          rows={3}
          style={{ width: '100%', boxSizing: 'border-box' }}
        />
      )}
      {(sourceType === 'upload' || sourceType === 'recorded') && (
        <>
          {data.storage_key ? (
            <p style={{ fontSize: 13 }}>
              Uploaded: <code>{data.storage_key as string}</code>
            </p>
          ) : (
            <p style={{ fontSize: 13, color: '#888' }}>
              Save with a file attached (use the multipart upload API).
            </p>
          )}
        </>
      )}
    </div>
  )
}

function FlashCardEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <label style={{ fontSize: 13 }}>Front</label>
      <textarea
        value={(data.front as string) ?? ''}
        onChange={(e) => onChange({ ...data, front: e.target.value })}
        rows={2}
        style={{ width: '100%', boxSizing: 'border-box' }}
      />
      <label style={{ fontSize: 13 }}>Back</label>
      <textarea
        value={(data.back as string) ?? ''}
        onChange={(e) => onChange({ ...data, back: e.target.value })}
        rows={2}
        style={{ width: '100%', boxSizing: 'border-box' }}
      />
      <label style={{ fontSize: 13 }}>Interaction mode</label>
      <select
        value={(data.interaction_mode as string) ?? 'flip'}
        onChange={(e) => onChange({ ...data, interaction_mode: e.target.value })}
      >
        <option value="flip">Flip</option>
        <option value="type">Type answer</option>
        <option value="choice">Multiple choice</option>
      </select>
    </div>
  )
}

function ImageEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  const sourceType = (data.source_type as string) ?? 'url'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <label style={{ fontSize: 13 }}>Source type</label>
      <select
        value={sourceType}
        onChange={(e) => onChange({ ...data, source_type: e.target.value })}
      >
        <option value="url">External URL</option>
        <option value="upload">Upload</option>
        <option value="ai">AI Generate</option>
      </select>
      {sourceType === 'url' && (
        <input
          type="url"
          placeholder="https://…"
          value={(data.external_url as string) ?? ''}
          onChange={(e) => onChange({ ...data, external_url: e.target.value })}
          style={{ width: '100%', boxSizing: 'border-box' }}
        />
      )}
      {sourceType === 'upload' && (
        <>
          {data.storage_key ? (
            <p style={{ fontSize: 13 }}>
              Uploaded: <code>{data.storage_key as string}</code>
            </p>
          ) : (
            <p style={{ fontSize: 13, color: '#888' }}>
              Save with a file attached (use the multipart upload API).
            </p>
          )}
        </>
      )}
      {sourceType === 'ai' && (
        <input
          type="text"
          placeholder="AI prompt…"
          value={(data.prompt as string) ?? ''}
          onChange={(e) => onChange({ ...data, prompt: e.target.value })}
          style={{ width: '100%', boxSizing: 'border-box' }}
        />
      )}
      <label style={{ fontSize: 13 }}>Alt text</label>
      <input
        type="text"
        value={(data.alt as string) ?? ''}
        onChange={(e) => onChange({ ...data, alt: e.target.value })}
        style={{ width: '100%', boxSizing: 'border-box' }}
      />
    </div>
  )
}

function DiagramEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  const raw = data.diagram_json
    ? JSON.stringify(data.diagram_json, null, 2)
    : ''
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <label style={{ fontSize: 13 }}>Diagram JSON (Excalidraw)</label>
      <textarea
        rows={8}
        value={raw}
        onChange={(e) => {
          try {
            const parsed = JSON.parse(e.target.value) as Record<string, unknown>
            onChange({ ...data, diagram_json: parsed })
          } catch {
            // allow partial input
          }
        }}
        style={{ width: '100%', boxSizing: 'border-box', fontFamily: 'monospace', fontSize: 12 }}
      />
    </div>
  )
}

function VideoEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  const sourceType = (data.source_type as string) ?? 'embed'
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <label style={{ fontSize: 13 }}>Source type</label>
      <select
        value={sourceType}
        onChange={(e) => onChange({ ...data, source_type: e.target.value })}
      >
        <option value="embed">Embed URL</option>
        <option value="upload">Upload</option>
        <option value="ai">AI Generate</option>
      </select>
      {sourceType === 'embed' && (
        <input
          type="url"
          placeholder="YouTube / Vimeo embed URL…"
          value={(data.embed_url as string) ?? ''}
          onChange={(e) => onChange({ ...data, embed_url: e.target.value })}
          style={{ width: '100%', boxSizing: 'border-box' }}
        />
      )}
      {sourceType === 'upload' && (
        <>
          {data.storage_key ? (
            <p style={{ fontSize: 13 }}>
              Uploaded: <code>{data.storage_key as string}</code>
            </p>
          ) : (
            <p style={{ fontSize: 13, color: '#888' }}>
              Save with a file attached (use the multipart upload API).
            </p>
          )}
        </>
      )}
    </div>
  )
}

interface TranslationEntry {
  language: string
  text: string
}

function TranslationEditor({
  data,
  onChange,
}: {
  data: Record<string, unknown>
  onChange: (d: Record<string, unknown>) => void
}) {
  const translations = (data.translations as TranslationEntry[]) ?? []

  function updateTranslation(i: number, field: keyof TranslationEntry, value: string) {
    const next = translations.map((t, idx) => (idx === i ? { ...t, [field]: value } : t))
    onChange({ ...data, translations: next })
  }

  function addTranslation() {
    onChange({ ...data, translations: [...translations, { language: '', text: '' }] })
  }

  function removeTranslation(i: number) {
    onChange({ ...data, translations: translations.filter((_, idx) => idx !== i) })
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <label style={{ fontSize: 13 }}>Source language</label>
      <input
        type="text"
        placeholder="en"
        value={(data.source_language as string) ?? ''}
        onChange={(e) => onChange({ ...data, source_language: e.target.value })}
      />
      <label style={{ fontSize: 13 }}>Source text</label>
      <textarea
        rows={2}
        value={(data.source as string) ?? ''}
        onChange={(e) => onChange({ ...data, source: e.target.value })}
        style={{ width: '100%', boxSizing: 'border-box' }}
      />
      <label style={{ fontSize: 13 }}>Translations</label>
      {translations.map((t, i) => (
        <div key={i} style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
          <input
            type="text"
            placeholder="lang"
            value={t.language}
            onChange={(e) => updateTranslation(i, 'language', e.target.value)}
            style={{ width: 60 }}
          />
          <input
            type="text"
            placeholder="translation"
            value={t.text}
            onChange={(e) => updateTranslation(i, 'text', e.target.value)}
            style={{ flex: 1 }}
          />
          <button onClick={() => removeTranslation(i)} style={{ color: '#c00' }}>
            ✕
          </button>
        </div>
      ))}
      <button onClick={addTranslation} style={{ alignSelf: 'flex-start' }}>
        + Add translation
      </button>
    </div>
  )
}

// ── Block type label map ───────────────────────────────────────────────────────

const blockTypeLabel: Record<BlockType, string> = {
  text: 'Text',
  article: 'Article',
  speech: 'Speech',
  flash_card: 'Flash Card',
  image: 'Image',
  diagram: 'Diagram',
  video: 'Video',
  translation: 'Translation',
}

// ── Main BlockEditor component ────────────────────────────────────────────────

export default function BlockEditor({
  block,
  index,
  onChange,
  onDelete,
  onMoveUp,
  onMoveDown,
  isFirst,
  isLast,
}: BlockEditorProps) {
  const label = blockTypeLabel[block.type] ?? block.type

  function renderInner() {
    switch (block.type) {
      case 'text':
      case 'article':
        return <TextEditor data={block.data} onChange={(d) => onChange(index, d)} />
      case 'speech':
        return <SpeechEditor data={block.data} onChange={(d) => onChange(index, d)} />
      case 'flash_card':
        return <FlashCardEditor data={block.data} onChange={(d) => onChange(index, d)} />
      case 'image':
        return <ImageEditor data={block.data} onChange={(d) => onChange(index, d)} />
      case 'diagram':
        return <DiagramEditor data={block.data} onChange={(d) => onChange(index, d)} />
      case 'video':
        return <VideoEditor data={block.data} onChange={(d) => onChange(index, d)} />
      case 'translation':
        return <TranslationEditor data={block.data} onChange={(d) => onChange(index, d)} />
      default:
        return <p style={{ color: '#888', fontSize: 13 }}>Unknown block type.</p>
    }
  }

  return (
    <div
      style={{
        border: '1px solid #ddd',
        borderRadius: 6,
        padding: '0.75rem',
        background: '#fafafa',
      }}
    >
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '0.5rem',
        }}
      >
        <span style={{ fontWeight: 600, fontSize: 13, color: '#555' }}>{label}</span>
        <div style={{ display: 'flex', gap: 4 }}>
          <button
            disabled={isFirst}
            onClick={() => onMoveUp(index)}
            title="Move up"
            style={{ padding: '2px 6px', cursor: isFirst ? 'not-allowed' : 'pointer' }}
          >
            ↑
          </button>
          <button
            disabled={isLast}
            onClick={() => onMoveDown(index)}
            title="Move down"
            style={{ padding: '2px 6px', cursor: isLast ? 'not-allowed' : 'pointer' }}
          >
            ↓
          </button>
          <button
            onClick={() => onDelete(index)}
            title="Delete block"
            style={{ padding: '2px 6px', color: '#c00' }}
          >
            ✕
          </button>
        </div>
      </div>
      {renderInner()}
    </div>
  )
}
