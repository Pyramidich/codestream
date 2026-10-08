import apiClient from './client'
import type { Project, ProjectMember } from '../types'

export interface ProjectsResponse {
  projects: Project[]
}

export interface MembersResponse {
  members: ProjectMember[]
}

export const projectsApi = {
  getProjects: () =>
    apiClient.get<Project[] | ProjectsResponse>('/projects'),

  createProject: (name: string) =>
    apiClient.post<Project>('/projects', { name }),

  getProject: (id: string) => apiClient.get<Project>(`/projects/${id}`),

  getMembers: (projectId: string) =>
    apiClient.get<MembersResponse>(`/projects/${projectId}/members`),

  addMember: (projectId: string, email: string, role: string) =>
    apiClient.post<ProjectMember>(`/projects/${projectId}/members`, {
      email,
      role,
    }),

  removeMember: (projectId: string, userId: string) =>
    apiClient.delete(`/projects/${projectId}/members/${userId}`),
}

export default projectsApi
