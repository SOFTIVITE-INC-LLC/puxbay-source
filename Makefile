.PHONY: test test-coverage build run

build:
	cd backend && go build -o bin/server cmd/server/main.go

run:
	cd backend && go run cmd/server/main.go

test:
	cd backend && go test ./... -v

test-coverage:
	cd backend && go test ./... -coverprofile=coverage.out
	cd backend && go tool cover -func=coverage.out
	@echo "Checking if coverage meets the 80% threshold..."
	@cd backend && go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}' | awk '{if ($$1 < 80.0) {print "Coverage failed! Expected >= 80%, got "$$1"%"; exit 1} else {print "Coverage passed with "$$1"%"}}'

migrate-up:
	migrate -path backend/db/migrations -database "$${DATABASE_URL}" -verbose up

migrate-down:
	migrate -path backend/db/migrations -database "$${DATABASE_URL}" -verbose down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir backend/db/migrations -seq $$name
