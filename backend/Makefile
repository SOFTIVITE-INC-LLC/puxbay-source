.PHONY: run test test-coverage build

run:
	go run cmd/server/main.go

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

build:
	go build -o bin/server cmd/server/main.go

migrate-up:
	migrate -path db/migrations -database "postgres://postgres:gentechco@localhost:5432/puxbay?sslmode=disable" -verbose up

migrate-down:
	migrate -path db/migrations -database "postgres://postgres:gentechco@localhost:5432/puxbay?sslmode=disable" -verbose down -all
