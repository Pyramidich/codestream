import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import Editor from '../components/Editor'
import { filesApi } from '../api/files'
import type { ProjectFile } from '../types'

const EditorPage: React.FC = () => {
  const { projectId, fileId } = useParams<{ projectId: string; fileId: string }>()
  const [file, setFile] = useState<ProjectFile | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!fileId) return

    filesApi
      .getFile(fileId)
      .then((response) => setFile(response.data))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Failed to load file')
      })
  }, [fileId])

  if (!projectId || !fileId) {
    return <div className="text-red-600">Invalid project or file</div>
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            {file?.name ?? `File ${fileId}`}
          </h1>
          <p className="text-sm text-gray-500">
            Project: {projectId} {file?.path ? `• ${file.path}` : ''}
          </p>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      <div className="flex-1 min-h-0">
        <Editor fileId={fileId} projectId={projectId} language={file?.language} />
      </div>
    </div>
  )
}

export default EditorPage
