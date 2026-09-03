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
| `dev`         | запустить бэкенд (air) и фронтенд (vite preview) вместе через concurrently |
| `dev-backend` | только air (hot-reload; .air.toml в gitignore) |
| `dev-frontend`| только фронтенд (`npm exec start-hexlet-url-shortener-frontend`) |
| `sqlc-generate` | сгенерировать код sqlc (`cd internal/storage/sqlc && sqlc generate`) |
| `migrate-up`  | применить миграции goose (`-dir db/migrations`) |
| `migrate-down`| откатить миграции goose                       |
| `tidy`, `clean`, `test-race`, `test-coverage`, `show-coverage`, `vuln` | см. Makefile |

Перед завершением задачи обязательно: `make test && make lint`.

## Общие правила

- Исправляй причину, а не следствие.

## Структура

- `cmd/server/main.go` — точка входа: godotenv → `config.Load` → pgxpool →
  `stdlib.OpenDBFromPool` → `app.NewRouter(db, cfg)` → `router.Run(":" + PORT)`.
- `internal/config/config.go` — конфиг через caarlos0/env: `DATABASE_URL`,
  `BASE_SHORT_URL`, `PORT` (default `8080`).
- `internal/app/app.go` — NewRouter(): CORS (gin-contrib/cors) + gin.Logger +
  gin.Recovery + маршруты; тонкий слой, только сборка роутера; принимает
  `sqlc.DBTX` (совместим с `*sql.DB` и `*sql.Tx`).
- `internal/app/app_test.go` — тесты маршрутов (testify + httptest) на реальной
  БД (skip, если нет `DATABASE_URL`).
- `internal/service/links/` — доменный слой сущности links (пакет `links`):
  - `handler.go` — HTTP-хендлеры, зависит от интерфейса `Repository`;
    пагинация List через `?range=[start,end]`, ответ с `Content-Range`;
  - `repository.go` — интерфейс `Repository` + реализация на sqlc,
    sentinel-ошибки `ErrNotFound` / `ErrShortNameTaken`;
    `List(ctx, offset, limit int32) ([]Link, int64, error)`;
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
- Интеграционные тесты, работающие с реальной БД, изолировать через транзакции:
  каждый тест в своей транзакции `Begin()` + `Rollback()` в `t.Cleanup`
  (без `TRUNCATE`/удаления чужих данных).
- Импорты: форматирует gci (std / default / localmodule).
- Линтер golangci-lint v2 строгий (gosec, errcheck, staticcheck, gocritic,
  revive, ...). `nolint` — только с обоснованием.
- HTTP-слой зависит от интерфейса `Repository`, не от sqlc напрямую.
- НЕ редактировать `.github/workflows/hexlet-check.yml` (автогенерируется).

## Фронтенд и деплой

- Фронтенд — пакет `@hexlet/project-url-shortener-frontend` (НЕ редактировать;
  только исследовать). Живёт в `frontend/`, зависимости ставятся `npm ci`
  (node_modules в .gitignore). Это предсобранный статический SPA (`dist/`),
  API-вызовы — на относительный путь `/api` своего origin.
- Запуск фронта: `npm exec start-hexlet-url-shortener-frontend` = `vite preview`
  (порт 5173). `vite preview` наследует `server.proxy`, поэтому локально `/api`
  проксируется на бэкенд — с точки зрения браузера всё same-origin, CORS
  фактически обходится.
- Планируемый деплой: render.com, в docker-контейнере. **Caddy раздаёт статику
  фронта и проксирует `/api` на бэкенд.** Это ещё НЕ реализовано (в проекте нет
  Caddyfile/nginx/Dockerfile/render.yaml) — реализация идёт отдельной веткой,
  отдельно от основного кода.
- CORS на бэкенде (gin-contrib/cors): `AllowOrigins` захардкожен
  `http://localhost:5173`, методы GET/POST/PUT/DELETE, заголовки Content-Type,
  expose `Content-Range` (нужен react-admin для total). Это dev-конфигурация;
  при реальном деплое origin фронта будет иным — держать в уме.

## Слой данных (PostgreSQL)

- Миграции goose в SQL: `db/migrations/`, последовательная нумерация.
- sqlc: конфиг `internal/storage/sqlc/sqlc.yaml`, команда `make sqlc-generate`;
  сгенерированный код править руками нельзя.
- В SQL-запросах всегда явно указывать поля, не использовать `*`.
- Пагинация: `GET /api/links?range=[start,end]` — **инклюзивный** диапазон
  (как у react-admin/`ra-data-simple-rest`): для `perPage=5` фронт шлёт
  `range=[0,4]`, бэкенд возвращает `limit = end-start+1` записей. Без
  параметра `range` — дефолт 10 записей, максимум 100.
  Ответ: JSON-массив + `Content-Range: links {start}-{end}/{total}`
  (end — последний индекс возвращённых записей).
- Типы: offset/limit = `int32` (sqlc LIMIT/OFFSET), total/ID = `int64`
  (COUNT::bigint / BIGSERIAL). BIGSERIAL оставлен: смена на SERIAL
  не устраняет приведения (total = int64), но сужает PK.
