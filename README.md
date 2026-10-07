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

Bootstrap backend завершён. Сервер запускается, `/health` работает, тесты проходят. Реализованы: конфигурация, логирование, Gin + health/readiness endpoints, Docker Compose с PostgreSQL и Redis, миграции, CI workflow.

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
