package selector

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	meta "github.com/WALL-EEEEEEE/proxy-service/gateway/internal/meta"
)

type Selector[T any] interface {
	Select(context.Context, *meta.Metadata, ...T) T
}

type Weight interface {
	Weight() int
}

type randomWeightedItem[T any] struct {
	item   T
	weight int
}

type randomWeighted[T any] struct {
	items []*randomWeightedItem[T]
	sum   int
	r     *rand.Rand
}

func newRandomWeighted[T any]() *randomWeighted[T] {
	return &randomWeighted[T]{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (rw *randomWeighted[T]) Add(item T, weight int) {
	ri := &randomWeightedItem[T]{item: item, weight: weight}
	rw.items = append(rw.items, ri)
	rw.sum += weight
}

func (rw *randomWeighted[T]) Next() (v T) {
	if len(rw.items) == 0 {
		return
	}
	if rw.sum <= 0 {
		return
	}
	weight := rw.r.Intn(rw.sum) + 1
	for _, item := range rw.items {
		weight -= item.weight
		if weight <= 0 {
			return item.item
		}
	}

	return rw.items[len(rw.items)-1].item
}

func (rw *randomWeighted[T]) Reset() {
	rw.items = nil
	rw.sum = 0
}

type roundRobin[T any] struct {
	counter uint64
}

// RoundRobinStrategy is a strategy for node selector.
// The node will be selected by round-robin algorithm.
func NewRoundRobin[T any]() *roundRobin[T] {
	return &roundRobin[T]{}
}

func (s *roundRobin[T]) Select(ctx context.Context, metadata *meta.Metadata, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}
	n := atomic.AddUint64(&s.counter, 1) - 1
	return vs[int(n%uint64(len(vs)))]
}

type random[T any] struct {
	rw *randomWeighted[T]
	mu sync.Mutex
}

// RandomStrategy is a strategy for node selector.
// The node will be selected randomly.
func Random[T any]() *random[T] {
	return &random[T]{
		rw: newRandomWeighted[T](),
	}
}

func (s *random[T]) Select(ctx context.Context, metadata *meta.Metadata, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rw.Reset()
	for i := range vs {
		weight := 0
		if weight <= 0 {
			weight = 1
		}
		s.rw.Add(vs[i], weight)
	}

	return s.rw.Next()
}

type fifo[T any] struct{}

// FIFOStrategy is a strategy for node selector.
// The node will be selected from first to last,
// and will stick to the selected node until it is failed.
func FIFO[T any]() *fifo[T] {
	return &fifo[T]{}
}

// Apply applies the fifo strategy for the nodes.
func (s *fifo[T]) Select(ctx context.Context, metadata *meta.Metadata, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}
	return vs[0]
}

type hash[T any] struct {
	r  *rand.Rand
	mu sync.Mutex
}

func Hash[T any]() *hash[T] {
	return &hash[T]{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *hash[T]) Select(ctx context.Context, metadata *meta.Metadata, vs ...T) (v T) {
	if len(vs) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return vs[s.r.Intn(len(vs))]
}
