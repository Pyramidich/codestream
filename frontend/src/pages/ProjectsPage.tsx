import React, { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import gsap from 'gsap'
import { projectsApi } from '../api/projects'
import type { Project } from '../types'
import { prefersReducedMotion } from '../utils/animations'

const ProjectsPage: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [newProjectName, setNewProjectName] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const cardRefs = useRef<(HTMLAnchorElement | null)[]>([])
  const buttonRef = useRef<HTMLButtonElement>(null)
  const modalRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLUListElement>(null)

  const fetchProjects = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await projectsApi.getProjects()
      const data = response.data
      const projects = Array.isArray(data) ? data : data.projects ?? []
      setProjects(projects)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchProjects()
  }, [])

  useEffect(() => {
    if (prefersReducedMotion()) return
    if (projects.length === 0 || isLoading) return

    const ctx = gsap.context(() => {
      gsap.fromTo(
        cardRefs.current.filter(Boolean),
        { y: 30, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.5, stagger: 0.1, ease: 'power2.out' }
      )
    })
    return () => ctx.revert()
  }, [projects, isLoading])

  useEffect(() => {
    if (!isModalOpen || !modalRef.current) return
    const ctx = gsap.context(() => {
      gsap.fromTo(
        modalRef.current,
        { scale: 0.95, opacity: 0 },
        { scale: 1, opacity: 1, duration: 0.3, ease: 'power2.out' }
      )
    })
    return () => ctx.revert()
  }, [isModalOpen])

  const handleCardEnter = (e: React.MouseEvent<HTMLAnchorElement>) => {
    gsap.to(e.currentTarget, {
      y: -4,
      borderColor: '#624EC2',
      boxShadow: '0 10px 25px rgba(0, 0, 0, 0.3)',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleCardLeave = (e: React.MouseEvent<HTMLAnchorElement>) => {
    gsap.to(e.currentTarget, {
      y: 0,
      borderColor: '#3A3A3A',
      boxShadow: '0 0 0 rgba(0, 0, 0, 0)',
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleButtonEnter = () => {
    if (buttonRef.current) {
      gsap.to(buttonRef.current, {
        scale: 1.05,
        boxShadow: '0 0 15px rgba(98, 78, 194, 0.5)',
        duration: 0.3,
        ease: 'power2.out',
        yoyo: true,
      })
    }
  }

  const handleButtonLeave = () => {
    if (buttonRef.current) {
      gsap.to(buttonRef.current, {
        scale: 1,
        boxShadow: '0 0 0 rgba(98, 78, 194, 0)',
        duration: 0.3,
        ease: 'power2.out',
      })
    }
  }

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newProjectName.trim()) return

    setIsCreating(true)
    try {
      await projectsApi.createProject(newProjectName.trim())
      setNewProjectName('')
      setIsModalOpen(false)
      await fetchProjects()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project')
    } finally {
      setIsCreating(false)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold text-[#F5F0EA]">Projects</h1>
        <button
          ref={buttonRef}
          onClick={() => setIsModalOpen(true)}
          onMouseEnter={handleButtonEnter}
          onMouseLeave={handleButtonLeave}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-[#624EC2] hover:bg-[#7B68D1]"
        >
          Create Project
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-900/20 text-red-400 rounded-md text-sm">
          {error}
        </div>
      )}

      {isLoading ? (
        <p className="text-[#A0A0A0]">Loading...</p>
      ) : projects.length === 0 ? (
        <p className="text-[#A0A0A0]">No projects yet.</p>
      ) : (
        <ul ref={listRef} className="space-y-3">
          {projects.map((project, index) => (
            <li key={project.id}>
              <Link
                ref={(el) => { cardRefs.current[index] = el }}
                to={`/projects/${project.id}`}
                onMouseEnter={handleCardEnter}
                onMouseLeave={handleCardLeave}
                className="block p-4 bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A]"
              >
                <h2 className="text-lg font-semibold text-[#F5F0EA]">
                  {project.name}
                </h2>
                <p className="text-sm text-[#A0A0A0]">
                  Created: {new Date(project.created_at).toLocaleDateString()}
                </p>
              </Link>
            </li>
          ))}
        </ul>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div
            ref={modalRef}
            className="bg-[#2A2A2A] rounded-lg shadow-lg p-6 w-full max-w-md border border-[#3A3A3A]"
          >
            <h2 className="text-xl font-bold text-[#F5F0EA] mb-4">
              Create Project
            </h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label
                  htmlFor="projectName"
                  className="block text-sm font-medium text-[#A0A0A0]"
                >
                  Project name
                </label>
                <input
                  id="projectName"
                  type="text"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  required
                  className="mt-1 block w-full rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
                />
              </div>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 border border-[#3A3A3A] rounded-md text-sm font-medium text-[#F5F0EA] bg-transparent hover:bg-[#2A2A2A]"
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

export default ProjectsPage
