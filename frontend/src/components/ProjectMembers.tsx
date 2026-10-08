import React, { useEffect, useRef, useState } from 'react'
import gsap from 'gsap'
import { projectsApi } from '../api/projects'
import type { ProjectMember } from '../types'
import { prefersReducedMotion } from '../utils/animations'

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
  const listRef = useRef<HTMLUListElement>(null)
  const formRef = useRef<HTMLFormElement>(null)

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

  useEffect(() => {
    if (loading) return
    if (prefersReducedMotion()) {
      if (formRef.current) gsap.set(formRef.current, { opacity: 1, y: 0 })
      if (listRef.current) gsap.set(listRef.current.children, { opacity: 1, y: 0 })
      return
    }

    const ctx = gsap.context(() => {
      if (formRef.current) {
        gsap.fromTo(
          formRef.current,
          { y: 20, opacity: 0 },
          { y: 0, opacity: 1, duration: 0.5, ease: 'power2.out' }
        )
      }
      if (listRef.current) {
        const items = Array.from(listRef.current.children)
        gsap.fromTo(
          items,
          { y: 20, opacity: 0 },
          { y: 0, opacity: 1, duration: 0.5, stagger: 0.08, ease: 'power2.out' }
        )
      }
    })
    return () => ctx.revert()
  }, [loading, members])

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

  const handleRowEnter = (e: React.MouseEvent<HTMLLIElement>) => {
    gsap.to(e.currentTarget, {
      backgroundColor: '#333333',
      x: 2,
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  const handleRowLeave = (e: React.MouseEvent<HTMLLIElement>) => {
    gsap.to(e.currentTarget, {
      backgroundColor: 'transparent',
      x: 0,
      duration: 0.2,
      ease: 'power2.out',
    })
  }

  return (
    <div className="space-y-6">
      {isOwner && (
        <form
          ref={formRef}
          onSubmit={handleAdd}
          className="bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A] p-4 space-y-4"
        >
          <h3 className="text-lg font-semibold text-[#F5F0EA]">Add Member</h3>
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="member@example.com"
              required
              className="flex-1 rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
            />
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="rounded-md border-[#3A3A3A] bg-[#202020] text-[#F5F0EA] shadow-sm focus:border-[#624EC2] focus:ring-[#624EC2] sm:text-sm px-3 py-2 border"
            >
              <option value="editor">Editor</option>
              <option value="viewer">Viewer</option>
            </select>
            <button
              type="submit"
              disabled={actionLoading}
              className="px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-[#624EC2] hover:bg-[#7B68D1] disabled:opacity-50"
            >
              Add
            </button>
          </div>
        </form>
      )}

      {error && (
        <div className="p-3 bg-red-900/20 text-red-400 rounded-md text-sm">
          {error}
        </div>
      )}

      <div className="bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A] overflow-hidden">
        <div className="px-4 py-3 border-b border-[#3A3A3A]">
          <h3 className="text-lg font-semibold text-[#F5F0EA]">Members</h3>
        </div>
        {loading ? (
          <p className="p-4 text-sm text-[#A0A0A0]">Loading...</p>
        ) : members.length === 0 ? (
          <p className="p-4 text-sm text-[#A0A0A0]">No members yet.</p>
        ) : (
          <ul ref={listRef} className="divide-y divide-[#3A3A3A]">
            {members.map((member) => (
              <li
                key={member.user_id}
                onMouseEnter={handleRowEnter}
                onMouseLeave={handleRowLeave}
                className="px-4 py-3 flex items-center justify-between"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <p className="text-sm font-medium text-[#F5F0EA]">
                      {member.display_name || member.email}
                    </p>
                    <span
                      className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium capitalize"
                      style={{
                        backgroundColor:
                          member.role === 'owner'
                            ? '#624EC2'
                            : member.role === 'editor'
                            ? '#3A3A3A'
                            : '#202020',
                        color:
                          member.role === 'owner'
                            ? '#F5F0EA'
                            : member.role === 'editor'
                            ? '#F5F0EA'
                            : '#A0A0A0',
                      }}
                    >
                      {member.role}
                    </span>
                  </div>
                  <p className="text-xs text-[#A0A0A0]">{member.email}</p>
                </div>
                {isOwner && (
                  <button
                    onClick={() => handleRemove(member.user_id)}
                    disabled={actionLoading}
                    className="text-sm text-red-400 hover:text-red-300 disabled:opacity-50"
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
