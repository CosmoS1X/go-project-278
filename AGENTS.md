# AGENTS.md

## О проекте

Go-веб-сервис на Gin. Модуль: `github.com/CosmoS1X/go-project-278` (временное
имя — к релизу заменится; массовых переименований сейчас не делать).

- Язык: Go 1.26.x, фреймворк gin v1.12, тесты testify.
- СУБД: PostgreSQL. Доступ к данным: sqlc (SQL как источник истины,
  кодогенерация). Миграции: goose. Драйвер/пул: pgx/v5 + pgxpool.
- Конфигурация: caarlos0/env (структуры с тегами), .env через godotenv
  (.air.toml уже грузит .env для разработки).
- Код и commit-сообщения — на английском. Комментарии в коде — только по запросу.

## Команды (make)

| Таргет        | Назначение                                    |
|---------------|-----------------------------------------------|
| `test`        | go test -v ./...                              |
| `lint`        | golangci-lint fmt + golangci-lint run         |
| `lint-fix`    | автоправки линтера                            |
| `build`       | go build -o ./bin/server ./cmd/server         |
| `run`         | собрать и запустить `./bin/server`            |
| `dev`         | air (hot-reload; .air.toml в gitignore)       |
| `sqlc-generate` | сгенерировать код sqlc (`cd internal/storage/sqlc && sqlc generate`) |
| `migrate-up`  | применить миграции goose (`-dir db/migrations`) |
| `migrate-down`| откатить миграции goose                       |
| `tidy`, `clean`, `test-race`, `test-coverage`, `show-coverage`, `vuln` | см. Makefile |

Перед завершением задачи обязательно: `make test && make lint`.

## Общие правила

- Исправляй причину, а не следствие.

## Структура

- `cmd/server/main.go` — точка входа: godotenv → `config.Load` → pgxpool →
  `app.NewRouter(pool, cfg)` → `router.Run(":" + PORT)`.
- `internal/config/config.go` — конфиг через caarlos0/env: `DATABASE_URL`,
  `BASE_SHORT_URL`, `PORT` (default `8080`).
- `internal/app/app.go` — NewRouter(): gin.Logger + gin.Recovery + маршруты;
  тонкий слой, только сборка роутера.
- `internal/app/app_test.go` — тесты маршрутов (testify + httptest) на реальной
  БД (skip, если нет `DATABASE_URL`).
- `internal/service/links/` — доменный слой сущности links (пакет `links`):
  - `handler.go` — HTTP-хендлеры, зависит от интерфейса `Repository`;
  - `repository.go` — интерфейс `Repository` + реализация на sqlc,
    sentinel-ошибки `ErrNotFound` / `ErrShortNameTaken`;
  - `links.go` — доменный тип `Link` + DTO (с вычисляемым `short_url`).
  - Тесты: `handler_test.go` — юнит с фейковым `Repository` (без БД);
    `repository_test.go` — интеграция на реальной БД.
- `internal/storage/sqlc/` — сгенерированный код sqlc (не редактировать руками);
  схема — `schema/schema.sql`, запросы — `query/`, конфиг — `sqlc.yaml`.
- `db/migrations/` — миграции goose (SQL, последовательная нумерация).
- Новые сущности — отдельные пакеты `internal/service/<name>` (пакет называется
  по имени сущности, не `<name>service`).

## Конвенции

- Conventional Commits: `feat:`, `fix:`, `chore:`, `test:`, `docs:`.
  Коммиты небольшие и логически раздельные; каждый коммит должен собираться.
- Тесты рядом с кодом, в том же пакете; `gin.SetMode(gin.TestMode)` в тестах.
- Импорты: форматирует gci (std / default / localmodule).
- Линтер golangci-lint v2 строгий (gosec, errcheck, staticcheck, gocritic,
  revive, ...). `nolint` — только с обоснованием.
- HTTP-слой зависит от интерфейса `Repository`, не от sqlc напрямую.
- НЕ редактировать `.github/workflows/hexlet-check.yml` (автогенерируется).

## Слой данных (PostgreSQL)

- Миграции goose в SQL: `db/migrations/`, последовательная нумерация.
- sqlc: конфиг `internal/storage/sqlc/sqlc.yaml`, команда `make sqlc-generate`;
  сгенерированный код править руками нельзя.
- В SQL-запросах всегда явно указывать поля, не использовать `*`.
