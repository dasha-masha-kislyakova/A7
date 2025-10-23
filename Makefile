.PHONY: local down tidy run-all

local:
	docker compose up --build

down:
	docker compose down -v

tidy:
	cd ./ && go mod tidy

# локальный запуск всех бэкендов и статики фронта одной командой (без Docker)
run-all:
docker compose up -d office-db logistic-db

	./scripts/wait-pg.sh office-db office
	./scripts/wait-pg.sh logistic-db logistic

	SERVICE=all JWT_SECRET=dev-secret \
	OFFICE_DSN=postgres://postgres:postgres@localhost:5433/office?sslmode=disable \
	LOGISTIC_DSN=postgres://postgres:postgres@localhost:5434/logistic?sslmode=disable \
	OFFICE_INTERNAL_URL=http://localhost:8081 FE_DIR=./FE \
	go run ./