import {
  startAuthentication,
  startRegistration,
  type PublicKeyCredentialCreationOptionsJSON,
  type PublicKeyCredentialRequestOptionsJSON,
} from '@simplewebauthn/browser'
import { api } from '@/api/client'
import type {
  AcceptInvitationRequest,
  AuthResponse,
  CreateInvitationRequest,
  CreateInvitationResponse,
  LoginRequest,
  RegisterAdminRequest,
  User,
} from './types'

export const authApi = {
  /** Check if first-run admin setup is needed */
  setupStatus: () => api.get<{ needs_setup: boolean }>('/auth/setup'),

  /** Register the first admin user (only works when no admin exists) */
  registerAdmin: (req: RegisterAdminRequest) => api.post<User>('/auth/setup', req),

  /** Password login */
  login: (req: LoginRequest) => api.post<AuthResponse>('/auth/login', req),

  /** Rotate tokens using refresh token */
  refresh: (refreshToken: string) =>
    api.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken }),

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
  passkeyLogin: async (): Promise<AuthResponse> => {
    const options = await api.post<PublicKeyCredentialRequestOptionsJSON>(
      '/auth/passkey/authenticate/begin',
    )
    const response = await startAuthentication({ optionsJSON: options })
    return api.post<AuthResponse>('/auth/passkey/authenticate/complete', response)
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
