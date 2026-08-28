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
| `tidy`, `clean`, `test-race`, `test-coverage`, `show-coverage`, `vuln` | см. Makefile |

Перед завершением задачи обязательно: `make test && make lint`.

## Структура

- `cmd/server/main.go` — точка входа (NewRouter → Run).
- `internal/app/app.go` — NewRouter(): gin.Logger + gin.Recovery + маршруты.
- `internal/app/app_test.go` — тесты маршрутов (testify + httptest).
- При росте: `internal/config`, `internal/storage` (sqlc) и т.п.
  Точные имена пакетов — согласовать на момент внедрения.

## Конвенции

- Conventional Commits: `feat:`, `fix:`, `chore:`, `test:`, `docs:`.
  Коммиты небольшие и логически раздельные; каждый коммит должен собираться.
- Тесты рядом с кодом, в том же пакете; `gin.SetMode(gin.TestMode)` в тестах.
- Импорты: форматирует gci (std / default / localmodule).
- Линтер golangci-lint v2 строгий (gosec, errcheck, staticcheck, gocritic,
  revive, ...). `nolint` — только с обоснованием.
- НЕ редактировать `.github/workflows/hexlet-check.yml` (автогенерируется).

## Планируемый слой данных (PostgreSQL)

- Миграции goose в SQL: последовательная нумерация, каталог согласовать
  (напр. `db/migrations/`).
- sqlc: запросы и схема в каталогах через `sqlc.yaml`, генерация командой
  `sqlc generate`; сгенерированный код не править руками.