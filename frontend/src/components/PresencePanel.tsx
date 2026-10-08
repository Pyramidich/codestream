import React from 'react'

export interface PresenceUser {
  userId: string
  displayName: string
  color: string
}

interface PresencePanelProps {
  users?: PresenceUser[]
}

const PresencePanel: React.FC<PresencePanelProps> = ({ users = [] }) => {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
      <h3 className="text-sm font-semibold text-gray-900 mb-2">Users online</h3>
      <p className="text-sm text-gray-600">Users online: {users.length}</p>
      {users.length > 0 && (
        <ul className="mt-2 space-y-1 max-h-32 overflow-auto">
          {users.map((user) => (
            <li key={user.userId} className="flex items-center gap-2 text-xs">
              <span
                className="inline-block w-3 h-3 rounded-full"
                style={{ backgroundColor: user.color }}
              />
              <span className="text-gray-700 truncate">{user.displayName}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default PresencePanel
