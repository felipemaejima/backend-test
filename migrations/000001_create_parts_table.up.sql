CREATE TABLE IF NOT EXISTS parts (
    id                  UUID PRIMARY KEY,
    name                VARCHAR(120) NOT NULL,
    category            VARCHAR(60)  NOT NULL,
    -- current_stock aceita valores negativos: representa estoque já comprometido
    -- por vendas em aberto, e é um cenário legítimo do domínio.
    current_stock       INTEGER      NOT NULL,
    minimum_stock       INTEGER      NOT NULL CHECK (minimum_stock >= 0),
    average_daily_sales NUMERIC(12,4) NOT NULL CHECK (average_daily_sales >= 0),
    lead_time_days      INTEGER      NOT NULL CHECK (lead_time_days >= 0),
    unit_cost           NUMERIC(12,2) NOT NULL CHECK (unit_cost >= 0),
    criticality_level   SMALLINT     NOT NULL CHECK (criticality_level BETWEEN 1 AND 5),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_parts_category ON parts (category);

-- Ordenação estável na listagem paginada de milhares de peças.
CREATE INDEX IF NOT EXISTS idx_parts_name_id ON parts (name, id);
