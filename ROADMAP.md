# Roadmap de construção da aplicação

Abstração/entendimento da proposta, além de decisões de implementação previamente decididas.

## Pontos de atenção/contexto

 - Estoque limitado
 - Capital de giro limitado (Saldo monetário disponível e preço dos produtos)
 - niveis de criticidade (Importancia da peça, data de entrega da venda,...)
 - padrões de venda distintos (Frequencia?)
 - tempo de reposição do fornecedor
 - lidar com estoque negativo

 ## A aplicação

Microsserviço de gerenciamento de peças em estoque, com calculo de prioridade e ordenação de nivel de urgercia. Endpoints: 

 - Criar peça 
 ```POST /parts```
 - Listar peças (com filtro de categoria)
 ```GET /parts?category={:optional}```
 - Atualizar peça
 ```PUT /parts/:id```
 - Remover peça
 ```DELETE /parts/:id```
 - Listar as peças ordenadas ṕor prioridade de reposição
```GET /restock/priorities```


### Estrutura da Entidade

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
| Campo | Descrição |
|--------|------------|
| `currentStock` | Estoque atual disponível |
| `minimumStock` | Estoque mínimo desejado |
| `averageDailySales` | Média de vendas por dia |
| `leadTimeDays` | Tempo (em dias) que o fornecedor demora para entregar a peça |
| `unitCost` | Custo unitário da peça |
| `criticalityLevel` | Nível de criticidade (1 a 5) |

---

## Regras de Negócio

### Peças que precisam de reposição

```expectedConsumption = averageDailySales * leadTimeDays``` 
```projectedStock = currentStock - expectedConsumption```

Uma peça precisa de reposição quando:
```projectedStock < minimumStock```

### Calcular Score de Prioridade

O score de prioridade deve ser calculado da seguinte forma:

```urgencyScore = (minimumStock - projectedStock) * criticalityLevel```

Quanto maior o `urgencyScore`, maior a prioridade de reposição.

#### Critérios de Desempate

Em caso de empate no `urgencyScore`, aplicar:

1. Maior `criticalityLevel`
2. Maior `averageDailySales`
3. Ordem alfabética pelo nome da peça

---


## Exemplo de Resposta

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

---

### Testes
- Testes unitários do cálculo de prioridade
- Testes de cenários extremos (estoque negativo, venda zero, lead time alto)

# TODO

 - Montar o ambiente em docker 
 - Definir a estrutura da aplicação (go)
 - migrations
 - documentação
 - logs
 - testes
 - seed 
 - Detalhar o uso de IA
 - Subir a aplicação para exemplificar a usabilidade (se sobrar tempo)