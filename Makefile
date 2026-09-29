.PHONY: help run migrate test swagger tidy

help:
	@echo "Available commands:"
	@echo "  make run        Start the API server"
	@echo "  make migrate    Run the PostgreSQL migrations"
	@echo "  make test       Run the Go test suite"
	@echo "  make swagger    Generate Swagger documentation"
	@echo "  make tidy       Download and tidy Go dependencies"

run:
	go run ./cmd/server

migrate:
	go run ./cmd/migrate

test:
	go test ./...

swagger:
	@which swag >/dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@v1.8.1
	$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs/swagger

tidy:
	go mod tidy
