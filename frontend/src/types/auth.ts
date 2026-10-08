export interface User {
  id: string
  email: string
  display_name?: string
}

export interface Tokens {
  access_token: string
  refresh_token: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  email: string
  password: string
  display_name?: string
}
