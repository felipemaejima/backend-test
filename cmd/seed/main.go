package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/google/uuid"

	"github.com/felipemaejima/backend-test/internal/config"
	"github.com/felipemaejima/backend-test/internal/domain"
	"github.com/felipemaejima/backend-test/internal/logger"
	"github.com/felipemaejima/backend-test/internal/repository/postgres"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed falhou", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, closeLog, err := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})
	if err != nil {
		return err
	}
	defer func() { _ = closeLog() }()

	ctx := context.Background()

	db, err := postgres.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		return err
	}
	defer func() { _ = postgres.Close(db) }()

	repo := postgres.NewPartRepository(db)

	created, skipped := 0, 0
	for _, input := range catalog() {
		part, err := domain.NewPart(input)
		if err != nil {
			return err
		}
		// ID derivado do nome: rodar o seed de novo não duplica nada.
		part.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(part.Name))

		switch _, err := repo.FindByID(ctx, part.ID); {
		case err == nil:
			skipped++
			continue
		case !errors.Is(err, domain.ErrPartNotFound):
			return err
		}

		if err := repo.Create(ctx, part); err != nil {
			return err
		}
		created++
	}

	log.Info("seed concluído", "criadas", created, "já existentes", skipped)
	return nil
}

// catalog cobre de propósito os casos que o cálculo de reposição precisa
// mostrar: estoque negativo, venda zero, lead time longo e peças saudáveis.
func catalog() []domain.PartInput {
	return []domain.PartInput{
		{Name: "Vela de Ignição", Category: "engine", CurrentStock: -6, MinimumStock: 30, AverageDailySales: 6, LeadTimeDays: 7, UnitCost: 24.90, CriticalityLevel: 4},
		{Name: "Correia Dentada Z", Category: "engine", CurrentStock: 3, MinimumStock: 12, AverageDailySales: 1.5, LeadTimeDays: 20, UnitCost: 210.00, CriticalityLevel: 5},
		{Name: "Junta do Cabeçote", Category: "engine", CurrentStock: 0, MinimumStock: 5, AverageDailySales: 0.2, LeadTimeDays: 90, UnitCost: 260.00, CriticalityLevel: 5},
		{Name: "Bomba d'Água", Category: "engine", CurrentStock: 2, MinimumStock: 6, AverageDailySales: 0.3, LeadTimeDays: 45, UnitCost: 380.00, CriticalityLevel: 5},
		{Name: "Bateria 60Ah", Category: "electrical", CurrentStock: 12, MinimumStock: 10, AverageDailySales: 1.2, LeadTimeDays: 15, UnitCost: 450.00, CriticalityLevel: 5},
		{Name: "Filtro de Óleo X", Category: "filters", CurrentStock: 15, MinimumStock: 20, AverageDailySales: 4, LeadTimeDays: 5, UnitCost: 18.50, CriticalityLevel: 3},
		{Name: "Disco de Freio", Category: "brakes", CurrentStock: 5, MinimumStock: 8, AverageDailySales: 0.5, LeadTimeDays: 30, UnitCost: 175.00, CriticalityLevel: 4},
		{Name: "Pastilha de Freio Y", Category: "brakes", CurrentStock: 8, MinimumStock: 10, AverageDailySales: 2, LeadTimeDays: 5, UnitCost: 89.90, CriticalityLevel: 5},
		{Name: "Mola Helicoidal", Category: "suspension", CurrentStock: 4, MinimumStock: 10, AverageDailySales: 0, LeadTimeDays: 60, UnitCost: 145.00, CriticalityLevel: 3},
		{Name: "Amortecedor Dianteiro", Category: "suspension", CurrentStock: 40, MinimumStock: 15, AverageDailySales: 0.8, LeadTimeDays: 10, UnitCost: 320.00, CriticalityLevel: 4},
		{Name: "Filtro de Ar", Category: "filters", CurrentStock: 60, MinimumStock: 25, AverageDailySales: 3, LeadTimeDays: 4, UnitCost: 32.00, CriticalityLevel: 2},
		{Name: "Lâmpada H4", Category: "electrical", CurrentStock: 200, MinimumStock: 50, AverageDailySales: 0, LeadTimeDays: 3, UnitCost: 12.00, CriticalityLevel: 1},
	}
}
