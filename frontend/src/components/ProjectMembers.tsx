import React, { useEffect, useState } from 'react'
import { projectsApi } from '../api/projects'
import type { ProjectMember } from '../types'

interface ProjectMembersProps {
  projectId: string
  currentUserRole?: string
}

const ProjectMembers: React.FC<ProjectMembersProps> = ({
  projectId,
  currentUserRole,
}) => {
  const [members, setMembers] = useState<ProjectMember[]>([])
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('editor')
  const [loading, setLoading] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isOwner = currentUserRole === 'owner'

  const fetchMembers = async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await projectsApi.getMembers(projectId)
      setMembers(response.data.members)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load members')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchMembers()
  }, [projectId])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim()) return

    setActionLoading(true)
    setError(null)
    try {
      await projectsApi.addMember(projectId, email.trim(), role)
      setEmail('')
      setRole('editor')
      await fetchMembers()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add member')
    } finally {
      setActionLoading(false)
    }
  }

  const handleRemove = async (userId: string) => {
    setActionLoading(true)
    setError(null)
    try {
      await projectsApi.removeMember(projectId, userId)
      await fetchMembers()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove member')
    } finally {
      setActionLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {isOwner && (
        <form
          onSubmit={handleAdd}
          className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 space-y-4"
        >
          <h3 className="text-lg font-semibold text-gray-900">Add Member</h3>
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="member@example.com"
              required
              className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
            />
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm px-3 py-2 border"
            >
              <option value="editor">Editor</option>
              <option value="viewer">Viewer</option>
            </select>
            <button
              type="submit"
              disabled={actionLoading}
              className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
            >
              Add
            </button>
          </div>
        </form>
      )}

      {error && (
        <div className="p-3 bg-red-50 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Members</h3>
        </div>
        {loading ? (
          <p className="p-4 text-sm text-gray-600">Loading...</p>
        ) : members.length === 0 ? (
          <p className="p-4 text-sm text-gray-600">No members yet.</p>
        ) : (
          <ul className="divide-y divide-gray-200">
            {members.map((member) => (
              <li
                key={member.user_id}
                className="px-4 py-3 flex items-center justify-between"
              >
                <div>
                <div className="flex items-center gap-2">
                    <p className="text-sm font-medium text-gray-900">
                      {member.display_name || member.email}
                    </p>
                    <span
                      className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium capitalize"
                      style={{
                        backgroundColor:
                          member.role === 'owner'
                            ? '#f3e8ff'
                            : member.role === 'editor'
                            ? '#dbeafe'
                            : '#f3f4f6',
                        color:
                          member.role === 'owner'
                            ? '#6b21a8'
                            : member.role === 'editor'
                            ? '#1e40af'
                            : '#374151',
                      }}
                    >
                      {member.role}
                    </span>
                  </div>
                  <p className="text-xs text-gray-500">{member.email}</p>
                </div>
                {isOwner && (
                  <button
                    onClick={() => handleRemove(member.user_id)}
                    disabled={actionLoading}
                    className="text-sm text-red-600 hover:text-red-800 disabled:opacity-50"
                  >
                    Remove
                  </button>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}

export default ProjectMembers
