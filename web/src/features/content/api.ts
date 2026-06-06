import { api } from '@/api/client'
import type { Subject, Topic, ContentItem, Page, Block, BlockInput } from './types'

// ── Subjects ──────────────────────────────────────────────────────────────────

export const contentApi = {
  // Subjects
  listSubjects: () => api.get<{ subjects: Subject[] }>('/subjects'),

  getSubject: (id: string) => api.get<{ subject: Subject }>(`/subjects/${id}`),

  createSubject: (form: FormData) =>
    fetch('/api/v1/subjects', {
      method: 'POST',
      credentials: 'include',
      headers: (() => {
        const h: Record<string, string> = {}
        const token = localStorage.getItem('librarie_access_token')
        if (token) h.Authorization = `Bearer ${token}`
        return h
      })(),
      body: form,
    }).then(async (res) => {
      if (!res.ok) {
        const json = await res.json().catch(() => ({})) as { message?: string }
        throw new Error(json.message ?? res.statusText)
      }
      return res.json() as Promise<{ subject: Subject }>
    }),

  updateSubject: (id: string, form: FormData) =>
    fetch(`/api/v1/subjects/${id}`, {
      method: 'PUT',
      credentials: 'include',
      headers: (() => {
        const h: Record<string, string> = {}
        const token = localStorage.getItem('librarie_access_token')
        if (token) h.Authorization = `Bearer ${token}`
        return h
      })(),
      body: form,
    }).then(async (res) => {
      if (!res.ok) {
        const json = await res.json().catch(() => ({})) as { message?: string }
        throw new Error(json.message ?? res.statusText)
      }
      return res.json() as Promise<{ subject: Subject }>
    }),

  deleteSubject: (id: string) => api.delete<void>(`/subjects/${id}`),

  // Topics
  listSubjectTopics: (subjectId: string) =>
    api.get<{ topics: Topic[] }>(`/subjects/${subjectId}/topics`),

  addSubjectTopic: (subjectId: string, body: { id?: string; name?: string; description?: string }) =>
    api.post<{ topic: Topic }>(`/subjects/${subjectId}/topics`, body),

  // Contents
  createContent: (body: { subject_id: string; title: string; description?: string }) =>
    api.post<{ content: ContentItem }>('/contents', body),

  getContent: (id: string) => api.get<{ content: ContentItem }>(`/contents/${id}`),

  updateContent: (id: string, body: { title?: string; description?: string; position?: number }) =>
    api.put<{ content: ContentItem }>(`/contents/${id}`, body),

  deleteContent: (id: string) => api.delete<void>(`/contents/${id}`),

  replaceContentTopics: (id: string, topicIds: string[]) =>
    api.put<{ topics: Topic[] }>(`/contents/${id}/topics`, { topic_ids: topicIds }),

  listContents: (subjectId: string) =>
    api.get<{ contents: ContentItem[] }>(`/subjects/${subjectId}/contents`),
  listPages: (contentId: string) =>
    api.get<{ pages: Page[] }>(`/contents/${contentId}/pages`),

  createPage: (contentId: string, name: string) =>
    api.post<{ page: Page }>(`/contents/${contentId}/pages`, { name }),

  getPage: (id: string) => api.get<{ page: Page }>(`/pages/${id}`),

  patchPage: (id: string, body: { name?: string; position?: number }) =>
    api.patch<{ page: Page }>(`/pages/${id}`, body),

  deletePage: (id: string) => api.delete<void>(`/pages/${id}`),

  // Blocks
  replaceBlocks: (pageId: string, blocks: BlockInput[]) =>
    api.put<{ blocks: Block[] }>(`/pages/${pageId}/blocks`, { blocks }),

  replaceBlocksWithFiles: (pageId: string, blocks: BlockInput[], files: Record<number, File>) => {
    const form = new FormData()
    form.append('blocks', JSON.stringify(blocks))
    for (const [idx, file] of Object.entries(files)) {
      form.append(`file_${idx}`, file)
    }
    return fetch(`/api/v1/pages/${pageId}/blocks`, {
      method: 'PUT',
      credentials: 'include',
      headers: (() => {
        const h: Record<string, string> = {}
        const token = localStorage.getItem('librarie_access_token')
        if (token) h.Authorization = `Bearer ${token}`
        return h
      })(),
      body: form,
    }).then(async (res) => {
      if (!res.ok) {
        const json = await res.json().catch(() => ({})) as { message?: string }
        throw new Error(json.message ?? res.statusText)
      }
      return res.json() as Promise<{ blocks: Block[] }>
    })
  },
}
