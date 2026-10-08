import React, { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import gsap from 'gsap'
import Editor from '../components/Editor'
import { filesApi } from '../api/files'
import type { ProjectFile } from '../types'
import { prefersReducedMotion } from '../utils/animations'

const EditorPage: React.FC = () => {
  const { projectId, fileId } = useParams<{ projectId: string; fileId: string }>()
  const [file, setFile] = useState<ProjectFile | null>(null)
  const [error, setError] = useState<string | null>(null)
  const headerRef = useRef<HTMLDivElement>(null)
  const editorRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!fileId) return

    filesApi
      .getFile(fileId)
      .then((response) => setFile(response.data))
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'Failed to load file')
      })
  }, [fileId])

  useEffect(() => {
    if (!headerRef.current || !editorRef.current) return
    if (prefersReducedMotion()) {
      gsap.set([headerRef.current, editorRef.current], { opacity: 1, y: 0, x: 0 })
      return
    }

    const ctx = gsap.context(() => {
      const tl = gsap.timeline({ defaults: { ease: 'power2.out' } })
      tl.fromTo(
        headerRef.current,
        { y: -20, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.5 }
      ).fromTo(
        editorRef.current,
        { y: 20, opacity: 0 },
        { y: 0, opacity: 1, duration: 0.6 },
        '-=0.3'
      )
    })
    return () => ctx.revert()
  }, [file])

  if (!projectId || !fileId) {
    return <div className="text-red-400">Invalid project or file</div>
  }

  return (
    <div className="h-full flex flex-col">
      <div
        ref={headerRef}
        className="flex items-center justify-between mb-4"
      >
        <div>
          <h1 className="text-2xl font-bold text-[#F5F0EA]">
            {file?.name ?? `File ${fileId}`}
          </h1>
          <p className="text-sm text-[#A0A0A0]">
            Project: {projectId} {file?.path ? `• ${file.path}` : ''}
          </p>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-900/20 text-red-400 rounded-md text-sm">
          {error}
        </div>
      )}

      <div ref={editorRef} className="flex-1 min-h-0">
        <Editor fileId={fileId} projectId={projectId} language={file?.language} />
      </div>
    </div>
  )
}

export default EditorPage
