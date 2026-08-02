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

**Único pré-requisito: Docker com Docker Compose v2.** Go, Postgres e as
ferramentas de migration rodam todos dentro de containers — nada precisa ser
instalado na máquina.

O `make` é uma conveniência, não uma exigência: cada alvo é um atalho para um
comando `docker compose`, e a [tabela de equivalências](#sem-make-windows-ou-qualquer-shell)
mais abaixo traz todos eles. Linux e macOS já vêm com `make`; no Windows use
WSL2, Git Bash, ou os comandos diretos.

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

### Populando o banco

```bash
make seed
curl 'http://localhost:8080/restock/priorities?n=10'
```

O seed cria **1.012 peças** em 7 categorias: 1.000 geradas a partir de 25
componentes × 40 modelos de veículo, mais 12 escritas à mão que garantem a
presença dos casos de borda do cálculo — estoque negativo, venda diária zero,
lead time de 90 dias e peças saudáveis que ficam de fora da fila.

É **determinístico e idempotente**: a geração usa uma semente fixa e o ID de cada
peça é derivado do nome, então o mesmo catálogo sai em qualquer máquina e rodar
de novo não duplica nada. Também é o volume que dá sentido à paginação e
sustenta o requisito de "centenas ou milhares de peças".

### Configuração

Nenhum arquivo é necessário: o `docker-compose.yml` traz defaults para tudo.
Para sobrescrever, copie `.env.example` para `.env` e rode `make up` de novo
(recriar o container é o que faz ele reler as variáveis).

| Variável | Default | Descrição |
|---|---|---|
| `APP_EXTERNAL_PORT` | `8080` | Porta da API no host |
| `DB_EXTERNAL_PORT` | `5432` | Porta do Postgres no host |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `restock` | Credenciais do banco |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` ou `error` |
| `LOG_FORMAT` | `json` | `json` ou `text` |
| `LOG_FILE` | vazio | Caminho para espelhar os logs em arquivo |

Valores inválidos falham no boot com mensagem nomeando a variável, em vez de
cair em default silencioso.

Se a porta 8080 já estiver ocupada: `APP_EXTERNAL_PORT=8081 make up`.

### Comandos

`make help` lista todos. Os principais:

```bash
make up          # sobe o ambiente completo em background
make seed        # popula o banco com o catálogo de exemplo
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

### Sem `make` (Windows ou qualquer shell)

Todo alvo é um atalho. Os equivalentes diretos:

| Alvo | Comando |
|---|---|
| `make up` | `docker compose up -d --build api` |
| `make seed` | `docker compose run --rm seed` |
| `make logs` | `docker compose logs -f api` |
| `make down` | `docker compose down` |
| `make down-v` | `docker compose down -v` |
| `make test` | `docker compose run --rm --no-deps test` |
| `make test-api` | `docker compose run --rm --no-deps -e BASE_URL=http://api:8080 test go run ./api-tests` |
| `make cover` | `docker compose run --rm --no-deps test go test ./... -coverprofile=coverage.out` |
| `make migrate-up` | `docker compose run --rm migrate` |
| `make psql` | `docker compose exec postgres psql -U restock -d restock` |
| `make build` | `docker build --target runtime -t restock-api:latest .` |

Em PowerShell as variáveis de ambiente mudam de sintaxe: use
`$env:APP_EXTERNAL_PORT="8081"` antes do `docker compose up`.

---

## API

Todas as respostas são JSON. Campos usam `camelCase`.

| Método | Rota | Sucesso |
|---|---|---|
| `GET` | `/health` | `200` / `503` se o banco não responder |
| `POST` | `/parts` | `201` |
| `GET` | `/parts?category=&page=&n=` | `200` |
| `GET` | `/parts/:id` | `200` |
| `PUT` | `/parts/:id` | `200` |
| `DELETE` | `/parts/:id` | `204` |
| `GET` | `/restock/priorities?page=&n=` | `200` |

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
curl 'http://localhost:8080/parts?page=3&n=10'
```

```json
{
  "parts": [ { "id": "...", "name": "Filtro de Óleo X", "...": "..." } ],
  "pagination": {
    "page": 3,
    "perPage": 10,
    "total": 137,
    "totalPages": 14,
    "hasNext": true,
    "hasPrevious": true
  }
}
```

Ordenação por `(name, id)`, estável entre páginas. Coleção vazia devolve `[]`,
nunca `null`.

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
para a menos, com o mesmo objeto `pagination` da listagem de peças.

A paginação aqui acontece **depois** do cálculo: a ordenação é por
`urgencyScore`, que só existe depois de percorrer todas as peças. Então
`total` é o tamanho da fila inteira, e a página 1 traz sempre as mais urgentes
de toda a base — não de um recorte dela.

Consequência que vale saber: `n` limita a **resposta**, não a memória. O
endpoint carrega todas as peças para calcular, independente do tamanho de
página pedido. Em `/parts` o teto de 500 limita a query e a resposta; aqui,
só a resposta.

### Erros

| Status | Quando |
|---|---|
| `400` | JSON malformado, ou UUID inválido na rota |
| `404` | Peça inexistente |
| `422` | JSON válido mas violando regra de negócio |
| `500` | Falha interna — o detalhe vai para o log, nunca para a resposta |
| `503` | Prazo da requisição excedido (10s), ou banco fora no `/health` |

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

## Logs

Log estruturado em JSON via `log/slog`, escrito em **stdout** — o destino que
`docker compose logs`, Loki, CloudWatch e afins consomem sem configuração extra.

Cada requisição gera exatamente uma linha, correlacionável pelo `request_id` —
um UUID v4 que também volta no header `X-Request-ID` da resposta. Se a requisição
já chegar com esse header (vindo de um gateway ou proxy), o valor é preservado em
vez de sobrescrito, mantendo a correlação de ponta a ponta.

```json
{
  "time": "2026-08-02T14:03:11.482Z",
  "level": "INFO",
  "msg": "requisição",
  "request_id": "5f2c8a1e-9b3d-4c7a-8e01-2d4f6a8b0c3e",
  "method": "POST",
  "path": "/parts",
  "status": 201,
  "duration_ms": 4.117,
  "ip": "172.18.0.1"
}
```

O nível acompanha o resultado: `5xx` vira **ERROR**, `4xx` vira **WARN**, o resto
**INFO**, e `/health` vira **DEBUG** para não poluir o output com healthcheck.
Quando há erro, o detalhe interno entra no campo `error` da linha de log — e só
lá; a resposta HTTP nunca o expõe.

### Lendo os logs

```bash
make logs                                          # follow, com prefixo do container
docker compose logs --tail 100 api                 # últimas 100 linhas
docker compose logs --no-log-prefix -f api | jq    # JSON formatado
```

O `--no-log-prefix` é o que remove o `restock-api  | ` de cada linha — sem ele o
`jq` não consegue parsear. A partir daí dá para filtrar:

```bash
# só as requisições, uma linha por chamada
docker compose logs --no-log-prefix api | jq -r \
  'select(.msg == "requisição") | [.status, .method, .path, .duration_ms] | @tsv'

# só o que falhou
docker compose logs --no-log-prefix api | jq 'select(.level == "ERROR")'

# tudo de uma requisição específica
docker compose logs --no-log-prefix api | jq 'select(.request_id == "COLE-O-ID")'

# requisições acima de 100 ms
docker compose logs --no-log-prefix api | jq 'select(.duration_ms > 100)'
```

Para desenvolvimento, `LOG_FORMAT=text` no `.env` deixa a saída legível sem
`jq`. E `LOG_LEVEL=debug` passa a mostrar as chamadas ao `/health`, ocultas por
padrão.

| Variável | Default | Valores |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json`, `text` (mais legível no desenvolvimento) |
| `LOG_FILE` | vazio | caminho para espelhar em arquivo |

### Por que não gravar em arquivo por padrão

Filesystem de container é efêmero, arquivo em disco exige rotação própria, e
múltiplas réplicas gravando no mesmo volume embaralham a saída. A rotação do
stdout fica a cargo do driver do Docker, já configurado no `docker-compose.yml`
em 3 arquivos de 10 MB.

Se você quiser os logs em disco mesmo assim, `LOG_FILE` liga o espelhamento — a
saída vai para stdout **e** para o arquivo:

```bash
echo 'LOG_FILE=/app/logs/api.log' >> .env
make up          # recria o container para pegar a nova variável
tail -f logs/api.log
```

> `make restart` **não** serve aqui: ele reinicia o container existente sem
> reler o `.env`. Mudança de variável de ambiente exige recriar, que é o que
> `make up` faz.

Como o projeto já é montado em `/app` dentro do container, o arquivo aparece em
`./logs` no host (diretório ignorado pelo Git). Vale lembrar que esse arquivo
**não tem rotação** — é conveniência de desenvolvimento, não configuração de
produção.

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

São 88 funções de teste em três níveis, mais um benchmark:

**Unitários** — domínio, service, repositório in-memory e camada HTTP via
`app.Test()` do Fiber. Não tocam rede nem banco. Cobrem os três cenários extremos
que o desafio pede (estoque negativo, venda zero, lead time alto), os dois
exemplos numéricos do enunciado conferidos passo a passo, os três critérios de
desempate, acesso concorrente ao repositório in-memory, e o caminho de erro
interno — incluindo um teste que falha se qualquer detalhe de infraestrutura
vazar no corpo da resposta.

Também cobrem robustez de entrada (tipos inválidos, corpo vazio, `Content-Type`
ausente, campos desconhecidos) e **integridade da paginação**: percorrer todas as
páginas e verificar que cada peça aparece exatamente uma vez, inclusive no
cenário com nomes duplicados, que é o que exercita o desempate por `(name, id)`.

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
- **`/restock/priorities` carrega a tabela inteira em memória** para calcular,
  antes de paginar. O `n` limita a resposta, não o consumo de memória. Para as
  escalas do enunciado é irrelevante; acima disso exigiria cálculo incremental
  ou materialização do score em coluna.
- **`unitCost` é `float64`.** Não entra em nenhum cálculo hoje, mas dinheiro em
  ponto flutuante pede inteiro em centavos ou tipo decimal.
- **Paginação por `OFFSET`** degrada em páginas muito profundas, porque o banco
  varre e descarta as linhas anteriores. A saída seria paginação por cursor
  sobre `(name, id)`, que o índice existente já suportaria.
- **`LOG_FILE` não rotaciona.** O espelhamento em arquivo é conveniência de
  desenvolvimento; em produção o destino é stdout, com rotação pelo driver do
  Docker.
- **O pacote `postgres` não tem testes em `go test`.** É coberto pelo `api-tests`,
  que roda no CI contra um banco real, mas não por testes de integração locais.

---

## Uso de IA no desenvolvimento

Este projeto foi desenvolvido em par com **Claude (Claude Code)**, e faz sentido
ser explícito sobre como.

### O que foi delegado

Praticamente toda a digitação: configuração do ambiente Docker, código das
camadas, suíte de testes e esta documentação. Também usei a IA como
interlocutora nas decisões de arquitetura — pedindo o mapa de opções com prós e
contras antes de escolher.

### O que continuou comigo

**Todas as decisões.** A escolha de Go com Fiber e GORM, a estrutura em camadas
pragmáticas em vez de Clean Architecture completa, o modelo de persistência
separado da entidade, a ordem de construção (ambiente → CRUD → regra de negócio
→ testes), a decisão de trocar testes de integração por um cliente HTTP em Go, e
o que ficou de fora conscientemente (bloqueio otimista, tipo monetário decimal).

**Toda a verificação.** Rodei os testes, subi os containers e conferi as
respostas eu mesmo. A IA não executou nada neste repositório.

### O que a revisão pegou

Vale registrar, porque é a parte que mostra que o código foi lido e não apenas
aceito:

- Um `defer cleanup()` seguido de `os.Exit()` no `api-tests` — `os.Exit` não roda
  defers, então a limpeza nunca acontecia e cada execução deixava peças órfãs no
  banco. Descobri lendo o arquivo de log e notando a ausência dos `DELETE` finais.
- O `api-tests` deduzindo o total de registros contando os itens da página, o que
  funciona até o banco passar de 50 registros.
- Uma asserção lendo o campo `parts` numa resposta que devolve `priorities`.
- O README instruindo `make restart` onde era necessário `make up`, porque
  `docker compose restart` não relê o `.env`.
- Uma refatoração apresentada como otimização de performance que, ao ser
  questionada, se revelou irrelevante — foi revertida.
- Testes acoplados à implementação: usar as constantes do próprio código como
  valor esperado torna o teste incapaz de detectar mudança de comportamento.

### Leitura

A IA acelerou muito a parte mecânica e foi útil para explorar alternativas antes
de decidir. Mas os erros acima não eram sutis, e nenhum apareceu sozinho — todos
saíram de revisão, de ler log de execução real e de perguntar "por que isso está
assim?". O ganho é proporcional ao rigor da revisão, não à quantidade de código
gerado.
