import React, { useEffect, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { projectsApi } from '../api/projects'
import { filesApi } from '../api/files'
import { useAuth } from '../contexts/AuthContext'
import { detectLanguage, LANGUAGE_OPTIONS } from '../utils/language'
import type { Project as ProjectType, ProjectFile, ProjectMember } from '../types'

const ProjectPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [project, setProject] = useState<ProjectType | null>(null)
  const [files, setFiles] = useState<ProjectFile[]>([])
  const [members, setMembers] = useState<ProjectMember[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [newFile, setNewFile] = useState({
    name: '',
    path: '',
    language: '',
  })
  const [isCreating, setIsCreating] = useState(false)

  const fetchProject = async () => {
    if (!id) return
    setIsLoading(true)
    setError(null)
    try {
      const [projectResponse, filesResponse, membersResponse] = await Promise.all([
        projectsApi.getProject(id),
        filesApi.getFiles(id),
        projectsApi.getMembers(id),
      ])
      const filesData = filesResponse.data
      setProject(projectResponse.data)
      setFiles(
        Array.isArray(filesData)
          ? filesData
          : (filesData as { files?: ProjectFile[] }).files ?? [],
      )
      setMembers(membersResponse.data.members)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load project')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchProject()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const currentUserRole =
    user && members.find((m) => m.user_id === user.id)?.role

  const handleCreateFile = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id) return

    setIsCreating(true)
    try {
      const name = newFile.name.trim()
      const path = newFile.path.trim()
      const language =
        newFile.language.trim() || detectLanguage(name || path)
      await filesApi.createFile(id, name, path, language)
      setNewFile({ name: '', path: '', language: '' })
      setIsModalOpen(false)
      await fetchProject()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create file')
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div className="h-full">
      {isLoading ? (
        <p className="text-gray-600">Loading...</p>
      ) : error ? (
        <div className="p-4 bg-red-50 text-red-700 rounded-md">{error}</div>
      ) : (
        <>
          <div className="flex items-center justify-between mb-6">
            <div>
              <button
                onClick={() => navigate('/projects')}
                className="text-sm text-indigo-600 hover:text-indigo-500 mb-2"
              >
                ← Back to projects
              </button>
              <h1 className="text-3xl font-bold text-gray-900">
                {project?.name}
              </h1>
              {currentUserRole && (
                <p className="text-sm text-gray-500 mt-1 capitalize">
                  Your role: {currentUserRole}
                </p>
              )}
            </div>
            <div className="flex items-center gap-3">
              <Link
                to={`/projects/${id}/settings`}
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
              >
                Settings
              </Link>
              {currentUserRole !== 'viewer' && (
                <button
                  onClick={() => setIsModalOpen(true)}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
                >
                  Create File
                </button>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-1">
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">
                  Files
                </h2>
                {files.length === 0 ? (
                  <p className="text-gray-600 text-sm">No files yet.</p>
                ) : (
                  <ul className="space-y-1">
                    {files.map((file) => (
                      <li key={file.id}>
                        <Link
                          to={`/projects/${id}/files/${file.id}`}
                          className="block px-3 py-2 rounded-md text-sm text-gray-700 hover:bg-gray-100"
                        >
                          <span className="font-medium">{file.name}</span>
                          <span className="block text-xs text-gray-500">
                            {file.path}
                          </span>
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="mt-6 bg-white rounded-lg shadow-sm border border-gray-200 p-4">
                <h2 className="text-lg font-semibold text-gray-900 mb-2">
                  Members
                </h2>
                {members.length === 0 ? (
                  <p className="text-sm text-gray-600">No members yet.</p>
                ) : (
                  <ul className="space-y-2">
                    {members.map((member) => (
                      <li
                        key={member.user_id}
                        className="flex items-center justify-between text-sm"
                      >
                        <span className="text-gray-700 truncate">
                          {member.display_name || member.email}
                        </span>
                        <span className="text-xs text-gray-500 capitalize">
                          {member.role}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
                <Link
                  to={`/projects/${id}/settings`}
                  className="mt-3 inline-block text-sm text-indigo-600 hover:text-indigo-500"
                >
                  Manage members →
                </Link>
              </div>
            </div>

            <div className="lg:col-span-2">
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center">
                <p className="text-gray-600">
                  Select a file to open the editor
                </p>
              </div>
            </div>
          </div>
        </>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-white rounded-lg shadow-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold text-gray-900 mb-4">
              Create File
            </h2>
            <form onSubmit={handleCreateFile} className="space-y-4">
              <div>
                <label
                  htmlFor="fileName"
                  className="block text-sm font-medium text-gray-700"
                >
                  Name
                </label>
                <input
                  id="fileName"
                  type="text"
                  value={newFile.name}
                  onChange={(e) =>
                    setNewFile({ ...newFile, name: e.target.value })
                  }
                  required
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
                />
              </div>
              <div>
                <label
                  htmlFor="filePath"
                  className="block text-sm font-medium text-gray-700"
                >
                  Path
                </label>
                <input
                  id="filePath"
                  type="text"
                  value={newFile.path}
                  onChange={(e) =>
                    setNewFile({ ...newFile, path: e.target.value })
                  }
                  required
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
                />
              </div>
              <div>
                <label
                  htmlFor="fileLanguage"
                  className="block text-sm font-medium text-gray-700"
                >
                  Language
                </label>
                <select
                  id="fileLanguage"
                  value={newFile.language}
                  onChange={(e) =>
                    setNewFile({ ...newFile, language: e.target.value })
                  }
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
                >
                  {LANGUAGE_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
                >
                  {isCreating ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default ProjectPage
