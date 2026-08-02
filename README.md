# 🚚 Motor de Priorização de Reposição de Estoque

Microserviço em Go para gestão de peças de autopeças, com cálculo automático de
prioridade de reposição.

O enunciado original do desafio está em [DESAFIO.md](DESAFIO.md), e as decisões
tomadas durante a construção em [ROADMAP.md](ROADMAP.md).

---

## Stack

| Camada | Escolha |
|---|---|
| Linguagem | Go (toolchain 1.26 no container, módulo em 1.25) |
| HTTP | [Fiber v2](https://github.com/gofiber/fiber) |
| Persistência | PostgreSQL 16 + [GORM](https://gorm.io) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Desenvolvimento | Docker Compose + [air](https://github.com/air-verse/air) (hot reload) |

Não há dependência de nenhuma API externa. **Go não precisa estar instalado na
máquina** — tudo roda em container.

---

## Rodando localmente

Pré-requisitos: Docker e Docker Compose.

```bash
git clone https://github.com/felipemaejima/backend-test.git
cd backend-test
make up
```

Isso sobe o Postgres, espera ele ficar saudável, aplica as migrations e só então
inicia a API — com hot reload ativo. A API fica em `http://localhost:8080`.

```bash
curl http://localhost:8080/health
# {"database":"up","status":"ok"}
```

Nenhum arquivo de configuração é necessário: o `docker-compose.yml` traz defaults
para tudo. Para sobrescrever, copie `.env.example` para `.env`.

| Variável | Default | Descrição |
|---|---|---|
| `APP_EXTERNAL_PORT` | `8080` | Porta da API no host |
| `DB_EXTERNAL_PORT` | `5432` | Porta do Postgres no host |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `restock` | Credenciais do banco |

### Comandos

`make help` lista tudo. Os principais:

```bash
make up          # sobe o ambiente completo em background
make logs        # acompanha os logs da API
make down        # derruba tudo, preservando os dados
make down-v      # derruba tudo e apaga o volume do banco

make test        # testes unitários, com -race
make test-api    # requisições reais contra a API rodando
make ci          # as mesmas verificações do CI (fmt + vet + test)
make cover       # cobertura por pacote

make psql        # abre o psql no banco
make build       # compila a imagem de produção (25 MB, non-root)
```

---

## API

Todas as respostas são JSON. Campos usam `camelCase`.

| Método | Rota | Sucesso |
|---|---|---|
| `GET` | `/health` | `200` / `503` se o banco não responder |
| `POST` | `/parts` | `201` |
| `GET` | `/parts?category=&limit=&offset=` | `200` |
| `GET` | `/parts/:id` | `200` |
| `PUT` | `/parts/:id` | `200` |
| `DELETE` | `/parts/:id` | `204` |
| `GET` | `/restock/priorities` | `200` |

### Criar peça

```bash
curl -X POST http://localhost:8080/parts \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Filtro de Óleo X",
    "category": "engine",
    "currentStock": 15,
    "minimumStock": 20,
    "averageDailySales": 4,
    "leadTimeDays": 5,
    "unitCost": 18.50,
    "criticalityLevel": 3
  }'
```

```json
{
  "id": "8f14e45f-ceea-467a-9f8a-3b2c1d4e5f60",
  "name": "Filtro de Óleo X",
  "category": "engine",
  "currentStock": 15,
  "minimumStock": 20,
  "averageDailySales": 4,
  "leadTimeDays": 5,
  "unitCost": 18.5,
  "criticalityLevel": 3,
  "createdAt": "2026-08-01T18:24:05.123456Z",
  "updatedAt": "2026-08-01T18:24:05.123456Z"
}
```

`name` é normalizado (espaços nas pontas removidos) e `category` é sempre
convertida para minúsculas, na escrita e na busca.

### Listar peças

```bash
curl 'http://localhost:8080/parts'
curl 'http://localhost:8080/parts?category=engine'
curl 'http://localhost:8080/parts?limit=10&offset=20'
```

```json
{ "parts": [ { "id": "...", "name": "Filtro de Óleo X", "...": "..." } ] }
```

Ordenação por `(name, id)`, estável entre páginas. `limit` tem default **50** e
teto **500**; valores inválidos ou ausentes caem no default em vez de erro.
Coleção vazia devolve `[]`, nunca `null`.

### Buscar, atualizar e remover

```bash
curl http://localhost:8080/parts/{id}

# PUT tem semântica de substituição total — envie o recurso inteiro
curl -X PUT http://localhost:8080/parts/{id} \
  -H 'Content-Type: application/json' \
  -d '{ "name": "Filtro de Óleo X", "category": "engine", "currentStock": 2,
        "minimumStock": 20, "averageDailySales": 4, "leadTimeDays": 5,
        "unitCost": 18.50, "criticalityLevel": 3 }'

curl -X DELETE http://localhost:8080/parts/{id}
```

### Prioridades de reposição

```bash
curl http://localhost:8080/restock/priorities
```

```json
{
  "priorities": [
    {
      "partId": "8f14e45f-ceea-467a-9f8a-3b2c1d4e5f60",
      "name": "Filtro de Óleo X",
      "currentStock": 15,
      "projectedStock": 5,
      "minimumStock": 20,
      "urgencyScore": 45
    },
    {
      "partId": "3c9a7b21-8d4e-4f10-a5b6-7c8d9e0f1a2b",
      "name": "Pastilha de Freio Y",
      "currentStock": 8,
      "projectedStock": -2,
      "minimumStock": 10,
      "urgencyScore": 36
    }
  ]
}
```

Retorna **apenas as peças que precisam de reposição**, ordenadas da mais urgente
para a menos.

### Erros

| Status | Quando |
|---|---|
| `400` | JSON malformado, ou UUID inválido na rota |
| `404` | Peça inexistente |
| `422` | JSON válido mas violando regra de negócio |
| `500` | Falha interna — o detalhe vai para o log, nunca para a resposta |

Um `422` devolve **todas** as violações de uma vez, não uma por requisição:

```json
{
  "error": "dados inválidos",
  "fields": [
    { "field": "name", "message": "é obrigatório" },
    { "field": "criticalityLevel", "message": "deve estar entre 1 e 5" }
  ]
}
```

---

## Regras de negócio

```
expectedConsumption = averageDailySales × leadTimeDays
projectedStock      = currentStock − expectedConsumption
precisa de reposição quando  projectedStock < minimumStock
urgencyScore        = (minimumStock − projectedStock) × criticalityLevel
```

Empates no `urgencyScore` são resolvidos, nesta ordem, por maior
`criticalityLevel`, maior `averageDailySales` e ordem alfabética do nome.

Todo esse cálculo vive em `internal/domain/restock.go`, como uma **função pura**
sobre um slice de entidades — sem I/O, sem relógio, sem `context`, sem qualquer
referência a HTTP ou banco:

```go
func CalculateRestockPriorities(parts []Part) []RestockPriority
```

### Decisões de domínio

**`currentStock` aceita valores negativos.** É o único campo numérico sem piso.
Estoque negativo representa peças já comprometidas com vendas em aberto. Os demais campos recusam negativos, e
`criticalityLevel` fica restrito a 1–5.

**`projectedStock == minimumStock` não entra na fila.** A regra é `<` estrito.

**Os desempates comparam `urgencyScore` com tolerância de `1e-9`.** Como
`averageDailySales` é ponto flutuante, dois scores matematicamente iguais podem
diferir no último bit — sem a tolerância, o critério de desempate nunca seria
alcançado nesses casos. Os valores calculados também são arredondados em 4 casas
na resposta, para o ruído não vazar no JSON.

---

## Arquitetura

Camadas com dependências apontando sempre para dentro. O domínio não importa
nada do projeto — nem Fiber, nem GORM.

```
cmd/api/main.go              composition root: monta e pluga tudo

internal/
  domain/                    ← não importa nada do projeto
    part.go                  entidade Part, invariantes, normalização
    restock.go               cálculo de prioridade (função pura)
    repository.go            interface PartRepository (a porta) + PartFilter
    errors.go                ErrPartNotFound, ValidationError
  service/                   ← importa domain
    part_service.go          orquestra o CRUD
    restock_service.go       busca as peças e delega o cálculo ao domínio
  repository/                ← importa domain (implementa a porta)
    postgres/                único lugar que conhece GORM
    memory/                  map + RWMutex, usado nos testes
    repositorytest/          bateria de contrato compartilhada
  http/                      ← importa service; único lugar que conhece Fiber
    router.go                middlewares e rotas
    part_handler.go          decodifica → chama service → serializa
    restock_handler.go
    health_handler.go
    dto.go                   contrato JSON
    errors.go                erro de domínio → status HTTP
  config/config.go           configuração via ambiente

api-tests/main.go            cliente HTTP que exercita a API rodando
migrations/                  SQL versionado
```

```
http ──▶ service ──▶ domain ◀── repository
                                    ▲
                          main ─────┘
```

Trocar de banco é fornecer outra implementação de `domain.PartRepository`. Em
`main.go` isso é uma linha:

```go
partRepository := postgres.NewPartRepository(db)   // ← trocar aqui, e só aqui
```

Essa não é uma afirmação de brochura: existe uma bateria de contrato em
`internal/repository/repositorytest` que roda **a mesma suíte de casos contra as
duas implementações**. Se as duas passam, elas são intercambiáveis por
construção.

O modelo de persistência (`partModel`, com as tags do GORM) é separado da
entidade de domínio, com conversão explícita nos dois sentidos.

---

## Testes

```bash
make test       # unitários, com detector de race
make test-api   # requisições reais (exige `make up` antes)
make cover      # cobertura por pacote
```

São ~60 testes em três níveis:

**Unitários** — domínio, service, repositório in-memory e camada HTTP via
`app.Test()` do Fiber. Não tocam rede nem banco. Cobrem os três cenários extremos
que o desafio pede (estoque negativo, venda zero, lead time alto), os dois
exemplos numéricos do enunciado conferidos passo a passo, os três critérios de
desempate, acesso concorrente ao repositório in-memory, e o caminho de erro
interno — incluindo um teste que falha se qualquer detalhe de infraestrutura
vazar no corpo da resposta.

**Contrato de repositório** — uma bateria de casos que qualquer implementação de
`PartRepository` precisa satisfazer, hoje executada contra a implementação
in-memory. O adapter Postgres é validado de ponta a ponta pelo `api-tests` no CI,
não por esta bateria.

**API** — `api-tests/` é um programa Go independente que fala HTTP com o serviço
rodando, sem conhecer nada do código interno. Cada execução usa uma categoria
própria (`smoke-HHMMSS`), então rodar várias vezes seguidas não interfere nos
dados existentes; as peças criadas são removidas ao final.

**Benchmark** do cálculo de prioridade com 100, 1.000 e 10.000 peças:

```bash
docker compose run --rm --no-deps test \
  go test ./internal/domain -bench=. -benchmem -run='^$'
```

---

## CI

`.github/workflows/ci.yml` roda em push na `main` e em pull request, com três
jobs paralelos:

- **test** — `gofmt`, verificação de que `go.mod`/`go.sum` estão em dia, `go vet`,
  build e testes com `-race` e cobertura
- **api** — sobe um Postgres real, aplica as migrations, sobe a API e roda o
  `api-tests` contra ela
- **image** — build da imagem de produção

O job `api` é o que dá cobertura automática ao adapter Postgres, ao SQL e às
constraints das migrations.

---

## Limitações conhecidas

Coisas que ficaram fora do escopo e que eu trataria antes de colocar em produção:

- **`PUT` não tem controle de concorrência.** É um read-modify-write em três
  passos; duas atualizações simultâneas na mesma peça resultam em *lost update*.
  A correção seria uma coluna `version` com bloqueio otimista e `409` no conflito.
- **`GET /restock/priorities` não tem teto.** Carrega todas as peças e serializa
  todas as que precisam de reposição. Com dezenas de milhares em déficit, a
  resposta cresce sem limite.
- **`GET /parts` não devolve o total de registros**, então o cliente pagina sem
  saber quando acabou.
- **`unitCost` é `float64`.** Não entra em nenhum cálculo hoje, mas dinheiro em
  ponto flutuante pede inteiro em centavos ou tipo decimal.
- **Sem seed e sem log de requisição** — ambos estão no TODO do
  [ROADMAP.md](ROADMAP.md).
- **O pacote `postgres` não tem testes em `go test`.** É coberto pelo `api-tests`,
  que roda no CI contra um banco real, mas não por testes de integração locais.
