# 🚚 Desafio Backend – Motor de Priorização de Reposição de Estoque

## 🧩 Contexto

Somos um distribuidor de autopeças. Diariamente precisamos decidir **quais peças devem ser priorizadas para reposição**, considerando:

- Estoque limitado
- Capital de giro limitado
- Diferentes níveis de criticidade
- Padrões de venda distintos
- Tempo de reposição do fornecedor

O objetivo é construir um microserviço capaz de:

1. Gerenciar peças em estoque
2. Calcular automaticamente quais peças devem ser priorizadas para reposição
3. Ordenar as peças por nível de urgência

---

# 🛠️ Requisitos Funcionais

## 1️⃣ CRUD de Peças

Criar uma API para:

- Criar peça
- Listar peças
- Atualizar peça
- Remover peça
- Buscar por categoria (opcional)

### 📦 Estrutura da Entidade

```json
{
  "id": "uuid",
  "name": "Filtro de Óleo X",
  "category": "engine",
  "currentStock": 15,
  "minimumStock": 20,
  "averageDailySales": 4,
  "leadTimeDays": 5,
  "unitCost": 18.50,
  "criticalityLevel": 3
}
```

## 📝 Descrição dos Campos

| Campo | Descrição |
|--------|------------|
| `currentStock` | Estoque atual disponível |
| `minimumStock` | Estoque mínimo desejado |
| `averageDailySales` | Média de vendas por dia |
| `leadTimeDays` | Tempo (em dias) que o fornecedor demora para entregar a peça |
| `unitCost` | Custo unitário da peça |
| `criticalityLevel` | Nível de criticidade (1 a 5) |

---

## 🧠 Endpoint de Priorização

Criar o endpoint:

```GET /restock/priorities```

Esse endpoint deve retornar as peças ordenadas por prioridade de reposição.

---

## 📐 Regras de Negócio

### 1️⃣ Calcular Consumo Esperado Durante o Lead Time

```expectedConsumption = averageDailySales * leadTimeDays```

---

### 2️⃣ Calcular Estoque Projetado

```projectedStock = currentStock - expectedConsumption```

---

### 3️⃣ Identificar Necessidade de Reposição

Uma peça precisa de reposição quando:
```projectedStock < minimumStock```


---

### 4️⃣ Calcular Score de Prioridade

O score de prioridade deve ser calculado da seguinte forma:

```urgencyScore = (minimumStock - projectedStock) * criticalityLevel```


Quanto maior o `urgencyScore`, maior a prioridade de reposição.

---

## 🟰 Critérios de Desempate

Em caso de empate no `urgencyScore`, aplicar:

1. Maior `criticalityLevel`
2. Maior `averageDailySales`
3. Ordem alfabética pelo nome da peça

---

## 📤 Exemplo de Resposta

```json
{
  "priorities": [
    {
      "partId": "uuid-1",
      "name": "Filtro de Óleo X",
      "currentStock": 15,
      "projectedStock": 5,
      "minimumStock": 20,
      "urgencyScore": 45
    },
    {
      "partId": "uuid-2",
      "name": "Pastilha de Freio Y",
      "currentStock": 8,
      "projectedStock": -2,
      "minimumStock": 10,
      "urgencyScore": 36
    }
  ]
}
```

### 📌 Regras Gerais

- Não utilizar APIs externas.
- O sistema deve estar preparado para suportar centenas ou milhares de peças.
- A solução deve permitir futura troca de banco de dados.
- O cálculo de prioridade deve estar isolado da camada HTTP.
- Tratar corretamente casos de estoque negativo.

### 🎯 O Que Será Avaliado
- 🧠 Modelagem de Domínio
- Clareza das entidades
- Separação de responsabilidades
- Organização das regras de negócio

### 🧪 Testes
- Testes unitários do cálculo de prioridade
- Testes de cenários extremos (estoque negativo, venda zero, lead time alto)

### 🏗️ Arquitetura
- Uso adequado de camadas (ex: Controller, Service, Domain, Repository)
- Código limpo e organizado
- Facilidade de manutenção

### 🧰 Tecnologias

Pode ser desenvolvido utilizando:

- Node.js (com TypeScript)
- Golang
- Frameworks e bibliotecas são livres

### 📄 Entrega

O projeto deve conter:

- Código-fonte organizado
- README com instruções para rodar localmente
- Exemplos de requisição
- Testes automatizados

Boa implementação 🚀
