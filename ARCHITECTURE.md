# CodeStream Backend Architecture

This document describes the high-level architecture of the CodeStream backend.

## Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         Clients (React + Monaco)                  │
└─────────────────────┬───────────────────────────────────────────┘
                      │ HTTP / WebSocket
                      ▼
            ┌──────────────────┐
            │   Gin Router     │
└─────────────────────────────┴───────────┐
            │                                │
     REST API (/auth, /projects, /files)    WebSocket (/ws)
            │                                │
            ▼                                ▼
   ┌────────────────┐              ┌─────────────────────────┐
   │   Handlers     │              │    WSHandler            │
   │ (Auth, Project,│              │  - JWT via query param  │
   │  File, Member) │              │  - Room per file        │
   └───────┬────────┘              │  - Heartbeat            │
           │                       └─────────────┬───────────┘
           ▼                                     │
   ┌────────────────┐                    ┌───────┴────────┐
   │    Services    │                    │  DocumentState │
   │ (business logic)│                   │     Manager    │
   └───────────────┘                   └───────┬────────┘
           │                                     │
           ▼                                     ▼
   ┌────────────────┐                    ┌───────────────┐
   │  Repositories  │                  │  CRDT snapshots│
   │   (GORM + DB)  │                  │  document_versions
   └───────┬────────┘                  └───────┬────────┘
           │                                   │
   ┌────────┴────────┐               ┌────────┴────────┐
   │   PostgreSQL     │               │   PostgreSQL     │
   └──────────────────┘               └──────────────────┘
```

## Collaborative Editing Flow

1. Client opens a file and connects to `GET /ws?token=<jwt>&file_id=<id>`.
2. Server validates JWT and checks project membership.
3. Server loads the CRDT snapshot from `files.content` and sends `doc:sync`.
4. Client edits generate CRDT updates (Yjs).
5. Client sends `doc:update` to the server.
6. Server saves the update in `document_versions` and broadcasts it to the room.
7. Server keeps an in-memory snapshot and persists it to `files.content` every 30 seconds or when the room becomes empty.
8. Old `document_versions` are compacted, keeping the last 50 entries.

## Data Flow

- **Auth**: JWT access token + SHA-256 hashed refresh tokens.
- **Projects**: owner/editor/viewer roles via `project_members`.
- **Files**: content stored as binary Yjs update (`yjs-binary`).
- **Change History**: async audit log for file and member changes.

## Detailed Design

For full technical details, see [docs/technical-design.md](docs/technical-design.md).
