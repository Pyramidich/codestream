# CodeStream Project Context

> «Figma для кода» — онлайн-редактор кода с совместным редактированием в реальном времени.

---

## 1. Overview

CodeStream is a pet-project demonstrating real-time collaborative code editing in the browser. The backend focus is on WebSocket server architecture, CRDT document synchronization, and reliable persistence. The frontend is a Vite + React + TypeScript + Tailwind CSS SPA with Monaco Editor and Yjs for collaborative editing.

---

## 2. Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.27, Gin, GORM |
| **Frontend** | Vite, React 19, TypeScript, Tailwind CSS |
| **Real-time** | WebSocket (`gorilla/websocket`) |
| **Collaborative Editing** | CRDT (Yjs on client, Go server stores/relays binary CRDT updates) |
| **Database** | PostgreSQL 16 |
| **Cache / Pub-Sub** | Redis 7 |
| **Auth** | JWT access/refresh tokens, bcrypt |
| **Migrations** | golang-migrate |
| **Containerization** | Docker, Docker Compose |
| **CI/CD** | GitHub Actions |

---

## 3. Repository Structure

```
.
├── cmd/server/main.go              # Entry point
├── internal/
│   ├── authz/                      # Authorization logic
│   ├── config/                     # App configuration
│   ├── handler/                    # HTTP handlers (REST + WS)
│   ├── middleware/                 # Gin middleware (auth, CORS, logging)
│   ├── models/                     # GORM models
│   ├── repository/                 # Data access layer (GORM)
│   ├── server/                     # Gin server setup, DI, routes
│   ├── service/                    # Business logic
│   ├── testutil/                   # Test helpers
│   └── ws/                         # WebSocket hub, connections, protocol
├── migrations/                     # SQL migrations (golang-migrate)
├── frontend/                       # Vite + React + TS + Tailwind SPA
│   ├── src/
│   │   ├── api/                    # Axios API clients (auth, projects, files)
│   │   ├── components/             # React components (Editor, PresencePanel, Layout, ProtectedRoute)
│   │   ├── contexts/               # AuthContext
│   │   ├── pages/                  # Page components (Home, Login, Register, Projects, Project, Editor)
│   │   ├── providers/              # WebSocket provider (CodestreamProvider)
│   │   ├── types/                  # TypeScript types
│   │   ├── App.tsx                 # Router + AuthProvider
│   │   └── main.tsx                # Entry point
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── .env.example
```

---

## 4. Backend Architecture

### 4.1 Layers

```
┌─────────────────────────────────────────────┐
│                  Clients                      │
└───────────────────┬─────────────────────────┘
                    │ HTTP / WebSocket
                    ▼
          ┌──────────────────┐
          │   Gin Router     │
          └────────┬─────────┘
                   │
     ┌─────────────┴──────────────┐
     │                            │
  REST API                  WebSocket (/ws)
     │                            │
     ▼                            ▼
┌──────────┐              ┌──────────────┐
│ Handlers │              │  WSHandler   │
│ (auth,   │              │  - JWT auth  │
│ projects,│              │  - Rooms     │
│ files)   │              │  - Heartbeat │
└────┬─────┘              └──────┬───────┘
     │                            │
     ▼                            ▼
┌──────────┐              ┌──────────────┐
│ Services │              │ DocumentState│
│ (business│              │ Manager      │
│  logic)  │              │ (CRDT)       │
└────┬─────┘              └──────┬───────┘
     │                            │
     ▼                            ▼
┌──────────              ┌──────────────
│Repositories│            │ PostgreSQL   │
│  (GORM)   │             │ (snapshots)  │
└────┬─────┘              └──────────────┘
     │
     ▼
┌──────────┐
│PostgreSQL│
└──────────┘
```

### 4.2 REST Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Register new user |
| POST | `/auth/login` | No | Login, get tokens |
| POST | `/auth/refresh` | No | Refresh access token |
| POST | `/auth/logout` | No | Revoke refresh token |
| GET | `/me` | Yes | Get current user `{id, email, display_name}` |
| POST | `/projects` | Yes | Create project |
| GET | `/projects` | Yes | List projects |
| GET | `/projects/:id` | Yes | Get project |
| PATCH | `/projects/:id` | Yes | Update project |
| DELETE | `/projects/:id` | Yes | Delete project |
| POST | `/projects/:id/members` | Yes | Add member |
| GET | `/projects/:id/members` | Yes | List members |
| DELETE | `/projects/:id/members/:user_id` | Yes | Remove member |
| POST | `/projects/:id/files` | Yes | Create file |
| GET | `/projects/:id/files` | Yes | List files |
| GET | `/files/:id` | Yes | Get file |
| PATCH | `/files/:id` | Yes | Update file |
| DELETE | `/files/:id` | Yes | Delete file |
| GET | `/ws` | JWT (query) | WebSocket upgrade |

### 4.3 API Response Formats

**Important:** Some endpoints wrap responses in objects. Frontend handles both formats.

```json
// GET /projects
{ "projects": [...] }
// or (older)
[...]

// GET /projects/:id/files
{ "files": [...] }
// or (older)
[...]
```

### 4.4 WebSocket Protocol

**Connection:**
```
ws://localhost:8080/ws?token=<jwt>&file_id=<file_id>
```

**Client → Server:**

| Event | Data | Description |
|-------|------|-------------|
| `join:file` | `{ fileId, stateVector }` | Join file room |
| `leave:file` | `{ fileId }` | Leave file room |
| `doc:update` | `{ fileId, update }` | CRDT update (base64) |
| `pong` | `{}` | Heartbeat response |

**Server → Client:**

| Event | Data | Description |
|-------|------|-------------|
| `doc:sync` | `{ fileId, state, stateVector }` | Full snapshot on join |
| `doc:update` | `{ fileId, update, userId }` | Broadcast CRDT update |
| `user:joined` | `{ userId }` | User joined room |
| `user:left` | `{ userId }` | User left room |
| `presence:list` | `{ users: [...] }` | Current room users list |
| `ping` | `{}` | Heartbeat ping |

### 4.5 Collaborative Editing Flow

1. Client opens file → connects to WebSocket with JWT + file_id.
2. Server validates JWT, checks project membership.
3. Server loads CRDT snapshot from `files.content`, sends `doc:sync`.
4. Client edits generate CRDT updates via Yjs.
5. Client sends `doc:update` to server.
6. Server saves update in `document_versions`, broadcasts to room.
7. Server keeps in-memory snapshot, persists to `files.content` every 30s or when room empties.

### 4.6 Authentication

- JWT access token (short-lived) + refresh token (long-lived, SHA-256 hashed, stored in DB).
- Access token in `Authorization: Bearer <token>` header.
- Refresh via `/auth/refresh`.
- WebSocket auth via `?token=<jwt>` query param.

### 4.7 CORS

Configured for `http://localhost:5173` and `http://localhost:3000`.

---

## 5. Frontend Architecture

### 5.1 Stack

- **Build tool:** Vite 8.3
- **Framework:** React 19, TypeScript 6.0
- **Styling:** Tailwind CSS 3.4
- **Editor:** Monaco Editor + `@monaco-editor/react`
- **Collaborative:** Yjs + y-monaco
- **HTTP:** Axios
- **Routing:** react-router-dom 7.18

### 5.2 Structure

```
frontend/src/
├── api/
│   ├── auth.ts          # Auth API (login, register, logout, refresh, me)
│   ├── client.ts        # Axios client with interceptors (JWT, refresh)
│   ├── files.ts         # Files API (list, create, get)
│   └── projects.ts      # Projects API (list, create, get)
├── components/
│   ├── Editor.tsx         # Monaco + Yjs + WebSocket integration
│   ├── Layout.tsx         # App layout with navbar
│   ├── PresencePanel.tsx  # Online users list with names and colors
│   ├── ProjectMembers.tsx # Project members management UI
│   └── ProtectedRoute.tsx # Auth guard
├── contexts/
│   └── AuthContext.tsx    # Auth state, login, register, logout
├── pages/
│   ├── EditorPage.tsx     # File editor page
│   ├── HomePage.tsx       # Landing page
│   ├── LoginPage.tsx      # Login form
│   ├── ProjectPage.tsx    # Project files list
│   ├── ProjectSettings.tsx # Project settings (members management)
│   ├── ProjectsPage.tsx   # Projects list
│   └── RegisterPage.tsx   # Registration form
├── providers/
│   └── websocket.ts       # CodestreamProvider (WebSocket client)
├── types/
│   ├── auth.ts            # User, Tokens, AuthResponse
│   └── index.ts           # Project, ProjectFile, ProjectMember
├── utils/
│   ├── colors.ts          # Per-user color generation for presence
│   └── language.ts        # Language detection and options
├── App.tsx                # Router + AuthProvider with lazy-loaded pages
└── main.tsx               # Entry point
```

### 5.3 Auth Flow

1. Login/Register → tokens saved to `localStorage`.
2. `AuthContext` checks `/me` on mount.
3. Axios interceptor adds `Authorization: Bearer <token>`.
4. On 401, attempts refresh; if fails → redirect to `/login`.
5. `ProtectedRoute` redirects unauthenticated users to `/login`.

### 5.4 WebSocket Provider

`src/providers/websocket.ts`:
- Manages WebSocket connection with auto-reconnect (max 3 attempts).
- Message format: `{ event: string, data: Record<string, unknown> }`.
- Methods: `connect()`, `disconnect()`, `sendJoin()`, `sendLeave()`, `sendUpdate()`, `onMessage()`.

### 5.5 Editor Integration

`src/components/Editor.tsx`:
- Creates `Y.Doc` and `Y.Text('monaco')`.
- Connects to WebSocket on mount.
- Sends `Y.encodeStateVector()` on `join:file`.
- Applies `doc:sync` and `doc:update` via `Y.applyUpdate()`.
- Sends local updates via `Y.encodeStateAsUpdate()` on `doc:update`.
- Uses `y-monaco` `MonacoBinding` for Monaco ↔ Yjs sync.
- Tracks online users via `user:joined`, `user:left`, `presence:list` events.
- Remote cursor and selection awareness via Yjs `Awareness` and `y-protocols/awareness`.
- Displays user names and colors in the presence panel.
- Monaco Editor language is synced from `files.language`.

### 5.6 Presence

`src/components/PresencePanel.tsx`:
- Displays `Users online: N` and list of user names with colored dots.
- Receives `users` prop from `Editor.tsx`.

---

## 6. Database Schema

### Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts (id, email, password_hash, display_name) |
| `projects` | Projects (id, name, owner_id) |
| `project_members` | Project membership (id, project_id, user_id, role) |
| `files` | Files (id, project_id, name, path, language, content, content_text, content_type) |
| `document_versions` | CRDT update history (id, file_id, update_payload, created_by) |
| `change_history` | Audit log (id, file_id, user_id, action, metadata) |
| `refresh_tokens` | Refresh tokens (id, user_id, token_hash, expires_at, revoked_at) |

### Key Design Decisions

- `files.content` stores binary Yjs update (`content_type = 'yjs-binary'`), not plain text.
- `files.content_text` stores the latest plain-text snapshot for save/restore and search.
- Plain text is derived on the client and persisted on save.
- `project_members.role` is `owner`, `editor`, or `viewer`. Viewers cannot create/edit files or add members.
- `document_versions` keeps last N updates for history; old ones are compacted.

---

## 7. Known Issues & Fixes

| Issue | Fix | File |
|-------|-----|------|
| `TypeError: projects.map is not a function` | Handle `{ projects: [] }` wrapper | `src/pages/ProjectsPage.tsx` |
| `TypeError: files.map is not a function` | Handle `{ files: [] }` wrapper | `src/pages/ProjectPage.tsx` |
| WebSocket disconnect in React StrictMode | Lazy disconnect with 100ms delay | `src/components/Editor.tsx` |
| Dark mode input fields | Removed `color-scheme: light dark` from `index.css` | `src/index.css` |
| Presence counter always 0 | Backend broadcasts `user:joined`/`user:left`, frontend tracks in state | `internal/ws/hub.go`, `src/components/Editor.tsx` |
| y-monaco import error | Added Vite alias for `monaco-editor/esm/vs/editor/editor.api.js` | `vite.config.ts` |
| Member list not enriched | Preload `User` relation in `FindByProjectIDWithUser` and return `MemberInfo` | `internal/repository/project_member.go`, `internal/service/project_member.go` |
| Member invite by user_id | Changed request to `email`, lookup user by email, allow owner/editor to invite | `internal/handler/project_member.go`, `internal/service/project_member.go` |
| Last owner removal | Count owners before removal, block if only one owner remains | `internal/service/project_member.go` |
| Monaco language hardcoded | Pass `language` prop from file to Monaco Editor | `src/components/Editor.tsx`, `src/pages/EditorPage.tsx` |
| Large single JS chunk | Split vendor/editor/yjs/react into separate manual chunks and lazy-loaded pages | `vite.config.ts`, `src/App.tsx` |

---

## 8. Current State

### What Works
- ✅ Auth (register, login, logout, refresh, JWT middleware)
- ✅ Projects CRUD
- ✅ Files CRUD
- ✅ WebSocket foundation (rooms, heartbeat, reconnect)
- ✅ Collaborative editing (Yjs CRDT updates, snapshots)
- ✅ Presence (user:joined, user:left, online counter)
- ✅ Change history logging
- ✅ CI/CD (GitHub Actions)
- ✅ Docker Compose setup
- ✅ Frontend: Monaco + Yjs + WebSocket integration
- ✅ Presence: remote cursors, user names, colors
- ✅ Project members: invite by email, roles, remove
- ✅ Monaco language sync with files.language
- ✅ Bundle optimization (lazy chunks, manual chunks)
- ✅ E2E tests (save/restore, collaboration, members, language)

### What's Left
- 🔄 Redis pub/sub for horizontal WebSocket scaling
- 🔄 Integration tests for WebSocket with PostgreSQL
- 🔄 Snapshot merge on server (Yjs/WASM)
- 🔄 Undo/redo history
- 🔄 File content plain-text derivation for search

---

## 9. How to Run

### Backend

```bash
# Copy env and fill values
cp .env.example .env

# Docker Compose (PostgreSQL + Redis + Backend)
make up

# Or run locally (requires PostgreSQL + Redis running)
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173
npm run build      # Production build
```

### Tests

```bash
# Backend tests
make test

# Frontend lint
npx oxlint
```

### Frontend Tests

```bash
cd frontend
npx playwright test --project=chromium
```

### Bundle Analysis

```bash
cd frontend
npm run analyze        # opens vite-bundle-visualizer report
```

---

## 10. Architecture Decisions

1. **CRDT over OT:** CRDT (Yjs) provides convergence without central transformation, simplifying server logic. Server treats updates as opaque blobs.
2. **Go server as opaque relay:** No CRDT interpretation on server. Avoids need for Go CRDT library.
3. **JWT in WebSocket query param:** Simple for MVP. Risk: token in logs. Future: short-lived WS tickets.
4. **In-memory Room Manager:** Sufficient for single-instance MVP. Future: Redis pub/sub + sticky sessions for scaling.
5. **Yjs binary in PostgreSQL:** `files.content` stores binary Yjs update. Plain text is derived on client.
6. **React StrictMode lazy disconnect:** 100ms delay in cleanup to survive StrictMode's double mount/unmount cycle.

---

## 11. Security Notes

- JWT secret must be changed in production.
- WebSocket tokens are passed via query param (may appear in logs).
- CORS allows `localhost:5173` and `localhost:3000` only.
- No secrets are committed to this file.
