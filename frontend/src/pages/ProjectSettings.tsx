import React, { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { projectsApi } from '../api/projects'
import ProjectMembers from '../components/ProjectMembers'
import type { Project as ProjectType, ProjectMember } from '../types'
import { useAuth } from '../contexts/AuthContext'

const ProjectSettings: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const [project, setProject] = useState<ProjectType | null>(null)
  const [members, setMembers] = useState<ProjectMember[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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

  const currentUserRole =
    user && members.find((m) => m.user_id === user.id)?.role

  return (
    <div className="h-full max-w-4xl mx-auto space-y-6">
      {loading ? (
        <p className="text-gray-600">Loading...</p>
      ) : error ? (
        <div className="p-4 bg-red-50 text-red-700 rounded-md">{error}</div>
      ) : (
        <>
          <div className="flex items-center justify-between">
            <div>
              <Link
                to={`/projects/${id}`}
                className="text-sm text-indigo-600 hover:text-indigo-500"
              >
                ← Back to project
              </Link>
              <h1 className="text-3xl font-bold text-gray-900 mt-1">
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
