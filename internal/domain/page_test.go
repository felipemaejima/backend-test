package domain

import (
	"slices"
	"testing"
)

func TestPageRequestNormalize(t *testing.T) {
	tests := []struct {
		name       string
		request    PageRequest
		wantNumber int
		wantSize   int
		wantOffset int
	}{
		{"zero value falls back to the first page", PageRequest{}, 1, DefaultPageSize, 0},
		{"page zero becomes one", PageRequest{Number: 0, Size: 10}, 1, 10, 0},
		{"negative page becomes one", PageRequest{Number: -3, Size: 10}, 1, 10, 0},
		{"negative size falls back to default", PageRequest{Number: 2, Size: -1}, 2, DefaultPageSize, DefaultPageSize},
		{"size above cap is clamped", PageRequest{Number: 1, Size: 10_000}, 1, MaxPageSize, 0},
		{"valid values are preserved", PageRequest{Number: 3, Size: 20}, 3, 20, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.request.Normalize()
			if got.Number != tt.wantNumber {
				t.Errorf("Number = %d, expected %d", got.Number, tt.wantNumber)
			}
			if got.Size != tt.wantSize {
				t.Errorf("Size = %d, expected %d", got.Size, tt.wantSize)
			}
			if offset := tt.request.Offset(); offset != tt.wantOffset {
				t.Errorf("Offset = %d, expected %d", offset, tt.wantOffset)
			}
		})
	}
}

func TestPaginate(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e", "f", "g"}

	tests := []struct {
		name           string
		request        PageRequest
		wantItems      []string
		wantTotalPages int
		wantNext       bool
		wantPrevious   bool
	}{
		{
			name:           "first page",
			request:        PageRequest{Number: 1, Size: 3},
			wantItems:      []string{"a", "b", "c"},
			wantTotalPages: 3,
			wantNext:       true,
			wantPrevious:   false,
		},
		{
			name:           "middle page",
			request:        PageRequest{Number: 2, Size: 3},
			wantItems:      []string{"d", "e", "f"},
			wantTotalPages: 3,
			wantNext:       true,
			wantPrevious:   true,
		},
		{
			name:           "last page is partial",
			request:        PageRequest{Number: 3, Size: 3},
			wantItems:      []string{"g"},
			wantTotalPages: 3,
			wantNext:       false,
			wantPrevious:   true,
		},
		{
			name:           "page beyond the end is empty but still reports the total",
			request:        PageRequest{Number: 99, Size: 3},
			wantItems:      []string{},
			wantTotalPages: 3,
			wantNext:       false,
			wantPrevious:   true,
		},
		{
			name:           "size larger than the collection returns everything",
			request:        PageRequest{Number: 1, Size: 100},
			wantItems:      items,
			wantTotalPages: 1,
			wantNext:       false,
			wantPrevious:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Paginate(items, tt.request)

			if !slices.Equal(page.Items, tt.wantItems) {
				t.Errorf("Items = %v, expected %v", page.Items, tt.wantItems)
			}
			if page.Total != len(items) {
				t.Errorf("Total = %d, expected %d", page.Total, len(items))
			}
			if got := page.TotalPages(); got != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, expected %d", got, tt.wantTotalPages)
			}
			if got := page.HasNext(); got != tt.wantNext {
				t.Errorf("HasNext = %v, expected %v", got, tt.wantNext)
			}
			if got := page.HasPrevious(); got != tt.wantPrevious {
				t.Errorf("HasPrevious = %v, expected %v", got, tt.wantPrevious)
			}
		})
	}
}

func TestPaginateEmptyCollection(t *testing.T) {
	page := Paginate([]string{}, PageRequest{Number: 1, Size: 10})

	if page.Items == nil {
		t.Error("expected an empty slice, got nil")
	}
	if len(page.Items) != 0 {
		t.Errorf("Items = %v, expected empty", page.Items)
	}
	if page.Total != 0 {
		t.Errorf("Total = %d, expected 0", page.Total)
	}
	if page.TotalPages() != 0 {
		t.Errorf("TotalPages = %d, expected 0", page.TotalPages())
	}
	if page.HasNext() || page.HasPrevious() {
		t.Error("an empty collection should have neither next nor previous")
	}
}

func TestNewPageNeverCarriesNilItems(t *testing.T) {
	page := NewPage[string](nil, PageRequest{}, 0)
	if page.Items == nil {
		t.Error("expected an empty slice so the JSON serializes as [] instead of null")
	}
}

func TestTotalPagesRoundsUp(t *testing.T) {
	tests := []struct {
		total, size, want int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{99, 10, 10},
		{100, 10, 10},
		{101, 10, 11},
	}

	for _, tt := range tests {
		page := Page[string]{Size: tt.size, Total: tt.total}
		if got := page.TotalPages(); got != tt.want {
			t.Errorf("total=%d size=%d: TotalPages = %d, expected %d", tt.total, tt.size, got, tt.want)
		}
	}
}

func TestPartFilterNormalize(t *testing.T) {
	filter := PartFilter{Category: "  ENGINE  "}.Normalize()

	if filter.Category != "engine" {
		t.Errorf("Category = %q, expected %q", filter.Category, "engine")
	}
	if filter.Page.Number != 1 || filter.Page.Size != DefaultPageSize {
		t.Errorf("Page = %+v, expected the defaults", filter.Page)
	}
}
