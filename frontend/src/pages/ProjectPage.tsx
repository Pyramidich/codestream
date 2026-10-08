import React, { useEffect, useRef, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import gsap from 'gsap'
import { projectsApi } from '../api/projects'
import { filesApi } from '../api/files'
import { useAuth } from '../contexts/AuthContext'
import { detectLanguage, LANGUAGE_OPTIONS } from '../utils/language'
import type { Project as ProjectType, ProjectFile, ProjectMember } from '../types'
import { prefersReducedMotion } from '../utils/animations'

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
  const containerRef = useRef<HTMLDivElement>(null)
  const modalRef = useRef<HTMLDivElement>(null)

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

  useEffect(() => {
    if (isLoading) return
    if (!containerRef.current) return
    if (prefersReducedMotion()) {
      gsap.set(containerRef.current, { opacity: 1, y: 0 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        containerRef.current,
        { y: 20, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.6, ease: 'power2.out' }
      )
    })
    return () => ctx.revert()
  }, [isLoading])

  useEffect(() => {
    if (!isModalOpen || !modalRef.current) return
    if (prefersReducedMotion()) {
      gsap.set(modalRef.current, { opacity: 1, scale: 1 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        modalRef.current,
        { scale: 0.95, opacity: 0 },
        { scale: 1, opacity: 1, duration: 0.3, ease: 'power2.out' }
      )
    })
    return () => ctx.revert()
  }, [isModalOpen])

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

  const handleRowEnter = (e: React.MouseEvent<HTMLAnchorElement>) => {
    gsap.to(e.currentTarget, {
      backgroundColor: '#333333',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleRowLeave = (e: React.MouseEvent<HTMLAnchorElement>) => {
    gsap.to(e.currentTarget, {
      backgroundColor: 'transparent',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  return (
    <div ref={containerRef} className="h-full">
      {isLoading ? (
        <p className="text-[#A0A0A0]">Loading...</p>
      ) : error ? (
        <div className="p-4 bg-red-900/20 text-red-400 rounded-md">{error}</div>
      ) : (
        <>
          <div className="flex items-center justify-between mb-6">
            <div>
              <button
                onClick={() => navigate('/projects')}
                className="text-sm text-[#624EC2] hover:text-[#7B68D1] mb-2"
              >
                ← Back to projects
              </button>
              <h1 className="text-3xl font-bold text-[#F5F0EA]">
                {project?.name}
              </h1>
              {currentUserRole && (
                <p className="text-sm text-[#A0A0A0] mt-1 capitalize">
                  Your role: {currentUserRole}
                </p>
              )}
            </div>
            <div className="flex items-center gap-3">
              <Link
                to={`/projects/${id}/settings`}
                className="inline-flex items-center px-4 py-2 border border-[#3A3A3A] text-sm font-medium rounded-md text-[#F5F0EA] bg-transparent hover:bg-[#2A2A2A]"
              >
                Settings
              </Link>
              {currentUserRole !== 'viewer' && (
                <button
                  onClick={() => setIsModalOpen(true)}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-[#624EC2] hover:bg-[#7B68D1]"
                >
                  Create File
                </button>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-1">
              <div className="bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A] p-4">
                <h2 className="text-lg font-semibold text-[#F5F0EA] mb-4">
                  Files
                </h2>
                {files.length === 0 ? (
                  <p className="text-[#A0A0A0] text-sm">No files yet.</p>
                ) : (
                  <ul className="space-y-1">
                    {files.map((file) => (
                      <li key={file.id}>
                        <Link
                          to={`/projects/${id}/files/${file.id}`}
                          onMouseEnter={handleRowEnter}
                          onMouseLeave={handleRowLeave}
                          className="block px-3 py-2 rounded-md text-sm text-[#F5F0EA]"
                        >
                          <span className="font-medium">{file.name}</span>
                          <span className="block text-xs text-[#A0A0A0]">
                            {file.path}
                          </span>
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="mt-6 bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A] p-4">
                <h2 className="text-lg font-semibold text-[#F5F0EA] mb-2">
                  Members
                </h2>
                {members.length === 0 ? (
                  <p className="text-sm text-[#A0A0A0]">No members yet.</p>
                ) : (
                  <ul className="space-y-2">
                    {members.map((member) => (
                      <li
                        key={member.user_id}
                        className="flex items-center justify-between text-sm"
                      >
                        <span className="text-[#F5F0EA] truncate">
                          {member.display_name || member.email}
                        </span>
                        <span className="text-xs text-[#A0A0A0] capitalize">
                          {member.role}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
                <Link
                  to={`/projects/${id}/settings`}
                  className="mt-3 inline-block text-sm text-[#624EC2] hover:text-[#7B68D1]"
                >
                  Manage members →
                </Link>
              </div>
            </div>

            <div className="lg:col-span-2">
              <div className="bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A] p-8 text-center">
                <p className="text-[#A0A0A0]">
                  Select a file to open the editor
                </p>
              </div>
            </div>
          </div>
        </>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div
            ref={modalRef}
            className="bg-[#2A2A2A] rounded-lg shadow-lg p-6 w-full max-w-md border border-[#3A3A3A]"
          >
            <h2 className="text-xl font-bold text-[#F5F0EA] mb-4">
              Create File
            </h2>
            <form onSubmit={handleCreateFile} className="space-y-4">
              <div>
                <label
                  htmlFor="fileName"
                  className="block text-sm font-medium text-[#A0A0A0]"
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
                  className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
                />
              </div>
              <div>
                <label
                  htmlFor="filePath"
                  className="block text-sm font-medium text-[#A0A0A0]"
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
                  className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
                />
              </div>
              <div>
                <label
                  htmlFor="fileLanguage"
                  className="block text-sm font-medium text-[#A0A0A0]"
                >
                  Language
                </label>
                <select
                  id="fileLanguage"
                  value={newFile.language}
                  onChange={(e) =>
                    setNewFile({ ...newFile, language: e.target.value })
                  }
                  className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
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
                  className="px-4 py-2 border border-[#3A3A3A] rounded-md text-sm font-medium text-[#F5F0EA] bg-transparent hover:bg-[#202020]"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-[#624EC2] hover:bg-[#7B68D1] disabled:opacity-50"
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
