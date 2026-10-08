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

interface EditorProps {
  fileId: string
  projectId: string
}

const Editor: React.FC<EditorProps> = ({ fileId }) => {
  const yDocRef = useRef(new Y.Doc())
  const providerRef = useRef<CodestreamProvider | null>(null)
  const bindingRef = useRef<MonacoBinding | null>(null)
  const editorRef = useRef<monacoEditor.IStandaloneCodeEditor | null>(null)
  const [onlineUsers, setOnlineUsers] = useState<string[]>([])

  useEffect(() => {
    const token = getAccessToken()
    if (!token) {
      window.location.href = '/login'
      return
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

    return () => {
      // Delay disconnect to survive React StrictMode's immediate remount.
      const p = provider
      setTimeout(() => {
        if (providerRef.current === p) {
          yDocRef.current.off('update', handleUpdate)
          bindingRef.current?.destroy()
          bindingRef.current = null
          p.sendLeave()
          p.disconnect()
          providerRef.current = null
        }
      }, 100)
    }
  }, [fileId])

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
  }

  return (
    <div>
      <div className="mb-2 flex justify-end">
        <PresencePanel users={onlineUsers} />
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
