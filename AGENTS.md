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
| `dev-backend` | только air (hot-reload) |
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
  `stdlib.OpenDBFromPool` → `app.NewRouter(db, cfg)` → `router.Run(":" + SERVER_PORT)`.
- `internal/config/config.go` — конфиг через caarlos0/env: `DATABASE_URL`,
  `BASE_SHORT_URL`, `SERVER_PORT` (default `8080`), `CORS_ORIGIN` (default
  `http://localhost:5173`).
- `internal/app/app.go` — NewRouter(): CORS (gin-contrib/cors) + gin.Logger +
  gin.Recovery + маршруты; тонкий слой, только сборка роутера; принимает
  `sqlc.DBTX` (совместим с `*sql.DB` и `*sql.Tx`).
  Маршруты: `/ping`, `GET /r/:code` (редирект), `/api/links*`, `GET /api/link_visits`.
  `router.TrustedPlatform = gin.PlatformCloudflare` — корректный ClientIP
  (нужен статистике посещений) при работе за Cloudflare.
- `internal/app/app_test.go` — тесты маршрутов (testify + httptest) на реальной
  БД (skip, если нет `DATABASE_URL`).
- `internal/service/links/` — доменный слой сущности links (пакет `links`):
  - `handler.go` — HTTP-хендлеры, зависит от интерфейсов `Repository` и
    `VisitRecorder` (запись визитов при редиректе); пагинация List через
    `?range=[start,end]` (общий хелпер `httpapi.ParseRangeParam`), ответ с
    `Content-Range`;
  - `repository.go` — интерфейс `Repository` + реализация на sqlc,
    sentinel-ошибки `ErrNotFound` / `ErrShortNameTaken`;
    `List(ctx, offset, limit int32) ([]Link, int64, error)`,
    `GetByShortName(ctx, shortName)` (для редиректа);
  - `links.go` — доменный тип `Link` + DTO (с вычисляемым `short_url`).
  - Тесты: `handler_test.go` — юнит с фейковым `Repository` и
    `fakeVisitRecorder` (без БД); `repository_test.go` — интеграция на реальной БД.
- `internal/service/visits/` — доменный слой сущности link_visits (пакет
  `visits`): тип `LinkVisit` + DTO (JSON-поле `reffer`), `handler.go`
  (`ListVisits`, пагинация как у links), `repository.go` — интерфейс
  `Repository` (`RecordVisit`, `ListVisits`) + реализация на sqlc.
  Тесты: `handler_test.go` юнит с фейком, `repository_test.go` интеграция.
- `internal/httpapi/` — общие HTTP-хелперы: `ParseRangeParam(c, errKey)`
  (инклюзивный `range=[start,end]`, дефолт 10, максимум 100).
- `internal/storage/sqlc/` — сгенерированный код sqlc (не редактировать руками);
  схема — `schema/schema.sql`, запросы — `query/`, конфиг — `sqlc.yaml`.
- `db/migrations/` — миграции goose (SQL, последовательная нумерация).
- Новые самостоятельные сущности — отдельные пакеты
  `internal/service/<name>` (пакет называется по имени сущности, не
  `<name>service`); связанная аналитика `link_visits` живёт в своём пакете
  `visits` и не пересекается с доменом `links`.

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
- Деплой: render.com, в docker-контейнере. **Caddy — точка входа**: раздаёт
  статику фронта из `/app/public` и проксирует `/api/*`, `/ping` и `/r/*` на
  бэкенд (localhost:8080). Caddyfile: `:80`, `handle /api/*` и `handle /r/*` →
  reverse_proxy, `try_files {path} /index.html` для SPA-роутинга,
  `file_server` для статики.
- `Dockerfile` — 3 стадии: (1) frontend-builder `node:22-alpine` (`npm ci`),
  (2) backend-builder `golang:1.26-alpine` (go build + goose
  `v3.27.3`), (3) runtime `alpine:3.22` (статический бинарник Caddy
  `v2.11.4` с GitHub Releases, goose из builder). См. также `bin/run.sh`:
  goose migrate-up → запуск Caddy в фоне → `exec /app/bin/server`.
- Конфиг: `DATABASE_URL`, `BASE_SHORT_URL` (required), `SERVER_PORT` (default
  8080, внутренний порт Go-сервера), `CORS_ORIGIN` (default
  `http://localhost:5173`, можно переопределить под origin фронта в деплое).
  Порт Caddy (`PORT` на Render) задаётся в Caddyfile (`:80`) и НЕ должен
  конфликтовать с `SERVER_PORT`.
- CORS на бэкенде (gin-contrib/cors): `AllowOrigins` берётся из
  `cfg.CORSOrigin`, методы GET/POST/PUT/DELETE, заголовки Content-Type,
  expose `Content-Range` (нужен react-admin для total).

## Слой данных (PostgreSQL)

- Миграции goose в SQL: `db/migrations/`, последовательная нумерация.
- sqlc: конфиг `internal/storage/sqlc/sqlc.yaml`, команда `make sqlc-generate`;
  сгенерированный код править руками нельзя.
- В SQL-запросах всегда явно указывать поля, не использовать `*`.
- Пагинация: `GET /api/links?range=[start,end]` и `GET /api/link_visits?range=[start,end]` —
  **инклюзивный** диапазон
  (как у react-admin/`ra-data-simple-rest`): для `perPage=5` фронт шлёт
  `range=[0,4]`, бэкенд возвращает `limit = end-start+1` записей. Без
  параметра `range` — дефолт 10 записей, максимум 100.
  Ответ: JSON-массив + `Content-Range: links {start}-{end}/{total}` (или
  `link_visits ...`) — end последний индекс возвращённых записей.
- Таблица `link_visits` (миграция `00002`): `link_id` на `links(id)` с
  `ON DELETE CASCADE`, колонки `ip`, `referer`, `user_agent`, `status`,
  `created_at`; индекс по `link_id`.
- **Gotcha фронта:** JSON-поле называется `reffer` (две «f») — так читает
  предсобранный бандл фронта; колонка БД — `referer`. В коде отдаём сырые
  ключи (data provider не маппит поля), комментарий-пояснение — у поля в DTO.
- Типы: offset/limit = `int32` (sqlc LIMIT/OFFSET), total/ID = `int64`
  (COUNT::bigint / BIGSERIAL). BIGSERIAL оставлен: смена на SERIAL
  не устраняет приведения (total = int64), но сужает PK.
