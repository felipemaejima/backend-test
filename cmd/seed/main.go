package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
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

	// Uma consulta em vez de uma por peça: com mil registros, checar existência
	// individualmente dobraria as idas ao banco.
	existing, err := repo.ListAll(ctx)
	if err != nil {
		return err
	}
	known := make(map[uuid.UUID]struct{}, len(existing))
	for _, part := range existing {
		known[part.ID] = struct{}{}
	}

	created, skipped := 0, 0
	for _, input := range catalog() {
		part, err := domain.NewPart(input)
		if err != nil {
			return fmt.Errorf("peça %q: %w", input.Name, err)
		}
		// ID derivado do nome: rodar o seed de novo não duplica nada.
		part.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(part.Name))

		if _, ok := known[part.ID]; ok {
			skipped++
			continue
		}
		if err := repo.Create(ctx, part); err != nil {
			return fmt.Errorf("peça %q: %w", part.Name, err)
		}
		created++
	}

	log.Info("seed concluído", "criadas", created, "já existentes", skipped)
	return nil
}

// catalogSeed fixa a geração pseudoaleatória: o mesmo catálogo sai em qualquer
// máquina e em qualquer execução, que é o que torna o seed idempotente.
const catalogSeed = 20260802

func catalog() []domain.PartInput {
	parts := make([]domain.PartInput, 0, len(highlights())+len(components)*len(vehicles))
	parts = append(parts, highlights()...)

	random := rand.New(rand.NewSource(catalogSeed))
	for _, component := range components {
		for _, vehicle := range vehicles {
			parts = append(parts, domain.PartInput{
				Name:              component.name + " " + vehicle,
				Category:          component.category,
				CurrentStock:      random.Intn(320) - 20,
				MinimumStock:      5 + random.Intn(56),
				AverageDailySales: float64(random.Intn(81)) / 10,
				LeadTimeDays:      1 + random.Intn(90),
				UnitCost:          float64(800+random.Intn(89200)) / 100,
				CriticalityLevel:  1 + random.Intn(5),
			})
		}
	}

	return parts
}

// highlights são escritas à mão para garantir que os casos de borda do cálculo
// apareçam na fila de reposição, independente do sorteio.
func highlights() []domain.PartInput {
	return []domain.PartInput{
		{Name: "Vela de Ignição X", Category: "engine", CurrentStock: -6, MinimumStock: 30, AverageDailySales: 6, LeadTimeDays: 7, UnitCost: 24.90, CriticalityLevel: 4},
		{Name: "Correia Dentada Z", Category: "engine", CurrentStock: 3, MinimumStock: 12, AverageDailySales: 1.5, LeadTimeDays: 20, UnitCost: 210.00, CriticalityLevel: 5},
		{Name: "Junta do Cabeçote X", Category: "engine", CurrentStock: 0, MinimumStock: 5, AverageDailySales: 0.2, LeadTimeDays: 90, UnitCost: 260.00, CriticalityLevel: 5},
		{Name: "Bomba d'Água X", Category: "cooling", CurrentStock: 2, MinimumStock: 6, AverageDailySales: 0.3, LeadTimeDays: 45, UnitCost: 380.00, CriticalityLevel: 5},
		{Name: "Bateria 60Ah X", Category: "electrical", CurrentStock: 12, MinimumStock: 10, AverageDailySales: 1.2, LeadTimeDays: 15, UnitCost: 450.00, CriticalityLevel: 5},
		{Name: "Filtro de Óleo X", Category: "filters", CurrentStock: 15, MinimumStock: 20, AverageDailySales: 4, LeadTimeDays: 5, UnitCost: 18.50, CriticalityLevel: 3},
		{Name: "Disco de Freio X", Category: "brakes", CurrentStock: 5, MinimumStock: 8, AverageDailySales: 0.5, LeadTimeDays: 30, UnitCost: 175.00, CriticalityLevel: 4},
		{Name: "Pastilha de Freio Y", Category: "brakes", CurrentStock: 8, MinimumStock: 10, AverageDailySales: 2, LeadTimeDays: 5, UnitCost: 89.90, CriticalityLevel: 5},
		{Name: "Mola Helicoidal X", Category: "suspension", CurrentStock: 4, MinimumStock: 10, AverageDailySales: 0, LeadTimeDays: 60, UnitCost: 145.00, CriticalityLevel: 3},
		{Name: "Amortecedor Dianteiro X", Category: "suspension", CurrentStock: 40, MinimumStock: 15, AverageDailySales: 0.8, LeadTimeDays: 10, UnitCost: 320.00, CriticalityLevel: 4},
		{Name: "Filtro de Ar X", Category: "filters", CurrentStock: 60, MinimumStock: 25, AverageDailySales: 3, LeadTimeDays: 4, UnitCost: 32.00, CriticalityLevel: 2},
		{Name: "Lâmpada H4 X", Category: "electrical", CurrentStock: 200, MinimumStock: 50, AverageDailySales: 0, LeadTimeDays: 3, UnitCost: 12.00, CriticalityLevel: 1},
	}
}

var components = []struct {
	name     string
	category string
}{
	{"Filtro de Óleo", "filters"},
	{"Filtro de Ar", "filters"},
	{"Filtro de Combustível", "filters"},
	{"Filtro de Cabine", "filters"},
	{"Pastilha de Freio", "brakes"},
	{"Disco de Freio", "brakes"},
	{"Cilindro Mestre", "brakes"},
	{"Flexível de Freio", "brakes"},
	{"Amortecedor Dianteiro", "suspension"},
	{"Amortecedor Traseiro", "suspension"},
	{"Mola Helicoidal", "suspension"},
	{"Bieleta Estabilizadora", "suspension"},
	{"Terminal de Direção", "suspension"},
	{"Rolamento de Roda", "suspension"},
	{"Correia Dentada", "engine"},
	{"Junta do Cabeçote", "engine"},
	{"Vela de Ignição", "engine"},
	{"Bobina de Ignição", "engine"},
	{"Coxim do Motor", "engine"},
	{"Bomba d'Água", "cooling"},
	{"Radiador", "cooling"},
	{"Válvula Termostática", "cooling"},
	{"Bateria 60Ah", "electrical"},
	{"Alternador", "electrical"},
	{"Kit de Embreagem", "transmission"},
}

var vehicles = []string{
	"Gol G5", "Gol G6", "Uno Mille", "Uno Way", "Palio Fire",
	"Palio Weekend", "Onix 1.0", "Onix Plus", "Prisma", "Celta",
	"Corsa Classic", "HB20 1.0", "HB20S", "Ka SE", "Ka Plus",
	"Fiesta 1.6", "EcoSport", "Ranger", "Strada", "Toro",
	"Argo", "Cronos", "Mobi", "Kwid", "Sandero",
	"Logan", "Duster", "Clio", "Corolla XEi", "Etios",
	"Hilux SW4", "Civic LXR", "Fit", "City", "Versa",
	"March", "Tiggo 5", "Compass", "Renegade", "Tracker",
}
