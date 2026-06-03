export interface SubjectSummary {
  id: string
  name: string
  description: string
}

export interface Teacher {
  id: string
  username: string
  email: string
  role: 'teacher'
  created_at: string
  subjects: SubjectSummary[]
}

export interface ListTeachersResponse {
  teachers: Teacher[]
}

export interface ReplaceTeacherSubjectsRequest {
  subject_ids: string[]
}

export interface ReplaceTeacherSubjectsResponse {
  teacher_id: string
  subjects: SubjectSummary[]
}

export interface InviteTeacherRequest {
  email: string
}

export interface InviteTeacherResponse {
  id: string
  email: string
  role: string
  expires_at: string
}

export interface CapabilityConfig {
  provider_key: string
  capability: string
  model: string
  is_enabled: boolean
}

export interface ListProviderCapabilitiesResponse {
  provider_key: string
  capabilities: CapabilityConfig[]
}

export interface UpsertCapabilityRequest {
  model: string
  is_enabled: boolean
  credentials: Record<string, unknown>
}
