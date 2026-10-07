# CodeStream

> «Figma для кода» — онлайн-редактор кода с совместным редактированием в реальном времени.

## О проекте

CodeStream — pet-проект, демонстрирующий real-time совместное редактирование кода в браузере. Основной фокус сделан на backend: архитектура WebSocket-сервера на Go, синхронизация документов с помощью CRDT и надёжная персистентность.

## Документация

- Основное техническое проектирование: **[docs/technical-design.md](docs/technical-design.md)**
- Обоснование выбора стека: **[docs/stack.md](docs/stack.md)**

## Ключевые решения (кратко)

- **Backend**: Go + Gin + GORM + PostgreSQL + Redis.
- **Real-time**: `gorilla/websocket`.
- **Collaborative editing**: CRDT на базе Yjs (клиент) + Go-сервер хранит и ретранслирует бинарные CRDT-updates.
- **Frontend**: React + Vite + Monaco Editor + `y-monaco` (реализуется позже).
- **Инфраструктура**: Docker, Docker Compose, GitHub Actions.

## Статус

Bootstrap, аутентификация и REST API для проектов/файлов завершены. Реализованы: регистрация, логин, refresh, logout, JWT middleware, проекты, участники, файлы, авторизация на уровне ресурсов.

## Запуск

```bash
# Копировать env и заполнить значения
cp .env.example .env

# Docker Compose
make up

# Тесты
make test

# Миграции (требуется DATABASE_URL)
make migrate-up
```

## Команды

- `make up` — поднять проект в Docker.
- `make down` — остановить Docker.
- `make test` — запустить тесты.
- `make migrate-up` — применить миграции.

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

## WebSocket

### Подключение

```bash
websocat "ws://localhost:8080/ws?token=<access_token>&file_id=<file_id>"
```

### Сообщения

**Client → Server:**

```json
{ "event": "join:file", "data": { "fileId": "..." } }
{ "event": "leave:file", "data": { "fileId": "..." } }
{ "event": "pong", "data": {} }
```

**Server → Client:**

```json
{ "event": "joined:file", "data": { "fileId": "..." } }
{ "event": "user:joined", "data": { "userId": "...", "fileId": "..." } }
{ "event": "user:left", "data": { "userId": "...", "fileId": "..." } }
{ "event": "ping", "data": {} }
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
