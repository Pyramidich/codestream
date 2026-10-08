import React, { useEffect, useRef, useState } from 'react'
import MonacoEditor from '@monaco-editor/react'
import * as Y from 'yjs'
import { MonacoBinding } from 'y-monaco'
import type { editor as monacoEditor } from 'monaco-editor'
import gsap from 'gsap'
import { CodestreamProvider, base64ToArrayBuffer } from '../providers/websocket'
import type { WebSocketMessage } from '../providers/websocket'
import PresencePanel, { type PresenceUser } from './PresencePanel'
import { getAccessToken } from '../api/client'
import { filesApi } from '../api/files'
import { useAuth } from '../contexts/AuthContext'
import { getUserColor } from '../utils/colors'
import { prefersReducedMotion } from '../utils/animations'

interface EditorProps {
  fileId: string
  projectId: string
  language?: string
}

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error'

interface AwarenessUserState {
  user?: {
    id?: string
    name?: string
    color?: string
  }
}

const Editor: React.FC<EditorProps> = ({ fileId, language = 'plaintext' }) => {
  const { user } = useAuth()
  const yDocRef = useRef(new Y.Doc())
  const providerRef = useRef<CodestreamProvider | null>(null)
  const bindingRef = useRef<MonacoBinding | null>(null)
  const editorRef = useRef<monacoEditor.IStandaloneCodeEditor | null>(null)
  const [onlineUsers, setOnlineUsers] = useState<PresenceUser[]>([])
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle')
  const [isLoading, setIsLoading] = useState(true)
  const statusRef = useRef<HTMLSpanElement>(null)
  const loadingRef = useRef<HTMLDivElement>(null)

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
      const provider = new CodestreamProvider(url, fileId, yDocRef.current)
      providerRef.current = provider

      const awareness = provider.getAwareness()

      const updatePresenceFromAwareness = () => {
        const states = Array.from(awareness.getStates().values()) as AwarenessUserState[]
        const users: PresenceUser[] = []
        for (const state of states) {
          const u = state.user
          if (u?.id && u.name && u.color) {
            users.push({ userId: u.id, displayName: u.name, color: u.color })
          }
        }
        setOnlineUsers(users)
      }

      awareness.on('change', updatePresenceFromAwareness)

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
          const joinedUserId = payload.userId
          if (joinedUserId) {
            setOnlineUsers((prev) =>
              prev.some((u) => u.userId === joinedUserId)
                ? prev
                : [...prev, { userId: joinedUserId, displayName: joinedUserId, color: getUserColor(joinedUserId) }],
            )
          }
        } else if (message.event === 'user:left') {
          const payload = message.data as { userId?: string }
          const leftUserId = payload.userId
          if (leftUserId) {
            setOnlineUsers((prev) => prev.filter((u) => u.userId !== leftUserId))
          }
        } else if (message.event === 'presence:list') {
          const payload = message.data as { users?: string[] }
          if (Array.isArray(payload.users)) {
            setOnlineUsers(
              payload.users.map((userId) => ({
                userId,
                displayName: userId,
                color: getUserColor(userId),
              })),
            )
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
        awareness.off('change', updatePresenceFromAwareness)
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

  useEffect(() => {
    if (!loadingRef.current) return
    if (prefersReducedMotion()) {
      gsap.set(loadingRef.current, { opacity: 1 })
      return
    }
    const tween = gsap.to(loadingRef.current, {
      opacity: 0.5,
      duration: 0.8,
      yoyo: true,
      repeat: -1,
      ease: 'power2.inOut',
    })
    return () => {
      tween.kill()
    }
  }, [])

  useEffect(() => {
    if (!statusRef.current) return
    if (saveStatus === 'idle') return
    if (prefersReducedMotion()) {
      gsap.set(statusRef.current, { opacity: 1 })
      return
    }

    const ctx = gsap.context(() => {
      gsap.fromTo(
        statusRef.current,
        { opacity: 0 },
        { opacity: 1, duration: 0.3, ease: 'power2.out' }
      )
      if (saveStatus === 'saved' || saveStatus === 'error') {
        gsap.to(statusRef.current, {
          opacity: 0,
          duration: 0.3,
          delay: 1.7,
          ease: 'power2.out',
        })
      }
    })
    return () => ctx.revert()
  }, [saveStatus])

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

    const provider = providerRef.current
    if (!provider) return

    const awareness = provider.getAwareness()

    bindingRef.current = new MonacoBinding(
      yText,
      model,
      new Set([editor]),
      awareness,
    )

    const updateAwareness = () => {
      const selection = editor.getSelection()
      const position = editor.getPosition()
      const state: Record<string, unknown> = {
        user: {
          id: user?.id ?? 'unknown',
          name: user?.display_name || user?.email || user?.id || 'Unknown',
          color: getUserColor(user?.id ?? 'unknown'),
        },
        cursor: position
          ? {
              lineNumber: position.lineNumber,
              column: position.column,
            }
          : null,
        selection: selection
          ? {
              startLineNumber: selection.startLineNumber,
              startColumn: selection.startColumn,
              endLineNumber: selection.endLineNumber,
              endColumn: selection.endColumn,
            }
          : null,
      }
      awareness.setLocalState(state)
      provider.sendAwarenessUpdate()
    }

    updateAwareness()

    const disposeCursor = editor.onDidChangeCursorPosition(updateAwareness)
    const disposeSelection = editor.onDidChangeCursorSelection(updateAwareness)

    editor.onKeyDown((e) => {
      if ((e.ctrlKey || e.metaKey) && e.code === 'KeyS') {
        e.preventDefault()
        e.stopPropagation()
        void saveContent()
      }
    })

    return () => {
      disposeCursor.dispose()
      disposeSelection.dispose()
    }
  }

  const statusText = {
    idle: '',
    saving: 'Saving...',
    saved: 'Saved',
    error: 'Save error',
  }

  const statusClass = {
    idle: 'text-[#A0A0A0]',
    saving: 'text-yellow-500',
    saved: 'text-green-500',
    error: 'text-red-400',
  }

  if (isLoading) {
    return (
      <div
        ref={loadingRef}
        className="h-[600px] flex items-center justify-center border border-[#3A3A3A] rounded-md bg-[#202020]"
      >
        <p className="text-[#A0A0A0]">Loading...</p>
      </div>
    )
  }

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <PresencePanel users={onlineUsers} />
        <span
          ref={statusRef}
          className={`text-sm font-medium ${statusClass[saveStatus]}`}
        >
          {statusText[saveStatus]}
        </span>
      </div>
      <div className="h-[600px] border border-[#3A3A3A] rounded-md overflow-hidden">
        <MonacoEditor
          defaultLanguage={language || 'plaintext'}
          language={language || 'plaintext'}
          theme="vs-dark"
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
