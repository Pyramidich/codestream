import React from 'react'
import { useParams } from 'react-router-dom'

const EditorPage: React.FC = () => {
  const { projectId, fileId } = useParams<{ projectId: string; fileId: string }>()

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-4">
        File {fileId}
      </h1>
      <p className="text-gray-600">Editor will be here</p>
      <p className="text-sm text-gray-500 mt-2">
        Project: {projectId}
      </p>
    </div>
  )
}

export default EditorPage
