import apiClient from './client'
import type { AuthResponse, LoginCredentials, RegisterCredentials } from '../types/auth'

export const authApi = {
  register: (credentials: RegisterCredentials) =>
    apiClient.post('/auth/register', credentials),

  login: (credentials: LoginCredentials) =>
    apiClient.post<AuthResponse>('/auth/login', credentials),

  logout: (refreshToken: string) =>
    apiClient.post('/auth/logout', { refresh_token: refreshToken }),

  refresh: (refreshToken: string) =>
    apiClient.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken }),

  me: () => apiClient.get('/me'),
}

export default authApi
