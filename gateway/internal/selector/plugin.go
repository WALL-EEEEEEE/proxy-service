package selector

import "context"

type SelectorPlugin[T any] struct {
}

func (s *SelectorPlugin[T]) Select(ctx context.Context, items ...T) *T {
	return nil
}
