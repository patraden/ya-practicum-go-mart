VETTOOL ?= $(shell which statictest)
DOCKER_COMPOSE_PATH := ./deployments/docker-compose.yml
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= postgres
POSTGRES_DB ?= praktikum
DATABASE_DSN ?= postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable

.PHONY: docker-up 
docker-up:
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker-down
docker-down:
	docker-compose -f $(DOCKER_COMPOSE_PATH) down

.PHONY: docker-rebuild
docker-rebuild: docker-down
	docker-compose -f $(DOCKER_COMPOSE_PATH) build --no-cache
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: test
test:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

.PHONY: clean
clean:
	@rm -f ./cmd/gothermart/gothermart
	@rm -f ./coverage.out

.PHONY: goose-status
goose-status:
	@goose -dir db/migrations postgres ${DATABASE_DSN} status

.PHONY: goose-up
goose-up:
	@goose -dir db/migrations postgres ${DATABASE_DSN} up

.PHONY: goose-reset
goose-reset:
	@goose -dir db/migrations postgres ${DATABASE_DSN} reset

.PHONY: sqlc
sqlc:
	@sqlc generate

.PHONY: code
code:
	@easyjson -all internal/app/dto/user_credentials.go
	@mockgen -source=internal/app/repository/user_repository.go -destination=internal/app/mock/user_repository.go -package=mock UserRepository
	@mockgen -source=internal/app/usecase/usecase.go -destination=internal/app/mock/usecase.go -package=mock IUserUseCase

.PHONY: lint
lint:
	@goimports -e -w -local "github.com/patraden/ya-practicum-go-mart" .
	@gofumpt -w ./cmd/gophermart ./internal/app
	@go vet -vettool=$(VETTOOL) ./...
	@golangci-lint run ./...
