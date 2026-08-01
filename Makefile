COMPOSE := docker compose

.PHONY: help up down down-v logs restart test test-api cover ci fmt-check migrate-up migrate-down migrate-create tidy fmt vet build sh psql

help: ## Lista os comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

up: ## Sobe Postgres + migrations + API (hot reload) em background
	$(COMPOSE) up -d --build api
	@echo "API em http://localhost:$${APP_EXTERNAL_PORT:-8080}"

down: ## Derruba tudo, preservando o volume do banco
	$(COMPOSE) down

down-v: ## Derruba tudo e apaga o volume do banco
	$(COMPOSE) down -v

logs: ## Acompanha os logs da API
	$(COMPOSE) logs -f api

restart: ## Reinicia a API
	$(COMPOSE) restart api

test: ## Testes unitários (sem banco, com -race)
	$(COMPOSE) run --rm --no-deps test

test-api: ## Requisições reais contra a API rodando (exige `make up` antes)
	$(COMPOSE) run --rm --no-deps -e BASE_URL=$${BASE_URL:-http://api:8080} test go run ./api-tests

ci: ## Roda localmente as mesmas verificações do CI
	@$(MAKE) fmt-check
	@$(MAKE) vet
	@$(MAKE) test

fmt-check: ## Falha se algum arquivo estiver fora do padrão do gofmt
	@out="$$($(COMPOSE) run --rm --no-deps -T test gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "Arquivos fora do padrão do gofmt:"; echo "$$out"; exit 1; \
	fi
	@echo "gofmt OK"

cover: ## Roda os testes com relatório de cobertura por pacote
	$(COMPOSE) run --rm --no-deps test go test ./... -coverprofile=coverage.out -covermode=atomic
	$(COMPOSE) run --rm --no-deps test go tool cover -func=coverage.out

migrate-up: ## Aplica as migrations pendentes
	$(COMPOSE) run --rm migrate

migrate-down: ## Reverte a última migration
	$(COMPOSE) run --rm migrate -path=/migrations -database="postgres://$${DB_USER:-restock}:$${DB_PASSWORD:-restock}@postgres:5432/$${DB_NAME:-restock}?sslmode=disable" down 1

migrate-create: ## Cria um par de migrations: make migrate-create name=add_x
	$(COMPOSE) run --rm --entrypoint migrate migrate create -ext sql -dir /migrations -seq $(name)

tidy: ## Sincroniza go.mod/go.sum
	$(COMPOSE) run --rm --no-deps test go mod tidy

fmt: ## Formata o código
	$(COMPOSE) run --rm --no-deps test go fmt ./...

vet: ## Análise estática do toolchain
	$(COMPOSE) run --rm --no-deps test go vet ./...

build: ## Compila a imagem de produção
	docker build --target runtime -t restock-api:latest .

sh: ## Abre um shell no container da API
	$(COMPOSE) run --rm --no-deps test sh

psql: ## Abre o psql no banco
	$(COMPOSE) exec postgres psql -U $${DB_USER:-restock} -d $${DB_NAME:-restock}
