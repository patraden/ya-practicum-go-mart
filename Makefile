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

.PHONY: docker-stop
docker-stop:
	docker-compose -f $(DOCKER_COMPOSE_PATH) stop

.PHONY: docker-rebuild
docker-rebuild: docker-down
	docker-compose -f $(DOCKER_COMPOSE_PATH) build --no-cache
	docker-compose -f $(DOCKER_COMPOSE_PATH) up -d

.PHONY: docker-build
docker-build: 
	@docker-compose -f $(DOCKER_COMPOSE_PATH) build gophermart

.PHONY: test
test:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

.PHONY: clean
clean:
	@rm -f ./cmd/gophermart/gophermart
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
	@easyjson -all -omit_empty internal/app/dto/order_status.go
	@easyjson -all internal/app/dto/withdrawal.go
	@easyjson -all internal/app/dto/user_balance.go
	@mockgen -source=internal/app/repository/user_repository.go -destination=internal/app/mock/user_repository.go -package=mock UserRepository
	@mockgen -source=internal/app/repository/order_repository.go -destination=internal/app/mock/order_repository.go -package=mock OrderRepository
	@mockgen -source=internal/app/repository/transactions_repository.go -destination=internal/app/mock/transactions_repository.go -package=mock OrderTransactionsRepository
	@mockgen -source=internal/app/usecase/usecase.go -destination=internal/app/mock/usecase.go -package=mock IUserUseCase,ITransactionsUseCase 
	@mockgen -source=internal/app/integration/accrual/client.go -destination=internal/app/mock/accrual_client.go -package=mock IClient

.PHONY: lint
lint:
	@goimports -e -w -local "github.com/patraden/ya-practicum-go-mart" .
	@gofumpt -w ./cmd/gophermart ./internal/app
	@gofumpt -w ./cmd/gophermart ./pkg
	@go vet -vettool=$(VETTOOL) ./...
	@golangci-lint run ./...


.PHONY: build
build: clean
	@go build -buildvcs=false -o cmd/gophermart/gophermart ./cmd/gophermart


.PHONY: gophermarttest
gophermarttest:
	@gophermarttest \
		-test.v -test.run=^TestGophermart$ \
		-gophermart-binary-path=cmd/gophermart/gophermart \
		-gophermart-host=localhost \
		-gophermart-port=8080 \
		-gophermart-database-uri="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" \
		-accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
		-accrual-host=localhost \
		-accrual-port=8081 \
		-accrual-database-uri="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
