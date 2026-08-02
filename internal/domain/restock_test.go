package domain

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func makePart(name string, currentStock, minimumStock int, averageDailySales float64, leadTimeDays, criticalityLevel int) Part {
	return Part{
		ID:                uuid.New(),
		Name:              name,
		Category:          "engine",
		CurrentStock:      currentStock,
		MinimumStock:      minimumStock,
		AverageDailySales: averageDailySales,
		LeadTimeDays:      leadTimeDays,
		UnitCost:          18.50,
		CriticalityLevel:  criticalityLevel,
	}
}

func priorityNames(priorities []RestockPriority) []string {
	names := make([]string, 0, len(priorities))
	for _, priority := range priorities {
		names = append(names, priority.Part.Name)
	}
	return names
}

func TestFormulas(t *testing.T) {
	tests := []struct {
		name                    string
		part                    Part
		wantExpectedConsumption float64
		wantProjectedStock      float64
		wantUrgencyScore        float64
		wantNeedsRestock        bool
	}{
		{
			name:                    "oil filter, from the entity fields published in the spec",
			part:                    makePart("Filtro de Óleo X", 15, 20, 4, 5, 3),
			wantExpectedConsumption: 20,
			wantProjectedStock:      -5,
			wantUrgencyScore:        75,
			wantNeedsRestock:        true,
		},
		{
			name:                    "brake pad, reproducing the response row published in the spec",
			part:                    makePart("Pastilha de Freio Y", 8, 10, 2, 5, 3),
			wantExpectedConsumption: 10,
			wantProjectedStock:      -2,
			wantUrgencyScore:        36,
			wantNeedsRestock:        true,
		},
		{
			name:                    "negative current stock",
			part:                    makePart("Negative", -42, 20, 4, 5, 5),
			wantExpectedConsumption: 20,
			wantProjectedStock:      -62,
			wantUrgencyScore:        410,
			wantNeedsRestock:        true,
		},
		{
			name:                    "zero daily sales",
			part:                    makePart("No Sales", 5, 20, 0, 30, 2),
			wantExpectedConsumption: 0,
			wantProjectedStock:      5,
			wantUrgencyScore:        30,
			wantNeedsRestock:        true,
		},
		{
			name:                    "zero daily sales with healthy stock",
			part:                    makePart("Idle", 100, 20, 0, 30, 2),
			wantExpectedConsumption: 0,
			wantProjectedStock:      100,
			wantUrgencyScore:        -160,
			wantNeedsRestock:        false,
		},
		{
			name:                    "very high lead time",
			part:                    makePart("Slow Supplier", 100, 10, 2, 365, 4),
			wantExpectedConsumption: 730,
			wantProjectedStock:      -630,
			wantUrgencyScore:        2560,
			wantNeedsRestock:        true,
		},
		{
			name:                    "zero lead time",
			part:                    makePart("Instant", 5, 10, 4, 0, 1),
			wantExpectedConsumption: 0,
			wantProjectedStock:      5,
			wantUrgencyScore:        5,
			wantNeedsRestock:        true,
		},
		{
			name:                    "projected stock equal to minimum does not need restock",
			part:                    makePart("Exact", 30, 10, 2, 10, 3),
			wantExpectedConsumption: 20,
			wantProjectedStock:      10,
			wantUrgencyScore:        0,
			wantNeedsRestock:        false,
		},
		{
			name:                    "fractional daily sales",
			part:                    makePart("Fractional", 10, 5, 2.5, 4, 2),
			wantExpectedConsumption: 10,
			wantProjectedStock:      0,
			wantUrgencyScore:        10,
			wantNeedsRestock:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.part.ExpectedConsumption(); got != tt.wantExpectedConsumption {
				t.Errorf("ExpectedConsumption = %v, expected %v", got, tt.wantExpectedConsumption)
			}
			if got := tt.part.ProjectedStock(); got != tt.wantProjectedStock {
				t.Errorf("ProjectedStock = %v, expected %v", got, tt.wantProjectedStock)
			}
			if got := tt.part.UrgencyScore(); got != tt.wantUrgencyScore {
				t.Errorf("UrgencyScore = %v, expected %v", got, tt.wantUrgencyScore)
			}
			if got := tt.part.NeedsRestock(); got != tt.wantNeedsRestock {
				t.Errorf("NeedsRestock = %v, expected %v", got, tt.wantNeedsRestock)
			}
		})
	}
}

func TestCalculateRestockPrioritiesFiltersHealthyParts(t *testing.T) {
	parts := []Part{
		makePart("Needs Restock", 5, 20, 1, 2, 3),
		makePart("Healthy", 500, 20, 1, 2, 3),
	}

	priorities := CalculateRestockPriorities(parts)

	if len(priorities) != 1 {
		t.Fatalf("expected 1 priority, got %d (%v)", len(priorities), priorityNames(priorities))
	}
	if priorities[0].Part.Name != "Needs Restock" {
		t.Errorf("part = %q, expected only the one needing restock", priorities[0].Part.Name)
	}
}

func TestCalculateRestockPrioritiesOrdersByUrgencyScore(t *testing.T) {
	parts := []Part{
		makePart("Low", 15, 20, 1, 1, 1),
		makePart("High", -42, 20, 4, 5, 5),
		makePart("Medium", 8, 20, 4, 5, 4),
	}

	priorities := CalculateRestockPriorities(parts)

	want := []string{"High", "Medium", "Low"}
	got := priorityNames(priorities)
	if len(got) != len(want) {
		t.Fatalf("list = %v, expected %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("list = %v, expected %v", got, want)
		}
	}
}

func TestCalculateRestockPrioritiesTieBreakers(t *testing.T) {
	tests := []struct {
		name  string
		parts []Part
		want  []string
	}{
		{
			name: "same score falls back to criticality",
			parts: []Part{
				makePart("Low Criticality", 0, 10, 0, 0, 1),
				makePart("High Criticality", 8, 10, 0, 0, 5),
			},
			want: []string{"High Criticality", "Low Criticality"},
		},
		{
			name: "same score and criticality fall back to daily sales",
			parts: []Part{
				makePart("Slow Mover", 10, 20, 0, 0, 2),
				makePart("Fast Mover", 10, 20, 7, 0, 2),
			},
			want: []string{"Fast Mover", "Slow Mover"},
		},
		{
			name: "full tie falls back to alphabetical order",
			parts: []Part{
				makePart("Zebra", 10, 20, 3, 0, 2),
				makePart("Alpha", 10, 20, 3, 0, 2),
				makePart("Meio", 10, 20, 3, 0, 2),
			},
			want: []string{"Alpha", "Meio", "Zebra"},
		},
		{
			name: "alphabetical order ignores case",
			parts: []Part{
				makePart("beta", 10, 20, 3, 0, 2),
				makePart("Alpha", 10, 20, 3, 0, 2),
			},
			want: []string{"Alpha", "beta"},
		},
		{
			name: "scores that differ only by float noise still tie",
			parts: []Part{
				makePart("Noisy", 10, 10, 0.1, 3, 2),
				makePart("Exact", 10, 10, 0.3, 1, 2),
			},
			want: []string{"Exact", "Noisy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priorityNames(CalculateRestockPriorities(tt.parts))
			if len(got) != len(tt.want) {
				t.Fatalf("list = %v, expected %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("list = %v, expected %v", got, tt.want)
				}
			}
		})
	}
}

func TestCalculateRestockPrioritiesCarriesComputedValues(t *testing.T) {
	part := makePart("Pastilha de Freio Y", 8, 10, 2, 5, 3)

	priorities := CalculateRestockPriorities([]Part{part})

	if len(priorities) != 1 {
		t.Fatalf("expected 1 priority, got %d", len(priorities))
	}
	got := priorities[0]
	if got.ExpectedConsumption != 10 {
		t.Errorf("ExpectedConsumption = %v, expected 10", got.ExpectedConsumption)
	}
	if got.ProjectedStock != -2 {
		t.Errorf("ProjectedStock = %v, expected -2", got.ProjectedStock)
	}
	if got.UrgencyScore != 36 {
		t.Errorf("UrgencyScore = %v, expected 36", got.UrgencyScore)
	}
	if got.Part.ID != part.ID {
		t.Error("expected the original part to be carried along")
	}
}

func TestCalculateRestockPrioritiesEmptyInput(t *testing.T) {
	priorities := CalculateRestockPriorities(nil)
	if priorities == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(priorities) != 0 {
		t.Errorf("expected 0 priorities, got %d", len(priorities))
	}
}

func TestCalculateRestockPrioritiesAllHealthy(t *testing.T) {
	parts := []Part{
		makePart("A", 500, 10, 1, 1, 3),
		makePart("B", 500, 10, 1, 1, 3),
	}

	priorities := CalculateRestockPriorities(parts)
	if len(priorities) != 0 {
		t.Errorf("expected 0 priorities, got %d (%v)", len(priorities), priorityNames(priorities))
	}
}

func TestCalculateRestockPrioritiesIsFullyOrdered(t *testing.T) {
	parts := benchmarkParts(2_000)

	priorities := CalculateRestockPriorities(parts)
	if len(priorities) == 0 {
		t.Fatal("expected the sample to produce priorities")
	}

	for i := 1; i < len(priorities); i++ {
		if comparePriorities(priorities[i-1], priorities[i]) > 0 {
			t.Fatalf("position %d breaks the ordering: %+v came before %+v",
				i, priorities[i-1].Part.Name, priorities[i].Part.Name)
		}
	}
}

func benchmarkParts(total int) []Part {
	parts := make([]Part, 0, total)
	for i := range total {
		parts = append(parts, makePart(
			fmt.Sprintf("Part %05d", i),
			i%60-10,
			20,
			float64(i%7)/2,
			i%15,
			i%5+1,
		))
	}
	return parts
}

func BenchmarkCalculateRestockPriorities(b *testing.B) {
	for _, total := range []int{100, 1_000, 10_000} {
		parts := benchmarkParts(total)
		b.Run(fmt.Sprintf("%d-parts", total), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				CalculateRestockPriorities(parts)
			}
		})
	}
}
