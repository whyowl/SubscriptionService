BINARY_DIR=bin
POSTGRES_URL=postgres://user:password@localhost:5432/projectdb?sslmode=disable

docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose -f ./docker-compose.yml up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

# Database targets
goose-install:
	@echo "Installing goose migration tool..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest

migrate:
	@echo "Running database migrations..."
	@goose -dir ./migrations postgres "$(POSTGRES_URL)" up

migrate-down:
	@echo "Rolling back database migrations..."
	@goose -dir ./migrations postgres "$(POSTGRES_URL)" down

migrate-status:
	@echo "Migration status:"
	@goose -dir ./migrations postgres "$(POSTGRES_URL)" status

migrate-create:
	@echo "Creating new migration file..."
	@goose -dir ./migrations postgres "$(POSTGRES_URL)" create rename_me sql

# Development targets

test: ## Run tests
	@echo "Running tests..."
	@go test ./... -coverprofile=cover.out
	@go tool cover -func=cover.out


docker-clean: ## Clean Docker
	@echo "Cleaning..."
	@docker-compose down -v --remove-orphans
	@docker system prune -f
	@echo "Cleaned!"
