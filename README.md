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

Bootstrap и аутентификация backend завершены. Реализованы: регистрация, логин, refresh, logout, JWT middleware, интеграция с PostgreSQL через GORM.

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
