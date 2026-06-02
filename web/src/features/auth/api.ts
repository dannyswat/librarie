import {
  startAuthentication,
  startRegistration,
  type PublicKeyCredentialCreationOptionsJSON,
  type PublicKeyCredentialRequestOptionsJSON,
} from '@simplewebauthn/browser'
import { api } from '@/api/client'
import type {
  AcceptInvitationRequest,
  CreateInvitationRequest,
  CreateInvitationResponse,
  LoginRequest,
  User,
} from './types'

export const authApi = {
  /** Password login */
  login: (req: LoginRequest) => api.post<User>('/auth/login', req),

  /** Logout */
  logout: () => api.post<void>('/auth/logout'),

  /** Get currently authenticated user */
  me: () => api.get<User>('/auth/me'),

  /** Create an invitation (admin only) */
  createInvitation: (req: CreateInvitationRequest) =>
    api.post<CreateInvitationResponse>('/invitations', req),

  /** Accept an invitation and register */
  acceptInvitation: (token: string, req: AcceptInvitationRequest) =>
    api.post<User>(`/invitations/${encodeURIComponent(token)}/accept`, req),

  /** Begin passkey authentication and complete the flow */
  passkeyLogin: async (): Promise<User> => {
    const options = await api.post<PublicKeyCredentialRequestOptionsJSON>(
      '/auth/passkey/authenticate/begin',
    )
    const response = await startAuthentication({ optionsJSON: options })
    return api.post<User>('/auth/passkey/authenticate/complete', response)
  },

  /** Register a new passkey for the currently authenticated user */
  registerPasskey: async (): Promise<void> => {
    const options = await api.post<PublicKeyCredentialCreationOptionsJSON>(
      '/auth/passkey/register/begin',
    )
    const response = await startRegistration({ optionsJSON: options })
    await api.post('/auth/passkey/register/complete', response)
  },
}
