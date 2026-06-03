export interface User {
  id: string
  username: string
  email: string
  role: 'admin' | 'teacher' | 'student'
  created_at: string
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface CreateInvitationRequest {
  email: string
  role: 'admin' | 'teacher' | 'student'
}

export interface CreateInvitationResponse {
  id: string
  email: string
  role: string
  expires_at: string
}

export interface AcceptInvitationRequest {
  username: string
  password: string
}

export interface RegisterAdminRequest {
  username: string
  email: string
  password: string
}
