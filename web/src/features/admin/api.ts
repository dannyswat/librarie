import { api } from '@/api/client'
import type {
  InviteTeacherRequest,
  InviteTeacherResponse,
  ListProviderCapabilitiesResponse,
  ListTeachersResponse,
  ReplaceTeacherSubjectsRequest,
  ReplaceTeacherSubjectsResponse,
  UpsertCapabilityRequest,
  CapabilityConfig,
} from './types'

export const adminApi = {
  listTeachers: () => api.get<ListTeachersResponse>('/admin/teachers'),

  inviteTeacher: (req: InviteTeacherRequest) =>
    api.post<InviteTeacherResponse>('/admin/teachers/invite', req),

  replaceTeacherSubjects: (teacherId: string, req: ReplaceTeacherSubjectsRequest) =>
    api.put<ReplaceTeacherSubjectsResponse>(`/admin/teachers/${encodeURIComponent(teacherId)}/subjects`, req),

  listProviderCapabilities: (providerKey: string) =>
    api.get<ListProviderCapabilitiesResponse>(`/admin/ai/providers/${encodeURIComponent(providerKey)}/capabilities`),

  upsertProviderCapability: (providerKey: string, capability: string, req: UpsertCapabilityRequest) =>
    api.put<CapabilityConfig>(
      `/admin/ai/providers/${encodeURIComponent(providerKey)}/capabilities/${encodeURIComponent(capability)}`,
      req,
    ),
}
