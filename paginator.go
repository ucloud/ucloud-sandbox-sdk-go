package sandbox

import "context"

// Paginator provides lazy pagination over API results.
type Paginator[T any] struct {
	hasNext   bool
	nextToken string
	limit     int
	fetchFunc func(ctx context.Context, token string, limit int) ([]T, string, error)
}

func newPaginator[T any](limit int, fetchFunc func(ctx context.Context, token string, limit int) ([]T, string, error)) *Paginator[T] {
	return &Paginator[T]{hasNext: true, fetchFunc: fetchFunc, limit: limit}
}

func (p *Paginator[T]) HasNext() bool { return p.hasNext }

func (p *Paginator[T]) NextItems(ctx context.Context) ([]T, error) {
	if !p.hasNext {
		return nil, nil
	}
	items, nextToken, err := p.fetchFunc(ctx, p.nextToken, p.limit)
	if err != nil {
		return nil, err
	}
	p.nextToken = nextToken
	p.hasNext = nextToken != ""
	return items, nil
}

func (p *Paginator[T]) All(ctx context.Context) ([]T, error) {
	var all []T
	for p.HasNext() {
		items, err := p.NextItems(ctx)
		if err != nil {
			return all, err
		}
		all = append(all, items...)
	}
	return all, nil
}
