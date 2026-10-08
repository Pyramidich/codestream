import React, { useEffect, useRef } from 'react'
import gsap from 'gsap'
import { prefersReducedMotion } from '../utils/animations'

export interface PresenceUser {
  userId: string
  displayName: string
  color: string
}

interface PresencePanelProps {
  users?: PresenceUser[]
}

const PresencePanel: React.FC<PresencePanelProps> = ({ users = [] }) => {
  const panelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!panelRef.current) return
    if (prefersReducedMotion()) {
      gsap.set(panelRef.current, { opacity: 1, x: 0 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        panelRef.current,
        { x: -20, opacity: 0 },
        { x: 0, opacity: 1, duration: 0.5, ease: 'power2.out' }
      )
    })
    return () => ctx.revert()
  }, [])

  return (
    <div
      ref={panelRef}
      className="bg-[#2A2A2A] rounded-lg shadow-sm border border-[#3A3A3A] p-4"
    >
      <h3 className="text-sm font-semibold text-[#F5F0EA] mb-2">Users online</h3>
      <p className="text-sm text-[#A0A0A0]">Users online: {users.length}</p>
      {users.length > 0 && (
        <ul className="mt-2 space-y-1 max-h-32 overflow-auto">
          {users.map((user) => (
            <li key={user.userId} className="flex items-center gap-2 text-xs">
              <span
                className="inline-block w-3 h-3 rounded-full"
                style={{ backgroundColor: user.color }}
              />
              <span className="text-[#F5F0EA] truncate">{user.displayName}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default PresencePanel
