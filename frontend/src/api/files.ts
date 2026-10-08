import apiClient from './client'
import type { ProjectFile } from '../types'

export interface FilesResponse {
  files: ProjectFile[]
}

export interface FileContentResponse {
  content: string
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

  getFileContent: (fileId: string) =>
    apiClient
      .get<FileContentResponse>(`/files/${fileId}/content`)
      .then((response) => response.data.content),

  saveFile: (fileId: string, content: string, language?: string) =>
    apiClient.patch<ProjectFile>(`/files/${fileId}`, {
      content_text: content,
      ...(language !== undefined && { language }),
    }),

  updateFile: (
    fileId: string,
    updates: Partial<{
      name: string
      path: string
      language: string
      content_text: string
    }>,
  ) => apiClient.patch<ProjectFile>(`/files/${fileId}`, updates),
}

export default filesApi
