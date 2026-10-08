import * as Y from 'yjs'
import { Awareness, applyAwarenessUpdate, encodeAwarenessUpdate } from 'y-protocols/awareness'

export interface WebSocketMessage {
  event: string
  data: Record<string, unknown>
}

export class CodestreamProvider {
  private url: string
  private fileId: string
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 3
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null
  private messageListeners: Array<(message: WebSocketMessage) => void> = []
  private shouldReconnect = true
  private yDoc: Y.Doc
  private awareness: Awareness

  constructor(url: string, fileId: string, yDoc: Y.Doc) {
    this.url = url
    this.fileId = fileId
    this.yDoc = yDoc
    this.awareness = new Awareness(this.yDoc)
  }

  getAwareness(): Awareness {
    return this.awareness
  }

  setLocalState(state: Record<string, unknown> | null) {
    this.awareness.setLocalState(state)
  }

  onAwarenessChange(callback: () => void) {
    this.awareness.on('change', callback)
    return () => {
      this.awareness.off('change', callback)
    }
  }

  sendAwarenessUpdate() {
    const update = encodeAwarenessUpdate(this.awareness, [this.awareness.clientID])
    this.send('awareness:update', {
      fileId: this.fileId,
      update: arrayBufferToBase64(update),
    })
  }

  applyAwarenessUpdate(updateBase64: string) {
    try {
      const update = base64ToArrayBuffer(updateBase64)
      applyAwarenessUpdate(this.awareness, update, 'remote')
    } catch {
      // ignore invalid awareness update
    }
  }

  connect() {
    this.shouldReconnect = true
    this.createConnection()
  }

  private createConnection() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {
      this.reconnectAttempts = 0
    }

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as WebSocketMessage
        this.handleServerMessage(message)
      } catch {
        // ignore malformed messages
      }
    }

    this.ws.onclose = () => {
      this.attemptReconnect()
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  private attemptReconnect() {
    if (!this.shouldReconnect) return
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return

    this.reconnectAttempts += 1
    this.reconnectTimeout = setTimeout(() => {
      this.createConnection()
    }, 1000 * this.reconnectAttempts)
  }

  private handleServerMessage(message: WebSocketMessage) {
    if (message.event === 'ping') {
      this.sendPong()
      return
    }

    if (message.event === 'awareness:update') {
      const payload = message.data as { update?: string }
      if (payload.update) {
        this.applyAwarenessUpdate(payload.update)
      }
      return
    }

    this.messageListeners.forEach((callback) => callback(message))
  }

  disconnect() {
    this.shouldReconnect = false
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    this.ws?.close()
    this.ws = null
  }

  onMessage(callback: (message: WebSocketMessage) => void) {
    this.messageListeners.push(callback)
  }

  private send(event: string, data: Record<string, unknown>) {
    if (this.ws?.readyState !== WebSocket.OPEN) return
    this.ws.send(JSON.stringify({ event, data }))
  }

  sendUpdate(update: Uint8Array) {
    this.send('doc:update', {
      fileId: this.fileId,
      update: arrayBufferToBase64(update),
    })
  }

  sendJoin(stateVector: Uint8Array | null) {
    this.send('join:file', {
      fileId: this.fileId,
      stateVector: stateVector ? arrayBufferToBase64(stateVector) : '',
    })
  }

  sendLeave() {
    this.send('leave:file', { fileId: this.fileId })
  }

  sendPong() {
    this.send('pong', {})
  }
}

function arrayBufferToBase64(buffer: Uint8Array): string {
  let binary = ''
  const len = buffer.byteLength
  for (let i = 0; i < len; i++) {
    binary += String.fromCharCode(buffer[i])
  }
  return btoa(binary)
}

export function base64ToArrayBuffer(base64: string): Uint8Array {
  const binary = atob(base64)
  const len = binary.length
  const buffer = new Uint8Array(len)
  for (let i = 0; i < len; i++) {
    buffer[i] = binary.charCodeAt(i)
  }
  return buffer
}

export default CodestreamProvider
