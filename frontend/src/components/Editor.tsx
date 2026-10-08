import React, { useEffect, useRef, useState } from 'react'
import MonacoEditor from '@monaco-editor/react'
import * as Y from 'yjs'
import { MonacoBinding } from 'y-monaco'
import type { editor as monacoEditor } from 'monaco-editor'
import type { Awareness } from 'y-protocols/awareness'
import { CodestreamProvider, base64ToArrayBuffer } from '../providers/websocket'
import type { WebSocketMessage } from '../providers/websocket'
import PresencePanel from './PresencePanel'
import { getAccessToken } from '../api/client'
import { filesApi } from '../api/files'

interface EditorProps {
  fileId: string
  projectId: string
}

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error'

const Editor: React.FC<EditorProps> = ({ fileId }) => {
  const yDocRef = useRef(new Y.Doc())
  const providerRef = useRef<CodestreamProvider | null>(null)
  const bindingRef = useRef<MonacoBinding | null>(null)
  const editorRef = useRef<monacoEditor.IStandaloneCodeEditor | null>(null)
  const [onlineUsers, setOnlineUsers] = useState<string[]>([])
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle')
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const token = getAccessToken()
    if (!token) {
      window.location.href = '/login'
      return
    }

    const setup = async () => {
      let initialContent = ''
      try {
        initialContent = await filesApi.getFileContent(fileId)
      } catch {
        initialContent = ''
      }

      const yText = yDocRef.current.getText('monaco')
      yText.delete(0, yText.length)
      if (initialContent) {
        yText.insert(0, initialContent)
      }

      const url = `ws://localhost:8080/ws?token=${encodeURIComponent(token)}&file_id=${encodeURIComponent(fileId)}`
      const provider = new CodestreamProvider(url, fileId)
      providerRef.current = provider

      provider.onMessage((message: WebSocketMessage) => {
        if (message.event === 'doc:sync') {
          const payload = message.data as { state?: string; stateVector?: string }
          if (payload.state) {
            try {
              const update = base64ToArrayBuffer(payload.state)
              Y.applyUpdate(yDocRef.current, update)
            } catch {
              // ignore invalid sync
            }
          }
        } else if (message.event === 'doc:update') {
          const payload = message.data as { update?: string }
          if (payload.update) {
            try {
              const update = base64ToArrayBuffer(payload.update)
              Y.applyUpdate(yDocRef.current, update)
            } catch {
              // ignore invalid update
            }
          }
        } else if (message.event === 'user:joined') {
          const payload = message.data as { userId?: string }
          if (payload.userId) {
            setOnlineUsers((prev) =>
              prev.includes(payload.userId as string)
                ? prev
                : [...prev, payload.userId as string],
            )
          }
        } else if (message.event === 'user:left') {
          const payload = message.data as { userId?: string }
          if (payload.userId) {
            setOnlineUsers((prev) =>
              prev.filter((id) => id !== payload.userId),
            )
          }
        } else if (message.event === 'presence:list') {
          const payload = message.data as { users?: string[] }
          if (Array.isArray(payload.users)) {
            setOnlineUsers(payload.users)
          }
        }
      })

      const handleUpdate = () => {
        const update = Y.encodeStateAsUpdate(yDocRef.current)
        provider.sendUpdate(update)
      }

      yDocRef.current.on('update', handleUpdate)

      provider.connect()

      const stateVector = Y.encodeStateVector(yDocRef.current)
      provider.sendJoin(stateVector)

      setIsLoading(false)

      return () => {
        yDocRef.current.off('update', handleUpdate)
        bindingRef.current?.destroy()
        bindingRef.current = null
        provider.sendLeave()
        provider.disconnect()
        providerRef.current = null
        yDocRef.current = new Y.Doc()
      }
    }

    let cleanupFn: (() => void) | undefined
    setup().then((cleanup) => {
      cleanupFn = cleanup
    })

    return () => {
      if (cleanupFn) {
        cleanupFn()
      }
    }
  }, [fileId])

  const saveContent = async () => {
    const editor = editorRef.current
    if (!editor) return

    const content = editor.getValue()
    setSaveStatus('saving')
    try {
      await filesApi.saveFile(fileId, content)
      setSaveStatus('saved')
      setTimeout(() => setSaveStatus('idle'), 2000)
    } catch (err) {
      setSaveStatus('error')
      setTimeout(() => setSaveStatus('idle'), 3000)
    }
  }

  const handleEditorMount = (editor: monacoEditor.IStandaloneCodeEditor) => {
    editorRef.current = editor

    const yText = yDocRef.current.getText('monaco')
    const model = editor.getModel()
    if (!model) return

    bindingRef.current = new MonacoBinding(
      yText,
      model,
      new Set([editor]),
      undefined as Awareness | undefined,
    )

    editor.onKeyDown((e) => {
      if ((e.ctrlKey || e.metaKey) && e.code === 'KeyS') {
        e.preventDefault()
        e.stopPropagation()
        void saveContent()
      }
    })
  }

  const statusText = {
    idle: '',
    saving: 'Saving...',
    saved: 'Saved',
    error: 'Save error',
  }

  const statusClass = {
    idle: 'text-gray-400',
    saving: 'text-yellow-600',
    saved: 'text-green-600',
    error: 'text-red-600',
  }

  if (isLoading) {
    return (
      <div className="h-[600px] flex items-center justify-center border border-gray-300 rounded-md">
        <p className="text-gray-600">Loading...</p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <PresencePanel users={onlineUsers} />
        <span className={`text-sm font-medium ${statusClass[saveStatus]}`}>
          {statusText[saveStatus]}
        </span>
      </div>
      <div className="h-[600px] border border-gray-300 rounded-md overflow-hidden">
        <MonacoEditor
          defaultLanguage="javascript"
          theme="vs-light"
          onMount={handleEditorMount}
          options={{
            automaticLayout: true,
            minimap: { enabled: false },
          }}
        />
      </div>
    </div>
  )
}

export default Editor
