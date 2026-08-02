package domain

const (
	DefaultPageSize = 50
	MaxPageSize     = 500
)

type PageRequest struct {
	Number int
	Size   int
}

func (r PageRequest) Normalize() PageRequest {
	if r.Number < 1 {
		r.Number = 1
	}
	if r.Size <= 0 {
		r.Size = DefaultPageSize
	}
	if r.Size > MaxPageSize {
		r.Size = MaxPageSize
	}
	return r
}

func (r PageRequest) Offset() int {
	normalized := r.Normalize()
	return (normalized.Number - 1) * normalized.Size
}

type Page[T any] struct {
	Items  []T
	Number int
	Size   int
	Total  int
}

func NewPage[T any](items []T, request PageRequest, total int) Page[T] {
	request = request.Normalize()
	if items == nil {
		items = []T{}
	}
	return Page[T]{
		Items:  items,
		Number: request.Number,
		Size:   request.Size,
		Total:  total,
	}
}

func Paginate[T any](items []T, request PageRequest) Page[T] {
	request = request.Normalize()

	total := len(items)
	start := min(request.Offset(), total)
	end := min(start+request.Size, total)

	return NewPage(items[start:end], request, total)
}

func (p Page[T]) TotalPages() int {
	if p.Size <= 0 || p.Total <= 0 {
		return 0
	}
	return (p.Total + p.Size - 1) / p.Size
}

func (p Page[T]) HasNext() bool {
	return p.Number < p.TotalPages()
}

func (p Page[T]) HasPrevious() bool {
	return p.Number > 1
}
