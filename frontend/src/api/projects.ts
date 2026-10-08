import apiClient from './client'
import type { Project, ProjectMember } from '../types'

export interface ProjectsResponse {
  projects: Project[]
}

export const projectsApi = {
  getProjects: () =>
    apiClient.get<Project[] | ProjectsResponse>('/projects'),

  createProject: (name: string) =>
    apiClient.post<Project>('/projects', { name }),

  getProject: (id: string) => apiClient.get<Project>(`/projects/${id}`),

  addMember: (projectId: string, userId: string, role: string) =>
    apiClient.post<ProjectMember>(`/projects/${projectId}/members`, {
      user_id: userId,
      role,
    }),
}

export default projectsApi
