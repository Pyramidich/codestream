export interface Project {
  id: string
  name: string
  owner_id: string
  created_at: string
  updated_at: string
}

export interface ProjectMember {
  user_id: string
  email: string
  display_name: string
  role: string
}

export interface ProjectFile {
  id: string
  project_id: string
  name: string
  path: string
  language: string
  content_type: string
  created_at: string
  updated_at: string
}
