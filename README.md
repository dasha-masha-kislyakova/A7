# A7 — микросервисы на Go (auth, office, logistic, proxy)

- Разделение на службы: `auth`, `office`, `logistic`, `proxy`
- Две отдельные БД PostgreSQL: для `office` и для `logistic`
- Межсервисное взаимодействие — HTTP
- Главный `main` — в корне репозитория; по переменной `SERVICE` запускает один сервис или все сразу

## Запуск через Docker
```bash
make local
# или
docker compose up --build
```
Прокси доступен на `http://localhost:8080`.

## Локальный запуск одной командой
```bash
make run-all
```
Запускает `auth` (8083), `office` (8081), `logistic` (8082), `proxy` (8080) и раздачу статических файлов из `FE/`.

## Переменные окружения (основные)
- `SERVICE`: `auth` | `office` | `logistic` | `proxy` | `all` (по умолчанию `proxy`)
- `PORT`: порт конкретного сервиса
- `DB_DSN`: DSN PostgreSQL для `office` или `logistic`
- `JWT_SECRET`: секрет для подписи JWT
- `OFFICE_INTERNAL_URL`: внутренний URL `office` для `logistic`
- `PLANNER_INTERVAL`: период планировщика (сек)
- `FE_DIR`: путь к статическим файлам фронтенда (по умолчанию `./FE`)

## Эндпойнты (через прокси `:8080`)
- `/auth/register`, `/auth/login`
- `/office/applications`, `/office/applications/{id}`, `/office/applications/{id}/accept`, `/office/applications/{id}/deliver`
- `/logistic/points`, `/logistic/shipments`, `/logistic/shipments/{id}`, `/logistic/shipments/{id}/send`, `/logistic/assignments`
- `/logistic/status/applications?ids=1,2,3`

Роли:
- `office_admin` — управление офисом, маршруты
- `logpoint_admin` — просмотр статусов заявок логточкой
