# CodeStream

> «Figma для кода» — онлайн-редактор кода с совместным редактированием в реальном времени.

## О проекте

CodeStream — pet-проект, демонстрирующий real-time совместное редактирование кода в браузере. Основной фокус сделан на backend: архитектура WebSocket-сервера на Go, синхронизация документов с помощью CRDT и надёжная персистентность.

## Стек

- **Backend**: Go 1.27, Gin, GORM
- **Frontend**: Vite + React + TypeScript + Tailwind CSS
- **База данных**: PostgreSQL 16
- **Кэш / pub-sub**: Redis 7
- **Real-time**: WebSocket (`gorilla/websocket`)
- **Collaborative editing**: CRDT (Yjs на клиенте, Go-сервер хранит/ретранслирует бинарные CRDT-updates)
- **Контейнеризация**: Docker, Docker Compose
- **CI/CD**: GitHub Actions

## Статус

MVP backend завершён. Frontend bootstrap создан (Vite + React + TypeScript + Tailwind CSS).
Реализованы: аутентификация (JWT + refresh tokens), проекты, участники, файлы, WebSocket foundation, collaborative editing, change history, CI/CD.

## Запуск

```bash
# Копировать env и заполнить значения
cp .env.example .env

# Docker Compose
make up

# Тесты
make test

# Линтер
make lint

# CI-пайплайн
make ci
```

## Команды

- `make up` — поднять проект в Docker.
- `make down` — остановить Docker.
- `make test` — запустить тесты.
- `make lint` — запустить golangci-lint.
- `make fmt` — форматировать код.
- `make ci` — запустить CI-проверки.
- `make migrate-up` — применить миграции.

## Frontend

```bash
cd frontend
npm install
npm run dev
```

## Архитектура

Краткая схема архитектуры и потока данных описаны в [ARCHITECTURE.md](ARCHITECTURE.md).

Подробное техническое проектирование: [docs/technical-design.md](docs/technical-design.md).
Обоснование стека: [docs/stack.md](docs/stack.md).

## Auth API

### Регистрация

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","display_name":"User"}'
```

### Логин

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Refresh

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

### Logout

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

### Защищённый endpoint

```bash
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/me
```

## Projects & Files API

### Проекты

```bash
# Создать проект
curl -X POST http://localhost:8080/projects \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Project"}'

# Получить список проектов
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/projects

# Получить проект
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/projects/<project_id>

# Обновить проект
curl -X PATCH http://localhost:8080/projects/<project_id> \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"New Name"}'

# Удалить проект
curl -X DELETE -H "Authorization: Bearer <access_token>" http://localhost:8080/projects/<project_id>
```

### Участники проекта

```bash
# Добавить участника
curl -X POST http://localhost:8080/projects/<project_id>/members \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"...","role":"editor"}'

# Получить список участников
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/projects/<project_id>/members

# Удалить участника
curl -X DELETE -H "Authorization: Bearer <access_token>" http://localhost:8080/projects/<project_id>/members/<user_id>
```

### Файлы

```bash
# Создать файл
curl -X POST http://localhost:8080/projects/<project_id>/files \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"main.go","path":"/main.go","language":"go"}'

# Получить список файлов проекта
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/projects/<project_id>/files

# Получить файл
curl -H "Authorization: Bearer <access_token>" http://localhost:8080/files/<file_id>

# Обновить файл
curl -X PATCH http://localhost:8080/files/<file_id> \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"updated.go","path":"/updated.go","language":"go"}'

# Удалить файл
curl -X DELETE -H "Authorization: Bearer <access_token>" http://localhost:8080/files/<file_id>
```

## WebSocket

### Подключение

```bash
websocat "ws://localhost:8080/ws?token=<access_token>&file_id=<file_id>"
```

### Сообщения

**Client → Server:**

```json
{ "event": "join:file", "data": { "fileId": "...", "stateVector": "..." } }
{ "event": "doc:update", "data": { "fileId": "...", "update": "AAABAA..." } }
{ "event": "leave:file", "data": { "fileId": "..." } }
{ "event": "pong", "data": {} }
```

**Server → Client:**

```json
{ "event": "doc:sync", "data": { "fileId": "...", "state": "AAABAA...", "stateVector": "..." } }
{ "event": "doc:update", "data": { "fileId": "...", "update": "AAABAA...", "userId": "..." } }
{ "event": "user:joined", "data": { "userId": "...", "fileId": "..." } }
{ "event": "user:left", "data": { "userId": "...", "fileId": "..." } }
{ "event": "ping", "data": {} }
```

## TODO / Следующие шаги

- [ ] Frontend на React + Monaco + Yjs
- [ ] Presence (курсоры, онлайн-статус)
- [ ] Redis pub/sub для горизонтального масштабирования WebSocket
- [ ] Интеграционные тесты WebSocket с PostgreSQL
- [ ] Snapshot merge на сервере через Yjs/WASM
- [ ] Undo/redo истории

## Contributing

См. [CONTRIBUTING.md](CONTRIBUTING.md).
