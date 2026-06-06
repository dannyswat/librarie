export interface Subject {
  id: string
  name: string
  description: string
  cover_image_key?: string | null
  cover_image_url?: string | null
  position: number
  created_at: string
}

export interface Topic {
  id: string
  name: string
  description: string
  created_at: string
}

export interface ContentItem {
  id: string
  subject_id: string
  title: string
  description: string
  position: number
  created_at: string
  updated_at: string
  topics?: Topic[]
}

export interface Page {
  id: string
  content_id: string
  name: string
  position: number
  created_at: string
  updated_at: string
  blocks?: Block[]
}

export type BlockType =
  | 'text'
  | 'article'
  | 'speech'
  | 'flash_card'
  | 'image'
  | 'diagram'
  | 'video'
  | 'translation'

export const BLOCK_TYPES: { type: BlockType; label: string }[] = [
  { type: 'text', label: 'Text' },
  { type: 'article', label: 'Article' },
  { type: 'speech', label: 'Speech' },
  { type: 'flash_card', label: 'Flash Card' },
  { type: 'image', label: 'Image' },
  { type: 'diagram', label: 'Diagram' },
  { type: 'video', label: 'Video' },
  { type: 'translation', label: 'Translation' },
]

export interface Block {
  id: string
  page_id: string
  type: BlockType
  position: number
  data: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface BlockInput {
  type: BlockType
  position: number
  data: Record<string, unknown>
}

// Block data types

export interface TextBlockData {
  content: string
}

export interface SpeechBlockData {
  source_type: 'tts' | 'stt' | 'upload' | 'recorded'
  audio_url?: string
  storage_key?: string
  text?: string
  voice?: string
}

export interface FlashCardBlockData {
  front: string
  back: string
  interaction_mode: 'flip' | 'type' | 'choice'
}

export interface ImageBlockData {
  source_type: 'url' | 'upload' | 'ai'
  external_url?: string
  storage_key?: string
  alt?: string
}

export interface DiagramBlockData {
  diagram_json: Record<string, unknown>
}

export interface VideoBlockData {
  source_type: 'embed' | 'upload' | 'ai'
  embed_url?: string
  storage_key?: string
}

export interface TranslationEntry {
  language: string
  text: string
}

export interface TranslationBlockData {
  source_language: string
  source: string
  translations: TranslationEntry[]
}
