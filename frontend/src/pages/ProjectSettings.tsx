import React, { useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import gsap from 'gsap'
import { projectsApi } from '../api/projects'
import ProjectMembers from '../components/ProjectMembers'
import type { Project as ProjectType, ProjectMember } from '../types'
import { useAuth } from '../contexts/AuthContext'
import { prefersReducedMotion } from '../utils/animations'

const ProjectSettings: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const [project, setProject] = useState<ProjectType | null>(null)
  const [members, setMembers] = useState<ProjectMember[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!id) return

    const fetchData = async () => {
      setLoading(true)
      setError(null)
      try {
        const [projectResponse, membersResponse] = await Promise.all([
          projectsApi.getProject(id),
          projectsApi.getMembers(id),
        ])
        setProject(projectResponse.data)
        setMembers(membersResponse.data.members)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load project')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [id])

  useEffect(() => {
    if (loading) return
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
  }, [loading])

  const currentUserRole =
    user && members.find((m) => m.user_id === user.id)?.role

  return (
    <div ref={containerRef} className="h-full max-w-4xl mx-auto space-y-6">
      {loading ? (
        <p className="text-[#A0A0A0]">Loading...</p>
      ) : error ? (
        <div className="p-4 bg-red-900/20 text-red-400 rounded-md">{error}</div>
      ) : (
        <>
          <div className="flex items-center justify-between">
            <div>
              <Link
                to={`/projects/${id}`}
                className="text-sm text-[#624EC2] hover:text-[#7B68D1]"
              >
                ← Back to project
              </Link>
              <h1 className="text-3xl font-bold text-[#F5F0EA] mt-1">
                {project?.name} Settings
              </h1>
            </div>
          </div>

          <ProjectMembers
            projectId={id!}
            currentUserRole={currentUserRole ?? undefined}
          />
        </>
      )}
    </div>
  )
}

export default ProjectSettings
