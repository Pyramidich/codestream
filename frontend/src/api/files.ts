import apiClient from './client'
import type { ProjectFile } from '../types'

export interface FilesResponse {
  files: ProjectFile[]
}

export const filesApi = {
  getFiles: (projectId: string) =>
    apiClient.get<ProjectFile[] | FilesResponse>(`/projects/${projectId}/files`),

  createFile: (
    projectId: string,
    name: string,
    path: string,
    language: string,
  ) =>
    apiClient.post<ProjectFile>(`/projects/${projectId}/files`, {
      name,
      path,
      language,
    }),

  getFile: (fileId: string) => apiClient.get<ProjectFile>(`/files/${fileId}`),
}

export default filesApi
