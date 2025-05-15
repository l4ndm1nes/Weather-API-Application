.PHONY: migrate-script-ready

migrate-script-ready:
	chmod +x scripts/wait-for-postgres.sh

migrate-up:
	docker run --rm -v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations -database "postgres://postgres:postgres@localhost:5432/weather?sslmode=disable" up

migrate-down:
	docker run --rm -v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations -database "postgres://postgres:postgres@localhost:5432/weather?sslmode=disable" down


run:
	docker-compose up --build

build:
	go build -o weather-api-application ./cmd/httpserver

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test ./...
