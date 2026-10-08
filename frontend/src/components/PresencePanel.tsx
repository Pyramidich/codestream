import React from 'react'

interface PresencePanelProps {
  users?: string[]
}

const PresencePanel: React.FC<PresencePanelProps> = ({ users = [] }) => {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
      <h3 className="text-sm font-semibold text-gray-900 mb-2">Users online</h3>
      <p className="text-sm text-gray-600">
        Users online: {users.length}
      </p>
      {users.length > 0 && (
        <ul className="mt-2 space-y-1 max-h-32 overflow-auto">
          {users.map((userId) => (
            <li key={userId} className="text-xs text-gray-500 font-mono truncate">
              {userId}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default PresencePanel
